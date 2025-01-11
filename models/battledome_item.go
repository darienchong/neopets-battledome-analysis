package models

import (
	"fmt"

	"github.com/darienchong/neopets-battledome-analysis/caches"
	"github.com/darienchong/neopets-battledome-analysis/helpers"
)

type ItemName string

type BattledomeItem struct {
	Metadata BattledomeItemMetadata
	Name     ItemName
	Quantity int32
}

func (first *BattledomeItem) Union(second *BattledomeItem) (*BattledomeItem, error) {
	if first.Name != second.Name {
		return nil, fmt.Errorf("tried to union two items that did not have the same name: %s and %s", first, second)
	}

	combined := &BattledomeItem{}
	combined.Name = first.Name
	combined.Quantity = first.Quantity + second.Quantity
	return combined, nil
}

func (item *BattledomeItem) GetProfit(itemPriceCache *caches.ItemPriceCache) float64 {
	return float64(item.Quantity) * itemPriceCache.GetPrice(string(item.Name))
}

func (item *BattledomeItem) GetPercentageProfit(itemPriceCache *caches.ItemPriceCache, items NormalisedBattledomeItems) (float64, error) {
	var defaultValue float64

	totalProfit, err := items.GetTotalProfit()
	if err != nil {
		return defaultValue, err
	}
	return item.GetProfit(itemPriceCache) / totalProfit, nil
}

func (item *BattledomeItem) GetDropRate(items NormalisedBattledomeItems) float64 {
	return float64(item.Quantity) / float64(helpers.Sum(helpers.Map(helpers.ToSlice(items), func(tuple helpers.Tuple) int32 {
		return tuple.Elements[1].(*BattledomeItem).Quantity
	})))
}

func (item *BattledomeItem) String() string {
	return fmt.Sprintf("[%s] %s × %d", item.Metadata.String(), item.Name, item.Quantity)
}

func (item *BattledomeItem) Copy() *BattledomeItem {
	return &BattledomeItem{
		Metadata: item.Metadata,
		Name:     item.Name,
		Quantity: item.Quantity,
	}
}
