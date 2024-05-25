package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/spf13/viper"
	cdn "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cdn/v20180606"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/regions"
)

// 定义一个结构体来保存配置信息
type Config struct {
	//fromat Secret_ID=xxx Secret_Key=xxx Region=xxx
	SecretID  string `mapstructure:"Secret_ID"`
	SecretKey string `mapstructure:"Secret_Key"`
	Region    string
}

// cache 文件名
var cacheFile = "cdn_refresh_tasks.txt"

// 初始化配置
func initConfig(configFile string) (*Config, error) {
	// 使用 viper 读取配置文件
	viper.SetConfigFile(configFile)
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	// 尝试读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %v", err)
	}

	// 从环境变量和配置文件中读取配置
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshalling config: %v", err)
	}
	config.Region = regions.Guangzhou
	//print config value
	fmt.Println("Region: ", config.Region)
	return &config, nil
}

// 创建 CDN 客户端
func createCDNClient(config *Config) (*cdn.Client, error) {
	credential := common.NewCredential(config.SecretID, config.SecretKey)
	clientProfile := profile.NewClientProfile()
	client, err := cdn.NewClient(credential, config.Region, clientProfile)
	if err != nil {
		return nil, fmt.Errorf("failed to create CDN client: %v", err)
	}
	return client, nil
}

// 读取 URL 列表
func readURLs(urls string, filePath string) ([]string, error) {
	var urlList []string
	if urls != "" {
		urlList = strings.Split(urls, ",")
	} else if filePath != "" {
		file, err := os.Open(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to open file: %v", err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			urlList = append(urlList, scanner.Text())
		}

		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("failed to read file: %v", err)
		}
	} else {
		return nil, fmt.Errorf("either --urls or --file must be provided")
	}
	return urlList, nil
}

// 刷新 URL
func refreshURLs(client *cdn.Client, urls []string) (*cdn.PurgeUrlsCacheResponse, error) {
	if len(urls) == 0 {
		return nil, nil
	}
	request := cdn.NewPurgeUrlsCacheRequest()
	request.Urls = common.StringPtrs(urls)

	response, err := client.PurgeUrlsCache(request)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh URLs: %v", err)
	}
	return response, nil
}

// 刷新路径
func refreshPaths(client *cdn.Client, paths []string) (*cdn.PurgePathCacheResponse, error) {
	if len(paths) == 0 {
		return nil, nil
	}
	request := cdn.NewPurgePathCacheRequest()
	request.FlushType = common.StringPtr("delete")
	request.Paths = common.StringPtrs(paths)

	response, err := client.PurgePathCache(request)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh paths: %v", err)
	}
	return response, nil
}

// 查询刷新状态
func queryRefreshHistory(client *cdn.Client, taskId string) (allTasksCompleted bool) {
	request := cdn.NewDescribePurgeTasksRequest()
	request.TaskId = &taskId
	response, err := client.DescribePurgeTasks(request)
	allTasksCompleted = true
	if err != nil {
		fmt.Printf("failed to query refresh history: %v", err)
		return allTasksCompleted
	}

	if response != nil {
		for _, detail := range response.Response.PurgeLogs {
			fmt.Printf("URL: %s, Status: %s\n", *detail.Url, *detail.Status)
			if *detail.Status == "process" {
				allTasksCompleted = false
				break
			}
		}
	}
	return allTasksCompleted

}

// 读取缓存文件中的任务 ID
func readCacheFile() ([]string, error) {

	// 读取文件内容
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		// 如果文件不存在，则返回空列表
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("读取缓存文件失败: %v", err)
	}

	// 按行分割文件内容，获取任务 ID 列表
	tasks := strings.Split(string(data), "\n")
	var validTasks []string
	for _, task := range tasks {
		if task != "" {
			validTasks = append(validTasks, task)
		}
	}
	return validTasks, nil
}

func updateCacheFile(taskId string, completed bool) error {

	// 读取已有任务 ID
	data, err := os.ReadFile(cacheFile)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("读取缓存文件失败: %v", err)
	}
	tasks := strings.Split(string(data), "\n")

	// 更新任务 ID 列表
	if completed {
		// 删除已完成的任务 ID
		newTasks := make([]string, 0)
		for _, t := range tasks {
			if t != taskId && t != "" {
				newTasks = append(newTasks, t)
			}
		}
		tasks = newTasks
	} else {
		// 添加新的任务 ID
		tasks = append(tasks, taskId)
	}

	// 写入更新后的任务 ID 列表
	err = os.WriteFile(cacheFile, []byte(strings.Join(tasks, "\n")), 0644)
	if err != nil {
		return fmt.Errorf("写入缓存文件失败: %v", err)
	}
	return nil
}

// 等待任务完成
func waitQueryStatusForTaskCompletion(client *cdn.Client, taskId string, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		if queryRefreshHistory(client, taskId) {
			fmt.Printf("Task %s completed.\n", taskId)
			if err := updateCacheFile(taskId, true); err != nil {
				fmt.Printf("update taskId to cache error %v\n", err)
			}
			return
		}
		time.Sleep(5 * time.Second) // 每次轮询间隔5秒
	}

}
func main() {
	urls := flag.String("urls", "", "Comma separated URLs to refresh")
	filePath := flag.String("file", "", "Path to file containing URLs to refresh")
	configFile := flag.String("envfile", ".env", "Path to configuration file, default .env")
	flag.Parse()

	config, err := initConfig(*configFile)
	if err != nil {
		fmt.Printf("Error initializing config: %v\n", err)
		return
	}

	client, err := createCDNClient(config)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	urlList, err := readURLs(*urls, *filePath)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	if len(urlList) == 0 {
		fmt.Println("No URLs to refresh")
		return
	}

	var urlsToPurge []string
	var pathsToPurge []string
	for _, url := range urlList {
		if strings.HasSuffix(url, "/") {
			pathsToPurge = append(pathsToPurge, url)
		} else {
			urlsToPurge = append(urlsToPurge, url)
		}
	}
	var wg sync.WaitGroup
	urlsResponse, _ := refreshURLs(client, urlsToPurge)
	if urlsResponse != nil {
		taskId := *urlsResponse.Response.TaskId
		updateCacheFile(taskId, false)
	}
	pathsResponse, _ := refreshPaths(client, pathsToPurge)
	if pathsResponse != nil {
		taskId := *urlsResponse.Response.TaskId
		updateCacheFile(taskId, false)
	}
	//load cache file for
	tasks, err := readCacheFile()
	if err != nil {
		fmt.Printf("read cache file error: %v\n", err)
	}
	for _, taskId := range tasks {
		wg.Add(1)
		go waitQueryStatusForTaskCompletion(client, taskId, &wg)
	}
	wg.Wait() // 等待所有 goroutine 完成
}
