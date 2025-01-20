package tests

import (
	"github.com/darienchong/neopets-battledome-analysis/caches"
)

var _ caches.ItemPriceCache = (*MockItemPriceCache)(nil)

type MockItemPriceCache struct{}

func (cache *MockItemPriceCache) GetPrice(itemName string) float64 {
	return 0.0
}

func (cache *MockItemPriceCache) Close() error {
	return nil
}
