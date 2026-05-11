package cmd

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zzerding/cdnfix/workflow"
)

var batchCmd = &cobra.Command{
	Use:   "batch",
	Short: "execute jobs from manifest",
	Long: `Execute jobs from the jobs manifest.

By default, the manifest path is <config-dir>/jobs.yaml, which is typically
/etc/cdnfix/jobs.yaml in a system install. Each job selects a site, an action,
and a URL file. Job file paths are resolved relative to the manifest file
itself, not the shell working directory.

Use --root only for portable deployments where config/, urls/, and var/ live
under one directory.`,
	Example: `  cdnfix batch
  cdnfix --config-dir /etc/cdnfix --state-dir /var/lib/cdnfix --log-dir /var/log/cdnfix batch
  cdnfix --root /opt/cdnfix batch
  cdnfix --manifest /opt/cdnfix/config/jobs.yaml batch`,
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
	return workflow.ExecuteBatch(runtimePaths(), manifest)
}
