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

func TestJellyNeoPrice(t *testing.T) {
	itemName := "Green Apple"
	target := NewJellyNeoDataSource()
	price, err := target.Price(itemName)

	if err != nil || price <= 0 {
		t.Fatalf("failed to retrieve price for %q from JellyNeo! The retrieved price was %f", itemName, price)
	}
}
