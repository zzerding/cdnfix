package cmd

import (
	"fmt"
	"io"
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

System-install defaults:

  config-dir: /etc/cdnfix
  state-dir:  /var/lib/cdnfix
  log-dir:    /var/log/cdnfix

Expected config files:

  <config-dir>/sites.yaml
  <config-dir>/jobs.yaml

Derived runtime paths:

  <state-dir>/cache
  <state-dir>/runs

Environment overrides:

  CDNFIX_CONFIG_DIR
  CDNFIX_STATE_DIR
  CDNFIX_LOG_DIR
  CDNFIX_ROOT

Use --root only for portable deployments. It acts as a shortcut for:

  <root>/config
  <root>/var/lib
  <root>/var/log`,
	Example: `  cdnfix batch
  cdnfix init
  cdnfix --config-dir /etc/cdnfix --state-dir /var/lib/cdnfix --log-dir /var/log/cdnfix query
  cdnfix --site prod-a -u https://example.com/a.js push
  cdnfix --site prod-a -f /etc/cdnfix/urls/prod-a/refresh.txt refresh
  printf '%s\n' https://example.com/a.js | cdnfix --site prod-a push
  cdnfix --root /opt/cdnfix batch`,
}

func Execute() {
	if code := execute(os.Args[1:], os.Stdout, os.Stderr); code != 0 {
		os.Exit(code)
	}
}

func execute(args []string, stdout io.Writer, stderr io.Writer) int {
	if wantsVersion(args) {
		_, _ = fmt.Fprintln(stdout, Version())
		return 0
	}

	rootCmd.SetOut(stdout)
	rootCmd.SetErr(stderr)
	rootCmd.SetArgs(args)
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}

func wantsVersion(args []string) bool {
	return len(args) == 1 && (args[0] == "-v" || args[0] == "--version")
}

func init() {
	rootCmd.PersistentFlags().String("root", "", "Portable deployment shortcut; maps default config/state/log directories to <root>/config, <root>/var/lib, and <root>/var/log")
	rootCmd.PersistentFlags().String("config-dir", "", "Configuration directory; defaults to /etc/cdnfix or <root>/config in portable mode")
	rootCmd.PersistentFlags().String("state-dir", "", "State directory for cache and run metadata; defaults to /var/lib/cdnfix or <root>/var/lib in portable mode")
	rootCmd.PersistentFlags().StringP("envfile", "e", "", "Path to site configuration file; defaults to <config-dir>/sites.yaml with .env fallbacks")
	rootCmd.PersistentFlags().StringP("site", "s", "", "Site name from configuration")
	rootCmd.PersistentFlags().StringP("manifest", "m", "", "Path to jobs manifest file; defaults to <config-dir>/jobs.yaml")
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "Debug mode")
	rootCmd.PersistentFlags().StringP("urls", "u", "", "Comma-separated URLs; ignored when stdin is piped in")
	rootCmd.PersistentFlags().StringP("urlfile", "f", "", "Path to URL file, one URL per line; relative paths are resolved from the current working directory, or from <root> in portable mode; ignored when stdin is piped in")
	rootCmd.PersistentFlags().String("log-dir", "", "Directory for run logs; defaults to /var/log/cdnfix or <root>/var/log in portable mode")
	rootCmd.Flags().BoolP("version", "v", false, "Print version and exit")

	_ = viper.BindPFlag("root_dir", rootCmd.PersistentFlags().Lookup("root"))
	_ = viper.BindPFlag("config_dir", rootCmd.PersistentFlags().Lookup("config-dir"))
	_ = viper.BindPFlag("state_dir", rootCmd.PersistentFlags().Lookup("state-dir"))
	_ = viper.BindPFlag("urls", rootCmd.PersistentFlags().Lookup("urls"))
	_ = viper.BindPFlag("urlfile", rootCmd.PersistentFlags().Lookup("urlfile"))
	_ = viper.BindPFlag("envfile", rootCmd.PersistentFlags().Lookup("envfile"))
	_ = viper.BindPFlag("site", rootCmd.PersistentFlags().Lookup("site"))
	_ = viper.BindPFlag("manifest", rootCmd.PersistentFlags().Lookup("manifest"))
	_ = viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
	_ = viper.BindPFlag("log_dir", rootCmd.PersistentFlags().Lookup("log-dir"))
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
		RootDir:    strings.TrimSpace(viper.GetString("root_dir")),
		ConfigDir:  strings.TrimSpace(viper.GetString("config_dir")),
		StateDir:   strings.TrimSpace(viper.GetString("state_dir")),
		SiteConfig: viper.GetString("envfile"),
		Manifest:   viper.GetString("manifest"),
		LogDir:     strings.TrimSpace(viper.GetString("log_dir")),
	})
	if err != nil {
		log.Fatal().Err(err).Msg("resolve runtime layout")
	}
	viper.Set("root_dir", layout.RootDir)
	viper.Set("config_dir", layout.ConfigDir)
	viper.Set("state_dir", layout.StateDir)
	viper.Set("envfile", layout.SiteConfig)
	viper.Set("manifest", layout.Manifest)
	viper.Set("log_dir", layout.LogDir)
	viper.Set("cache_dir", layout.CacheDir)
	viper.Set("run_dir", layout.RunDir)
	if urlFile := viper.GetString("urlfile"); strings.TrimSpace(urlFile) != "" {
		baseDir := ""
		if layout.RootDir == "" {
			if wd, err := os.Getwd(); err == nil {
				baseDir = wd
			}
		} else {
			baseDir = layout.RootDir
		}
		viper.Set("urlfile", workflow.ResolvePath(baseDir, urlFile))
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

func selectedSite() string {
	return strings.TrimSpace(viper.GetString("site"))
}
