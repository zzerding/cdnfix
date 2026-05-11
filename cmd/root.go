package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zzerding/cdnfix/logger"
	"github.com/zzerding/cdnfix/workflow"
)

var rootCmd = &cobra.Command{
	Use:   "cdnfix",
	Short: "Manage Tencent Cloud CDN refresh and push jobs",
	Long: `cdnfix manages Tencent Cloud CDN refresh and push operations with an
explicit site/job model.

By default, all paths are resolved from the application root, which is the
directory containing the cdnfix executable. The default layout is:

  <root>/config/sites.yaml
  <root>/config/jobs.yaml
  <root>/var/logs
  <root>/var/cache
  <root>/var/runs

Use --root when the binary is not deployed inside the application root.`,
	Example: `  cdnfix --root /opt/cdnfix batch
  cdnfix --site prod-a -f urls/prod-a/refresh.txt refresh
  cdnfix --site prod-a -u https://example.com/a.js push
  cdnfix query
  cdnfix --site prod-a query`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().String("root", "", "Application root directory; defaults to the executable directory")
	rootCmd.PersistentFlags().String("config-dir", "", "Configuration directory; defaults to <root>/config")
	rootCmd.PersistentFlags().StringP("envfile", "e", "", "Path to site configuration file; defaults to <config-dir>/sites.yaml with .env fallbacks")
	rootCmd.PersistentFlags().StringP("site", "s", "", "Site name from configuration")
	rootCmd.PersistentFlags().StringP("manifest", "m", "", "Path to jobs manifest file; defaults to <config-dir>/jobs.yaml")
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "Debug mode")
	rootCmd.PersistentFlags().StringP("urls", "u", "", "Comma-separated URLs")
	rootCmd.PersistentFlags().StringP("urlfile", "f", "", "Path to URL file, one URL per line; relative paths are resolved from <root>")
	rootCmd.PersistentFlags().String("log-dir", "", "Directory for run logs; defaults to <root>/var/logs")
	rootCmd.PersistentFlags().String("cache-dir", "", "Directory for task cache files; defaults to <root>/var/cache")
	rootCmd.PersistentFlags().String("run-dir", "", "Directory for run metadata; defaults to <root>/var/runs")

	_ = viper.BindPFlag("root_dir", rootCmd.PersistentFlags().Lookup("root"))
	_ = viper.BindPFlag("config_dir", rootCmd.PersistentFlags().Lookup("config-dir"))
	_ = viper.BindPFlag("urls", rootCmd.PersistentFlags().Lookup("urls"))
	_ = viper.BindPFlag("urlfile", rootCmd.PersistentFlags().Lookup("urlfile"))
	_ = viper.BindPFlag("envfile", rootCmd.PersistentFlags().Lookup("envfile"))
	_ = viper.BindPFlag("site", rootCmd.PersistentFlags().Lookup("site"))
	_ = viper.BindPFlag("manifest", rootCmd.PersistentFlags().Lookup("manifest"))
	_ = viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
	_ = viper.BindPFlag("log_dir", rootCmd.PersistentFlags().Lookup("log-dir"))
	_ = viper.BindPFlag("cache_dir", rootCmd.PersistentFlags().Lookup("cache-dir"))
	_ = viper.BindPFlag("run_dir", rootCmd.PersistentFlags().Lookup("run-dir"))
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	logger.InitLog()
	execPath, err := os.Executable()
	if err != nil {
		log.Fatal().Err(err).Msg("resolve executable path")
	}
	if resolvedPath, err := filepath.EvalSymlinks(execPath); err == nil {
		execPath = resolvedPath
	}
	layout, err := workflow.ResolveLayout(execPath, workflow.LayoutOptions{
		RootDir:    viper.GetString("root_dir"),
		ConfigDir:  viper.GetString("config_dir"),
		SiteConfig: viper.GetString("envfile"),
		Manifest:   viper.GetString("manifest"),
		LogDir:     viper.GetString("log_dir"),
		CacheDir:   viper.GetString("cache_dir"),
		RunDir:     viper.GetString("run_dir"),
	})
	if err != nil {
		log.Fatal().Err(err).Msg("resolve runtime layout")
	}
	viper.Set("root_dir", layout.RootDir)
	viper.Set("config_dir", layout.ConfigDir)
	viper.Set("envfile", layout.SiteConfig)
	viper.Set("manifest", layout.Manifest)
	viper.Set("log_dir", layout.Paths.LogDir)
	viper.Set("cache_dir", layout.Paths.CacheDir)
	viper.Set("run_dir", layout.Paths.RunDir)
	if urlFile := viper.GetString("urlfile"); strings.TrimSpace(urlFile) != "" {
		viper.Set("urlfile", workflow.ResolvePath(layout.RootDir, urlFile))
	}

	envfile := layout.SiteConfig
	log.Debug().Msgf("env file path is %s", envfile)
	if envfile == "" {
		return
	}
	viper.SetConfigFile(envfile)
	if configType := strings.TrimPrefix(filepath.Ext(envfile), "."); configType != "" {
		viper.SetConfigType(configType)
	} else {
		viper.SetConfigType("env")
	}
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		log.Info().Msgf("reading config file: %s", err.Error())
	}
}

func runtimePaths() workflow.Paths {
	return workflow.Paths{
		LogDir:   viper.GetString("log_dir"),
		CacheDir: viper.GetString("cache_dir"),
		RunDir:   viper.GetString("run_dir"),
	}
}

func commandLogPath(command string, site string) string {
	now := time.Now()
	dateDir := now.Format("2006-01-02")
	name := strings.TrimSpace(site)
	if name == "" {
		name = "all-sites"
	}
	file := fmt.Sprintf("%s.%s.%s.log", name, command, now.Format("20060102T150405"))
	return filepath.Join(viper.GetString("log_dir"), dateDir, file)
}
