package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var pushCacheCmd = &cobra.Command{
	Use:   "push",
	Short: "push cache for one site",
	Long:  "push cache for one site by --site and --urls/--file",
	Run:   pushCacheFunc,
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
