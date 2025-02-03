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
	"github.com/palantir/stacktrace"
)

var _ ItemPriceDataSource = (*JellyNeoDataSource)(nil)

type JellyNeoDataSource struct {
}

func NewJellyNeoDataSource() ItemPriceDataSource {
	return &JellyNeoDataSource{}
}

func (ds *JellyNeoDataSource) FilePath() string {
	return constants.CombineRelativeFolderAndFilename(constants.DataFolder, constants.JellyNeoItemPriceCacheFile)
}

func normalisedJellyNeoItemName(itemName string) string {
	return url.QueryEscape(itemName)
}

func jellyNeoPriceUrl(itemName string) string {
	return fmt.Sprintf("https://items.jellyneo.net/search/?name=%s&name_type=3", normalisedJellyNeoItemName(itemName))
}

func (ds *JellyNeoDataSource) Price(itemName string) (float64, error) {
	if slices.Contains(bannedItems, itemName) {
		return 0.0, stacktrace.NewError(fmt.Sprintf("item %q was a banned item", itemName))
	}

	url := jellyNeoPriceUrl(itemName)
	slog.Debug(fmt.Sprintf(`Calling "%s" for price`, url))

	res, err := helpers.HumanlikeGet(url)
	if err != nil {
		return 0.0, stacktrace.Propagate(err, "failed to reach JellyNeo")
	}
	defer res.Body.Close()

	bodyCopy, err := io.ReadAll(res.Body)
	if err != nil {
		return 0.0, stacktrace.Propagate(err, "failed to read response body")
	}

	reader := strings.NewReader(string(bodyCopy))
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return 0.0, stacktrace.Propagate(err, "failed to create reader from response body")
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
		return 0.0, stacktrace.NewError(fmt.Sprintf("failed to retrieve price for %q from JellyNeo", itemName))
	}
	return price, nil
}
