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

var refreshCmd = &cobra.Command{
	Use:   "refresh cdn",
	Short: "refresh cdn",
	Long:  "refresh cnd for tencent use -f or -u input url list",
	Run:   refreshCommand,
}

func init() {
	refreshCmd.Flags().StringP("urls", "u", "", "urls ")
	refreshCmd.Flags().StringP("urlfile", "f", "", "urls file ,this file on line on one url")
	viper.BindPFlag("urls", refreshCmd.Flags().Lookup("urls"))
	viper.BindPFlag("urlfile", refreshCmd.Flags().Lookup("urlfile"))
	rootCmd.AddCommand(refreshCmd)
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
		return nil, fmt.Errorf("either --urls %s or --file %s must be provided", urls, filePath)
	}
	return urlList, nil
}

func refresh() error {
	urls := viper.GetString("urls")
	filePath := viper.GetString("urlfile")
	log.Debug().Msgf("refresh urls %s,urlfile: %s", urls, filePath)
	urlList, err := readURLs(urls, filePath)
	if err != nil || len(urlList) == 0 {
		return err
	}

	var urlsToPurge, pathsToPurge []string
	for _, url := range urlList {
		if strings.HasSuffix(url, "/") {
			pathsToPurge = append(pathsToPurge, url)
		} else {
			urlsToPurge = append(urlsToPurge, url)
		}
	}

	client, err := tencent.CreateCDNClient()
	if err != nil {
		return err
	}

	if err := client.RefreshPaths(urlsToPurge); err != nil {
		return err
	}
	if err := client.RefreshPaths(pathsToPurge); err != nil {
		return err
	}
	log.Print("urls is push to cloud cdn")
	return nil
}
func refreshCommand(cmd *cobra.Command, args []string) {
	if err := refresh(); err != nil {
		log.Error().Msgf(" %s!!!use -v env file or set system env", err.Error())
	}
}
