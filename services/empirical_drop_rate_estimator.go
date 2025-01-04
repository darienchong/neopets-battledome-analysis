package services

import (
	"github.com/darienchong/neopetsbattledomeanalysis/caches"
	"github.com/darienchong/neopetsbattledomeanalysis/models"
)

type EmpiricalDropRateEstimator struct{}

func (analyser *EmpiricalDropRateEstimator) Analyse(drops *models.BattledomeDrops) *models.DropDataAnalysisResult {
	itemPriceCache := caches.GetItemPriceCacheInstance()
	defer itemPriceCache.Close()
	drops.AddPrices(itemPriceCache)
	return models.NewAnalysisResultFromDrops(drops)
}
