package services

import (
	"fmt"
	"log/slog"
	"strconv"

	"github.com/darienchong/neopets-battledome-analysis/caches"
	"github.com/darienchong/neopets-battledome-analysis/constants"
	"github.com/darienchong/neopets-battledome-analysis/helpers"
	"github.com/darienchong/neopets-battledome-analysis/models"
)

type ArenaDropsLogger struct{}

func (dropsLogger *ArenaDropsLogger) Log(dataFolderPath string) {
	if constants.FILTER_ARENA != "" {
		slog.Info(fmt.Sprintf("Only displaying data related to \"%s\"", constants.FILTER_ARENA))
	}

	parser := new(DropDataParser)
	analyser := new(EmpiricalDropRateEstimator)
	itemPriceCache := caches.GetItemPriceCacheInstance()
	defer itemPriceCache.Close()

	files, err := helpers.GetFilesInFolder(dataFolderPath)
	if err != nil {
		slog.Error("Failed to get files in folder!")
		panic(err)
	}

	samplesByArena := map[string][]*models.BattledomeDrops{}
	for _, file := range files {
		drops, err := parser.Parse(constants.GetDropDataFilePath(file))
		if err != nil {
			slog.Error(fmt.Sprintf("Failed to parse drop data file (%s)", file))
			panic(err)
		}

		_, isKeyExists := samplesByArena[drops.Metadata.Arena]
		if !isKeyExists {
			samplesByArena[drops.Metadata.Arena] = []*models.BattledomeDrops{}
		}
		samplesByArena[drops.Metadata.Arena] = append(samplesByArena[drops.Metadata.Arena], drops)

		if constants.FILTER_ARENA != "" && constants.FILTER_ARENA != drops.Metadata.Arena {
			continue
		}

		itemCount := 0
		res := analyser.Analyse(drops)
		profitBreakdownTable := helpers.NewTable([]string{
			"i",
			"Item Name",
			"Qty",
			"Price",
			"Profit",
			"%-age",
		})
		profitBreakdownTable.IsLastRowDistinct = true

		for i, item := range res.GetItemsOrderedByProfit() {
			itemCount += int(item.Quantity)
			itemProfit, err := item.GetProfit()
			if err != nil {
				panic(err)
			}
			itemPercentageProfit, err := item.GetPercentageProfit(res)
			if err != nil {
				panic(err)
			}
			if itemPercentageProfit < 0.01 {
				continue
			}
			profitBreakdownTable.AddRow([]string{
				strconv.Itoa(i + 1),
				item.Name,
				strconv.Itoa(int(item.Quantity)),
				helpers.FormatFloat(item.IndividualPrice) + " NP",
				helpers.FormatFloat(itemProfit) + " NP",
				helpers.FormatPercentage(itemPercentageProfit) + "%",
			})
		}
		profitBreakdownTable.AddRow([]string{
			"",
			"Total",
			helpers.FormatFloat(float64(itemCount)),
			"",
			helpers.FormatFloat(res.GetTotalProfit()) + " NP",
			"",
		})

		slog.Info(res.Metadata.String())
		profitBreakdownTable.LogWithPrefix("\t")
		slog.Info("")
	}
}
