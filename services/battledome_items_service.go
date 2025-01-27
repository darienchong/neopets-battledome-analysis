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
	ItemGenerationService         *BattledomeItemGenerationService
	GeneratedBattledomeItemParser *parsers.GeneratedBattledomeItemParser
	BattledomeItemDropDataParser  *parsers.BattledomeItemDropDataParser
}

func NewBattledomeItemsService(
	itemGenerationService *BattledomeItemGenerationService,
	generatedBattledomeItemParser *parsers.GeneratedBattledomeItemParser,
	battledomeItemDropDataParser *parsers.BattledomeItemDropDataParser,
) *BattledomeItemsService {
	return &BattledomeItemsService{
		ItemGenerationService:         itemGenerationService,
		GeneratedBattledomeItemParser: generatedBattledomeItemParser,
		BattledomeItemDropDataParser:  battledomeItemDropDataParser,
	}
}

func (s *BattledomeItemsService) AllDrops() (map[models.Arena]models.BattledomeItems, error) {
	files, err := helpers.FilesInFolder(constants.BattledomeDropsFolder)
	if err != nil {
		// Could be due to inconsistent caller, try going down one level
		newPath := strings.Replace(constants.BattledomeDropsFolder, "../", "", 1)
		files, err = helpers.FilesInFolder(newPath)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to get files in %q", newPath)
		}
	}

	itemsByArena := map[models.Arena]models.BattledomeItems{}
	for _, file := range files {
		dto, err := s.BattledomeItemDropDataParser.Parse(constants.DropDataFilePath(file))
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to parse %q as battledome drop data", file)
		}
		_, exists := itemsByArena[models.Arena(dto.Metadata.Arena)]
		if !exists {
			itemsByArena[dto.Metadata.Arena] = models.BattledomeItems{}
		}

		itemsByArena[dto.Metadata.Arena] = append(itemsByArena[dto.Metadata.Arena], dto.Items...)
	}

	return itemsByArena, nil
}

func (s *BattledomeItemsService) DropsByMetadata(metadata models.BattledomeItemMetadata) (models.NormalisedBattledomeItems, error) {
	allDrops, err := s.AllDrops()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get all drops")
	}

	arenaItems := allDrops[metadata.Arena]
	matchingArenaDrops := helpers.Filter(arenaItems, func(item *models.BattledomeItem) bool {
		return item.Metadata == metadata
	})

	return models.BattledomeItems(matchingArenaDrops).Normalise()
}

func (s *BattledomeItemsService) DropsGroupedByMetadata() (map[models.BattledomeItemMetadata]models.NormalisedBattledomeItems, error) {
	itemsByArena, err := s.AllDrops()
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

func (s *BattledomeItemsService) DropsByArena(arena models.Arena) (models.NormalisedBattledomeItems, error) {
	allDrops, err := s.AllDrops()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get all drops")
	}
	normalisedDrops, err := allDrops[arena].Normalise()
	if err != nil {
		return nil, helpers.PropagateWithSerialisedValue(err, "failed to normalise %s", "failed to normalise items; additional encountered an error while trying to serialise the item for logging: %s", allDrops[arena])
	}
	return normalisedDrops, nil
}

func (s *BattledomeItemsService) GeneratedDropsByArena(arena models.Arena) (models.NormalisedBattledomeItems, error) {
	if helpers.IsFileExists(constants.GeneratedDropsFilePath(string(arena))) {
		parsedDrops, err := s.GeneratedBattledomeItemParser.Parse(constants.GeneratedDropsFilePath(string(arena)))
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to parse %q as battledome drops", arena)
		}

		return parsedDrops, nil
	} else {
		items, err := s.ItemGenerationService.GenerateItems(arena, constants.NumberOfItemsToGenerate)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to generate items for %q", arena)
		}

		err = s.GeneratedBattledomeItemParser.Save(items, constants.GeneratedDropsFilePath(string(arena)))
		if err != nil {
			return nil, stacktrace.Propagate(err, "falled to save generated drops to %q", constants.GeneratedDropsFilePath(string(arena)))
		}

		return items, nil
	}
}
