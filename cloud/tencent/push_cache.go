package tencent

import (
	"fmt"

	"github.com/rs/zerolog/log"
	cdn "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cdn/v20180606"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
)

func (c *TencentCloudClient) PushUrlsCache(urls []string) (string, error) {
	if len(urls) == 0 {
		log.Info().Msgf("no urls to refresh %v", urls)
		return "", nil
	}

	request := cdn.NewPushUrlsCacheRequest()
	request.Urls = common.StringPtrs(urls)
	request.UrlEncode = common.BoolPtr(true)
	response, err := c.client.PushUrlsCache(request)
	if err != nil {
		return "", err
	}

	if response == nil || response.Response == nil || response.Response.TaskId == nil {
		return "", fmt.Errorf("failed to refresh URLs: invalid response")
	}
	return *response.Response.TaskId, nil
}

func (c *TencentCloudClient) QueryPushTaskCompleted(taskID string) (bool, error) {
	request := cdn.NewDescribePushTasksRequest()
	request.TaskId = &taskID
	response, err := c.client.DescribePushTasks(request)
	if err != nil {
		return false, fmt.Errorf("failed to query push history: %w", err)
	}
	if response == nil {
		return false, fmt.Errorf("empty response from DescribePushTasks")
	}
	for _, detail := range response.Response.PushLogs {
		log.Info().Msgf("query push cache task url: %s, status: %s", *detail.Url, *detail.Status)
		if *detail.Status == "process" {
			return false, nil
		}
	}
	return true, nil
}
