package main

import (
	"fmt"

	"github.com/spf13/viper"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/regions"
	"github.com/zzerding/refresh-cdn/cmd"
)

// 定义一个结构体来保存配置信息
type Config struct {
	//fromat Secret_ID=xxx Secret_Key=xxx Region=xxx
	SecretID  string `mapstructure:"Secret_ID"`
	SecretKey string `mapstructure:"Secret_Key"`
	Region    string //https://github.com/TencentCloud/tencentcloud-sdk-go/blob/master/tencentcloud/common/regions/regions.go
}

// cache 文件名
var cacheFile = "cdn_refresh_tasks.txt"

// 读取配置
func readConfig() (*Config, error) {
	// 从环境变量和配置文件中读取配置
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshalling config: %v", err)
	}
	if config.Region == "" {
		config.Region = regions.Guangzhou
	}

	//print config value
	fmt.Println("Region: ", config.Region)
	return &config, nil
}

func main() {
	cmd.Execute()
}
