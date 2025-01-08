package models

import (
	"fmt"
	"log/slog"

	"github.com/darienchong/neopets-battledome-analysis/helpers"
)

type DropsAnalysis struct {
	Metadata *DropsMetadata
	Items    map[string]*BattledomeItem
}

func NewAnalysisResultFromDrops(drops *BattledomeDrops) *DropsAnalysis {
	res := new(DropsAnalysis)
	res.Metadata = drops.Metadata.Copy()
	res.Items = drops.Items
	return res
}

func (result *DropsAnalysis) GetItemsOrderedByPrice() []*BattledomeItem {
	items := []*BattledomeItem{}
	for _, v := range result.Items {
		items = append(items, v)
	}
	return helpers.OrderByDescending(items, func(item *BattledomeItem) float64 {
		return item.IndividualPrice
	})
}

func (result *DropsAnalysis) GetItemsOrderedByProfit() []*BattledomeItem {
	items := []*BattledomeItem{}
	for _, v := range result.Items {
		items = append(items, v)
	}
	return helpers.OrderByDescending(items, func(item *BattledomeItem) float64 {
		return float64(item.Quantity) * item.IndividualPrice
	})
}

func (result *DropsAnalysis) GetTotalProfit() float64 {
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

func (res *DropsAnalysis) EstimateDropRates() []*ItemDropRate {
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
