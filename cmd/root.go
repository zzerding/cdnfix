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
	Short: "refresh and push cache of tencent cloud cdn",
	Long:  `This is a CDN management application that allows you to query refresh history.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringP("envfile", "e", ".env", "Path to site configuration file")
	rootCmd.PersistentFlags().StringP("site", "s", "", "Site name from configuration")
	rootCmd.PersistentFlags().StringP("manifest", "m", "", "Path to jobs manifest file")
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "Debug mode")
	rootCmd.PersistentFlags().StringP("urls", "u", "", "Comma-separated URLs")
	rootCmd.PersistentFlags().StringP("urlfile", "f", "", "Path to URL file, one URL per line")
	rootCmd.PersistentFlags().String("log-dir", "./var/logs", "Directory for run logs")
	rootCmd.PersistentFlags().String("cache-dir", "./var/cache", "Directory for task cache files")
	rootCmd.PersistentFlags().String("run-dir", "./var/runs", "Directory for run metadata")

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
	envfile := viper.GetString("envfile")
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
