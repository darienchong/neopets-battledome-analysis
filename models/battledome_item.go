package models

import (
	"fmt"

	"github.com/darienchong/neopets-battledome-analysis/caches"
	"github.com/darienchong/neopets-battledome-analysis/helpers"
	"github.com/palantir/stacktrace"
)

type ItemName string

type BattledomeItem struct {
	Metadata BattledomeItemMetadata
	Name     ItemName
	Quantity int32
}

func (first *BattledomeItem) Union(second *BattledomeItem) (*BattledomeItem, error) {
	if first.Name != second.Name {
		return nil, fmt.Errorf("tried to union two items that did not have the same name: %s and %s", first.Name, second.Name)
	}

	combined := &BattledomeItem{}
	combinedMetadata, err := first.Metadata.Combine(second.Metadata)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to combine metadata \"%s\" and \"%s\"", first.Metadata, second.Metadata)
	}
	combined.Metadata = combinedMetadata
	combined.Name = first.Name
	combined.Quantity = first.Quantity + second.Quantity
	return combined, nil
}

func (i *BattledomeItem) Profit(itemPriceCache caches.ItemPriceCache) float64 {
	return float64(i.Quantity) * itemPriceCache.Price(string(i.Name))
}

func (i *BattledomeItem) PercentageProfit(itemPriceCache caches.ItemPriceCache, items NormalisedBattledomeItems) (float64, error) {
	var defaultValue float64

	totalProfit, err := items.TotalProfit(itemPriceCache)
	if err != nil {
		return defaultValue, helpers.PropagateWithSerialisedValue(err, "failed to get total profit for \"%s\"", "failed to get total profit for a battledome item; additionally encountered an error while trying to serialise the value to log: %s", i)
	}
	return i.Profit(itemPriceCache) / totalProfit, nil
}

func (i *BattledomeItem) DropRate(items NormalisedBattledomeItems) float64 {
	return float64(i.Quantity) / float64(items.TotalItemQuantity())
}

func (i *BattledomeItem) String() string {
	return fmt.Sprintf("[%s] %s × %d", i.Metadata.String(), i.Name, i.Quantity)
}

func (i *BattledomeItem) Copy() *BattledomeItem {
	return &BattledomeItem{
		Metadata: i.Metadata,
		Name:     i.Name,
		Quantity: i.Quantity,
	}
}
