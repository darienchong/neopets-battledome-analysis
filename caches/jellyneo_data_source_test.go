package caches

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

func TestGetJellyNeoPrice(t *testing.T) {
	itemName := "Green Apple"
	target := NewJellyNeoDataSource()
	price := target.GetPrice(itemName)

	if price <= 0 {
		t.Fatalf("failed to retrieve price for \"%s\" from JellyNeo! The retrieved price was %f", itemName, price)
	}
}
