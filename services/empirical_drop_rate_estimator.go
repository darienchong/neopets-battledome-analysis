package services

import (
	"github.com/darienchong/neopets-battledome-analysis/caches"
	"github.com/darienchong/neopets-battledome-analysis/models"
)

type EmpiricalDropRateEstimator struct{}

func (analyser *EmpiricalDropRateEstimator) Analyse(drops *models.BattledomeDrops) *models.DropDataAnalysisResult {
	itemPriceCache := caches.GetItemPriceCacheInstance()
	defer itemPriceCache.Close()
	drops.AddPrices(itemPriceCache)
	return models.NewAnalysisResultFromDrops(drops)
}
