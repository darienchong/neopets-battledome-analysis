package services

import (
	"fmt"
	"log/slog"

	"github.com/darienchong/neopets-battledome-analysis/caches"
	"github.com/darienchong/neopets-battledome-analysis/constants"
	"github.com/darienchong/neopets-battledome-analysis/helpers"
	"github.com/darienchong/neopets-battledome-analysis/models"
)

type DropDataService struct{}

func NewDropDataService() *DropDataService {
	return &DropDataService{}
}

func (service *DropDataService) GetAllDrops(dataFolderPath string) ([]*models.BattledomeDrops, error) {
	parser := NewDropDataParser()
	files, err := helpers.GetFilesInFolder(dataFolderPath)
	if err != nil {
		slog.Error("Failed to get files in folder!")
		panic(err)
	}

	drops := []*models.BattledomeDrops{}
	for _, file := range files {
		drop, err := parser.Parse(constants.GetDropDataFilePath(file))
		if err != nil {
			return nil, fmt.Errorf("DropDataService.GetAllDrops(%s): %w", file, err)
		}
		drops = append(drops, drop)
	}

	return drops, nil
}

func (service *DropDataService) GetItems(drops []*models.BattledomeDrops) (map[string]*models.BattledomeItem, error) {
	itemPriceCache, err := caches.GetItemPriceCacheInstance()
	if err != nil {
		return nil, fmt.Errorf("failed to get item price cache instance: %w", err)
	}
	defer itemPriceCache.Close()

	items := map[string]*models.BattledomeItem{}
	for _, drop := range drops {
		for _, item := range drop.Items {
			_, inItems := items[item.Name]
			if !inItems {
				items[item.Name] = &models.BattledomeItem{
					Name:            item.Name,
					Quantity:        item.Quantity,
					IndividualPrice: helpers.LazyWhen(item.IndividualPrice > 0, func() float64 { return item.IndividualPrice }, func() float64 { return itemPriceCache.GetPrice(item.Name) }),
				}
			} else {
				items[item.Name].Quantity += item.Quantity
			}
		}
	}

	return items, nil
}
