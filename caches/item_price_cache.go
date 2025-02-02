package caches

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/darienchong/neopets-battledome-analysis/constants"
	"github.com/darienchong/neopets-battledome-analysis/helpers"
	"github.com/palantir/stacktrace"
)

type ItemPriceCache interface {
	Price(itemName string) float64
	Close() error
}

type RealItemPriceCache struct {
	retryPolicy   helpers.RetryPolicy[float64]
	dataSource    ItemPriceDataSource
	expiry        time.Time
	cachedPrices  map[string]float64
	specialPrices map[string]float64
}

var (
	_           ItemPriceCache = (*RealItemPriceCache)(nil)
	bannedItems                = []string{
		"nothing",
	}
)

func ItemPriceCacheInstance(dataSource ItemPriceDataSource) (ItemPriceCache, error) {
	var err error
	realCacheInstance := &RealItemPriceCache{
		retryPolicy: helpers.RetryPolicy[float64]{
			Backoff: func(retryCount int) int {
				return 500
			},
			MaxTries: 3,
		},
		dataSource:    dataSource,
		cachedPrices:  map[string]float64{},
		specialPrices: map[string]float64{},
	}
	realCacheInstance.generateExpiry()
	if err = realCacheInstance.loadFromFile(); err != nil {
		err = stacktrace.Propagate(err, "failed to load cache data from file")
		return nil, err
	}
	if err = realCacheInstance.loadSpecialPrices(); err != nil {
		err = stacktrace.Propagate(err, "failed to load special prices")
		return nil, err
	}
	return ItemPriceCache(realCacheInstance), nil
}

func (c *RealItemPriceCache) loadSpecialPrices() error {
	return nil
}

func (c *RealItemPriceCache) generateExpiry() {
	c.expiry = time.Now().AddDate(0, 0, 7)
}

func (c *RealItemPriceCache) Price(itemName string) float64 {
	if itemName == "nothing" {
		return 0.0
	}

	if maybeCachedValue, existsInCache := c.cachedPrices[itemName]; existsInCache {
		return maybeCachedValue
	}

	if maybeSpecialPrice, existsInSpecialPrices := c.specialPrices[itemName]; existsInSpecialPrices {
		return maybeSpecialPrice
	}

	price, err := c.retryPolicy.Execute(func() (float64, error) {
		return c.dataSource.Price(itemName)
	})

	if err != nil {
		slog.Error(fmt.Sprintf("%+v", err))
		return 0.0
	}

	if price > 0 {
		c.cachedPrices[itemName] = price
	}
	return price
}

func (c *RealItemPriceCache) flushToFile() error {
	file, err := os.OpenFile(c.dataSource.FilePath(), os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0755)
	if err != nil {
		return stacktrace.Propagate(err, "failed to open item price cache file (%s) when flushing to disk", c.dataSource.FilePath())
	}
	defer file.Close()

	file.WriteString(fmt.Sprintf("%s\n", c.expiry.Format(constants.DataExpiryTimeLayout)))
	for key, value := range c.cachedPrices {
		file.WriteString(fmt.Sprintf("%s|%f\n", key, value))
	}

	return nil
}

func (c *RealItemPriceCache) Close() error {
	return c.flushToFile()
}

func (c *RealItemPriceCache) loadFromFile() error {
	if c.cachedPrices == nil {
		c.cachedPrices = map[string]float64{}
	}

	_, err := os.Stat(c.dataSource.FilePath())
	if os.IsNotExist(err) {
		c.generateExpiry()
		return nil
	}

	file, err := os.Open(c.dataSource.FilePath())
	if err != nil {
		return stacktrace.Propagate(err, "failed to open the item price cache file path: %s", c.dataSource.FilePath())
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		isPriceData := strings.Contains(line, "|")
		if !isPriceData {
			// It's the expiry date
			parsedExpiry, err := time.Parse(constants.DataExpiryTimeLayout, line)
			if err != nil {
				slog.Warn(fmt.Sprintf("Failed to parse expiry for item price cache file; the line was %q: %s", line, err))
				c.generateExpiry()
			} else {
				c.expiry = parsedExpiry
			}

			if parsedExpiry.Before(time.Now()) {
				slog.Info("Deleting item price cache file as it was expired.")
				// We don't really care if it succeeds or not
				os.Remove(c.dataSource.FilePath())
				c.generateExpiry()
			}
		} else {
			data := strings.Split(line, "|")
			itemName := data[0]
			itemPrice, err := strconv.ParseFloat(data[1], 64)
			if err != nil {
				continue
			}
			c.cachedPrices[itemName] = itemPrice
		}
	}

	return nil
}
