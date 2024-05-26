package tencent

type Config struct {
	//fromat Secret_ID=xxx Secret_Key=xxx Region=xxx
	SecretID  string `mapstructure:"Secret_ID"`
	SecretKey string `mapstructure:"Secret_Key"`
	Region    string //https://github.com/TencentCloud/tencentcloud-sdk-go/blob/master/tencentcloud/common/regions/regions.go
}
