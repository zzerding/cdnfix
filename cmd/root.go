package cmd

import (
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "refresh-cdn",
	Short: "refresh cdn of tencentcloud",
	Long:  `This is a CDN management application that allows you to query refresh history.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringP("urls", "u", "", "Comma separated URLs to refresh")
	rootCmd.PersistentFlags().StringP("urlfile", "f", "", "Path to file containing URLs to refresh")
	rootCmd.PersistentFlags().StringP("envfile", "e", ".env", "Path to configuration file, default .env")
	viper.BindPFlag("urls", rootCmd.PersistentFlags().Lookup("urls"))
	viper.BindPFlag("urlfile", rootCmd.PersistentFlags().Lookup("urlfile"))
	viper.BindPFlag("envfile", rootCmd.PersistentFlags().Lookup("envfile"))
	initConfig()
}

// 定义一个结构体来保存配置信息
type Config struct {
	//fromat Secret_ID=xxx Secret_Key=xxx Region=xxx
	SecretID  string `mapstructure:"Secret_ID"`
	SecretKey string `mapstructure:"Secret_Key"`
	Region    string //https://github.com/TencentCloud/tencentcloud-sdk-go/blob/master/tencentcloud/common/regions/regions.go
}

func initConfig() {
	// 使用 viper 读取配置文件
	envfile := viper.GetString("envfile")
	log.Debug().Msgf("env file path is %s", envfile)
	if envfile == "" {
		log.Info().Msgf("no configuration file specified %s", envfile)
	}
	viper.SetConfigFile(envfile)
	viper.SetConfigType("env")
	viper.AutomaticEnv()
	// 尝试读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		log.Error().Msgf("error reading config file: %s", err.Error())
		os.Exit(1)
	}
}
