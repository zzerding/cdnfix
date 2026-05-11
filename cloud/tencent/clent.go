package tencent

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/viper"
	cdn "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cdn/v20180606"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/regions"
)

func ReadConfig() (*Config, error) {
	config := &Config{}
	config.SecretID = viper.GetString("SECRET_ID")
	config.SecretKey = viper.GetString("SECRET_KEY")
	config.Region = viper.GetString("REGION")
	if config.SecretID == "" || config.SecretKey == "" {
		return nil, fmt.Errorf("SECRET_ID or SECRET_KEY is not set .you can set system env or use .env file")
	}
	if config.Region == "" {
		config.Region = regions.Guangzhou
	}
	config.Name = "default"
	return config, nil
}

func ReadConfigs() (map[string]Config, error) {
	var sites map[string]Config
	if err := viper.UnmarshalKey("sites", &sites); err == nil && len(sites) > 0 {
		normalized := make(map[string]Config, len(sites))
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

	config, err := ReadConfig()
	if err != nil {
		return nil, err
	}
	return map[string]Config{config.Name: *config}, nil
}

func SortedSiteNames(configs map[string]Config) []string {
	names := make([]string, 0, len(configs))
	for name := range configs {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func CreateCDNClient() (*TencentCloudClient, error) {
	config, err := ReadConfig()
	if err != nil {
		return nil, err
	}
	return CreateCDNClientForConfig(*config)
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
