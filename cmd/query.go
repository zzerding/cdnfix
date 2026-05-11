package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/zzerding/cdnfix/logger"
	"github.com/zzerding/cdnfix/workflow"
)

func init() {
	rootCmd.AddCommand(queryCmd)
}

func query() error {
	siteFilter := selectedSite()
	runLogger, err := logger.NewRunLogger("")
	if err != nil {
		return err
	}
	defer func() {
		_ = runLogger.Close()
	}()
	runLog := runLogger.Logger()

	return workflow.QueryTasks(runtimePaths(), siteFilter, runLog)
}

var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "query pending task status from structured cache",
	Long: `Query pending task status from the structured cache under <state-dir>/cache.

Without --site, cdnfix scans all configured sites. With --site, it only checks
the selected site.

System-install defaults use /etc/cdnfix for configuration, /var/lib/cdnfix for
task state, and /var/log/cdnfix for logs. Use --root only for portable
deployments.`,
	Example: `  cdnfix query
  cdnfix --site prod-a query
  cdnfix --config-dir /etc/cdnfix --state-dir /var/lib/cdnfix query
  cdnfix --root /opt/cdnfix query`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := query(); err != nil {
			log.Error().Msgf("command query error %s", err.Error())
		}
	},
}
