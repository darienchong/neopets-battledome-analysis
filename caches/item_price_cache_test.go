package caches

import (
	"os"
	"testing"
)

func TestSaveToFile(t *testing.T) {
	dataSource := NewJellyNeoDataSource()
	target, err := ItemPriceCacheInstance(dataSource)
	if err != nil {
		t.Fatalf("%s", err)
	}
	target.Price("Green Apple")
	target.Close()
	_, err = os.Stat(dataSource.FilePath())
	if os.IsNotExist(err) {
		t.Fatalf("Cache file does not exist")
	}
}
