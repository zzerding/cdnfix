package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
	cdn "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cdn/v20180606"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/regions"
)

func main() {
	// 解析命令行参数
	urls := flag.String("urls", "", "Comma separated URLs to refresh")
	filePath := flag.String("file", "", "Path to file containing URLs to refresh")
	configFile := flag.String("envfile", ".env", "Path to configuration file, default .env")
	flag.Parse()

	// 使用 viper 读取 .env 配置文件
	viper.SetConfigFile(*configFile)
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	// 尝试读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Error reading config file: %v\n", err)
		return
	}

	// 从环境变量和配置文件中读取 SecretId 和 SecretKey
	secretId := viper.GetString("SECRET_ID")
	secretKey := viper.GetString("SECRET_KEY")

	if secretId == "" || secretKey == "" {
		fmt.Println("SecretId and SecretKey are required")
		return
	}

	// 创建 CDN 客户端
	client, err := cdn.NewClient(common.NewCredential(secretId, secretKey), regions.Guangzhou, profile.NewClientProfile())
	if err != nil {
		fmt.Printf("Failed to create CDN client: %v\n", err)
		return
	}

	var urlList []string

	// 从命令行参数获取 URL 列表
	if *urls != "" {
		urlList = strings.Split(*urls, ",")
	} else if *filePath != "" {
		// 从文件中读取 URL 列表
		file, err := os.Open(*filePath)
		if err != nil {
			fmt.Printf("Failed to open file: %v\n", err)
			return
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			urlList = append(urlList, scanner.Text())
		}

		if err := scanner.Err(); err != nil {
			fmt.Printf("Failed to read file: %v\n", err)
			return
		}
	} else {
		fmt.Println("Either --urls or --file must be provided")
		return
	}

	if len(urlList) == 0 {
		fmt.Println("No URLs to refresh")
		return
	}

	// 分别处理 URL 和路径
	var urlsToPurge []string
	var pathsToPurge []string
	for _, url := range urlList {
		if strings.HasSuffix(url, "/") {
			pathsToPurge = append(pathsToPurge, url)
		} else {
			urlsToPurge = append(urlsToPurge, url)
		}
	}

	if len(urlsToPurge) > 0 {
		// 创建并发送刷新 URL 请求
		urlRequest := cdn.NewPurgeUrlsCacheRequest()
		urlRequest.Urls = common.StringPtrs(urlsToPurge)

		urlResponse, err := client.PurgeUrlsCache(urlRequest)
		if err != nil {
			fmt.Printf("Failed to refresh URLs: %v\n", err)
			return
		}

		fmt.Printf("URL refresh request submitted successfully: %v\n", urlResponse.Response)
	}

	if len(pathsToPurge) > 0 {
		// 创建并发送刷新路径请求

		pathRequest := cdn.NewPurgePathCacheRequest()
    pathRequest.FlushType = common.StringPtr("delete") 
		pathRequest.Paths = common.StringPtrs(pathsToPurge)

		pathResponse, err := client.PurgePathCache(pathRequest)
		if err != nil {
			fmt.Printf("Failed to refresh paths: %v\n", err)
			return
		}

		fmt.Printf("Path refresh request submitted successfully: %v\n", pathResponse.Response)
	}
}

