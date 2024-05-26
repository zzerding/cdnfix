package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zzerding/refresh-cdn/cloud/tencent"
)

var puahCacheCmd = &cobra.Command{
	Use:   "push",
	Short: "push cache to cdn",
	Long:  "push cache  cnd for tencent use -f or -u input url list",
	Run:   pushCacheFunc,
}

func init() {
	puahCacheCmd.Flags().StringP("cachefile", "c", ".task_push.cache", "push cache task file")
	viper.BindPFlag("push_task_cache_file", puahCacheCmd.Flags().Lookup("cachefile"))

	rootCmd.AddCommand(puahCacheCmd)
}

func pushCache() error {
	urls := viper.GetString("urls")
	filePath := viper.GetString("urlfile")
	log.Debug().Msgf("refresh urls %s,urlfile: %s", urls, filePath)
	urlList, err := readURLs(urls, filePath)
	if err != nil || len(urlList) == 0 {
		return err
	}

	client, err := tencent.CreateCDNClient()
	if err != nil {
		return err
	}

	if err := client.PushUrlsCache(urlList); err != nil {
		return err
	}
	log.Info().Msg("push cache tasks urls is push to cloud cdn")
	return nil
}
func pushCacheFunc(cmd *cobra.Command, args []string) {
	if err := pushCache(); err != nil {
		log.Error().Msgf(" %s", err.Error())
	}
}
