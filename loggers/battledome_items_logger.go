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

func (dropsLogger *BattledomeItemsLogger) Log(dataFolderPath string) error {
	if constants.FILTER_ARENA != "" {
		slog.Info(fmt.Sprintf("Only displaying data related to \"%s\"", constants.FILTER_ARENA))
	}

	itemPriceCache, err := caches.GetItemPriceCacheInstance()
	if err != nil {
		return fmt.Errorf("failed to get item price cache instance: %w", err)
	}
	defer itemPriceCache.Close()

	files, err := helpers.GetFilesInFolder(dataFolderPath)
	if err != nil {
		return fmt.Errorf("failed to get files in %s: %w", dataFolderPath, err)
	}

	if constants.NUMBER_OF_DROPS_TO_PRINT > 0 {
		files = files[int(math.Max(float64(len(files)-constants.NUMBER_OF_DROPS_TO_PRINT), 0)):]
	}

	samplesByArena := map[models.Arena]models.BattledomeItems{}
	for _, file := range files {
		items, err := dropsLogger.BattledomeItemDropDataParser.Parse(constants.GetDropDataFilePath(file))
		if err != nil {
			slog.Error(fmt.Sprintf("Failed to parse drop data file (%s)", file))
			panic(err)
		}

		_, isKeyExists := samplesByArena[items.Metadata.Arena]
		if !isKeyExists {
			samplesByArena[items.Metadata.Arena] = models.BattledomeItems{}
		}
		samplesByArena[items.Metadata.Arena] = append(samplesByArena[items.Metadata.Arena], items.Items...)

		if constants.FILTER_ARENA != "" && constants.FILTER_ARENA != items.Metadata.Arena {
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
			panic(err)
		}
		orderedNormalisedItems, err := normalisedItems.GetItemsOrderedByProfit()
		if err != nil {
			panic(err)
		}

		for i, item := range orderedNormalisedItems {
			itemCount += int(item.Quantity)
			itemProfit := item.GetProfit(itemPriceCache)
			itemPercentageProfit, err := item.GetPercentageProfit(itemPriceCache, normalisedItems)
			if err != nil {
				panic(err)
			}
			if itemPercentageProfit < 0.01 {
				continue
			}
			profitBreakdownTable.AddRow([]string{
				strconv.Itoa(i + 1),
				string(item.Name),
				strconv.Itoa(int(item.Quantity)),
				helpers.FormatFloat(itemPriceCache.GetPrice(string(item.Name))) + " NP",
				helpers.FormatFloat(itemProfit) + " NP",
				helpers.FormatPercentage(itemPercentageProfit) + "%",
			})
		}

		totalProfit, err := normalisedItems.GetTotalProfit()
		if err != nil {
			panic(err)
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
		for _, line := range profitBreakdownTable.GetLines() {
			slog.Info("\t" + line)
		}
		slog.Info("")
	}

	return nil
}
