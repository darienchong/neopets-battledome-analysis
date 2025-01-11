package services

import (
	"fmt"
	"strings"

	"github.com/darienchong/neopets-battledome-analysis/caches"
	"github.com/darienchong/neopets-battledome-analysis/constants"
	"github.com/darienchong/neopets-battledome-analysis/helpers"
	"github.com/darienchong/neopets-battledome-analysis/models"
	"github.com/darienchong/neopets-battledome-analysis/parsers"
)

type BattledomeItemsService struct {
	ItemGenerationService         *ItemGenerationService
	GeneratedBattledomeItemParser *parsers.GeneratedBattledomeItemParser
	BattledomeItemDropDataParser  *parsers.BattledomeItemDropDataParser
}

func NewBattledomeItemsService() *BattledomeItemsService {
	return &BattledomeItemsService{
		ItemGenerationService:         NewItemGenerationService(),
		GeneratedBattledomeItemParser: parsers.NewGeneratedBattledomeItemParser(),
		BattledomeItemDropDataParser:  parsers.NewBattledomeItemDropDataParser(),
	}
}

func (service *BattledomeItemsService) GetAllDrops() (map[models.Arena]models.NormalisedBattledomeItems, error) {
	files, err := helpers.GetFilesInFolder(constants.BATTLEDOME_DROPS_FOLDER)
	if err != nil {
		// Could be due to inconsistent caller, try going down one level
		files, err = helpers.GetFilesInFolder(strings.Replace(constants.BATTLEDOME_DROPS_FOLDER, "../", "", 1))
		if err != nil {
			return nil, err
		}
	}

	itemsByArena := map[models.Arena]models.BattledomeItems{}
	for _, file := range files {
		dto, err := service.BattledomeItemDropDataParser.Parse(constants.GetDropDataFilePath(file))
		if err != nil {
			return nil, fmt.Errorf("BattledomeItemsService.GetAllDrops(%s): %w", file, err)
		}
		_, exists := itemsByArena[models.Arena(dto.Metadata.Arena)]
		if !exists {
			itemsByArena[dto.Metadata.Arena] = models.BattledomeItems{}
		}

		itemsByArena[dto.Metadata.Arena] = append(itemsByArena[dto.Metadata.Arena], dto.Items...)
	}

	normalisedItemsByArena := map[models.Arena]models.NormalisedBattledomeItems{}
	for arena, items := range itemsByArena {
		normalisedItems, err := items.Normalise()
		if err != nil {
			return nil, err
		}
		normalisedItemsByArena[arena] = normalisedItems
	}

	return normalisedItemsByArena, nil
}

func (service *BattledomeItemsService) GetDropsByMetadata(metadata models.BattledomeItemMetadata) (models.NormalisedBattledomeItems, error) {
	allDrops, err := service.GetAllDrops()
	if err != nil {
		return nil, err
	}

	arenaDrops := allDrops[metadata.Arena]
	matchingArenaDrops := helpers.Filter(helpers.Values(arenaDrops), func(item *models.BattledomeItem) bool {
		return item.Metadata == metadata
	})

	return models.BattledomeItems(matchingArenaDrops).Normalise()
}

func (service *BattledomeItemsService) GetDropsGroupedByMetadata() (map[models.BattledomeItemMetadata]models.NormalisedBattledomeItems, error) {
	dropsByArena, err := service.GetAllDrops()
	if err != nil {
		return nil, err
	}
	allDrops := helpers.FlatMap(helpers.Values(dropsByArena), func(items models.NormalisedBattledomeItems) []*models.BattledomeItem {
		return helpers.Values(items)
	})
	allDropsGroupedByMetadata := helpers.GroupBy(allDrops, func(item *models.BattledomeItem) models.BattledomeItemMetadata {
		return item.Metadata
	})
	normalisedItemsGroupedByMetadata := map[models.BattledomeItemMetadata]models.NormalisedBattledomeItems{}
	for k, v := range allDropsGroupedByMetadata {
		normalisedItems, err := models.BattledomeItems(v).Normalise()
		if err != nil {
			return nil, err
		}
		normalisedItemsGroupedByMetadata[k] = normalisedItems
	}
	return normalisedItemsGroupedByMetadata, nil
}

func (service *BattledomeItemsService) GetDropsByArena(arena models.Arena) (models.NormalisedBattledomeItems, error) {
	allDrops, err := service.GetAllDrops()
	if err != nil {
		return nil, err
	}
	return allDrops[arena], nil
}

func (service *BattledomeItemsService) GenerateDropsByArena(arena models.Arena) (models.NormalisedBattledomeItems, error) {
	if helpers.IsFileExists(constants.GetGeneratedDropsFilePath(string(arena))) {
		parsedDrops, err := service.GeneratedBattledomeItemParser.Parse(constants.GetGeneratedDropsFilePath(string(arena)))
		if err != nil {
			return nil, err
		}

		return parsedDrops, nil
	} else {
		itemPriceCache, err := caches.GetItemPriceCacheInstance()
		if err != nil {
			return nil, err
		}
		defer itemPriceCache.Close()

		items, err := service.ItemGenerationService.GenerateItems(arena, constants.NUMBER_OF_ITEMS_TO_GENERATE_FOR_ESTIMATING_PROFIT_STATISTICS)
		if err != nil {
			return nil, err
		}

		err = service.GeneratedBattledomeItemParser.Save(items, constants.GetGeneratedDropsFilePath(string(arena)))
		if err != nil {
			return nil, err
		}

		return items, nil
	}
}
