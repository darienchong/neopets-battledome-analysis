package services

import (
	"github.com/darienchong/neopets-battledome-analysis/caches"
	"github.com/darienchong/neopets-battledome-analysis/helpers"
	"github.com/darienchong/neopets-battledome-analysis/models"
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
		return
	}
	defer itemPriceCache.Close()

	realData, err = service.BattledomeItemsService.GetDropsByMetadata(metadata)
	if err != nil {
		return
	}

	generatedData, err = service.BattledomeItemsService.GenerateDropsByArena(metadata.Arena)
	if err != nil {
		return
	}

	return
}

func (service *DataComparisonService) CompareArena(arena models.Arena) (realData models.NormalisedBattledomeItems, generatedData models.NormalisedBattledomeItems, err error) {
	itemPriceCache, err := caches.GetItemPriceCacheInstance()
	if err != nil {
		return
	}
	defer itemPriceCache.Close()

	realData, err = service.BattledomeItemsService.GetDropsByArena(arena)
	if err != nil {
		return
	}

	generatedData, err = service.BattledomeItemsService.GenerateDropsByArena(arena)
	if err != nil {
		return
	}

	return
}

func (service *DataComparisonService) CompareAllChallengers() (challengerData []models.NormalisedBattledomeItems, err error) {
	data, err := service.BattledomeItemsService.GetDropsGroupedByMetadata()
	if err != nil {
		return nil, err
	}

	challengerData = helpers.OrderByDescending(helpers.Values(data), func(normalisedItems models.NormalisedBattledomeItems) float64 {
		meanDropsProfit, err := normalisedItems.GetMeanDropsProfit()
		if err != nil {
			panic(err)
		}
		return meanDropsProfit
	})

	return challengerData, nil
}
