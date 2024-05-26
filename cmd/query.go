package cmd

import (
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/zzerding/refresh-cdn/cloud/tencent"
)

func init() {
	rootCmd.AddCommand(queryCmd)
}
func query() error {
	var c *tencent.TencentCloudClient
	var err error
	if c, err = tencent.CreateCDNClient(); err != nil {
		return err
	}
	var wg sync.WaitGroup
	c.QueryRefreshHistoryForTasks(c.PushTackCacheFile, tencent.REFRESH, &wg)
	c.QueryRefreshHistoryForTasks(c.RefreshCacheFile, tencent.PUSHCACHE, &wg)
	defer log.Printf("task query complete")
	wg.Wait() // Wait for all goroutines to complete
	return nil
}

var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "query cdn refresh status fro task cache file",
	Run: func(cmd *cobra.Command, args []string) {
		if err := query(); err != nil {
			log.Error().Msgf("command query error %s", err.Error())
		}
	},
}
