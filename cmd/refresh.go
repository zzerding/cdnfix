package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var refreshCacheCmd = &cobra.Command{
	Use:   "refresh",
	Short: "refresh cdn cache for one site",
	Long:  "refresh cdn cache for one site by --site and --urls/--file",
	Run:   refreshCommand,
}

func init() {
	rootCmd.AddCommand(refreshCacheCmd)
}

func refresh() error {
	urlList, err := readURLs(viper.GetString("urls"), viper.GetString("urlfile"))
	if err != nil || len(urlList) == 0 {
		return err
	}
	return executeAction(selectedSite(), "refresh", viper.GetString("urlfile"), urlList, "")
}

func refreshCommand(cmd *cobra.Command, args []string) {
	if err := refresh(); err != nil {
		log.Error().Msgf("%s", err.Error())
	}
}
