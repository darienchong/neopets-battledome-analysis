package services

import (
	"fmt"
	"log/slog"
	"math"
	"slices"
	"strconv"

	"github.com/darienchong/neopets-battledome-analysis/caches"
	"github.com/darienchong/neopets-battledome-analysis/constants"
	"github.com/darienchong/neopets-battledome-analysis/helpers"
	"github.com/darienchong/neopets-battledome-analysis/models"
)

type ProfitEstimationLogger struct{}

func NewProfitEstimationLogger() *ProfitEstimationLogger {
	return &ProfitEstimationLogger{}
}

func (logger *ProfitEstimationLogger) Log() {
	itemPriceCache, err := caches.GetItemPriceCacheInstance()
	if err != nil {
		panic(err)
	}
	defer itemPriceCache.Close()

	statsEstimator := NewArenaProfitStatisticsEstimator()
	estimator := NewDropRateEstimator()
	dropRates, err := estimator.Estimate()
	if err != nil {
		panic(err)
	}

	itemProfits := helpers.Map(dropRates, func(dropRate models.ItemDropRate) models.ItemProfit {
		return models.ItemProfit{
			ItemDropRate:    dropRate,
			IndividualPrice: itemPriceCache.GetPrice(dropRate.ItemName),
		}
	})
	profitsByArena := helpers.GroupBy(itemProfits, func(itemProfit models.ItemProfit) string {
		return itemProfit.Arena
	})
	orderedProfitsByArena := helpers.OrderByDescending(helpers.ToSlice(profitsByArena), func(tuple helpers.Tuple) float64 {
		currProfits := tuple.Elements[1].([]*models.ItemProfit)
		return helpers.Sum(helpers.Map(currProfits, func(profit *models.ItemProfit) float64 {
			return profit.GetProfit()
		}))
	})
	arenaStats, err := statsEstimator.Estimate()
	if err != nil {
		panic(err)
	}

	slog.Info("")
	for i, tuple := range orderedProfitsByArena {
		arena := tuple.Elements[0].(string)
		profits := tuple.Elements[1].([]*models.ItemProfit)
		totalProfit := helpers.Sum(helpers.Map(profits, func(profit *models.ItemProfit) float64 {
			return profit.GetProfit()
		}))
		arenaStat := arenaStats[arena]
		slog.Info(
			fmt.Sprintf("%d. %s (%s ± %s NP)",
				i+1,
				arena,
				helpers.FormatFloat(totalProfit*constants.BATTLEDOME_DROPS_PER_DAY),
				helpers.FormatFloat(arenaStat.StandardDeviation*math.Sqrt(constants.BATTLEDOME_DROPS_PER_DAY))))
		table := helpers.NewTable([]string{
			"i",
			"Item Name",
			"Price",
			"Drop Rate",
			"Dry Chance",
			"Expectation",
			"Profit %-age",
		})
		profits = helpers.OrderByDescending(profits, func(profit *models.ItemProfit) float64 {
			return profit.GetPercentageProfit(totalProfit)
		})
		for j, itemProfit := range profits {
			if j > constants.NUMBER_OF_ITEMS_TO_PRINT-1 {
				break
			}

			err = table.AddRow([]string{
				strconv.Itoa(j + 1),
				itemProfit.ItemName,
				helpers.FormatFloat(itemProfit.IndividualPrice) + " NP",
				helpers.FormatPercentage(itemProfit.DropRate) + "%",
				helpers.FormatPercentage(GetDryChance(itemProfit.DropRate, 30*constants.BATTLEDOME_DROPS_PER_DAY)) + "%",
				helpers.FormatFloat(itemProfit.GetProfit()) + " NP",
				helpers.FormatPercentage(itemProfit.GetPercentageProfit(totalProfit)) + "%",
			})
			if err != nil {
				panic(err)
			}
		}
		table.Log()
		slog.Info("")
		additionalDataTable := helpers.NewTable([]string{
			"Drop Rate",
			"Value",
		})
		additionalDataTable.IsLastRowDistinct = true

		brownCodestoneDropRate := helpers.Sum(helpers.Map(helpers.Filter(profits, func(profit *models.ItemProfit) bool {
			return slices.Contains(constants.BROWN_CODESTONES, profit.ItemName)
		}), func(profit *models.ItemProfit) float64 {
			return profit.DropRate
		}))
		redCodestoneDropRate := helpers.Sum(helpers.Map(helpers.Filter(profits, func(profit *models.ItemProfit) bool {
			return slices.Contains(constants.RED_CODESTONES, profit.ItemName)
		}), func(profit *models.ItemProfit) float64 {
			return profit.DropRate
		}))
		additionalDataTable.AddRow([]string{
			"Brown Codestones",
			helpers.FormatPercentage(brownCodestoneDropRate) + "%",
		})
		additionalDataTable.AddRow([]string{
			"Red Codestones",
			helpers.FormatPercentage(redCodestoneDropRate) + "%",
		})
		additionalDataTable.AddRow([]string{
			"Sum",
			helpers.FormatPercentage(brownCodestoneDropRate+redCodestoneDropRate) + "%",
		})
		additionalDataTable.Log()
		slog.Info("")
	}
}
