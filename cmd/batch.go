package cmd

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var batchCmd = &cobra.Command{
	Use:   "batch",
	Short: "execute jobs from manifest",
	Run: func(cmd *cobra.Command, args []string) {
		if err := batch(); err != nil {
			log.Error().Msgf("command batch error %s", err.Error())
		}
	},
}

func init() {
	rootCmd.AddCommand(batchCmd)
}

func batch() error {
	manifest := viper.GetString("manifest")
	if manifest == "" {
		return fmt.Errorf("--manifest is required")
	}
	return executeBatch(manifest)
}
