package workflow

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/zzerding/cdnfix/cloud/tencent"
)

type TaskGroup struct {
	Client    *tencent.TencentCloudClient
	CacheFile string
	Site      string
	Action    string
	TaskIDs   []string
}

func QueryTasks(paths Paths, siteFilter string, runLog *zerolog.Logger) error {
	configs, err := ResolveSites()
	if err != nil {
		return err
	}

	siteNames := tencent.SortedSiteNames(configs)
	var groups []TaskGroup
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
			cacheFile := TaskStatePath(paths, siteName, action)
			pending, err := PendingTasks(cacheFile, siteName, action)
			if err != nil {
				return err
			}
			if len(pending) == 0 {
				continue
			}
			group := TaskGroup{
				Client:    client,
				CacheFile: cacheFile,
				Site:      siteName,
				Action:    action,
			}
			for _, task := range pending {
				group.TaskIDs = append(group.TaskIDs, task.ID)
			}
			groups = append(groups, group)
		}
	}

	if err := QueryTaskGroups(runLog, groups); err != nil {
		return err
	}
	runLog.Info().Msg("task query complete")
	return nil
}

func QueryTaskGroups(runLog *zerolog.Logger, groups []TaskGroup) error {
	if len(groups) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	errCh := make(chan error, len(groups)*10) // buffer to avoid blocking
	var wg sync.WaitGroup
	sem := make(chan struct{}, 10) // Limit API concurrency to 10

	for _, group := range groups {
		for _, taskID := range group.TaskIDs {
			group := group
			taskID := taskID
			wg.Add(1)
			go func() {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()
				if err := WaitAndMarkTask(ctx, runLog, group.Client, group.CacheFile, group.Site, group.Action, taskID); err != nil {
					errCh <- err
				}
			}()
		}
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

func WaitAndMarkTask(ctx context.Context, runLog *zerolog.Logger, client *tencent.TencentCloudClient, cacheFile string, site string, action string, taskID string) error {
	runLog.Info().Msgf("waiting task completion site=%s action=%s task=%s", site, action, taskID)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout or cancelled waiting for task %s", taskID)
		case <-ticker.C:
			completed, err := QueryTaskStatus(client, action, taskID)
			if err != nil {
				return err
			}
			if completed {
				now := time.Now()
				if err := MarkTask(cacheFile, site, action, taskID, "completed", now); err != nil {
					return err
				}
				runLog.Info().Msgf("task %s completed", taskID)
				return nil
			}
		}
	}
}

func QueryTaskStatus(client *tencent.TencentCloudClient, action string, taskID string) (bool, error) {
	switch strings.ToLower(action) {
	case "refresh":
		return client.QueryRefreshTaskCompleted(taskID)
	case "push":
		return client.QueryPushTaskCompleted(taskID)
	default:
		return false, fmt.Errorf("unsupported action %q", action)
	}
}
