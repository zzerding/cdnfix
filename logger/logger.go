package logger

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func InitLog() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	log.Logger = log.With().Logger()
	log.Logger = log.Logger.Level(zerolog.InfoLevel)
	if viper.GetBool("debug") {
		SetLogLevel("debug")
		fmt.Println("debug mod enbale")
	}
}
func SetLogLevel(level string) {
	switch level {
	case "debug":
		log.Logger = log.Logger.Level(zerolog.DebugLevel)
	case "info":
		log.Logger = log.Logger.Level(zerolog.InfoLevel)
	case "warn":
		log.Logger = log.Logger.Level(zerolog.WarnLevel)
	case "error":
		log.Logger = log.Logger.Level(zerolog.ErrorLevel)
	case "fatal":
		log.Logger = log.Logger.Level(zerolog.FatalLevel)
	default:
		log.Logger = log.Logger.Level(zerolog.InfoLevel)
	}
}
