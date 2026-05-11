package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var pushCacheCmd = &cobra.Command{
	Use:   "push",
	Short: "push cache for one site",
	Long: `Push cache for a single site.

Use --site to choose a configured site. Provide URLs with --urls or load them
from a file with --urlfile.

By default, cdnfix reads site configuration from /etc/cdnfix/sites.yaml and
writes runtime state under /var/lib/cdnfix and /var/log/cdnfix. Use --root for
a portable deployment where config/, urls/, and var/ live under one directory.`,
	Example: `  cdnfix --site prod-a -u https://example.com/a.js push
  cdnfix --site prod-a -f /etc/cdnfix/urls/prod-a/push.txt push
  cdnfix --config-dir /etc/cdnfix --site prod-a -f /srv/cdnfix/urls/prod-a/push.txt push
  cdnfix --root /opt/cdnfix --site prod-a -f urls/prod-a/push.txt push`,
	Run: pushCacheFunc,
}

func init() {
	rootCmd.AddCommand(pushCacheCmd)
}

func pushCache() error {
	urlList, err := readURLs(viper.GetString("urls"), viper.GetString("urlfile"))
	if err != nil || len(urlList) == 0 {
		return err
	}
	return executeAction(selectedSite(), "push", viper.GetString("urlfile"), urlList, "")
}

func pushCacheFunc(cmd *cobra.Command, args []string) {
	if err := pushCache(); err != nil {
		log.Error().Msgf("%s", err.Error())
	}
}
