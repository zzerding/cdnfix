package tencent

import (
	"fmt"

	"github.com/rs/zerolog/log"
	cdn "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cdn/v20180606"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
)

func (c *TencentCloudClient) RefreshURLs(urls []string) (string, error) {
	if len(urls) == 0 {
		log.Info().Msgf("no urls to refresh %v", urls)
		return "", nil
	}

	request := cdn.NewPurgeUrlsCacheRequest()
	request.Urls = common.StringPtrs(urls)
	request.UrlEncode = common.BoolPtr(true)
	response, err := c.client.PurgeUrlsCache(request)
	if err != nil {
		return "", err
	}
	if response == nil || response.Response == nil || response.Response.TaskId == nil {
		return "", fmt.Errorf("failed to refresh URLs: invalid response")
	}
	return *response.Response.TaskId, nil
}

func (c *TencentCloudClient) RefreshPaths(paths []string) (string, error) {
	if len(paths) == 0 {
		log.Info().Msgf("no paths to refresh %v", paths)
		return "", nil
	}
	request := cdn.NewPurgePathCacheRequest()
	request.FlushType = common.StringPtr("delete")
	request.Paths = common.StringPtrs(paths)

	response, err := c.client.PurgePathCache(request)
	if err != nil {
		return "", fmt.Errorf("failed to refresh paths: %w", err)
	}
	if response.Response == nil || response.Response.TaskId == nil {
		return "", fmt.Errorf("failed to refresh paths: invalid response")
	}
	return *response.Response.TaskId, nil
}

func (c *TencentCloudClient) QueryRefreshTaskCompleted(taskID string) (bool, error) {
	request := cdn.NewDescribePurgeTasksRequest()
	request.TaskId = &taskID
	response, err := c.client.DescribePurgeTasks(request)
	if err != nil {
		return false, fmt.Errorf("failed to query refresh history: %w", err)
	}
	if response == nil {
		return false, fmt.Errorf("empty response from DescribePurgeTasks")
	}
	for _, detail := range response.Response.PurgeLogs {
		log.Info().Msgf("query refresh cache task url: %s, status: %s", *detail.Url, *detail.Status)
		if *detail.Status == "process" {
			return false, nil
		}
	}
	return true, nil
}
