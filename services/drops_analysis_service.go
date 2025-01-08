package services

import (
	"github.com/darienchong/neopets-battledome-analysis/caches"
	"github.com/darienchong/neopets-battledome-analysis/models"
)

type DropsAnalysisService struct{}

func NewDropsAnalysisService() *DropsAnalysisService {
	return &DropsAnalysisService{}
}

func (analyser *DropsAnalysisService) Analyse(drops *models.BattledomeDrops) *models.DropsAnalysis {
	itemPriceCache, err := caches.GetItemPriceCacheInstance()
	if err != nil {
		panic(err)
	}
	defer itemPriceCache.Close()
	drops.AddPrices(itemPriceCache)
	return models.NewAnalysisResultFromDrops(drops)
}
