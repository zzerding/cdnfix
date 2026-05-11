package workflow

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/regions"
	"github.com/zzerding/cdnfix/cloud/tencent"
	"github.com/zzerding/cdnfix/logger"
)

func ReadURLs(urls string, filePath string) ([]string, error) {
	var urlList []string

	// Read from stdin if available
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			if value := strings.TrimSpace(scanner.Text()); value != "" {
				urlList = append(urlList, value)
			}
		}
		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("failed to read from stdin: %v", err)
		}
	}

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
	}

	if len(urlList) == 0 {
		return nil, fmt.Errorf("either --urls, --file, or stdin must be provided")
	}

	return urlList, nil
}

func ResolveSites() (map[string]tencent.Config, error) {
	var sites map[string]tencent.Config
	if err := viper.UnmarshalKey("sites", &sites); err == nil && len(sites) > 0 {
		normalized := make(map[string]tencent.Config, len(sites))
		for name, config := range sites {
			siteName := strings.TrimSpace(name)
			if siteName == "" {
				return nil, fmt.Errorf("site name is empty")
			}
			if config.SecretID == "" || config.SecretKey == "" {
				return nil, fmt.Errorf("site %q missing secret_id or secret_key", siteName)
			}
			if config.Region == "" {
				config.Region = regions.Guangzhou
			}
			config.Name = siteName
			normalized[siteName] = config
		}
		return normalized, nil
	}

	config := tencent.Config{}
	config.SecretID = viper.GetString("SECRET_ID")
	config.SecretKey = viper.GetString("SECRET_KEY")
	config.Region = viper.GetString("REGION")
	if config.SecretID == "" || config.SecretKey == "" {
		return nil, fmt.Errorf("SECRET_ID or SECRET_KEY is not set. You can set system env or use .env file")
	}
	if config.Region == "" {
		config.Region = regions.Guangzhou
	}
	config.Name = "default"
	return map[string]tencent.Config{config.Name: config}, nil
}

func ResolveSiteConfig(siteName string) (tencent.Config, error) {
	configs, err := ResolveSites()
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

func SubmitAndRecord(paths Paths, config tencent.Config, action string, sourceFile string, urls []string, jobName string) error {
	run := NewRun(paths, config.Name, action, jobName, sourceFile, time.Now())
	runLogger, err := logger.NewRunLogger(run.LogFile)
	if err != nil {
		return err
	}
	defer func() {
		_ = runLogger.Close()
	}()
	runLog := runLogger.Logger()

	if err := SaveRun(run); err != nil {
		return err
	}

	finalize := func(status string, runErr error) error {
		run.Status = status
		run.FinishedAt = time.Now().Format(time.RFC3339)
		if runErr != nil {
			run.Error = runErr.Error()
		}
		if err := SaveRun(run); err != nil && runErr == nil {
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
			if err := AddTask(run.CacheFile, config.Name, action, TaskRecord{
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

		recordTask := func(taskID string, count int) error {
			if taskID == "" {
				return nil
			}
			taskIDs = append(taskIDs, taskID)
			return AddTask(run.CacheFile, config.Name, action, TaskRecord{
				ID:          taskID,
				Action:      action,
				Status:      "pending",
				JobName:     jobName,
				SourceFile:  sourceFile,
				RunID:       run.ID,
				URLCount:    count,
				SubmittedAt: submittedAt.Format(time.RFC3339),
			})
		}

		if len(refreshURLs) > 0 {
			urlTaskID, err := client.RefreshURLs(refreshURLs)
			if err != nil {
				return finalize("failed", err)
			}
			if err := recordTask(urlTaskID, len(refreshURLs)); err != nil {
				return finalize("failed", err)
			}
		}

		if len(refreshPaths) > 0 {
			pathTaskID, err := client.RefreshPaths(refreshPaths)
			if err != nil {
				return finalize("failed", err)
			}
			if err := recordTask(pathTaskID, len(refreshPaths)); err != nil {
				return finalize("failed", err)
			}
		}
	}

	run.SubmittedIDs = taskIDs
	runLog.Info().Msgf("submitted %d task(s): %v", len(taskIDs), taskIDs)
	return finalize("submitted", nil)
}

func ExecuteBatch(paths Paths, manifestPath string, opts ExecuteBatchOptions) error {
	jobs, err := ReadJobs(manifestPath)
	if err != nil {
		return err
	}
	jobs, err = SelectBatchJobs(jobs, opts)
	if err != nil {
		return err
	}
	for _, job := range jobs {
		urls, err := ReadURLs("", job.File)
		if err != nil {
			return fmt.Errorf("job %s: %w", job.Name, err)
		}
		config, err := ResolveSiteConfig(job.Site)
		if err != nil {
			return fmt.Errorf("job %s: %w", job.Name, err)
		}
		if err := SubmitAndRecord(paths, config, job.Action, job.File, urls, job.Name); err != nil {
			return fmt.Errorf("job %s: %w", job.Name, err)
		}
	}
	return nil
}
