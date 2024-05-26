package cmd

import (
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
	c.QueryRefreshHistoryForTasks()
	return nil
}

var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "query cdn refresh status fro task cache file",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := query(); err != nil {
			log.Error().Msgf("command query error %s", err.Error())
			return err
		}
		return nil
	},
}
