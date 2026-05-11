package tencent

import (
	cdn "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cdn/v20180606"
)

type TencentCloudClient struct {
	client *cdn.Client
}

type Config struct {
	Name      string `mapstructure:"-"`
	SecretID  string `mapstructure:"secret_id"`
	SecretKey string `mapstructure:"secret_key"`
	Region    string `mapstructure:"region"`
}

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
