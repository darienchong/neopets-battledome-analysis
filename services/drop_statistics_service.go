package services

import (
	"github.com/darienchong/neopets-battledome-analysis/models"
	"github.com/montanaflynn/stats"
)

type DropStatisticsService struct{}

func NewDropStatisticsService() *DropStatisticsService {
	return &DropStatisticsService{}
}

func (estimator *DropStatisticsService) Estimate(drop *models.BattledomeDrops) (*models.DropsStatistics, error) {
	var arena string = drop.Metadata.Arena
	profitData := []float64{}
	for _, item := range drop.Items {
		for j := 0; j < int(item.Quantity); j++ {
			profitData = append(profitData, item.IndividualPrice)
		}
	}

	if len(profitData) == 0 {
		return &models.DropsStatistics{
			Arena:                       arena,
			MeanItemProfit:              0,
			MedianItemProfit:            0,
			ItemProfitStandardDeviation: 0,
		}, nil
	}

	mean, err := stats.Mean(profitData)
	if err != nil {
		return nil, err
	}
	median, err := stats.Median(profitData)
	if err != nil {
		return nil, err
	}
	stdev, err := stats.StandardDeviationSample(profitData)
	if err != nil {
		return nil, err
	}
	return &models.DropsStatistics{Arena: arena, MeanItemProfit: mean, MedianItemProfit: median, ItemProfitStandardDeviation: stdev}, nil
}
