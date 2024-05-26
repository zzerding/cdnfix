package tencent

import cdn "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cdn/v20180606"

type TencentCloudClient struct {
	client        *cdn.Client // CDN 客户端
	taskCacheFile string      // 任务缓存文件
}

type Config struct {
	//fromat Secret_ID=xxx Secret_Key=xxx Region=xxx
	SecretID  string `mapstructure:"SECRET_ID"`
	SecretKey string `mapstructure:"SECRET_KEY"`
	Region    string //https://github.com/TencentCloud/tencentcloud-sdk-go/blob/master/tencentcloud/common/regions/regions.go
}

var tencentCloudClient TencentCloudClient
var taskCacheFile string = ".tasks.cache"

func init() {
	tencentCloudClient.client = nil
	tencentCloudClient.taskCacheFile = taskCacheFile
}
