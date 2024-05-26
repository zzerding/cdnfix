package tencent

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	cdn "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cdn/v20180606"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/regions"
)

type TencentCloudClient struct {
	client        *cdn.Client // CDN 客户端
	taskCacheFile string      // 任务缓存文件
}

var tencentCloudClient TencentCloudClient
var taskCacheFile string = ".tasks.cache"

// ReadConfig reads the configuration from environment variables and a config file.
func ReadConfig() (*Config, error) {
	config := &Config{}
	if err := viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed to read config: %v", err)
	}

	if config.Region == "" {
		config.Region = regions.Guangzhou
	}

	return config, nil
}

// CreateCDNClient creates a CDN client.
func CreateCDNClient() (*TencentCloudClient, error) {
	config, err := ReadConfig()
	log.Debug().Msg("config region " + config.Region)
	if err != nil {
		return nil, fmt.Errorf("config is nil")
	}
	if config.SecretID == "" || config.SecretKey == "" {
		return nil, fmt.Errorf("SecretID or SecretKey is empty")
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
func init() {
	tencentCloudClient.client = nil
	tencentCloudClient.taskCacheFile = taskCacheFile
}
