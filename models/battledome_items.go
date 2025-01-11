package models

import (
	"fmt"
	"math"

	"github.com/darienchong/neopets-battledome-analysis/caches"
	"github.com/darienchong/neopets-battledome-analysis/constants"
	"github.com/darienchong/neopets-battledome-analysis/helpers"
	"github.com/montanaflynn/stats"
)

type BattledomeItems []*BattledomeItem

func (items BattledomeItems) Normalise() (NormalisedBattledomeItems, error) {
	normalisedItems := NormalisedBattledomeItems{}
	for _, item := range items {
		_, exists := normalisedItems[item.Name]
		if !exists {
			normalisedItems[item.Name] = item
		} else {
			combined, err := normalisedItems[item.Name].Union(item)
			if err != nil {
				return nil, err
			}
			normalisedItems[item.Name] = combined
		}
	}

	return normalisedItems, nil
}

type NormalisedBattledomeItems map[ItemName]*BattledomeItem

func (normalisedItems NormalisedBattledomeItems) GetMetadata() (BattledomeItemMetadata, error) {
	for _, v := range normalisedItems {
		return v.Metadata, nil
	}

	return BattledomeItemMetadata{}, fmt.Errorf("there was no item to get metadata from")
}

func generateProfitData(items NormalisedBattledomeItems) ([]float64, error) {
	itemPriceCache, err := caches.GetItemPriceCacheInstance()
	if err != nil {
		return nil, err
	}
	defer itemPriceCache.Close()

	profitData := []float64{}
	for _, item := range items {
		if item.Name == "nothing" {
			continue
		}
		for j := 0; j < int(item.Quantity); j++ {
			profitData = append(profitData, itemPriceCache.GetPrice(string(item.Name)))
		}
	}

	return profitData, nil
}

func (items NormalisedBattledomeItems) GetMeanDropsProfit() (float64, error) {
	profitData, err := generateProfitData(items)
	if err != nil {
		return 0.0, err
	}

	mean, err := stats.Mean(profitData)
	if err != nil {
		return 0.0, err
	}

	return mean * constants.BATTLEDOME_DROPS_PER_DAY, nil
}

func (items NormalisedBattledomeItems) GetDropsProfitStdev() (float64, error) {
	profitData, err := generateProfitData(items)
	if err != nil {
		return 0.0, err
	}

	stdev, err := stats.StandardDeviationSample(profitData)
	if err != nil {
		return 0.0, err
	}
	return stdev * math.Sqrt(constants.BATTLEDOME_DROPS_PER_DAY), nil
}

func (items NormalisedBattledomeItems) GetItemsOrderedByPrice() ([]*BattledomeItem, error) {
	itemPriceCache, err := caches.GetItemPriceCacheInstance()
	if err != nil {
		return nil, err
	}
	defer itemPriceCache.Close()

	orderedItems := []*BattledomeItem{}
	for _, v := range items {
		orderedItems = append(orderedItems, v)
	}
	return helpers.OrderByDescending(orderedItems, func(item *BattledomeItem) float64 {
		return itemPriceCache.GetPrice(string(item.Name))
	}), nil
}

func (items NormalisedBattledomeItems) GetItemsOrderedByProfit() ([]*BattledomeItem, error) {
	itemPriceCache, err := caches.GetItemPriceCacheInstance()
	if err != nil {
		return nil, err
	}
	defer itemPriceCache.Close()

	orderedItems := []*BattledomeItem{}
	for _, v := range items {
		orderedItems = append(orderedItems, v)
	}

	return helpers.OrderByDescending(orderedItems, func(item *BattledomeItem) float64 {
		return item.GetProfit(itemPriceCache)
	}), nil
}

func (items NormalisedBattledomeItems) GetTotalProfit() (float64, error) {
	var defaultValue float64

	itemPriceCache, err := caches.GetItemPriceCacheInstance()
	if err != nil {
		return defaultValue, err
	}
	defer itemPriceCache.Close()

	totalProfit := 0.0
	for _, item := range items {
		if item.Name == "nothing" || itemPriceCache.GetPrice(string(item.Name)) <= 0 {
			continue
		} else {
			totalProfit += item.GetProfit(itemPriceCache)
		}
	}
	return totalProfit, nil
}

func (items NormalisedBattledomeItems) GetTotalItemQuantity() int {
	quantities := helpers.Map(helpers.Values(items), func(item *BattledomeItem) int {
		return int(item.Quantity)
	})
	return helpers.Sum(quantities)
}

func (items NormalisedBattledomeItems) EstimateDropRates() []*BattledomeItemDropRate {
	totalItemCount := helpers.Sum(helpers.Map(helpers.Values(items), func(item *BattledomeItem) int32 {
		return item.Quantity
	}))

	return helpers.Map(helpers.Values(items), func(item *BattledomeItem) *BattledomeItemDropRate {
		return &BattledomeItemDropRate{
			Metadata: item.Metadata,
			ItemName: item.Name,
			DropRate: float64(item.Quantity) / float64(totalItemCount),
		}
	})
}
