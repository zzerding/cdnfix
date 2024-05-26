package tencent

import (
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	cdn "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cdn/v20180606"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
)

// 刷新 URL
func (c *TencentCloudClient) RefreshURLs(urls []string) error {
	if len(urls) == 0 {
		log.Info().Msgf("no urls to refresh %v", urls)
		return nil
	}

	request := cdn.NewPurgeUrlsCacheRequest()
	request.Urls = common.StringPtrs(urls)
	request.UrlEncode = common.BoolPtr(true)
	response, err := c.client.PurgeUrlsCache(request)
	if err != nil {
		return err
	}

	if response == nil || response.Response == nil || response.Response.TaskId == nil {
		return fmt.Errorf("failed to refresh URLs: invalid response")
	}

	taskId := *response.Response.TaskId
	if err := updateCacheFile(c.RefreshCacheFile, taskId, false); err != nil {
		return fmt.Errorf("failed to update cache file: %w", err)
	}
	return nil
}

// RefreshPaths refreshes the given paths.
func (c *TencentCloudClient) RefreshPaths(paths []string) error {
	if len(paths) == 0 {
		log.Info().Msgf("no paths to refresh %v", paths)
		return nil
	}
	request := cdn.NewPurgePathCacheRequest()
	request.FlushType = common.StringPtr("delete")
	request.Paths = common.StringPtrs(paths)

	response, err := c.client.PurgePathCache(request)
	if err != nil {
		return fmt.Errorf("failed to refresh paths: %w", err)
	}
	if response.Response == nil || response.Response.TaskId == nil {
		return fmt.Errorf("failed to refresh paths: invalid response")
	}
	taskId := *response.Response.TaskId
	if err := updateCacheFile(c.RefreshCacheFile, taskId, false); err != nil {
		return fmt.Errorf("failed to update cache file: %w", err)
	}
	return nil
}

// 查询刷新状态
func queryRefreshHistory(client *cdn.Client, taskId string) (allTasksCompleted bool) {
	request := cdn.NewDescribePurgeTasksRequest()
	request.TaskId = &taskId
	response, err := client.DescribePurgeTasks(request)
	if err != nil {
		log.Error().Msgf("failed to query refresh history: %v", err)
		return true
	}
	if response == nil {
		log.Error().Msgf("empty response from DescribePurgeTasks")
		return true
	}
	for _, detail := range response.Response.PurgeLogs {
		log.Info().Msgf("query refreshs cache task url: %s, status: %s", *detail.Url, *detail.Status)
		if *detail.Status == "process" {
			return false
		}
	}
	return true
}

// 等待任务完成
func (c *TencentCloudClient) waitQueryStatusForTaskCompletion(taskId string, taskType TaskType) {
	if c == nil || c.client == nil || c.RefreshCacheFile == "" {
		log.Error().Msgf("waitQueryStatusForTaskCompletion called with invalid arguments")
		return
	}
	for {
		switch taskType {
		case PUSHCACHE:
			if queryRefreshHistory(c.client, taskId) {
				log.Info().Msgf("Task %s completed.\n", taskId)
				if err := updateCacheFile(c.RefreshCacheFile, taskId, true); err != nil {
					log.Error().Msgf("update taskId to cache error %v\n", err)
				}
				return // 任务完成，退出循环
			}
		case REFRESH:
			if queryPushCachehHistory(c.client, taskId) {
				log.Info().Msgf("Task %s completed.\n", taskId)
				if err := updateCacheFile(c.PushTackCacheFile, taskId, true); err != nil {
					log.Error().Msgf("update taskId to cache error %v\n", err)
				}
				return // 任务完成，退出循环
			}
		default:
			log.Error().Msgf("waitQueryStatusForTaskCompletion called with invalid task type %s", taskType)
			return
		}

		time.Sleep(10 * time.Second) // 每次轮询间隔5秒
	}
}

// QueryRefreshHistoryForTasks queries CDN refresh status by reading tasks from a file.
func (c *TencentCloudClient) QueryRefreshHistoryForTasks(f string, taskType TaskType, wg *sync.WaitGroup) {
	if c == nil || c.client == nil {
		log.Error().Msgf("QueryRefreshHistoryForTasks called with invalid arguments: tencent cloud client is nil")
		return
	}
	if f == "" {
		log.Info().Msgf("QueryRefreshHistoryForTasks called with invalid arguments: taskCacheFile is empty")
		return
	}
	tasks, err := readCacheFile(f)
	if err != nil || len(tasks) == 0 {
		log.Info().Msgf("failed to read cache file: %v,len tasks: %d", f, len(tasks))
		return
	}

	for _, taskId := range tasks {
		wg.Add(1)
		go func(taskId string) {
			defer wg.Done()
			c.waitQueryStatusForTaskCompletion(taskId, taskType)
		}(taskId)
	}

}
