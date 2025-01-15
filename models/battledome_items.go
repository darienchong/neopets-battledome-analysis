package models

import (
	"fmt"
	"math"
	"math/rand/v2"
	"slices"
	"sort"

	"github.com/darienchong/neopets-battledome-analysis/caches"
	"github.com/darienchong/neopets-battledome-analysis/constants"
	"github.com/darienchong/neopets-battledome-analysis/helpers"
	"github.com/montanaflynn/stats"
	"github.com/palantir/stacktrace"
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
				return nil, helpers.PropagateWithSerialisedValue(err, "failed to add %s to normalised items", "failed to add item to normalised items; additionally encountered an error while trying to serialise the item to log: %s", item)
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
		return nil, stacktrace.Propagate(err, "failed to get item price cache")
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
	if len(profitData) == 0 {
		return 0.0, nil
	}

	if err != nil {
		return 0.0, stacktrace.Propagate(err, "failed to generate profit data")
	}

	mean, err := stats.Mean(profitData)
	if err != nil {
		return 0.0, stacktrace.Propagate(err, "failed to get mean of profit data")
	}

	return mean * constants.BATTLEDOME_DROPS_PER_DAY, nil
}

func (items NormalisedBattledomeItems) GetDropsProfitStdev() (float64, error) {
	profitData, err := generateProfitData(items)
	if len(profitData) == 0 {
		return 0.0, nil
	}

	if err != nil {
		return 0.0, stacktrace.Propagate(err, "failed to generate profit data")
	}

	stdev, err := stats.StandardDeviationSample(profitData)
	if err != nil {
		return 0.0, stacktrace.Propagate(err, "failed to get sample standard deviation")
	}
	return stdev * math.Sqrt(constants.BATTLEDOME_DROPS_PER_DAY), nil
}

func (items NormalisedBattledomeItems) GetItemsOrderedByPrice() ([]*BattledomeItem, error) {
	itemPriceCache, err := caches.GetItemPriceCacheInstance()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get item price cache")
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
		return nil, stacktrace.Propagate(err, "failed to get item price cache")
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
		return defaultValue, stacktrace.Propagate(err, "failed to get item price cache")
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
		return helpers.When(string(item.Name) == "nothing", 0, int(item.Quantity))
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

// Percentile calculates the Pth percentile of a slice of float64 values.
// P should be between 0 and 1 (e.g., 0.25 for the 25th percentile).
func percentile(values []float64, p float64) float64 {
	if len(values) == 0 {
		panic("cannot calculate percentile of an empty slice")
	}
	if p < 0 || p > 1 {
		panic("p must be between 0 and 1")
	}

	// Sort the slice in ascending order
	sort.Float64s(values)

	// Calculate the rank
	n := float64(len(values))
	rank := p * (n - 1)
	lower := int(math.Floor(rank))
	upper := int(math.Ceil(rank))

	// Interpolate if the rank is not an integer
	if lower == upper {
		return values[lower]
	}
	weight := rank - float64(lower)
	return values[lower]*(1-weight) + values[upper]*weight
}

func (items NormalisedBattledomeItems) GetProfitConfidenceInterval() (float64, float64, error) {
	profitData, err := generateProfitData(items)
	if err != nil {
		return 0.0, 0.0, stacktrace.Propagate(err, "failed to generate profit data")
	}

	if len(profitData) == 0 {
		return 0.0, 0.0, nil
	}

	bootstrap_sums := []float64{}
	for _ = range constants.NUMBER_OF_BOOTSTRAP_SAMPLES_FOR_ESTIMATING_CONFIDENCE_INTERVAL {
		bootstrap_sum := 0.0

		for _ = range constants.BATTLEDOME_DROPS_PER_DAY {
			bootstrap_sum += profitData[rand.IntN(len(profitData))]
		}

		bootstrap_sums = append(bootstrap_sums, bootstrap_sum)
	}

	slices.Sort(bootstrap_sums)
	return percentile(bootstrap_sums, constants.SIGNIFICANCE_LEVEL/2), percentile(bootstrap_sums, 1-constants.SIGNIFICANCE_LEVEL/2), nil
}
