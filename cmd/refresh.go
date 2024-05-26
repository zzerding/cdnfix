package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zzerding/refresh-cdn/cloud/tencent"
)

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

func init() {
	rootCmd.AddCommand(refreshCmd)
}
func refresh() error {

	urls := viper.GetString("urls")
	filePath := viper.GetString("urlfile")

	urlList, err := readURLs(urls, filePath)
	if err != nil {
		log.Printf("%v\n", err)
		return nil
	}

	if len(urlList) == 0 {
		log.Printf("No URLs to refresh")
		return nil
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
	var c *tencent.TencentCloudClient
	if c, err = tencent.CreateCDNClient(); err != nil {
		return err
	}

	if err := c.RefreshPaths(urlsToPurge); err != nil {
		return err
	}
	if err := c.RefreshPaths(pathsToPurge); err != nil {
		return err
	}
	log.Print("all refresh task is push")
	return nil
}

var refreshCmd = &cobra.Command{
	Use:   "query",
	Short: "query cdn refresh status",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := refresh(); err != nil {
			log.Error().Msg(err.Error())
			return err
		}
		return nil
	},
}
