package tests

import (
	"log/slog"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	code := m.Run()
	os.Exit(code)
}
