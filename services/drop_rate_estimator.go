package services

import (
	"fmt"
	"log/slog"
	"math/rand/v2"
	"sort"
	"sync"

	"github.com/darienchong/neopets-battledome-analysis/constants"
	"github.com/darienchong/neopets-battledome-analysis/helpers"
	"github.com/darienchong/neopets-battledome-analysis/models"
	"github.com/dustin/go-humanize"
	"github.com/schollz/progressbar/v3"
)

type DropRateEstimator struct{}

func (estimator *DropRateEstimator) GenerateItem(weights []models.ItemWeight) string {
	rand.Shuffle(len(weights), func(i int, j int) {
		weights[i], weights[j] = weights[j], weights[i]
	})
	sort.SliceStable(weights, func(i int, j int) bool {
		return weights[i].Weight < weights[j].Weight
	})
	total := 0.0
	for _, weight := range weights {
		total += weight.Weight * 100
	}
	sample := float64(1 + rand.IntN(int(total)))
	tot := 0.0
	for _, weight := range weights {
		tot += weight.Weight * 100
		if tot >= sample {
			return weight.Name
		}
	}
	panic(fmt.Errorf("failed to generate an item - this should not happen; total was %f, sample was %f", total, sample))
}

func (estimator *DropRateEstimator) GenerateItems(weights []models.ItemWeight, count int) []string {
	progressBarMutex := &sync.Mutex{}
	itemChannel := make(chan string, count)
	wg := &sync.WaitGroup{}

	progressBar := progressbar.Default(int64(count))
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(p *progressbar.ProgressBar) {
			defer wg.Done()

			item := estimator.GenerateItem(append([]models.ItemWeight(nil), weights...))
			itemChannel <- item
			progressBarMutex.Lock()
			defer progressBarMutex.Unlock()
			p.Add(1)
		}(progressBar)
	}
	wg.Wait()
	close(itemChannel)

	items := []string{}
	for range itemChannel {
		items = append(items, <-itemChannel)
	}
	return items
}

func (estimator *DropRateEstimator) generateItems(weights []models.ItemWeight) map[string]*models.BattledomeItem {
	arena := weights[0].Arena
	items := map[string]*models.BattledomeItem{}
	parsedItems, err := new(GeneratedDropsParser).Parse(constants.GetGeneratedDropsFilePath(arena))
	if err == nil {
		items = parsedItems[arena].Items
	} else {
		itemNames := estimator.GenerateItems(weights, constants.NUMBER_OF_ITEMS_TO_GENERATE_FOR_ESTIMATING_DROP_RATES)
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
	err = new(GeneratedDropsParser).Save(dropsMap, constants.GetGeneratedDropsFilePath(arena))
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
		return new(DropRateParser).Parse(constants.GetDropRatesFilePath())
	}

	itemWeights, err := new(ItemWeightParser).Parse(constants.GetItemWeightsFilePath())
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
	err = new(DropRateParser).Save(dropRates, constants.GetDropRatesFilePath())
	if err != nil {
		return nil, err
	}
	return dropRates, nil
}
