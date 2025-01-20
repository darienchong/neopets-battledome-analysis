package tests

import (
	"fmt"
	"log/slog"
	"os"
	"testing"

	"github.com/darienchong/neopets-battledome-analysis/caches"
	"github.com/darienchong/neopets-battledome-analysis/helpers"
)

func TestSaveToFile(t *testing.T) {
	dataSource := caches.NewJellyNeoDataSource()
	target, err := caches.GetItemPriceCacheInstance(dataSource)
	if err != nil {
		t.Fatalf("%s", err)
	}
	target.GetPrice("Green Apple")
	target.Close()
	_, err = os.Stat(dataSource.GetFilePath())
	if os.IsNotExist(err) {
		t.Fatalf("Cache file does not exist")
	}
}

func testGetPriceFromItemDb(itemName string, t *testing.T) {
	dataSource := caches.NewItemDbDataSource()
	target, err := caches.GetItemPriceCacheInstance(dataSource)
	if err != nil {
		t.Fatalf("%s", err)
	}
	defer target.Close()
	price := target.GetPrice(itemName)

	if price <= 0 {
		t.Fatalf(`Failed to retrieve price for "%s" from ItemDb! The retrieved price was %f`, itemName, price)
	}

	slog.Info(fmt.Sprintf(`The price from ItemDb for "%s" was %s NP`, itemName, helpers.FormatFloat(price)))
}

func testGetPriceFromJellyNeo(itemName string, t *testing.T) {
	dataSource := caches.NewJellyNeoDataSource()
	target, err := caches.GetItemPriceCacheInstance(dataSource)
	if err != nil {
		t.Fatalf("%s", err)
	}
	defer target.Close()
	price := target.GetPrice(itemName)

	if price <= 0 {
		t.Fatalf(`Failed to retrieve price for "%s" from JellyNeo! The retrieved price was %f`, itemName, price)
	}

	slog.Info(fmt.Sprintf(`The price from JellyNeo for "%s" was %s NP`, itemName, helpers.FormatFloat(price)))
}

func TestGetPriceFromJellyNeo(t *testing.T) {
	items := []string{
		"Green Apple",
		"The Rock Pool and You: Your guide to Mystery Island Petpets",
		"Eo Codestone",
	}
	for _, item := range items {
		testGetPriceFromJellyNeo(item, t)
	}
}

func TestGetPriceFromItemDb(t *testing.T) {
	items := []string{
		"Green Apple",
		"The Rock Pool and You: Your guide to Mystery Island Petpets",
		"Eo Codestone",
	}
	for _, item := range items {
		testGetPriceFromItemDb(item, t)
	}
}
