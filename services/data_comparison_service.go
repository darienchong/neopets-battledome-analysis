package services

import (
	"github.com/darienchong/neopets-battledome-analysis/caches"
	"github.com/darienchong/neopets-battledome-analysis/helpers"
	"github.com/darienchong/neopets-battledome-analysis/models"
)

type DataComparisonService struct {
	GeneratedDropsService *GeneratedDropsService
	EmpiricalDropsService *EmpiricalDropsService
	DropsAnalysisService  *DropsAnalysisService
	DropRateService       *DropRateService
}

func NewDataComparisonService() *DataComparisonService {
	return &DataComparisonService{
		GeneratedDropsService: NewGeneratedDropsService(),
		EmpiricalDropsService: NewEmpiricalDropsService(),
		DropsAnalysisService:  NewDropsAnalysisService(),
		DropRateService:       NewDropRateService(),
	}
}

func (service *DataComparisonService) ToComparisonResult(drop *models.BattledomeDrops) (*models.ComparisonResult, error) {
	itemPriceCache, err := caches.GetItemPriceCacheInstance()
	if err != nil {
		return nil, err
	}
	defer itemPriceCache.Close()

	analysis := service.DropsAnalysisService.Analyse(drop)
	dropRates, err := service.DropRateService.CalculateDropRates(&models.BattledomeDrops{
		Metadata: models.DropsMetadataWithSource{
			Source:        "(multiple sources)",
			DropsMetadata: analysis.Metadata,
		},
		Items: helpers.ToPointerMap(
			analysis.GetItemsOrderedByProfit(),
			func(item *models.BattledomeItem) string {
				return item.Name
			},
			func(item *models.BattledomeItem) *models.BattledomeItem {
				return item
			}),
	})
	if err != nil {
		return nil, err
	}

	profits := map[string]*models.ItemProfit{}
	for _, itemDropRate := range dropRates[analysis.Metadata.Arena] {
		profits[itemDropRate.ItemName] = &models.ItemProfit{
			ItemDropRate:    *itemDropRate,
			IndividualPrice: itemPriceCache.GetPrice(itemDropRate.ItemName),
		}
	}

	return &models.ComparisonResult{
		Analysis: analysis,
		Profit:   profits,
	}, nil
}

func (service *DataComparisonService) CompareByMetadata(metadata models.DropsMetadata) (*models.ComparisonResult, *models.ComparisonResult, error) {
	itemPriceCache, err := caches.GetItemPriceCacheInstance()
	if err != nil {
		return nil, nil, err
	}
	defer itemPriceCache.Close()

	realDrops, err := service.EmpiricalDropsService.GetDropsByMetadata(metadata)
	if err != nil {
		return nil, nil, err
	}
	for _, drop := range realDrops {
		for _, item := range drop.Items {
			item.IndividualPrice = itemPriceCache.GetPrice(item.Name)
		}
	}

	var combinedRealDrops *models.BattledomeDrops
	if len(realDrops) > 0 {
		combinedRealDrops = helpers.Reduce(realDrops, func(first *models.BattledomeDrops, second *models.BattledomeDrops) *models.BattledomeDrops {
			combined, err := first.Union(second)
			if err != nil {
				panic(err)
			}
			return combined
		})
		combinedRealDrops.Metadata = realDrops[0].Metadata
		combinedRealDrops.Metadata.Source = "(multiple sources)"
	} else {
		combinedRealDrops = models.NewBattledomeDrops()
		combinedRealDrops.Metadata = models.DropsMetadataWithSource{
			Source:        "(none)",
			DropsMetadata: metadata,
		}
	}

	combinedGeneratedDrops, err := service.GeneratedDropsService.GenerateDrops(metadata.Arena)
	if err != nil {
		return nil, nil, err
	}

	realComparisonResult, err := service.ToComparisonResult(combinedRealDrops)
	if err != nil {
		return nil, nil, err
	}
	generatedComparisonResult, err := service.ToComparisonResult(combinedGeneratedDrops)
	if err != nil {
		return nil, nil, err
	}

	return realComparisonResult, generatedComparisonResult, nil
}

func (service *DataComparisonService) CompareArena(arena string) (*models.ComparisonResult, *models.ComparisonResult, error) {
	itemPriceCache, err := caches.GetItemPriceCacheInstance()
	if err != nil {
		return nil, nil, err
	}
	defer itemPriceCache.Close()

	realDrops, err := service.EmpiricalDropsService.GetDropsByArena(arena)
	if err != nil {
		return nil, nil, err
	}
	for _, drop := range realDrops {
		for _, item := range drop.Items {
			item.IndividualPrice = itemPriceCache.GetPrice(item.Name)
		}
	}

	var combinedRealDrops *models.BattledomeDrops
	if len(realDrops) > 0 {
		combinedRealDrops = helpers.Reduce(realDrops, func(first *models.BattledomeDrops, second *models.BattledomeDrops) *models.BattledomeDrops {
			combined, err := first.Union(second)
			if err != nil {
				panic(err)
			}
			return combined
		})
	} else {
		combinedRealDrops = models.NewBattledomeDrops()
		combinedRealDrops.Metadata = models.DropsMetadataWithSource{
			Source: "(none)",
			DropsMetadata: models.DropsMetadata{
				Arena:      arena,
				Challenger: "(none)",
				Difficulty: "(none)",
			},
		}
	}

	combinedGeneratedDrops, err := service.GeneratedDropsService.GenerateDrops(arena)
	if err != nil {
		return nil, nil, err
	}

	realComparisonResult, err := service.ToComparisonResult(combinedRealDrops)
	if err != nil {
		return nil, nil, err
	}
	generatedComparisonResult, err := service.ToComparisonResult(combinedGeneratedDrops)
	if err != nil {
		return nil, nil, err
	}

	return realComparisonResult, generatedComparisonResult, nil
}

func (service *DataComparisonService) CompareAllChallengers() ([]*models.ComparisonResult, error) {
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
		res, err := service.ToComparisonResult(combinedDrops)
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
