package services

import (
	"fmt"

	"github.com/darienchong/neopets-battledome-analysis/constants"
	"github.com/darienchong/neopets-battledome-analysis/helpers"
	"github.com/darienchong/neopets-battledome-analysis/models"
)

type DropRateService struct {
	DropRateParser        *DropRateParser
	GeneratedDropsService *GeneratedDropsService
}

func NewDropRateService() *DropRateService {
	return &DropRateService{
		DropRateParser:        NewDropRateParser(),
		GeneratedDropsService: NewGeneratedDropsService(),
	}
}

func (service *DropRateService) GetPredictedDropRates(arena string) (map[string]*models.ItemDropRate, error) {
	if helpers.IsFileExists(constants.GetDropRatesFilePath(arena)) {
		itemDropRates, err := service.DropRateParser.Parse(constants.GetDropRatesFilePath(arena))
		if err != nil {
			return nil, err
		}

		filteredItemDropRates := helpers.Filter(itemDropRates, func(dropRate models.ItemDropRate) bool {
			return dropRate.Arena == arena
		})

		return helpers.ToMap(
			filteredItemDropRates,
			func(dropRate models.ItemDropRate) string {
				return dropRate.ItemName
			},
			func(dropRate models.ItemDropRate) *models.ItemDropRate {
				return &dropRate
			}), nil
	}

	drop, err := service.GeneratedDropsService.GenerateDrops(arena)
	if err != nil {
		return nil, err
	}

	totalItemCount := float64(drop.GetTotalItemQuantity())
	itemDropRates := map[string]*models.ItemDropRate{}
	for _, item := range drop.Items {
		_, exists := itemDropRates[item.Name]
		if exists {
			return nil, fmt.Errorf("DropRateService.GetPredictedDropRates: tried to generate drop rates for the same item (%s) more than once: %w", item.Name, err)
		}

		itemDropRates[item.Name] = &models.ItemDropRate{
			Arena:    arena,
			ItemName: item.Name,
			DropRate: float64(item.Quantity) / totalItemCount,
		}
	}

	err = service.DropRateParser.Save(helpers.AsLiteral(helpers.Values(itemDropRates)), constants.GetDropRatesFilePath(arena))
	if err != nil {
		return nil, err
	}

	return itemDropRates, nil
}

func (service *DropRateService) CalculateDropRates(drops ...*models.BattledomeDrops) (map[string][]*models.ItemDropRate, error) {
	dropsByArena := helpers.GroupBy(drops, func(drop *models.BattledomeDrops) string {
		return drop.Metadata.Arena
	})

	dropRatesByArena := map[string][]*models.ItemDropRate{}
	for arena, arenaDrops := range dropsByArena {
		_, exists := dropRatesByArena[arena]
		if !exists {
			dropRatesByArena[arena] = []*models.ItemDropRate{}
		}

		combinedDrops := helpers.Reduce(arenaDrops, func(first *models.BattledomeDrops, second *models.BattledomeDrops) *models.BattledomeDrops {
			combined, err := first.Union(second)
			if err != nil {
				panic(err)
			}
			return combined
		})
		totalNumberOfItems := helpers.Sum(
			helpers.Map(
				helpers.FilterPointers(helpers.PointerValues(combinedDrops.Items), func(item *models.BattledomeItem) bool {
					return item.Name != "nothing"
				}),
				func(item *models.BattledomeItem) int32 {
					return item.Quantity
				},
			),
		)

		for _, item := range combinedDrops.Items {
			dropRatesByArena[arena] = append(dropRatesByArena[arena], &models.ItemDropRate{
				Arena:    arena,
				ItemName: item.Name,
				DropRate: float64(item.Quantity) / float64(totalNumberOfItems),
			})
		}
	}

	return dropRatesByArena, nil
}
