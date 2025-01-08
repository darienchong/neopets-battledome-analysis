package services

import (
	"github.com/darienchong/neopets-battledome-analysis/helpers"
	"github.com/darienchong/neopets-battledome-analysis/models"
)

type ChallengerComparisonService struct {
	EmpiricalDropsService *EmpiricalDropsService
	DataComparisonService *DataComparisonService
}

func NewChallengerComparisonService() *ChallengerComparisonService {
	return &ChallengerComparisonService{
		EmpiricalDropsService: NewEmpiricalDropsService(),
		DataComparisonService: NewDataComparisonService(),
	}
}

func (service *ChallengerComparisonService) CompareAll() ([]*models.ComparisonResult, error) {
	data, err := service.EmpiricalDropsService.GetDropsGroupedByMetadata()
	if err != nil {
		return nil, err
	}

	comparisonResultsByMetadata := map[models.DropsMetadata]*models.ComparisonResult{}
	for metadata, drops := range data {
		combinedDrops := helpers.Reduce(drops, func(first *models.BattledomeDrops, second *models.BattledomeDrops) *models.BattledomeDrops {
			combined, err := first.Union(second)
			if err != nil {
				panic(err)
			}
			return combined
		})
		res, err := service.DataComparisonService.ToComparisonResult(combinedDrops)
		if err != nil {
			return nil, err
		}

		comparisonResultsByMetadata[metadata] = res
	}

	orderedResults := helpers.OrderByDescending(helpers.Values(comparisonResultsByMetadata), func(res *models.ComparisonResult) float64 {
		meanDropsProfit, err := res.Analysis.GetMeanDropsProfit()
		if err != nil {
			panic(err)
		}
		return meanDropsProfit
	})

	return orderedResults, nil
}
