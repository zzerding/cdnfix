package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(queryCmd)
}

func query() error {
	return queryTasks(selectedSite())
}

var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "query pending task status from structured cache",
	Run: func(cmd *cobra.Command, args []string) {
		if err := query(); err != nil {
			log.Error().Msgf("command query error %s", err.Error())
		}
	},
}
