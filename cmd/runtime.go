package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	"github.com/zzerding/cdnfix/cloud/tencent"
	"github.com/zzerding/cdnfix/logger"
	"github.com/zzerding/cdnfix/workflow"
)

type taskGroup struct {
	client    *tencent.TencentCloudClient
	cacheFile string
	site      string
	action    string
	taskIDs   []string
}

func readURLs(urls string, filePath string) ([]string, error) {
	var urlList []string
	if urls != "" {
		for _, rawURL := range strings.Split(urls, ",") {
			if value := strings.TrimSpace(rawURL); value != "" {
				urlList = append(urlList, value)
			}
		}
	} else if filePath != "" {
		file, err := os.Open(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to open file: %v", err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			if value := strings.TrimSpace(scanner.Text()); value != "" {
				urlList = append(urlList, value)
			}
		}

		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("failed to read file: %v", err)
		}
	} else {
		return nil, fmt.Errorf("either --urls or --file must be provided")
	}
	return urlList, nil
}

func resolveSites() (map[string]tencent.Config, error) {
	return tencent.ReadConfigs()
}

func resolveSiteConfig(siteName string) (tencent.Config, error) {
	configs, err := resolveSites()
	if err != nil {
		return tencent.Config{}, err
	}
	requested := strings.TrimSpace(siteName)
	if requested == "" {
		if len(configs) == 1 {
			for _, config := range configs {
				return config, nil
			}
		}
		return tencent.Config{}, fmt.Errorf("--site is required when multiple sites are configured")
	}
	config, ok := configs[requested]
	if !ok {
		return tencent.Config{}, fmt.Errorf("site %q not found in config", requested)
	}
	return config, nil
}

func executeAction(siteName string, action string, sourceFile string, urls []string, jobName string) error {
	action = strings.ToLower(strings.TrimSpace(action))
	if err := workflow.ValidateAction(action); err != nil {
		return err
	}
	config, err := resolveSiteConfig(siteName)
	if err != nil {
		return err
	}
	paths := runtimePaths()
	run := workflow.NewRun(paths, config.Name, action, jobName, sourceFile, time.Now())
	runLogger, err := logger.NewRunLogger(run.LogFile)
	if err != nil {
		return err
	}
	defer func() {
		_ = runLogger.Close()
	}()
	runLog := runLogger.Logger()

	if err := workflow.SaveRun(run); err != nil {
		return err
	}

	finalize := func(status string, runErr error) error {
		run.Status = status
		run.FinishedAt = time.Now().Format(time.RFC3339)
		if runErr != nil {
			run.Error = runErr.Error()
		}
		if err := workflow.SaveRun(run); err != nil && runErr == nil {
			return err
		}
		return runErr
	}

	client, err := tencent.CreateCDNClientForConfig(config)
	if err != nil {
		return finalize("failed", err)
	}

	runLog.Info().Msgf("start %s for site=%s job=%s source=%s url_count=%d", action, config.Name, jobName, sourceFile, len(urls))
	var taskIDs []string
	submittedAt := time.Now()

	switch action {
	case "push":
		taskID, err := client.PushUrlsCache(urls)
		if err != nil {
			return finalize("failed", err)
		}
		if taskID != "" {
			taskIDs = append(taskIDs, taskID)
			if err := workflow.AddTask(run.CacheFile, config.Name, action, workflow.TaskRecord{
				ID:          taskID,
				Action:      action,
				Status:      "pending",
				JobName:     jobName,
				SourceFile:  sourceFile,
				RunID:       run.ID,
				URLCount:    len(urls),
				SubmittedAt: submittedAt.Format(time.RFC3339),
			}); err != nil {
				return finalize("failed", err)
			}
		}
	case "refresh":
		var refreshURLs []string
		var refreshPaths []string
		for _, target := range urls {
			if strings.HasSuffix(target, "/") {
				refreshPaths = append(refreshPaths, target)
			} else {
				refreshURLs = append(refreshURLs, target)
			}
		}
		urlTaskID, err := client.RefreshURLs(refreshURLs)
		if err != nil {
			return finalize("failed", err)
		}
		if urlTaskID != "" {
			taskIDs = append(taskIDs, urlTaskID)
			if err := workflow.AddTask(run.CacheFile, config.Name, action, workflow.TaskRecord{
				ID:          urlTaskID,
				Action:      action,
				Status:      "pending",
				JobName:     jobName,
				SourceFile:  sourceFile,
				RunID:       run.ID,
				URLCount:    len(refreshURLs),
				SubmittedAt: submittedAt.Format(time.RFC3339),
			}); err != nil {
				return finalize("failed", err)
			}
		}
		pathTaskID, err := client.RefreshPaths(refreshPaths)
		if err != nil {
			return finalize("failed", err)
		}
		if pathTaskID != "" {
			taskIDs = append(taskIDs, pathTaskID)
			if err := workflow.AddTask(run.CacheFile, config.Name, action, workflow.TaskRecord{
				ID:          pathTaskID,
				Action:      action,
				Status:      "pending",
				JobName:     jobName,
				SourceFile:  sourceFile,
				RunID:       run.ID,
				URLCount:    len(refreshPaths),
				SubmittedAt: submittedAt.Format(time.RFC3339),
			}); err != nil {
				return finalize("failed", err)
			}
		}
	}

	run.SubmittedIDs = taskIDs
	runLog.Info().Msgf("submitted %d task(s): %v", len(taskIDs), taskIDs)
	return finalize("submitted", nil)
}

func executeBatch(manifestPath string) error {
	jobs, err := workflow.ReadJobs(manifestPath)
	if err != nil {
		return err
	}
	jobs = workflow.SortedJobs(jobs)
	for _, job := range jobs {
		urls, err := readURLs("", job.File)
		if err != nil {
			return fmt.Errorf("job %s: %w", job.Name, err)
		}
		if err := executeAction(job.Site, job.Action, job.File, urls, job.Name); err != nil {
			return fmt.Errorf("job %s: %w", job.Name, err)
		}
	}
	return nil
}

func queryTasks(siteFilter string) error {
	configs, err := resolveSites()
	if err != nil {
		return err
	}
	runLogger, err := logger.NewRunLogger(commandLogPath("query", siteFilter))
	if err != nil {
		return err
	}
	defer func() {
		_ = runLogger.Close()
	}()
	runLog := runLogger.Logger()

	siteNames := tencent.SortedSiteNames(configs)
	var groups []taskGroup
	for _, siteName := range siteNames {
		if siteFilter != "" && siteFilter != siteName {
			continue
		}
		config := configs[siteName]
		client, err := tencent.CreateCDNClientForConfig(config)
		if err != nil {
			return err
		}
		for _, action := range []string{"refresh", "push"} {
			cacheFile := workflow.TaskStatePath(runtimePaths(), siteName, action)
			pending, err := workflow.PendingTasks(cacheFile, siteName, action)
			if err != nil {
				return err
			}
			if len(pending) == 0 {
				continue
			}
			group := taskGroup{
				client:    client,
				cacheFile: cacheFile,
				site:      siteName,
				action:    action,
			}
			for _, task := range pending {
				group.taskIDs = append(group.taskIDs, task.ID)
			}
			groups = append(groups, group)
		}
	}

	if err := queryTaskGroups(runLog, groups); err != nil {
		return err
	}
	runLog.Info().Msg("task query complete")
	return nil
}

func queryTaskGroups(runLog *zerolog.Logger, groups []taskGroup) error {
	if len(groups) == 0 {
		return nil
	}

	errCh := make(chan error, len(groups))
	var wg sync.WaitGroup
	for _, group := range groups {
		group := group
		wg.Add(1)
		go func() {
			defer wg.Done()
			for _, taskID := range group.taskIDs {
				if err := waitAndMarkTask(runLog, group.client, group.cacheFile, group.site, group.action, taskID); err != nil {
					errCh <- err
					return
				}
			}
		}()
	}
	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			return err
		}
	}
	return nil
}

func waitAndMarkTask(runLog *zerolog.Logger, client *tencent.TencentCloudClient, cacheFile string, site string, action string, taskID string) error {
	runLog.Info().Msgf("waiting task completion site=%s action=%s task=%s", site, action, taskID)
	for {
		completed, err := queryTaskStatus(client, action, taskID)
		if err != nil {
			return err
		}
		now := time.Now()
		status := "pending"
		if completed {
			status = "completed"
		}
		if err := workflow.MarkTask(cacheFile, site, action, taskID, status, now); err != nil {
			return err
		}
		if completed {
			runLog.Info().Msgf("task %s completed", taskID)
			return nil
		}
		time.Sleep(10 * time.Second)
	}
}

func queryTaskStatus(client *tencent.TencentCloudClient, action string, taskID string) (bool, error) {
	switch strings.ToLower(action) {
	case "refresh":
		return client.QueryRefreshTaskCompleted(taskID)
	case "push":
		return client.QueryPushTaskCompleted(taskID)
	default:
		return false, fmt.Errorf("unsupported action %q", action)
	}
}

func selectedSite() string {
	return strings.TrimSpace(viper.GetString("site"))
}
