package services

import (
	"fmt"
	"log/slog"

	"github.com/darienchong/neopets-battledome-analysis/constants"
	"github.com/darienchong/neopets-battledome-analysis/helpers"
	"github.com/darienchong/neopets-battledome-analysis/models"
	"github.com/dustin/go-humanize"
)

type DropRateEstimator struct {
	itemGenerator *ItemGenerator
}

func NewDropRateEstimator() *DropRateEstimator {
	return &DropRateEstimator{
		itemGenerator: &ItemGenerator{},
	}
}

func (estimator *DropRateEstimator) generateItems(weights []models.ItemWeight) map[string]*models.BattledomeItem {
	arena := weights[0].Arena
	items := map[string]*models.BattledomeItem{}
	parsedItems, err := NewGeneratedDropsParser().Parse(constants.GetGeneratedDropsFilePath(arena))
	if err == nil {
		items = parsedItems[arena].Items
	} else {
		itemNames := estimator.itemGenerator.GenerateItems(weights, constants.NUMBER_OF_ITEMS_TO_GENERATE_FOR_ESTIMATING_DROP_RATES)
		for _, item := range itemNames {
			_, isEntryExists := items[item]
			if !isEntryExists {
				items[item] = &models.BattledomeItem{
					Name:     item,
					Quantity: 1,
				}
			} else {
				items[item].Quantity += 1
			}
		}
	}

	drops := models.NewBattledomeDrops()
	drops.Metadata = models.DropsMetadata{
		Source:     "(generated)",
		Arena:      arena,
		Challenger: "(generated)",
		Difficulty: "(generated)",
	}
	drops.Items = items
	dropsMap := map[string]*models.BattledomeDrops{}
	dropsMap[arena] = drops
	err = NewGeneratedDropsParser().Save(dropsMap, constants.GetGeneratedDropsFilePath(arena))
	if err != nil {
		panic(err)
	}

	return items
}

func (estimator *DropRateEstimator) generateItemDropRates(weights []models.ItemWeight) []models.ItemDropRate {
	arena := weights[0].Arena
	slog.Info(fmt.Sprintf("Generating drop rates for %s @ %s items", arena, humanize.Comma(constants.NUMBER_OF_ITEMS_TO_GENERATE_FOR_ESTIMATING_DROP_RATES)))
	items := estimator.generateItems(weights)
	numberOfItemsGenerated := helpers.Sum(helpers.Map(helpers.ToSlice(items), func(tuple helpers.Tuple) int32 {
		if tuple.Elements[0].(string) == "nothing" {
			return 0
		} else {
			return tuple.Elements[1].(*models.BattledomeItem).Quantity
		}
	}))
	dropRates := []models.ItemDropRate{}
	for _, v := range items {
		dropRates = append(dropRates, models.ItemDropRate{
			Arena:    arena,
			ItemName: v.Name,
			DropRate: float64(v.Quantity) / float64(numberOfItemsGenerated),
		})
	}

	return dropRates
}

func (estimator *DropRateEstimator) Estimate() ([]models.ItemDropRate, error) {
	if helpers.IsFileExists(constants.GetDropRatesFilePath()) {
		return NewDropRateParser().Parse(constants.GetDropRatesFilePath())
	}

	itemWeights, err := NewItemWeightParser().Parse(constants.GetItemWeightsFilePath())
	if err != nil {
		return nil, err
	}
	arenas := helpers.Distinct(helpers.Map(itemWeights, func(weight models.ItemWeight) string {
		return weight.Arena
	}))
	dropRates := []models.ItemDropRate{}
	for _, arena := range arenas {
		currWeights := helpers.Filter(itemWeights, func(weight models.ItemWeight) bool {
			return weight.Arena == arena
		})
		currDropRates := estimator.generateItemDropRates(currWeights)
		dropRates = append(dropRates, currDropRates...)
	}

	slog.Info(fmt.Sprintf("Saving generated drop rate data to \"%s\"", constants.GetDropRatesFilePath()))
	err = NewDropRateParser().Save(dropRates, constants.GetDropRatesFilePath())
	if err != nil {
		return nil, err
	}
	return dropRates, nil
}
