package logger

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func TestNewRunLoggerDoesNotMutateGlobalLogger(t *testing.T) {
	prev := log.Logger
	defer func() {
		log.Logger = prev
	}()

	var global bytes.Buffer
	log.Logger = zerolog.New(&global)

	logPath := filepath.Join(t.TempDir(), "run.log")
	runLogger, err := NewRunLogger(logPath)
	if err != nil {
		t.Fatalf("NewRunLogger() error = %v", err)
	}
	defer func() {
		_ = runLogger.Close()
	}()

	log.Info().Msg("global message")
	runLogger.Logger().Info().Msg("run message")
	if _, err := runLogger.Writer().Write([]byte("plain line\n")); err != nil {
		t.Fatalf("Writer().Write() error = %v", err)
	}

	globalOutput := global.String()
	if !strings.Contains(globalOutput, "global message") {
		t.Fatalf("global logger output missing message: %q", globalOutput)
	}
	if strings.Contains(globalOutput, "run message") {
		t.Fatalf("run logger unexpectedly wrote to global logger: %q", globalOutput)
	}

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	fileOutput := string(content)
	if !strings.Contains(fileOutput, "run message") {
		t.Fatalf("run log file missing message: %q", fileOutput)
	}
	if !strings.Contains(fileOutput, "plain line") {
		t.Fatalf("run log file missing direct writer output: %q", fileOutput)
	}
	if strings.Contains(fileOutput, "global message") {
		t.Fatalf("global logger unexpectedly wrote to run log file: %q", fileOutput)
	}
}
