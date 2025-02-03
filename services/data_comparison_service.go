package services

import (
	"github.com/darienchong/neopets-battledome-analysis/caches"
	"github.com/darienchong/neopets-battledome-analysis/helpers"
	"github.com/darienchong/neopets-battledome-analysis/models"
	"github.com/palantir/stacktrace"
)

type BattledomeItems interface {
	DropsByArena(arena models.Arena) (models.NormalisedBattledomeItems, error)
	GeneratedDropsByArena(arena models.Arena) (models.NormalisedBattledomeItems, error)
	DropsByMetadata(metadata models.BattledomeItemMetadata) (models.NormalisedBattledomeItems, error)
	DropsGroupedByMetadata() (map[models.BattledomeItemMetadata]models.NormalisedBattledomeItems, error)
}

type DataComparisonService struct {
	BattledomeItems
}

func NewDataComparisonService(battledomeItemsService BattledomeItems) *DataComparisonService {
	return &DataComparisonService{
		BattledomeItems: battledomeItemsService,
	}
}

func (s *DataComparisonService) CompareByMetadata(metadata models.BattledomeItemMetadata) (realData models.NormalisedBattledomeItems, generatedData models.NormalisedBattledomeItems, err error) {
	realData, err = s.BattledomeItems.DropsByMetadata(metadata)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "failed to get drops by metadata for %q", metadata.String())
	}

	generatedData, err = s.BattledomeItems.GeneratedDropsByArena(metadata.Arena)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "failed to generate drops by arena for %q", metadata.Arena)
	}

	return realData, generatedData, nil
}

func (s *DataComparisonService) CompareArena(arena models.Arena) (realData models.NormalisedBattledomeItems, generatedData models.NormalisedBattledomeItems, err error) {
	realData, err = s.BattledomeItems.DropsByArena(arena)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "failed to get drops by arena for %q", arena)
	}

	generatedData, err = s.BattledomeItems.GeneratedDropsByArena(arena)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "failed to generate drops by arena for %q", arena)
	}

	return
}

func (s *DataComparisonService) CompareAllChallengers(itemPriceCache caches.ItemPriceCache) (challengerData []models.NormalisedBattledomeItems, err error) {
	data, err := s.BattledomeItems.DropsGroupedByMetadata()
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
