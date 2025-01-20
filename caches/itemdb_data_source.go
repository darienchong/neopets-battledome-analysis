package caches

import (
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/darienchong/neopets-battledome-analysis/constants"
)

var _ ItemPriceDataSource = (*ItemDbDataSource)(nil)

type ItemDbDataSource struct{}

func NewItemDbDataSource() ItemPriceDataSource {
	return &ItemDbDataSource{}
}

func getItemDbPriceUrl(itemName string) string {
	return fmt.Sprintf("https://itemdb.com.br/item/%s", getNormalisedItemName(itemName))
}

func (cache *ItemDbDataSource) GetFilePath() string {
	return constants.ITEMDB_ITEM_PRICE_CACHE_FILE
}

func (cache *ItemDbDataSource) GetPrice(itemName string) float64 {
	if slices.Contains(bannedItems, itemName) {
		return 0.0
	}

	res, err := http.Get(getItemDbPriceUrl(itemName))
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
