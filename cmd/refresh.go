package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var refreshCacheCmd = &cobra.Command{
	Use:   "refresh",
	Short: "refresh cdn cache for one site",
	Long: `Refresh CDN cache for a single site.

Use --site to choose a configured site. Provide URLs with --urls or load them
from a file with --urlfile. URLs ending with "/" are submitted as path refresh
requests; all other URLs are submitted as URL refresh requests.`,
	Example: `  cdnfix --site prod-a -u https://example.com/a.js refresh
  cdnfix --site prod-a -f urls/prod-a/refresh.txt refresh
  cdnfix --root /opt/cdnfix --site prod-a -f urls/prod-a/refresh.txt refresh`,
	Run: refreshCommand,
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
