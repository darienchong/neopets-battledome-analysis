package services

import (
	"fmt"
	"strings"

	"github.com/darienchong/neopets-battledome-analysis/caches"
	"github.com/darienchong/neopets-battledome-analysis/constants"
	"github.com/darienchong/neopets-battledome-analysis/helpers"
	"github.com/darienchong/neopets-battledome-analysis/models"
)

type EmpiricalDropsService struct{}

func NewEmpiricalDropsService() *EmpiricalDropsService {
	return &EmpiricalDropsService{}
}

func (service *EmpiricalDropsService) GetAllDrops(dataFolderPath string) ([]*models.BattledomeDrops, error) {
	parser := NewDropDataParser()
	files, err := helpers.GetFilesInFolder(dataFolderPath)
	if err != nil {
		// Could be due to inconsistent caller, try going down one level
		files, err = helpers.GetFilesInFolder(strings.Replace(constants.BATTLEDOME_DROPS_FOLDER, "../", "", 1))
		if err != nil {
			return nil, err
		}
	}

	drops := []*models.BattledomeDrops{}
	for _, file := range files {
		drop, err := parser.Parse(constants.GetDropDataFilePath(file))
		if err != nil {
			return nil, fmt.Errorf("DropDataService.GetAllDrops(%s): %w", file, err)
		}
		drops = append(drops, drop.ToBattledomeDrops())
	}

	return drops, nil
}

func (service *EmpiricalDropsService) GetDrops(arena string) ([]*models.BattledomeDrops, error) {
	allDrops, err := service.GetAllDrops(constants.BATTLEDOME_DROPS_FOLDER)
	if err != nil {
		return nil, err
	}
	return helpers.Filter(allDrops, func(drop *models.BattledomeDrops) bool {
		return drop.Metadata.Arena == arena
	}), nil
}

func (service *EmpiricalDropsService) GetItems(drops []*models.BattledomeDrops) (map[string]*models.BattledomeItem, error) {
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
