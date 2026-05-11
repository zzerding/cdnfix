package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func InitLog() {
	logger, _, err := buildLogger("")
	if err == nil {
		log.Logger = logger
	}
}

func Configure(logFile string) (func(), error) {
	prev := log.Logger
	logger, file, err := buildLogger(logFile)
	if err != nil {
		return nil, err
	}
	log.Logger = logger
	return func() {
		if file != nil {
			_ = file.Close()
		}
		log.Logger = prev
	}, nil
}

func SetLogLevel(level string) {
	log.Logger = log.Logger.Level(parseLevel(level))
}

func buildLogger(logFile string) (zerolog.Logger, *os.File, error) {
	level := zerolog.InfoLevel
	if viper.GetBool("debug") {
		level = zerolog.DebugLevel
	}

	var file *os.File
	var output io.Writer = os.Stderr
	if logFile != "" {
		if err := os.MkdirAll(filepath.Dir(logFile), 0755); err != nil {
			return zerolog.Logger{}, nil, fmt.Errorf("failed to create log dir: %w", err)
		}
		f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return zerolog.Logger{}, nil, fmt.Errorf("failed to open log file: %w", err)
		}
		file = f
		output = io.MultiWriter(os.Stderr, file)
	}

	writer := zerolog.ConsoleWriter{Out: output, NoColor: true}
	logger := zerolog.New(writer).With().Timestamp().Logger().Level(level)
	return logger, file, nil
}

func parseLevel(level string) zerolog.Level {
	switch level {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	default:
		return zerolog.InfoLevel
	}
}
