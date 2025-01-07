package services

import (
	"fmt"
	"math/rand/v2"
	"sort"
	"sync"

	"github.com/darienchong/neopets-battledome-analysis/models"
	"github.com/schollz/progressbar/v3"
)

type ItemGenerator struct{}

func NewItemGenerator() *ItemGenerator {
	return &ItemGenerator{}
}

func (generator *ItemGenerator) GenerateItem(weights []models.ItemWeight) string {
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

func (generator *ItemGenerator) GenerateItems(weights []models.ItemWeight, count int) []string {
	progressBarMutex := &sync.Mutex{}
	itemChannel := make(chan string, count)
	wg := &sync.WaitGroup{}

	progressBar := progressbar.Default(int64(count))
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(p *progressbar.ProgressBar) {
			defer wg.Done()

			item := generator.GenerateItem(append([]models.ItemWeight(nil), weights...))
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
