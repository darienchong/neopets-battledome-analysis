package caches

import (
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/darienchong/neopets-battledome-analysis/constants"
	"github.com/darienchong/neopets-battledome-analysis/helpers"
)

var _ ItemPriceDataSource = (*JellyNeoDataSource)(nil)

type JellyNeoDataSource struct {
}

func NewJellyNeoDataSource() ItemPriceDataSource {
	return &JellyNeoDataSource{}
}

func (dataSource *JellyNeoDataSource) GetFilePath() string {
	return constants.CombineRelativeFolderAndFilename(constants.DATA_FOLDER, constants.JELLYNEO_ITEM_PRICE_CACHE_FILE)
}

func getNormalisedJellyNeoItemName(itemName string) string {
	return url.QueryEscape(itemName)
}

func getJellyNeoPriceUrl(itemName string) string {
	return fmt.Sprintf("https://items.jellyneo.net/search/?name=%s&name_type=3", getNormalisedJellyNeoItemName(itemName))
}

func (dataSource JellyNeoDataSource) GetPrice(itemName string) float64 {
	if slices.Contains(bannedItems, itemName) {
		return 0.0
	}

	url := getJellyNeoPriceUrl(itemName)
	slog.Debug(fmt.Sprintf(`Calling "%s" for price`, url))

	res, err := helpers.HumanlikeGet(url)
	if err != nil {
		return 0.0
	}
	defer res.Body.Close()

	bodyCopy, err := io.ReadAll(res.Body)
	if err != nil {
		return 0.0
	}

	reader := strings.NewReader(string(bodyCopy))
	doc, err := goquery.NewDocumentFromReader(reader)
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
		slog.Debug(fmt.Sprintf(`Response from JellyNeo for %s: %s`, url, string(bodyCopy)))
		slog.Error(fmt.Sprintf(`Failed to retrieve price for "%s" from JellyNeo!`, itemName))
	}
	return price
}
