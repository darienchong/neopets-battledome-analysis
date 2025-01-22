package infra

import (
	"github.com/darienchong/neopets-battledome-analysis/caches"
	"github.com/darienchong/neopets-battledome-analysis/loggers"
	"github.com/darienchong/neopets-battledome-analysis/parsers"
	"github.com/darienchong/neopets-battledome-analysis/services"
	"github.com/darienchong/neopets-battledome-analysis/viewers"
)

type ServiceContainer struct {
	ItemPriceCache *caches.ItemPriceCache

	BattledomeItemsLogger *loggers.BattledomeItemsLogger
	DataComparisonLogger  *loggers.DataComparisonLogger

	BattledomeItemDropDataParser  *parsers.BattledomeItemDropDataParser
	BattledomeItemWeightParser    *parsers.BattledomeItemWeightParser
	GeneratedBattledomeItemParser *parsers.GeneratedBattledomeItemParser

	BattledomeItemGenerationService *services.BattledomeItemGenerationService
	BattledomeItemWeightService     *services.BattledomeItemWeightService
	BattledomeItemsService          *services.BattledomeItemsService
	DataComparisonService           *services.DataComparisonService
	StatisticsService               *services.StatisticsService

	DataComparisonViewer *viewers.DataComparisonViewer
}

func NewServiceContainer() ServiceContainer {
	return ServiceContainer{}
}

func GetBattledomeItemsService() *services.BattledomeItemsService {
	return nil
}
