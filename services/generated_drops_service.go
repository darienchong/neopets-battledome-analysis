package services

import (
	"fmt"

	"github.com/darienchong/neopets-battledome-analysis/caches"
	"github.com/darienchong/neopets-battledome-analysis/constants"
	"github.com/darienchong/neopets-battledome-analysis/helpers"
	"github.com/darienchong/neopets-battledome-analysis/models"
)

type GeneratedDropsService struct {
	GeneratedDropsParser *GeneratedDropsParser
	ItemWeightService    *ItemWeightService
	ItemGenerator        *ItemGenerator
}

func NewGeneratedDropsService() *GeneratedDropsService {
	return &GeneratedDropsService{
		GeneratedDropsParser: NewGeneratedDropsParser(),
		ItemWeightService:    NewItemWeightService(),
		ItemGenerator:        NewItemGenerator(),
	}
}

func (service *GeneratedDropsService) GenerateDrops(arena string) (*models.BattledomeDrops, error) {
	if helpers.IsFileExists(constants.GetGeneratedDropsFilePath(arena)) {
		parsedDrops, err := service.GeneratedDropsParser.Parse(constants.GetGeneratedDropsFilePath(arena))
		if err != nil {
			return nil, err
		}

		if len(parsedDrops) > 1 {
			panic(fmt.Errorf("encountered mixed arena data in generated drops; there should only be a single arena's data per file"))
		}

		return parsedDrops[arena], nil
	} else {
		itemPriceCache, err := caches.GetItemPriceCacheInstance()
		if err != nil {
			return nil, err
		}
		defer itemPriceCache.Close()

		items, err := service.ItemGenerator.GenerateItems(arena, constants.NUMBER_OF_ITEMS_TO_GENERATE_FOR_ESTIMATING_PROFIT_STATISTICS)
		if err != nil {
			return nil, err
		}

		generatedDrops := models.NewBattledomeDrops()
		generatedDrops.Metadata = *models.GeneratedMetadata(arena)
		generatedDrops.Items = items

		err = service.GeneratedDropsParser.Save(map[string]*models.BattledomeDrops{arena: generatedDrops}, constants.GetGeneratedDropsFilePath(arena))
		if err != nil {
			return nil, err
		}

		return generatedDrops, nil
	}
}
