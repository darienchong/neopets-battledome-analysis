package services

import (
	"fmt"
	"math/rand/v2"
	"sort"
	"sync"

	"github.com/darienchong/neopets-battledome-analysis/models"
	"github.com/palantir/stacktrace"
	"github.com/schollz/progressbar/v3"
)

type BattledomeItemGenerationService struct {
	ItemWeightService *BattledomeItemWeightService
}

func NewItemGenerationService() *BattledomeItemGenerationService {
	return &BattledomeItemGenerationService{
		ItemWeightService: NewBattledomeItemWeightService(),
	}
}

func (service *BattledomeItemGenerationService) generateItem(weights []models.BattledomeItemWeight) string {
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

func (service *BattledomeItemGenerationService) GenerateItems(arena models.Arena, count int) (models.NormalisedBattledomeItems, error) {
	weights, err := service.ItemWeightService.GetItemWeights(string(arena))
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get item weights for \"%s\"", arena)
	}

	progressBarMutex := &sync.Mutex{}
	itemChannel := make(chan string, count)
	wg := &sync.WaitGroup{}

	progressBar := progressbar.Default(int64(count))
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(p *progressbar.ProgressBar) {
			defer wg.Done()

			item := service.generateItem(append([]models.BattledomeItemWeight(nil), weights...))
			itemChannel <- item
			progressBarMutex.Lock()
			defer progressBarMutex.Unlock()
			p.Add(1)
		}(progressBar)
	}
	wg.Wait()
	close(itemChannel)

	itemNames := []models.ItemName{}
	for range itemChannel {
		itemNames = append(itemNames, models.ItemName(<-itemChannel))
	}

	items := models.NormalisedBattledomeItems{}
	for _, generatedItem := range itemNames {
		item, isInItems := items[generatedItem]
		if !isInItems {
			items[generatedItem] = &models.BattledomeItem{
				Name:     generatedItem,
				Quantity: 1,
			}
		} else {
			item.Quantity++
		}
	}

	return items, nil
}
