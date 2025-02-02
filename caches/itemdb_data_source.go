package caches

import (
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/darienchong/neopets-battledome-analysis/constants"
	"github.com/palantir/stacktrace"
)

var _ ItemPriceDataSource = (*ItemDBDataSource)(nil)

type ItemDBDataSource struct{}

func NewItemDBDataSource() ItemPriceDataSource {
	return &ItemDBDataSource{}
}

func normalisedItemDBItemName(itemName string) string {
	itemName = strings.ToLower(itemName)
	itemName = strings.ReplaceAll(itemName, " ", "-")
	itemName = strings.ReplaceAll(itemName, ":", "")
	itemName = strings.ReplaceAll(itemName, "!", "")
	return itemName
}

func itemDBPriceUrl(itemName string) string {
	return fmt.Sprintf("https://itemdb.com.br/item/%s", normalisedItemDBItemName(itemName))
}

func (ds *ItemDBDataSource) FilePath() string {
	return constants.CombineRelativeFolderAndFilename(constants.DataFolder, constants.ItemDBItemPriceCacheFile)
}

func (ds *ItemDBDataSource) Price(itemName string) (float64, error) {
	if slices.Contains(bannedItems, itemName) {
		return 0.0, stacktrace.NewError(fmt.Sprintf("item %q was banned from search", itemName))
	}

	res, err := http.Get(itemDBPriceUrl(itemName))
	if err != nil {
		return 0.0, stacktrace.Propagate(err, "failed to reach ItemDB")
	}
	defer res.Body.Close()
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return 0.0, stacktrace.Propagate(err, "failed to parse HTML response from ItemDB")
	}
	price := 0.0
	doc.Find(".chakra-stat__number").Each(func(index int, item *goquery.Selection) {
		curr_price, err := strconv.ParseFloat(strings.ReplaceAll(strings.ReplaceAll(item.Text(), " NP", ""), ",", ""), 64)
		if err == nil {
			price = curr_price
		}
	})
	if price == 0.0 {
		return price, stacktrace.NewError(fmt.Sprintf("Failed to retrieve price for %q from ItemDB!", itemName))
	}

	return price, nil
}
