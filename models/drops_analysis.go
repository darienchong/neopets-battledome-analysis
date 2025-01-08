package models

import (
	"fmt"
	"log/slog"
	"math"

	"github.com/darienchong/neopets-battledome-analysis/caches"
	"github.com/darienchong/neopets-battledome-analysis/constants"
	"github.com/darienchong/neopets-battledome-analysis/helpers"
	"github.com/montanaflynn/stats"
)

type BattledomeDropsAnalysis struct {
	Metadata DropsMetadata
	Items    map[string]*BattledomeItem
}

func NewAnalysisResultFromDrops(drops *BattledomeDrops) *BattledomeDropsAnalysis {
	res := new(BattledomeDropsAnalysis)
	res.Metadata = drops.Metadata
	res.Items = drops.Items
	return res
}

func generateProfitData(items map[string]*BattledomeItem) ([]float64, error) {
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
			profitData = append(profitData, helpers.LazyWhen(item.IndividualPrice <= 0, func() float64 { return itemPriceCache.GetPrice(item.Name) }, func() float64 { return item.IndividualPrice }))
		}
	}

	return profitData, nil
}

func (analysis *BattledomeDropsAnalysis) GetMeanDropsProfit() (float64, error) {
	profitData, err := generateProfitData(analysis.Items)
	if err != nil {
		return 0.0, err
	}

	mean, err := stats.Mean(profitData)
	if err != nil {
		return 0.0, err
	}

	return mean * constants.BATTLEDOME_DROPS_PER_DAY, nil
}

func (analysis *BattledomeDropsAnalysis) GetDropsProfitStdev() (float64, error) {
	profitData, err := generateProfitData(analysis.Items)
	if err != nil {
		return 0.0, err
	}

	stdev, err := stats.StandardDeviationSample(profitData)
	if err != nil {
		return 0.0, err
	}
	return stdev * math.Sqrt(constants.BATTLEDOME_DROPS_PER_DAY), nil
}

func (analysis *BattledomeDropsAnalysis) GetItemsOrderedByPrice() []*BattledomeItem {
	items := []*BattledomeItem{}
	for _, v := range analysis.Items {
		items = append(items, v)
	}
	return helpers.OrderByDescending(items, func(item *BattledomeItem) float64 {
		return item.IndividualPrice
	})
}

func (analysis *BattledomeDropsAnalysis) GetItemsOrderedByProfit() []*BattledomeItem {
	items := []*BattledomeItem{}
	for _, v := range analysis.Items {
		items = append(items, v)
	}
	return helpers.OrderByDescending(items, func(item *BattledomeItem) float64 {
		return float64(item.Quantity) * item.IndividualPrice
	})
}

func (analysis *BattledomeDropsAnalysis) GetTotalProfit() float64 {
	totalProfit := 0.0
	for _, item := range analysis.Items {
		if item.IndividualPrice <= 0 {
			if item.Name == "nothing" {
				continue
			}
			slog.Warn(fmt.Sprintf("Did not include \"%s\" in total profit calculations as its price was not available.", item.Name))
		} else {
			totalProfit += float64(item.Quantity) * item.IndividualPrice
		}
	}
	return totalProfit
}

func (analysis *BattledomeDropsAnalysis) EstimateDropRates() []*ItemDropRate {
	items := helpers.Map(helpers.ToSlice(analysis.Items), func(tuple helpers.Tuple) *BattledomeItem {
		return tuple.Elements[1].(*BattledomeItem)
	})
	totalItemCount := helpers.Sum(helpers.Map(items, func(item *BattledomeItem) int32 {
		return item.Quantity
	}))

	return helpers.Map(items, func(item *BattledomeItem) *ItemDropRate {
		return &ItemDropRate{
			Arena:    analysis.Metadata.Arena,
			ItemName: item.Name,
			DropRate: float64(item.Quantity) / float64(totalItemCount),
		}
	})
}
