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

func NewDataComparisonService(battledomeItemsService *BattledomeItemsService) *DataComparisonService {
	return &DataComparisonService{
		BattledomeItemsService: battledomeItemsService,
	}
}

func (s *DataComparisonService) CompareByMetadata(metadata models.BattledomeItemMetadata) (realData models.NormalisedBattledomeItems, generatedData models.NormalisedBattledomeItems, err error) {
	realData, err = s.BattledomeItemsService.DropsByMetadata(metadata)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "failed to get drops by metadata for %q", metadata.String())
	}

	generatedData, err = s.BattledomeItemsService.GeneratedDropsByArena(metadata.Arena)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "failed to generate drops by arena for %q", metadata.Arena)
	}

	return realData, generatedData, nil
}

func (s *DataComparisonService) CompareArena(arena models.Arena) (realData models.NormalisedBattledomeItems, generatedData models.NormalisedBattledomeItems, err error) {
	realData, err = s.BattledomeItemsService.DropsByArena(arena)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "failed to get drops by arena for %q", arena)
	}

	generatedData, err = s.BattledomeItemsService.GeneratedDropsByArena(arena)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "failed to generate drops by arena for %q", arena)
	}

	return
}

func (s *DataComparisonService) CompareAllChallengers(itemPriceCache caches.ItemPriceCache) (challengerData []models.NormalisedBattledomeItems, err error) {
	data, err := s.BattledomeItemsService.DropsGroupedByMetadata()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get drops grouped by metadata")
	}

	challengerData = helpers.OrderByDescending(helpers.Values(data), func(normalisedItems models.NormalisedBattledomeItems) float64 {
		meanDropsProfit, err := normalisedItems.MeanDropsProfit(itemPriceCache)
		if err != nil {
			return 0.0
		}
		return meanDropsProfit
	})

	return challengerData, nil
}
