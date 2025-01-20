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

var _ ItemPriceDataSource = (*JellyNeoDataSource)(nil)

type JellyNeoDataSource struct {
}

func NewJellyNeoDataSource() ItemPriceDataSource {
	return &JellyNeoDataSource{}
}

func (dataSource *JellyNeoDataSource) GetFilePath() string {
	return constants.JELLYNEO_ITEM_PRICE_CACHE_FILE
}

func getJellyNeoPriceUrl(itemName string) string {
	return fmt.Sprintf("https://items.jellyneo.net/search/?name=%s&name_type=3", getNormalisedItemName(itemName))
}

func (dataSource JellyNeoDataSource) GetPrice(itemName string) float64 {
	if slices.Contains(bannedItems, itemName) {
		return 0.0
	}

	res, err := http.Get(getJellyNeoPriceUrl(itemName))
	if err != nil {
		return 0.0
	}
	defer res.Body.Close()
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return 0.0
	}

	price := 0.0
	doc.Find(".price-history-link").Each(func(index int, item *goquery.Selection) {
		currPrice, err := strconv.ParseFloat(strings.ReplaceAll(strings.ReplaceAll(item.Text(), " NP", ""), ",", ""), 64)
		if err == nil {
			price = currPrice
		}
	})
	if price == 0.0 {
		slog.Warn(fmt.Sprintf("Failed to retrieve price for \"%s\" from JellyNeo!", itemName))
	}
	return price
}
