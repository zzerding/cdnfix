package tencent

import (
	"fmt"
	"sort"

	cdn "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cdn/v20180606"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
)

func SortedSiteNames(configs map[string]Config) []string {
	names := make([]string, 0, len(configs))
	for name := range configs {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func CreateCDNClientForConfig(config Config) (*TencentCloudClient, error) {
	credential := common.NewCredential(config.SecretID, config.SecretKey)
	clientProfile := profile.NewClientProfile()
	client, err := cdn.NewClient(credential, config.Region, clientProfile)
	if err != nil {
		return nil, fmt.Errorf("failed to create CDN client: %w", err)
	}
	return &TencentCloudClient{client: client}, nil
}
