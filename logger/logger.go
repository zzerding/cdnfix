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

// RunLogger owns a run-scoped logger and any file handle opened for it.
// Callers can use Logger() directly without mutating global process state.
type RunLogger struct {
	logger zerolog.Logger
	writer io.Writer
	file   *os.File
}

func InitLog() {
	runLogger, err := NewRunLogger("")
	if err == nil {
		log.Logger = *runLogger.Logger()
	}
}

// NewRunLogger builds an isolated logger/writer pair for one run.
func NewRunLogger(logFile string) (*RunLogger, error) {
	logger, writer, file, err := buildLogger(logFile)
	if err != nil {
		return nil, err
	}
	return &RunLogger{
		logger: logger,
		writer: writer,
		file:   file,
	}, nil
}

func (l *RunLogger) Logger() *zerolog.Logger {
	return &l.logger
}

func (l *RunLogger) Writer() io.Writer {
	return l.writer
}

func (l *RunLogger) Close() error {
	if l == nil || l.file == nil {
		return nil
	}
	return l.file.Close()
}

// Configure is a compatibility wrapper for existing call sites that still use
// the package-global zerolog logger. Prefer NewRunLogger for new code.
func Configure(logFile string) (func(), error) {
	prev := log.Logger
	runLogger, err := NewRunLogger(logFile)
	if err != nil {
		return nil, err
	}
	log.Logger = *runLogger.Logger()
	return func() {
		_ = runLogger.Close()
		log.Logger = prev
	}, nil
}

func SetLogLevel(level string) {
	log.Logger = log.Logger.Level(parseLevel(level))
}

func buildLogger(logFile string) (zerolog.Logger, io.Writer, *os.File, error) {
	level := zerolog.InfoLevel
	if viper.GetBool("debug") {
		level = zerolog.DebugLevel
	}

	var file *os.File
	var output io.Writer = os.Stderr
	if logFile != "" {
		if err := os.MkdirAll(filepath.Dir(logFile), 0755); err != nil {
			return zerolog.Logger{}, nil, nil, fmt.Errorf("failed to create log dir: %w", err)
		}
		f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return zerolog.Logger{}, nil, nil, fmt.Errorf("failed to open log file: %w", err)
		}
		file = f
		output = io.MultiWriter(os.Stderr, file)
	}

	writer := zerolog.ConsoleWriter{Out: output, NoColor: true}
	logger := zerolog.New(writer).With().Timestamp().Logger().Level(level)
	return logger, output, file, nil
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
