package services

import (
	"fmt"
	"log/slog"
	"math"
	"slices"
	"strconv"
	"strings"

	"github.com/darienchong/neopets-battledome-analysis/caches"
	"github.com/darienchong/neopets-battledome-analysis/constants"
	"github.com/darienchong/neopets-battledome-analysis/helpers"
	"github.com/darienchong/neopets-battledome-analysis/models"
	"github.com/dustin/go-humanize"
)

func getPrefix(indentLevel int) string {
	return strings.Repeat("  ", indentLevel)
}

func getDryChance(dropRate float64, trials int) float64 {
	return math.Pow(1-dropRate, float64(trials))
}

type ArenaDataComparisonLogger struct {
	DropDataParser                 *DropDataParser
	EmpiricalDropRateEstimator     *EmpiricalDropRateEstimator
	ArenaProfitStatisticsEstimator *ArenaProfitStatisticsEstimator
	DropRateService                *DropRateService
}

func NewArenaDataComparisonLogger() *ArenaDataComparisonLogger {
	return &ArenaDataComparisonLogger{
		DropDataParser:                 NewDropDataParser(),
		EmpiricalDropRateEstimator:     NewEmpiricalDropRateEstimator(),
		ArenaProfitStatisticsEstimator: NewArenaProfitStatisticsEstimator(),
		DropRateService:                NewDropRateService(),
	}
}

func (comparisonLogger *ArenaDataComparisonLogger) getDropsByArena(dataFolderPath string) (map[string][]*models.BattledomeDrops, error) {
	files, err := helpers.GetFilesInFolder(dataFolderPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get files in %s: %w", dataFolderPath, err)
	}

	dropsByArena := map[string][]*models.BattledomeDrops{}
	for _, file := range files {
		drops, err := comparisonLogger.DropDataParser.Parse(constants.GetDropDataFilePath(file))
		if err != nil {
			return nil, fmt.Errorf("failed to parse drop data file %s: %w", file, err)
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

	return dropsByArena, nil
}

// TODO: Add error handling for this
func (comparisonLogger *ArenaDataComparisonLogger) getDropDataAnalysisResultsByArena(dropsByArena map[string][]*models.BattledomeDrops) map[string]*models.DropDataAnalysisResult {
	return helpers.ToMap(
		helpers.ToSlice(dropsByArena),
		func(tuple helpers.Tuple) string {
			return tuple.Elements[0].(string)
		},
		func(tuple helpers.Tuple) *models.DropDataAnalysisResult {
			return comparisonLogger.EmpiricalDropRateEstimator.Analyse(
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

func (comparisonLogger *ArenaDataComparisonLogger) getPredictedProfitStatisticsByArena() (map[string]*models.ArenaProfitStatistics, error) {
	arenaStatistics, err := comparisonLogger.ArenaProfitStatisticsEstimator.Estimate()
	if err != nil {
		return nil, fmt.Errorf("failed to generated estimated profit statistics: %w", err)
	}
	return arenaStatistics, nil
}

func (comparisonLogger *ArenaDataComparisonLogger) getPredictedProfitsByArena(itemPriceCache *caches.ItemPriceCache) (map[string][]*models.ItemProfit, error) {
	estimatedProfitsByArena := map[string][]*models.ItemProfit{}
	for _, arena := range constants.ARENAS {
		predictedDropRates, err := comparisonLogger.DropRateService.GetPredictedDropRates(arena)
		if err != nil {
			return nil, err
		}

		estimatedProfitsByArena[arena] = []*models.ItemProfit{}
		for _, dropRate := range predictedDropRates {
			estimatedProfitsByArena[arena] = append(estimatedProfitsByArena[arena], &models.ItemProfit{
				ItemDropRate:    *dropRate,
				IndividualPrice: itemPriceCache.GetPrice(dropRate.ItemName),
			})
		}
	}

	return estimatedProfitsByArena, nil
}

func (comparisonLogger *ArenaDataComparisonLogger) LogComparison(dataFolderPath string) error {
	itemPriceCache, err := caches.GetItemPriceCacheInstance()
	if err != nil {
		return fmt.Errorf("ArenaDataComparisonLogger.LogComparison: %w", err)
	}
	defer itemPriceCache.Close()

	dropsByArena, err := comparisonLogger.getDropsByArena(dataFolderPath)
	if err != nil {
		return fmt.Errorf("ArenaDataComparisonLogger.LogComparison: %w", err)
	}

	analysisResultsByArena := comparisonLogger.getDropDataAnalysisResultsByArena(dropsByArena)
	estimatedProfitsByArena, err := comparisonLogger.getPredictedProfitsByArena(itemPriceCache)
	if err != nil {
		return fmt.Errorf("ArenaDataComparisonLogger.LogComparison: %w", err)
	}
	estimatedProfitStatisticsByArena, err := comparisonLogger.getPredictedProfitStatisticsByArena()
	if err != nil {
		return fmt.Errorf("ArenaDataComparisonLogger.LogComparison: %w", err)
	}

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
			return fmt.Errorf("ArenaDataComparisonLogger.LogComparison: %w", err)
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
				helpers.FormatPercentage(getDryChance(itemProfit.DropRate, 30*constants.NUMBER_OF_ITEMS_TO_PRINT)) + "%",
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
		profitComparisonTable.LogWithPrefix(getPrefix(1))
		slog.Info("")

		slog.Info(getPrefix(1) + fmt.Sprintf("Top %d most profitable items in %s", constants.NUMBER_OF_ITEMS_TO_PRINT, arena))
		predictedProfitableItemsTable.LogAlongsideWithPrefix(actualProfitableItemsTable, getPrefix(1), getPrefix(1))
		slog.Info("")

		slog.Info(getPrefix(1) + fmt.Sprintf("Brown Codestone Drop Rates in %s", arena))
		predictedBrownCodestoneDropRateTable.LogAlongsideWithPrefix(actualBrownCodestoneDropRateTable, getPrefix(1), getPrefix(1))
		slog.Info("")

		slog.Info(getPrefix(1) + fmt.Sprintf("Red Codestone Drop Rates in %s", arena))
		predictedRedCodestoneDropRateTable.LogAlongsideWithPrefix(actualRedCodestoneDropRateTable, getPrefix(1), getPrefix(1))
		slog.Info("")
	}

	return nil
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
