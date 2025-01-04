package caches

import (
	"bufio"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/darienchong/neopets-battledome-analysis/constants"
)

var lock = &sync.Mutex{}

type ItemPriceCache struct {
	expiry       time.Time
	cachedPrices map[string]float64
}

var cacheInstance *ItemPriceCache

var bannedItems = []string{"nothing"}

func GetItemPriceCacheInstance() *ItemPriceCache {
	if cacheInstance == nil {
		lock.Lock()
		defer lock.Unlock()
		if cacheInstance == nil {
			cacheInstance = &ItemPriceCache{}
			cacheInstance.generateExpiry()
			cacheInstance.cachedPrices = map[string]float64{}
			cacheInstance.loadFromFile()
		}
	}
	return cacheInstance
}

func (cache *ItemPriceCache) generateExpiry() {
	cache.expiry = time.Now().AddDate(0, 0, 7)
}

func (cache ItemPriceCache) getNormalisedItemName(itemName string) string {
	itemName = strings.ToLower(itemName)
	itemName = strings.ReplaceAll(itemName, " ", "-")
	itemName = strings.ReplaceAll(itemName, ":", "")
	itemName = strings.ReplaceAll(itemName, "!", "")
	return itemName
}

func (cache ItemPriceCache) getItemDbPriceUrl(itemName string) string {
	return fmt.Sprintf("https://itemdb.com.br/item/%s", cache.getNormalisedItemName(itemName))
}

func (cache *ItemPriceCache) getPriceFromItemDb(itemName string) float64 {
	if slices.Contains(bannedItems, itemName) {
		return 0.0
	}

	res, err := http.Get(cache.getItemDbPriceUrl(itemName))
	if err != nil {
		return 0.0
	}
	defer res.Body.Close()
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return 0.0
	}
	price := 0.0
	doc.Find(".chakra-stat__number").Each(func(index int, item *goquery.Selection) {
		curr_price, err := strconv.ParseFloat(strings.ReplaceAll(strings.ReplaceAll(item.Text(), " NP", ""), ",", ""), 64)
		if err == nil {
			price = curr_price
		}
	})
	if price == 0.0 {
		slog.Warn(fmt.Sprintf("Failed to retrieve price for \"%s\" from ItemDb!", itemName))
	}
	return price
}

func (cache *ItemPriceCache) GetPrice(itemName string) float64 {
	maybeCachedValue, existsInCache := cache.cachedPrices[itemName]
	if existsInCache {
		return maybeCachedValue
	}

	priceFromItemDb := cache.getPriceFromItemDb(itemName)
	cache.cachedPrices[itemName] = priceFromItemDb
	return priceFromItemDb
}

func (cache *ItemPriceCache) flushToFile() {
	file, err := os.OpenFile(constants.GetItemPriceCacheFilePath(), os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0755)
	if err != nil {
		slog.Error("Failed to open item price cache file when flushing.")
		panic(err)
	}
	defer file.Close()
	file.WriteString(fmt.Sprintf("%s\n", cache.expiry.Format(constants.DATA_EXPIRY_TIME_LAYOUT)))
	for key, value := range cache.cachedPrices {
		file.WriteString(fmt.Sprintf("%s|%f\n", key, value))
	}
}

func (cache *ItemPriceCache) Close() {
	cache.flushToFile()
}

func (cache *ItemPriceCache) loadFromFile() {
	if cache.cachedPrices == nil {
		cache.cachedPrices = map[string]float64{}
	}

	_, err := os.Stat(constants.GetItemPriceCacheFilePath())
	if os.IsNotExist(err) {
		cache.generateExpiry()
		return
	}

	file, err := os.Open(constants.GetItemPriceCacheFilePath())
	if err != nil {
		panic(err)
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
				slog.Warn(fmt.Sprintf(`Failed to parse expiry for item price cache file; the line was %s`, line))
				cache.generateExpiry()
			} else {
				cache.expiry = parsedExpiry
			}

			if parsedExpiry.Before(time.Now()) {
				slog.Info("Deleting item price cache file as it was expired.")
				err := os.Remove(constants.GetItemPriceCacheFilePath())
				cache.generateExpiry()
				if err != nil {
					slog.Any("error", err)
				}
				return
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
}
