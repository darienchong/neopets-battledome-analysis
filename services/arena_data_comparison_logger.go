package services

import (
	"fmt"
	"log/slog"
	"math"
	"slices"
	"strconv"
	"strings"

	"github.com/darienchong/neopetsbattledomeanalysis/caches"
	"github.com/darienchong/neopetsbattledomeanalysis/constants"
	"github.com/darienchong/neopetsbattledomeanalysis/helpers"
	"github.com/darienchong/neopetsbattledomeanalysis/models"
	"github.com/dustin/go-humanize"
)

type ArenaDataComparisonLogger struct {
	dropDataParser                 *DropDataParser
	empiricalDropRateEstimator     *EmpiricalDropRateEstimator
	arenaProfitStatisticsEstimator *ArenaProfitStatisticsEstimator
	dropRateEstimator              *DropRateEstimator
}

func GetPrefix(indentLevel int) string {
	return strings.Repeat("  ", indentLevel)
}

func GetDryChance(dropRate float64, trials int) float64 {
	return math.Pow(1-dropRate, float64(trials))
}

func NewArenaDataComparisonLogger() *ArenaDataComparisonLogger {
	instance := new(ArenaDataComparisonLogger)
	instance.dropDataParser = new(DropDataParser)
	instance.empiricalDropRateEstimator = new(EmpiricalDropRateEstimator)
	instance.arenaProfitStatisticsEstimator = new(ArenaProfitStatisticsEstimator)
	instance.dropRateEstimator = new(DropRateEstimator)
	return instance
}

func (comparisonLogger *ArenaDataComparisonLogger) getDropsByArena(dataFolderPath string) map[string][]*models.BattledomeDrops {
	files, err := helpers.GetFilesInFolder(dataFolderPath)
	if err != nil {
		slog.Error("Failed to get files in folder!")
		panic(err)
	}

	dropsByArena := map[string][]*models.BattledomeDrops{}
	for _, file := range files {
		drops, err := comparisonLogger.dropDataParser.Parse(constants.GetDropDataFilePath(file))
		if err != nil {
			slog.Error(fmt.Sprintf("Failed to parse drop data file (%s)", file))
			panic(err)
		}
		_, isKeyExists := dropsByArena[drops.Metadata.Arena]
		if !isKeyExists {
			dropsByArena[drops.Metadata.Arena] = []*models.BattledomeDrops{}
		}

		if !drops.Validate() {
			slog.Error(fmt.Sprintf("WARNING! The drops recorded in %s were not valid!", file))
		}

		dropsByArena[drops.Metadata.Arena] = append(dropsByArena[drops.Metadata.Arena], drops)
	}

	return dropsByArena
}

func (comparisonLogger *ArenaDataComparisonLogger) getDropDataAnalysisResultsByArena(dropsByArena map[string][]*models.BattledomeDrops) map[string]*models.DropDataAnalysisResult {
	return helpers.ToMap(
		helpers.ToSlice(dropsByArena),
		func(tuple helpers.Tuple) string {
			return tuple.Elements[0].(string)
		},
		func(tuple helpers.Tuple) *models.DropDataAnalysisResult {
			return comparisonLogger.empiricalDropRateEstimator.Analyse(
				helpers.Reduce(
					tuple.Elements[1].([]*models.BattledomeDrops),
					func(first *models.BattledomeDrops, second *models.BattledomeDrops) *models.BattledomeDrops {
						combined, err := first.Union(second)
						if err != nil {
							panic(err)
						}
						return combined
					},
				),
			)
		},
	)
}

func (comparisonLogger *ArenaDataComparisonLogger) getEstimatedProfitStatisticsByArena() map[string]*models.ArenaProfitStatistics {
	arenaStatistics, err := comparisonLogger.arenaProfitStatisticsEstimator.Estimate()
	if err != nil {
		panic(err)
	}
	return arenaStatistics
}

func (comparisonLogger *ArenaDataComparisonLogger) getEstimatedProfitsByArena(predictedDropRates []models.ItemDropRate, itemPriceCache *caches.ItemPriceCache) map[string][]*models.ItemProfit {
	return helpers.GroupBy(
		helpers.Map(
			predictedDropRates,
			func(dropRate models.ItemDropRate) models.ItemProfit {
				return models.ItemProfit{
					ItemDropRate:    dropRate,
					IndividualPrice: itemPriceCache.GetPrice(dropRate.ItemName),
				}
			},
		),
		func(itemProfit models.ItemProfit) string {
			return itemProfit.Arena
		},
	)
}

func (comparisonLogger *ArenaDataComparisonLogger) LogComparison(dataFolderPath string) {
	itemPriceCache := caches.GetItemPriceCacheInstance()
	defer itemPriceCache.Close()

	predictedDropRates, err := comparisonLogger.dropRateEstimator.Estimate()
	if err != nil {
		panic(err)
	}

	dropsByArena := comparisonLogger.getDropsByArena(dataFolderPath)
	analysisResultsByArena := comparisonLogger.getDropDataAnalysisResultsByArena(dropsByArena)

	estimatedProfitsByArena := comparisonLogger.getEstimatedProfitsByArena(predictedDropRates, itemPriceCache)
	estimatedProfitStatisticsByArena := comparisonLogger.getEstimatedProfitStatisticsByArena()

	orderedArenas := helpers.OrderByDescending(helpers.Keys(analysisResultsByArena), func(arena string) float64 {
		stats, err := analysisResultsByArena[arena].GetStatistics()
		if err != nil {
			panic(err)
		}
		return stats.GetMean(1)
	})

	for i, arena := range orderedArenas {
		totalItemCount := helpers.Sum(helpers.Map(dropsByArena[arena], func(drops *models.BattledomeDrops) int {
			return drops.GetTotalItemQuantity()
		}))

		predictedProfit := helpers.Sum(helpers.Map(estimatedProfitsByArena[arena], func(profit *models.ItemProfit) float64 { return profit.GetProfit() })) * constants.BATTLEDOME_DROPS_PER_DAY
		predictedStdev := estimatedProfitStatisticsByArena[arena].StandardDeviation * math.Sqrt(constants.BATTLEDOME_DROPS_PER_DAY)
		stats, err := analysisResultsByArena[arena].GetStatistics()
		if err != nil {
			panic(err)
		}
		actualProfit := stats.GetMean(constants.BATTLEDOME_DROPS_PER_DAY)
		actualStdev := stats.GetStandardDeviationSample(constants.BATTLEDOME_DROPS_PER_DAY)

		profitComparisonTable := helpers.NewNamedTable("Profit", []string{
			"Type",
			"Value",
		})
		profitComparisonTable.IsLastRowDistinct = true
		profitComparisonTable.AddRow([]string{
			"Predicted",
			fmt.Sprintf("%s ± %s NP", helpers.FormatFloat(predictedProfit), helpers.FormatFloat(predictedStdev)),
		})
		profitComparisonTable.AddRow([]string{
			"Actual",
			fmt.Sprintf("%s ± %s NP", helpers.FormatFloat(actualProfit), helpers.FormatFloat(actualStdev)),
		})
		profitComparisonTable.AddRow([]string{
			"Difference",
			fmt.Sprintf("%s NP", helpers.FormatFloat(actualProfit-predictedProfit)),
		})

		predictedProfitableItemsTable := helpers.NewNamedTable("Predicted", []string{
			"i",
			"ItemName",
			"Drop Rate",
			"Dry Chance",
			"Price",
			"Expectation",
			"%age",
		})
		predictedProfitableItems := helpers.OrderByDescending(estimatedProfitsByArena[arena], func(profit *models.ItemProfit) float64 {
			return profit.GetProfit()
		})
		for i, itemProfit := range predictedProfitableItems {
			if i > constants.NUMBER_OF_ITEMS_TO_PRINT-1 {
				break
			}
			predictedProfitableItemsTable.AddRow([]string{
				strconv.Itoa(i + 1),
				itemProfit.ItemName,
				helpers.FormatPercentage(itemProfit.DropRate) + "%",
				helpers.FormatPercentage(GetDryChance(itemProfit.DropRate, 30*constants.NUMBER_OF_ITEMS_TO_PRINT)) + "%",
				helpers.FormatFloat(itemProfit.IndividualPrice) + " NP",
				helpers.FormatFloat(itemProfit.GetProfit()) + " NP",
				helpers.FormatPercentage(itemProfit.GetProfit()*constants.BATTLEDOME_DROPS_PER_DAY/predictedProfit) + "%",
			})
		}

		actualProfitableItemsTable := helpers.NewNamedTable("Actual", []string{
			"ItemName",
			"Drop Rate",
			"Price",
			"Expectation",
			"%age",
		})
		actualProfitableItems := analysisResultsByArena[arena].GetItemsOrderedByProfit()
		for i, item := range actualProfitableItems {
			if i > constants.NUMBER_OF_ITEMS_TO_PRINT-1 {
				break
			}

			dropRate := item.GetDropRate(analysisResultsByArena[arena])
			itemPrice := itemPriceCache.GetPrice(item.Name)
			expectation := dropRate * itemPrice
			percentageProfit := expectation * constants.BATTLEDOME_DROPS_PER_DAY / actualProfit
			actualProfitableItemsTable.AddRow([]string{
				item.Name,
				helpers.FormatPercentage(dropRate) + "%",
				helpers.FormatFloat(itemPrice) + " NP",
				helpers.FormatFloat(expectation) + " NP",
				helpers.FormatPercentage(percentageProfit) + "%",
			})
		}

		predictedBrownCodestoneDropRateTable, actualBrownCodestoneDropRateTable := comparisonLogger.generateCodestoneDropRateTables(arena, analysisResultsByArena, predictedProfitableItems, actualProfitableItems, constants.BROWN_CODESTONES)
		predictedRedCodestoneDropRateTable, actualRedCodestoneDropRateTable := comparisonLogger.generateCodestoneDropRateTables(arena, analysisResultsByArena, predictedProfitableItems, actualProfitableItems, constants.RED_CODESTONES)

		slog.Info(fmt.Sprintf("%d. %s (%s items dropped)", i+1, arena, humanize.Comma(int64(totalItemCount))))
		profitComparisonTable.LogWithPrefix(GetPrefix(1))
		slog.Info("")

		slog.Info(GetPrefix(1) + fmt.Sprintf("Top %d most profitable items in %s", constants.NUMBER_OF_ITEMS_TO_PRINT, arena))
		predictedProfitableItemsTable.LogAlongsideWithPrefix(actualProfitableItemsTable, GetPrefix(1), GetPrefix(1))
		slog.Info("")

		slog.Info(GetPrefix(1) + fmt.Sprintf("Brown Codestone Drop Rates in %s", arena))
		predictedBrownCodestoneDropRateTable.LogAlongsideWithPrefix(actualBrownCodestoneDropRateTable, GetPrefix(1), GetPrefix(1))
		slog.Info("")

		slog.Info(GetPrefix(1) + fmt.Sprintf("Red Codestone Drop Rates in %s", arena))
		predictedRedCodestoneDropRateTable.LogAlongsideWithPrefix(actualRedCodestoneDropRateTable, GetPrefix(1), GetPrefix(1))
		slog.Info("")
	}
}

func (*ArenaDataComparisonLogger) generateCodestoneDropRateTables(arena string, analysisResultsByArena map[string]*models.DropDataAnalysisResult, predictedProfitableItems []*models.ItemProfit, actualProfitableItems []*models.BattledomeItem, codestoneList []string) (*helpers.Table, *helpers.Table) {
	predictedCodestoneDropRateTable := helpers.NewNamedTable("Predicted", []string{
		"Item Name",
		"Drop Rate",
	})
	predictedCodestoneDropRateTable.IsLastRowDistinct = true

	predictedCodestoneDropRates := helpers.ToMap(helpers.Filter(predictedProfitableItems, func(itemProfit *models.ItemProfit) bool {
		return slices.Contains(codestoneList, itemProfit.ItemName)
	}), func(itemProfit *models.ItemProfit) string {
		return itemProfit.ItemName
	}, func(itemProfit *models.ItemProfit) *models.ItemProfit {
		return itemProfit
	})

	totalPredictedCodestoneDropRate := 0.0
	for _, codestoneName := range codestoneList {
		itemProfit := predictedCodestoneDropRates[codestoneName]
		predictedCodestoneDropRateTable.AddRow([]string{
			itemProfit.ItemName,
			helpers.FormatPercentage(itemProfit.DropRate) + "%",
		})
		totalPredictedCodestoneDropRate += itemProfit.DropRate
	}

	predictedCodestoneDropRateTable.AddRow([]string{
		"Total",
		helpers.FormatPercentage(totalPredictedCodestoneDropRate) + "%",
	})

	actualCodestoneDropRateTable := helpers.NewNamedTable("Actual", []string{
		"Item Name",
		"Drop Rate",
	})
	actualCodestoneDropRateTable.IsLastRowDistinct = true

	actualCodestoneDropRates := helpers.ToMap(helpers.Filter(actualProfitableItems, func(item *models.BattledomeItem) bool {
		return slices.Contains(codestoneList, item.Name)
	}), func(item *models.BattledomeItem) string {
		return item.Name
	}, func(item *models.BattledomeItem) *models.BattledomeItem {
		return item
	})

	totalActualCodestoneDropRate := 0.0
	for _, codestoneName := range codestoneList {
		item, isInData := actualCodestoneDropRates[codestoneName]
		if !isInData {
			actualCodestoneDropRateTable.AddRow([]string{
				codestoneName,
				helpers.FormatPercentage(0) + "%",
			})
		} else {
			dropRate := item.GetDropRate(analysisResultsByArena[arena])
			actualCodestoneDropRateTable.AddRow([]string{
				item.Name,
				helpers.FormatPercentage(dropRate) + "%",
			})
			totalActualCodestoneDropRate += dropRate
		}
	}

	actualCodestoneDropRateTable.AddRow([]string{
		"Total",
		helpers.FormatPercentage(totalActualCodestoneDropRate) + "%",
	})
	return predictedCodestoneDropRateTable, actualCodestoneDropRateTable
}
