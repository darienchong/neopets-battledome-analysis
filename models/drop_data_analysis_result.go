package models

import (
	"fmt"
	"log/slog"
	"math"

	"github.com/darienchong/neopets-battledome-analysis/helpers"
	"github.com/montanaflynn/stats"
)

type DropDataAnalysisResult struct {
	Metadata *DropsMetadata
	Items    map[string]*BattledomeItem
}

func NewAnalysisResultFromDrops(drops *BattledomeDrops) *DropDataAnalysisResult {
	res := new(DropDataAnalysisResult)
	res.Metadata = drops.Metadata.Copy()
	res.Items = drops.Items
	return res
}

func (result *DropDataAnalysisResult) GetItemsOrderedByPrice() []*BattledomeItem {
	items := []*BattledomeItem{}
	for _, v := range result.Items {
		items = append(items, v)
	}
	return helpers.OrderByDescending(items, func(item *BattledomeItem) float64 {
		return item.IndividualPrice
	})
}

func (result *DropDataAnalysisResult) GetItemsOrderedByProfit() []*BattledomeItem {
	items := []*BattledomeItem{}
	for _, v := range result.Items {
		items = append(items, v)
	}
	return helpers.OrderByDescending(items, func(item *BattledomeItem) float64 {
		return float64(item.Quantity) * item.IndividualPrice
	})
}

func (result *DropDataAnalysisResult) GetTotalProfit() float64 {
	totalProfit := 0.0
	for _, item := range result.Items {
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

func (result *DropDataAnalysisResult) GetStatistics() (ResultStatistics, error) {
	resultStatistics := new(ResultStatistics)
	resultStatistics.Min = math.MaxFloat64
	resultStatistics.Max = 0
	data := []float64{}
	for _, item := range result.Items {
		for i := 0; i < int(item.Quantity); i++ {
			data = append(data, item.IndividualPrice)
			if item.IndividualPrice > resultStatistics.Max {
				resultStatistics.Max = item.IndividualPrice
			}
			if item.IndividualPrice < resultStatistics.Min {
				resultStatistics.Min = item.IndividualPrice
			}
		}
	}

	resultStatistics.SampleSize = len(data)
	mean, err := stats.Mean(data)
	if err != nil {
		return *resultStatistics, err
	}

	median, err := stats.Median(data)
	if err != nil {
		return *resultStatistics, err
	}

	stdev, err := stats.StandardDeviationSample(data)
	if err != nil {
		return *resultStatistics, err
	}

	resultStatistics.Mean = mean
	resultStatistics.Median = median
	resultStatistics.StandardDeviationSample = stdev
	return *resultStatistics, err
}

func (res *DropDataAnalysisResult) EstimateDropRates() []*ItemDropRate {
	items := helpers.Map(helpers.ToSlice(res.Items), func(tuple helpers.Tuple) *BattledomeItem {
		return tuple.Elements[1].(*BattledomeItem)
	})
	totalItemCount := helpers.Sum(helpers.Map(items, func(item *BattledomeItem) int32 {
		return item.Quantity
	}))

	return helpers.Map(items, func(item *BattledomeItem) *ItemDropRate {
		return &ItemDropRate{
			Arena:    res.Metadata.Arena,
			ItemName: item.Name,
			DropRate: float64(item.Quantity) / float64(totalItemCount),
		}
	})
}
