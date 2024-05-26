package tencent

import (
	"fmt"

	"github.com/spf13/viper"
	cdn "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cdn/v20180606"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/regions"
)

// ReadConfig reads the configuration from environment variables and a config file.
func ReadConfig() (*Config, error) {
	config := &Config{}
	config.SecretID = viper.GetString("SECRET_ID")
	config.SecretKey = viper.GetString("SECRET_KEY")
	if config.SecretID == "" || config.SecretKey == "" {
		return nil, fmt.Errorf("SECRET_ID or SECRET_KEY is not set .you can set system env or use .env file")
	}
	if config.Region == "" {
		config.Region = regions.Guangzhou
	}
	return config, nil
}

// CreateCDNClient creates a CDN client.
func CreateCDNClient() (*TencentCloudClient, error) {
	config, err := ReadConfig()
	if err != nil {
		return nil, err
	}
	credential := common.NewCredential(config.SecretID, config.SecretKey)
	clientProfile := profile.NewClientProfile()
	client, err := cdn.NewClient(credential, config.Region, clientProfile)
	if err != nil {
		return nil, fmt.Errorf("failed to create CDN client: %w", err)
	}
	tencentCloudClient.client = client
	return &tencentCloudClient, nil
}
