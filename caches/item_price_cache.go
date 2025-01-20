package caches

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/darienchong/neopets-battledome-analysis/constants"
	"github.com/palantir/stacktrace"
)

type ItemPriceCache interface {
	GetPrice(itemName string) float64
	Close() error
}

type RealItemPriceCache struct {
	dataSource    ItemPriceDataSource
	expiry        time.Time
	cachedPrices  map[string]float64
	specialPrices map[string]float64
}

var (
	_             ItemPriceCache = (*RealItemPriceCache)(nil)
	cacheInstance ItemPriceCache
	lock          = &sync.Mutex{}
	bannedItems   = []string{
		"nothing",
	}
)

func GetCurrentItemPriceCacheInstance() (ItemPriceCache, error) {
	if cacheInstance == nil {
		return nil, stacktrace.NewError("tried to get current item price cache instance when it was not yet initialised")
	}
	return cacheInstance, nil
}

// Note: the first invocation of this method will determine what the data source is
func GetItemPriceCacheInstance(dataSource ItemPriceDataSource) (ItemPriceCache, error) {
	if cacheInstance == nil {
		lock.Lock()
		defer lock.Unlock()
		if cacheInstance == nil {
			realCacheInstance := &RealItemPriceCache{
				dataSource:    dataSource,
				cachedPrices:  map[string]float64{},
				specialPrices: map[string]float64{},
			}
			realCacheInstance.generateExpiry()
			err := realCacheInstance.loadFromFile()
			if err != nil {
				return nil, stacktrace.Propagate(err, "failed to load cache data from file")
			}
			err = realCacheInstance.loadSpecialPrices()
			if err != nil {
				return nil, stacktrace.Propagate(err, "failed to load special prices")
			}

			cacheInstance = realCacheInstance
		}
	}
	return cacheInstance, nil
}

func (cache RealItemPriceCache) loadSpecialPrices() error {
	return nil
}

func (cache *RealItemPriceCache) generateExpiry() {
	cache.expiry = time.Now().AddDate(0, 0, 7)
}

func (cache *RealItemPriceCache) GetPrice(itemName string) float64 {
	if itemName == "nothing" {
		return 0.0
	}

	if maybeCachedValue, existsInCache := cache.cachedPrices[itemName]; existsInCache {
		return maybeCachedValue
	}

	if maybeSpecialPrice, existsInSpecialPrices := cache.specialPrices[itemName]; existsInSpecialPrices {
		return maybeSpecialPrice
	}

	cache.cachedPrices[itemName] = cache.dataSource.GetPrice(itemName)
	return cache.cachedPrices[itemName]
}

func (cache *RealItemPriceCache) flushToFile() error {
	file, err := os.OpenFile(cache.dataSource.GetFilePath(), os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0755)
	if err != nil {
		return stacktrace.Propagate(err, "failed to open item price cache file (%s) when flushing to disk", cache.dataSource.GetFilePath())
	}
	defer file.Close()

	file.WriteString(fmt.Sprintf("%s\n", cache.expiry.Format(constants.DATA_EXPIRY_TIME_LAYOUT)))
	for key, value := range cache.cachedPrices {
		file.WriteString(fmt.Sprintf("%s|%f\n", key, value))
	}

	return nil
}

func (cache *RealItemPriceCache) Close() error {
	return cache.flushToFile()
}

func (cache *RealItemPriceCache) loadFromFile() error {
	if cache.cachedPrices == nil {
		cache.cachedPrices = map[string]float64{}
	}

	_, err := os.Stat(cache.dataSource.GetFilePath())
	if os.IsNotExist(err) {
		cache.generateExpiry()
		return nil
	}

	file, err := os.Open(cache.dataSource.GetFilePath())
	if err != nil {
		return stacktrace.Propagate(err, "failed to open the item price cache file path: %s", cache.dataSource.GetFilePath())
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		isPriceData := strings.Contains(line, "|")
		if !isPriceData {
			// It's the expiry date
			parsedExpiry, err := time.Parse(constants.DATA_EXPIRY_TIME_LAYOUT, line)
			if err != nil {
				slog.Warn(fmt.Sprintf("Failed to parse expiry for item price cache file; the line was \"%s\": %s", line, err))
				cache.generateExpiry()
			} else {
				cache.expiry = parsedExpiry
			}

			if parsedExpiry.Before(time.Now()) {
				slog.Info("Deleting item price cache file as it was expired.")
				// We don't really care if it succeeds or not
				os.Remove(cache.dataSource.GetFilePath())
				cache.generateExpiry()
			}
		} else {
			data := strings.Split(line, "|")
			itemName := data[0]
			itemPrice, err := strconv.ParseFloat(data[1], 64)
			if err != nil {
				continue
			}
			cache.cachedPrices[itemName] = itemPrice
		}
	}

	return nil
}
