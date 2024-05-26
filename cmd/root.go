package cmd

import (
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zzerding/refresh-cdn/logger"
)

var rootCmd = &cobra.Command{
	Use:   "refresh-cdn",
	Short: "refresh cdn of tencentcloud",
	Long:  `This is a CDN management application that allows you to query refresh history.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringP("envfile", "e", ".env", "Path to configuration file, default .env.or you can set system env SECRE_ID and SECRE_KEY")
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "Debug mode")
	viper.BindPFlag("envfile", rootCmd.PersistentFlags().Lookup("envfile"))
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
	cobra.OnInitialize(initConfig)
}

// 定义一个结构体来保存配置信息
type Config struct {
	//fromat Secret_ID=xxx Secret_Key=xxx Region=xxx
	SecretID  string `mapstructure:"Secret_ID"`
	SecretKey string `mapstructure:"Secret_Key"`
	Region    string //https://github.com/TencentCloud/tencentcloud-sdk-go/blob/master/tencentcloud/common/regions/regions.go
}

func initConfig() {
	logger.InitLog()
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
		log.Info().Msgf(" reading config file: %s", err.Error())
	}

}
