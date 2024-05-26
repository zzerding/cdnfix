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

	response, err := c.client.PurgeUrlsCache(request)
	if err != nil {
		return err
	}

	if response == nil || response.Response == nil || response.Response.TaskId == nil {
		return fmt.Errorf("failed to refresh URLs: invalid response")
	}

	taskId := *response.Response.TaskId
	if err := updateCacheFile(c.taskCacheFile, taskId, false); err != nil {
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
	if err := updateCacheFile(c.taskCacheFile, taskId, false); err != nil {
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
		log.Debug().Msgf("task: %s, status: %s", *detail.TaskId, *detail.Status)
		if *detail.Status == "process" {
			return false
		}
	}
	return true
}

// 等待任务完成
func (c *TencentCloudClient) waitQueryStatusForTaskCompletion(taskId string) {
	if c == nil || c.client == nil || c.taskCacheFile == "" {
		log.Error().Msgf("waitQueryStatusForTaskCompletion called with invalid arguments")
		return
	}
	for {
		if queryRefreshHistory(c.client, taskId) {
			log.Info().Msgf("Task %s completed.\n", taskId)
			if err := updateCacheFile(c.taskCacheFile, taskId, true); err != nil {
				log.Error().Msgf("update taskId to cache error %v\n", err)
			}
			return // 任务完成，退出循环
		}
		time.Sleep(5 * time.Second) // 每次轮询间隔5秒
	}
}

// QueryRefreshHistoryForTasks queries CDN refresh status by reading tasks from a file.
func (c *TencentCloudClient) QueryRefreshHistoryForTasks() {
	if c == nil || c.client == nil {
		log.Error().Msgf("QueryRefreshHistoryForTasks called with invalid arguments: tencent cloud client is nil")
		return
	}
	if c.taskCacheFile == "" {
		log.Error().Msgf("QueryRefreshHistoryForTasks called with invalid arguments: taskCacheFile is empty")
		return
	}
	tasks, err := readCacheFile(c.taskCacheFile)
	if err != nil || len(tasks) == 0 {
		log.Error().Msgf("failed to read cache file: %v,len tasks: %d", c.taskCacheFile, len(tasks))
		return
	}
	var wg sync.WaitGroup

	for _, taskId := range tasks {
		wg.Add(1)
		go func(taskId string) {
			defer wg.Done()
			c.waitQueryStatusForTaskCompletion(taskId)
		}(taskId)
	}
	defer log.Printf("task query complete")
	wg.Wait() // Wait for all goroutines to complete
}
