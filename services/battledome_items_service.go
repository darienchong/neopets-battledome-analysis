package services

import (
	"strings"

	"github.com/darienchong/neopets-battledome-analysis/constants"
	"github.com/darienchong/neopets-battledome-analysis/helpers"
	"github.com/darienchong/neopets-battledome-analysis/models"
	"github.com/darienchong/neopets-battledome-analysis/parsers"
	"github.com/palantir/stacktrace"
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

func (service *BattledomeItemsService) GetAllDrops() (map[models.Arena]models.BattledomeItems, error) {
	files, err := helpers.GetFilesInFolder(constants.BATTLEDOME_DROPS_FOLDER)
	if err != nil {
		// Could be due to inconsistent caller, try going down one level
		newPath := strings.Replace(constants.BATTLEDOME_DROPS_FOLDER, "../", "", 1)
		files, err = helpers.GetFilesInFolder(newPath)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to get files in \"%s\"", newPath)
		}
	}

	itemsByArena := map[models.Arena]models.BattledomeItems{}
	for _, file := range files {
		dto, err := service.BattledomeItemDropDataParser.Parse(constants.GetDropDataFilePath(file))
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to parse \"%s\" as battledome drop data", file)
		}
		_, exists := itemsByArena[models.Arena(dto.Metadata.Arena)]
		if !exists {
			itemsByArena[dto.Metadata.Arena] = models.BattledomeItems{}
		}

		itemsByArena[dto.Metadata.Arena] = append(itemsByArena[dto.Metadata.Arena], dto.Items...)
	}

	return itemsByArena, nil
}

func (service *BattledomeItemsService) GetDropsByMetadata(metadata models.BattledomeItemMetadata) (models.NormalisedBattledomeItems, error) {
	allDrops, err := service.GetAllDrops()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get all drops")
	}

	arenaItems := allDrops[metadata.Arena]
	matchingArenaDrops := helpers.Filter(arenaItems, func(item *models.BattledomeItem) bool {
		return item.Metadata == metadata
	})

	return models.BattledomeItems(matchingArenaDrops).Normalise()
}

func (service *BattledomeItemsService) GetDropsGroupedByMetadata() (map[models.BattledomeItemMetadata]models.NormalisedBattledomeItems, error) {
	itemsByArena, err := service.GetAllDrops()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get all drops")
	}
	allDrops := helpers.FlatMap(helpers.Values(itemsByArena), func(items models.BattledomeItems) []*models.BattledomeItem {
		return items
	})
	allDropsGroupedByMetadata := helpers.GroupBy(allDrops, func(item *models.BattledomeItem) models.BattledomeItemMetadata {
		return item.Metadata
	})
	normalisedItemsGroupedByMetadata := map[models.BattledomeItemMetadata]models.NormalisedBattledomeItems{}
	for k, v := range allDropsGroupedByMetadata {
		normalisedItems, err := models.BattledomeItems(v).Normalise()
		if err != nil {
			return nil, helpers.PropagateWithSerialisedValue(err, "failed to normalise %s", "failed to normalise items; additionally encountered an error while trying to serialise the item for logging: %s", v)
		}
		normalisedItemsGroupedByMetadata[k] = normalisedItems
	}
	return normalisedItemsGroupedByMetadata, nil
}

func (service *BattledomeItemsService) GetDropsByArena(arena models.Arena) (models.NormalisedBattledomeItems, error) {
	allDrops, err := service.GetAllDrops()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get all drops")
	}
	normalisedDrops, err := allDrops[arena].Normalise()
	if err != nil {
		return nil, helpers.PropagateWithSerialisedValue(err, "failed to normalise %s", "failed to normalise items; additional encountered an error while trying to serialise the item for logging: %s", allDrops[arena])
	}
	return normalisedDrops, nil
}

func (service *BattledomeItemsService) GenerateDropsByArena(arena models.Arena) (models.NormalisedBattledomeItems, error) {
	if helpers.IsFileExists(constants.GetGeneratedDropsFilePath(string(arena))) {
		parsedDrops, err := service.GeneratedBattledomeItemParser.Parse(constants.GetGeneratedDropsFilePath(string(arena)))
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to parse \"%s\" as battledome drops", arena)
		}

		return parsedDrops, nil
	} else {
		items, err := service.ItemGenerationService.GenerateItems(arena, constants.NUMBER_OF_ITEMS_TO_GENERATE_FOR_ESTIMATING_PROFIT_STATISTICS)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to generate items for \"%s\"", arena)
		}

		err = service.GeneratedBattledomeItemParser.Save(items, constants.GetGeneratedDropsFilePath(string(arena)))
		if err != nil {
			return nil, stacktrace.Propagate(err, "falled to save generated drops to \"%s\"", constants.GetGeneratedDropsFilePath(string(arena)))
		}

		return items, nil
	}
}
