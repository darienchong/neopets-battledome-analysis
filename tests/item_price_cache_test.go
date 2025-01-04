package tests

import (
	"os"
	"testing"

	"github.com/darienchong/neopetsbattledomeanalysis/caches"
	"github.com/darienchong/neopetsbattledomeanalysis/constants"
)

func TestSaveToFile(t *testing.T) {
	target := caches.GetItemPriceCacheInstance()
	target.GetPrice("Green Apple")
	target.Close()
	_, err := os.Stat(constants.GetItemPriceCacheFilePath())
	if os.IsNotExist(err) {
		t.Fatalf("Cache file does not exist")
	}
}

func TestGetPriceFromItemDb(t *testing.T) {
	target := caches.GetItemPriceCacheInstance()
	defer target.Close()
	price := target.GetPrice("Green Apple")

	if price < 0 {
		t.Fatalf(`Failed to retrieve price from ItemDb! The retrieved price was %f`, price)
	}
}
