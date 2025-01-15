package services

import (
	"github.com/darienchong/neopets-battledome-analysis/caches"
	"github.com/darienchong/neopets-battledome-analysis/helpers"
	"github.com/darienchong/neopets-battledome-analysis/models"
	"github.com/palantir/stacktrace"
)

type DataComparisonService struct {
	BattledomeItemsService *BattledomeItemsService
}

func NewDataComparisonService() *DataComparisonService {
	return &DataComparisonService{
		BattledomeItemsService: NewBattledomeItemsService(),
	}
}

func (service *DataComparisonService) CompareByMetadata(metadata models.BattledomeItemMetadata) (realData models.NormalisedBattledomeItems, generatedData models.NormalisedBattledomeItems, err error) {
	itemPriceCache, err := caches.GetItemPriceCacheInstance()
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "failed to get item price cache instance")
	}
	defer itemPriceCache.Close()

	realData, err = service.BattledomeItemsService.GetDropsByMetadata(metadata)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "failed to get drops by metadata for \"%s\"", metadata.String())
	}

	generatedData, err = service.BattledomeItemsService.GenerateDropsByArena(metadata.Arena)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "failed to generate drops by arena for \"%s\"", metadata.Arena)
	}

	return realData, generatedData, nil
}

func (service *DataComparisonService) CompareArena(arena models.Arena) (realData models.NormalisedBattledomeItems, generatedData models.NormalisedBattledomeItems, err error) {
	itemPriceCache, err := caches.GetItemPriceCacheInstance()
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "failed to get item price cache instance")
	}
	defer itemPriceCache.Close()

	realData, err = service.BattledomeItemsService.GetDropsByArena(arena)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "failed to get drops by arena for \"%s\"", arena)
	}

	generatedData, err = service.BattledomeItemsService.GenerateDropsByArena(arena)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "failed to generate drops by arena for \"%s\"", arena)
	}

	return
}

func (service *DataComparisonService) CompareAllChallengers() (challengerData []models.NormalisedBattledomeItems, err error) {
	data, err := service.BattledomeItemsService.GetDropsGroupedByMetadata()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get drops grouped by metadata")
	}

	challengerData = helpers.OrderByDescending(helpers.Values(data), func(normalisedItems models.NormalisedBattledomeItems) float64 {
		meanDropsProfit, err := normalisedItems.GetMeanDropsProfit()
		if err != nil {
			return 0.0
		}
		return meanDropsProfit
	})

	return challengerData, nil
}
