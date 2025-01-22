package loggers

import (
	"fmt"
	"log/slog"
	"math"
	"strconv"

	"github.com/darienchong/neopets-battledome-analysis/caches"
	"github.com/darienchong/neopets-battledome-analysis/constants"
	"github.com/darienchong/neopets-battledome-analysis/helpers"
	"github.com/darienchong/neopets-battledome-analysis/models"
	"github.com/darienchong/neopets-battledome-analysis/parsers"
	"github.com/darienchong/neopets-battledome-analysis/services"
	"github.com/palantir/stacktrace"
)

type BattledomeItemsLogger struct {
	BattledomeItemsService       *services.BattledomeItemsService
	BattledomeItemDropDataParser *parsers.BattledomeItemDropDataParser
}

func NewArenaDropsLogger() *BattledomeItemsLogger {
	return &BattledomeItemsLogger{
		BattledomeItemsService:       services.NewBattledomeItemsService(),
		BattledomeItemDropDataParser: parsers.NewBattledomeItemDropDataParser(),
	}
}

func (l *BattledomeItemsLogger) Log(dataFolderPath string, numDropsToPrint int) error {
	if numDropsToPrint <= 0 {
		numDropsToPrint = constants.NumberOfDropsToPrint
	}

	if constants.FilterArena != "" {
		slog.Info(fmt.Sprintf("Only displaying data related to \"%s\"", constants.FilterArena))
	}

	itemPriceCache, err := caches.CurrentItemPriceCacheInstance()
	if err != nil {
		return stacktrace.Propagate(err, "failed to get item price cache instance")
	}

	files, err := helpers.FilesInFolder(dataFolderPath)
	if err != nil {
		return stacktrace.Propagate(err, "failed to get files in %s", dataFolderPath)
	}

	if numDropsToPrint > 0 {
		files = files[int(math.Max(float64(len(files)-numDropsToPrint), 0)):]
	}

	samplesByArena := map[models.Arena]models.BattledomeItems{}
	for _, file := range files {
		items, err := l.BattledomeItemDropDataParser.Parse(constants.DropDataFilePath(file))
		if err != nil {
			return stacktrace.Propagate(err, "failed to parse drop data file: %s", file)
		}

		_, isKeyExists := samplesByArena[items.Metadata.Arena]
		if !isKeyExists {
			samplesByArena[items.Metadata.Arena] = models.BattledomeItems{}
		}
		samplesByArena[items.Metadata.Arena] = append(samplesByArena[items.Metadata.Arena], items.Items...)

		if constants.FilterArena != "" && constants.FilterArena != items.Metadata.Arena {
			continue
		}

		itemCount := 0
		profitBreakdownTable := helpers.NewTable([]string{
			"i",
			"Item Name",
			"Qty",
			"Price",
			"Profit",
			"%-age",
		})
		profitBreakdownTable.IsLastRowDistinct = true

		normalisedItems, err := items.Items.Normalise()
		if err != nil {
			return helpers.PropagateWithSerialisedValue(err, "failed to normalise items: %s", "failed to normalise items; another error occurred while trying to serialise the input: %s", items)
		}

		orderedNormalisedItems, err := normalisedItems.ItemsOrderedByProfit()
		if err != nil {
			return helpers.PropagateWithSerialisedValue(err, "failed to get items ordered by profit: %s", "failed to get items ordered by profit; another error occurred while trying to serialise the input: %s", normalisedItems)
		}

		for i, item := range orderedNormalisedItems {
			itemCount += int(item.Quantity)
			itemProfit := item.Profit(itemPriceCache)
			itemPercentageProfit, err := item.PercentageProfit(itemPriceCache, normalisedItems)
			if err != nil {
				return helpers.PropagateWithSerialisedValue(err, "failed to get percentage profit: %s", "failed to get percentage profit; another error occurred while trying to serialise the input: %s", items)
			}
			if itemPercentageProfit < 0.01 {
				continue
			}
			profitBreakdownTable.AddRow([]string{
				strconv.Itoa(i + 1),
				string(item.Name),
				strconv.Itoa(int(item.Quantity)),
				helpers.FormatFloat(itemPriceCache.Price(string(item.Name))) + " NP",
				helpers.FormatFloat(itemProfit) + " NP",
				helpers.FormatPercentage(itemPercentageProfit) + "%",
			})
		}

		totalProfit, err := normalisedItems.TotalProfit()
		if err != nil {
			return helpers.PropagateWithSerialisedValue(err, "failed to get total profit: %s", "failed to get total profit; an error occurred while trying to serialise the input to log: %s", normalisedItems)
		}

		profitBreakdownTable.AddRow([]string{
			"",
			"Total",
			helpers.FormatFloat(float64(itemCount)),
			"",
			helpers.FormatFloat(totalProfit) + " NP",
			"",
		})

		slog.Info(items.Metadata.String())
		for _, line := range profitBreakdownTable.Lines() {
			slog.Info("\t" + line)
		}
		slog.Info("")
	}

	return nil
}
