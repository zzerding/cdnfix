package tencent

import (
	cdn "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cdn/v20180606"
)

type TencentCloudClient struct {
	client            *cdn.Client // CDN client
	RefreshCacheFile  string      // refresh cache file
	PushTackCacheFile string      // push cache file
}

type Config struct {
	//fromat Secret_ID=xxx Secret_Key=xxx Region=xxx
	SecretID  string `mapstructure:"SECRET_ID"`
	SecretKey string `mapstructure:"SECRET_KEY"`
	Region    string //https://github.com/TencentCloud/tencentcloud-sdk-go/blob/master/tencentcloud/common/regions/regions.go
}

var tencentCloudClient TencentCloudClient

type TaskType uint16

const (
	PUSHCACHE TaskType = iota
	REFRESH
)

func (s TaskType) String() string {
	switch s {
	case PUSHCACHE:
		return "push"
	case REFRESH:
		return "refresh"
	default:
		return "unknown"
	}
}
