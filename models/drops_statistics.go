package models

import (
	"fmt"
	"math"

	"github.com/darienchong/neopets-battledome-analysis/constants"
)

type DropsStatistics struct {
	Arena                       string
	MeanItemProfit              float64
	MedianItemProfit            float64
	ItemProfitStandardDeviation float64
}

func (stats *DropsStatistics) GetDropsProfitMean() float64 {
	return stats.MeanItemProfit * constants.BATTLEDOME_DROPS_PER_DAY
}

func (stats *DropsStatistics) GetDropsProfitStandardDeviation() float64 {
	return stats.ItemProfitStandardDeviation * math.Sqrt(constants.BATTLEDOME_DROPS_PER_DAY)
}

func (stats *DropsStatistics) String() string {
	return fmt.Sprintf("%s|%f|%f|%f\n", stats.Arena, stats.MeanItemProfit, stats.MedianItemProfit, stats.ItemProfitStandardDeviation)
}
