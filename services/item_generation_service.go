package services

import (
	"fmt"
	"math/rand/v2"
	"sort"
	"sync"

	"github.com/darienchong/neopets-battledome-analysis/caches"
	"github.com/darienchong/neopets-battledome-analysis/models"
	"github.com/schollz/progressbar/v3"
)

type ItemGenerationService struct {
	ItemWeightService *ItemWeightService
}

func NewItemGenerationService() *ItemGenerationService {
	return &ItemGenerationService{
		ItemWeightService: NewItemWeightService(),
	}
}

func (generator *ItemGenerationService) generateItem(weights []models.ItemWeight) string {
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

func (generator *ItemGenerationService) GenerateItems(arena string, count int) (map[string]*models.BattledomeItem, error) {
	weights, err := generator.ItemWeightService.GetItemWeights(arena)
	if err != nil {
		return nil, err
	}

	itemPriceCache, err := caches.GetItemPriceCacheInstance()
	if err != nil {
		return nil, err
	}
	defer itemPriceCache.Close()

	progressBarMutex := &sync.Mutex{}
	itemChannel := make(chan string, count)
	wg := &sync.WaitGroup{}

	progressBar := progressbar.Default(int64(count))
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(p *progressbar.ProgressBar) {
			defer wg.Done()

			item := generator.generateItem(append([]models.ItemWeight(nil), weights...))
			itemChannel <- item
			progressBarMutex.Lock()
			defer progressBarMutex.Unlock()
			p.Add(1)
		}(progressBar)
	}
	wg.Wait()
	close(itemChannel)

	itemNames := []string{}
	for range itemChannel {
		itemNames = append(itemNames, <-itemChannel)
	}

	items := map[string]*models.BattledomeItem{}
	for _, generatedItem := range itemNames {
		item, isInItems := items[generatedItem]
		if !isInItems {
			items[generatedItem] = &models.BattledomeItem{
				Name:            generatedItem,
				Quantity:        1,
				IndividualPrice: itemPriceCache.GetPrice(generatedItem),
			}
		} else {
			item.Quantity++
		}
	}

	return items, nil
}
