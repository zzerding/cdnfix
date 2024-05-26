package tencent

import (
	"fmt"

	"github.com/rs/zerolog/log"
	cdn "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cdn/v20180606"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
)

func (c *TencentCloudClient) PushUrlsCache(urls []string) error {
	if len(urls) == 0 {
		log.Info().Msgf("no urls to refresh %v", urls)
		return nil
	}

	request := cdn.NewPushUrlsCacheRequest()
	request.Urls = common.StringPtrs(urls)
	request.UrlEncode = common.BoolPtr(true)
	response, err := c.client.PushUrlsCache(request)
	if err != nil {
		return err
	}

	if response == nil || response.Response == nil || response.Response.TaskId == nil {
		return fmt.Errorf("failed to refresh URLs: invalid response")
	}

	taskId := *response.Response.TaskId
	if err := updateCacheFile(c.PushTackCacheFile, taskId, false); err != nil {
		return fmt.Errorf("failed to update cache file: %w", err)
	}
	return nil
}

// query push cache task status
func queryPushCachehHistory(client *cdn.Client, taskId string) bool {
	request := cdn.NewDescribePushTasksRequest()
	request.TaskId = &taskId
	response, err := client.DescribePushTasks(request)
	if err != nil {
		log.Error().Msgf("failed to query refresh history: %v", err)
		return true
	}
	if response == nil {
		log.Error().Msgf("empty response from DescribePurgeTasks")
		return true
	}
	for _, detail := range response.Response.PushLogs {
		log.Info().Msgf("query push cache task.url: %s, status: %s", *detail.Url, *detail.Status)
		if *detail.Status == "process" {
			return false
		}
	}
	return true
}
