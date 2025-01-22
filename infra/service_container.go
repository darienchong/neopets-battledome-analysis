package infra

import (
	"reflect"
	"sync"

	"github.com/darienchong/neopets-battledome-analysis/caches"
	"github.com/darienchong/neopets-battledome-analysis/constants"
	"github.com/darienchong/neopets-battledome-analysis/loggers"
	"github.com/darienchong/neopets-battledome-analysis/parsers"
	"github.com/darienchong/neopets-battledome-analysis/services"
	"github.com/darienchong/neopets-battledome-analysis/viewers"
	"github.com/palantir/stacktrace"
)

type ServiceContainer struct {
	onces sync.Map

	ItemPriceCache caches.ItemPriceCache

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

var (
	containerOnce     *sync.Once = &sync.Once{}
	containerInstance *ServiceContainer
)

func ServiceContainerInstance() *ServiceContainer {
	containerOnce.Do(func() {
		containerInstance = NewServiceContainer()
	})
	return containerInstance
}

func NewServiceContainer() *ServiceContainer {
	return &ServiceContainer{}
}

func (sc *ServiceContainer) GetBattledomeItemsService() *services.BattledomeItemsService {
	once, _ := sc.onces.LoadOrStore(reflect.TypeOf(services.BattledomeItemsService{}), &sync.Once{})
	once.(*sync.Once).Do(func() {
		sc.BattledomeItemsService = services.NewBattledomeItemsService(
			sc.GetBattledomeItemGenerationService(),
			sc.GetGeneratedBattledomeItemParser(),
			sc.GetBattledomeItemDropDataParser(),
		)
	})
	return sc.BattledomeItemsService
}

func (sc *ServiceContainer) GetItemPriceCache() caches.ItemPriceCache {
	once, _ := sc.onces.LoadOrStore(reflect.TypeOf(caches.RealItemPriceCache{}), &sync.Once{})
	once.(*sync.Once).Do(func() {
		var dataSource caches.ItemPriceDataSource
		switch constants.ItemPriceDataSource {
		case constants.Unknown:
			panic(stacktrace.NewError("the data source for the item price cache wasn't specified"))
		case constants.JellyNeo:
			dataSource = caches.NewJellyNeoDataSource()
		case constants.ItemDB:
			dataSource = caches.NewItemDBDataSource()
		}
		cache, err := caches.ItemPriceCacheInstance(dataSource)
		if err != nil {
			panic(stacktrace.Propagate(err, "failed to get item price cache instance"))
		}
		sc.ItemPriceCache = cache
	})
	return sc.ItemPriceCache
}

func (sc *ServiceContainer) GetBattledomeItemsLogger() *loggers.BattledomeItemsLogger {
	once, _ := sc.onces.LoadOrStore(reflect.TypeOf(loggers.BattledomeItemsLogger{}), &sync.Once{})
	once.(*sync.Once).Do(func() {
		sc.BattledomeItemsLogger = loggers.NewBattledomeItemsLogger(
			sc.GetBattledomeItemsService(),
			sc.GetBattledomeItemDropDataParser(),
		)
	})
	return sc.BattledomeItemsLogger
}

func (sc *ServiceContainer) GetDataComparisonLogger() *loggers.DataComparisonLogger {
	once, _ := sc.onces.LoadOrStore(reflect.TypeOf(loggers.DataComparisonLogger{}), &sync.Once{})
	once.(*sync.Once).Do(func() {
		sc.DataComparisonLogger = loggers.NewDataComparisonLogger(
			sc.GetDataComparisonService(),
			sc.GetDataComparisonViewer(),
		)
	})
	return sc.DataComparisonLogger
}

func (sc *ServiceContainer) GetBattledomeItemDropDataParser() *parsers.BattledomeItemDropDataParser {
	once, _ := sc.onces.LoadOrStore(reflect.TypeOf(parsers.BattledomeItemDropDataParser{}), &sync.Once{})
	once.(*sync.Once).Do(func() {
		sc.BattledomeItemDropDataParser = parsers.NewBattledomeItemDropDataParser()
	})
	return sc.BattledomeItemDropDataParser
}

func (sc *ServiceContainer) GetBattledomeItemWeightParser() *parsers.BattledomeItemWeightParser {
	once, _ := sc.onces.LoadOrStore(reflect.TypeOf(parsers.BattledomeItemWeightParser{}), &sync.Once{})
	once.(*sync.Once).Do(func() {
		sc.BattledomeItemWeightParser = parsers.NewBattledomeItemWeightParser()
	})
	return sc.BattledomeItemWeightParser
}

func (sc *ServiceContainer) GetGeneratedBattledomeItemParser() *parsers.GeneratedBattledomeItemParser {
	once, _ := sc.onces.LoadOrStore(reflect.TypeOf(parsers.GeneratedBattledomeItemParser{}), &sync.Once{})
	once.(*sync.Once).Do(func() {
		sc.GeneratedBattledomeItemParser = parsers.NewGeneratedBattledomeItemParser()
	})
	return sc.GeneratedBattledomeItemParser
}

func (sc *ServiceContainer) GetBattledomeItemGenerationService() *services.BattledomeItemGenerationService {
	once, _ := sc.onces.LoadOrStore(reflect.TypeOf(services.BattledomeItemGenerationService{}), &sync.Once{})
	once.(*sync.Once).Do(func() {
		sc.BattledomeItemGenerationService = services.NewBattledomeItemGenerationService(
			sc.GetBattledomeItemWeightService(),
		)
	})
	return sc.BattledomeItemGenerationService
}

func (sc *ServiceContainer) GetBattledomeItemWeightService() *services.BattledomeItemWeightService {
	once, _ := sc.onces.LoadOrStore(reflect.TypeOf(services.BattledomeItemWeightService{}), &sync.Once{})
	once.(*sync.Once).Do(func() {
		sc.BattledomeItemWeightService = services.NewBattledomeItemWeightService(
			sc.GetBattledomeItemWeightParser(),
		)
	})
	return sc.BattledomeItemWeightService
}

func (sc *ServiceContainer) GetDataComparisonService() *services.DataComparisonService {
	once, _ := sc.onces.LoadOrStore(reflect.TypeOf(services.DataComparisonService{}), &sync.Once{})
	once.(*sync.Once).Do(func() {
		sc.DataComparisonService = services.NewDataComparisonService(
			sc.GetBattledomeItemsService(),
		)
	})
	return sc.DataComparisonService
}

func (sc *ServiceContainer) GetStatisticsService() *services.StatisticsService {
	once, _ := sc.onces.LoadOrStore(reflect.TypeOf(services.StatisticsService{}), &sync.Once{})
	once.(*sync.Once).Do(func() {
		sc.StatisticsService = services.NewStatisticsService()
	})
	return sc.StatisticsService
}

func (sc *ServiceContainer) GetDataComparisonViewer() *viewers.DataComparisonViewer {
	once, _ := sc.onces.LoadOrStore(reflect.TypeOf(viewers.DataComparisonViewer{}), &sync.Once{})
	once.(*sync.Once).Do(func() {
		sc.DataComparisonViewer = viewers.NewDataComparisonViewer(
			sc.GetBattledomeItemsService(),
			sc.GetDataComparisonService(),
			sc.GetStatisticsService(),
		)
	})
	return sc.DataComparisonViewer
}
