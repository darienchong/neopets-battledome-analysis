package viewers

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
	"github.com/darienchong/neopets-battledome-analysis/services"
	"github.com/palantir/stacktrace"
)

type DataComparisonViewer struct {
	BattledomeItemsService *services.BattledomeItemsService
	DataComparisonService  *services.DataComparisonService
	StatisticsService      *services.StatisticsService
}

func NewDataComparisonViewer() *DataComparisonViewer {
	return &DataComparisonViewer{
		BattledomeItemsService: services.NewBattledomeItemsService(),
		DataComparisonService:  services.NewDataComparisonService(),
		StatisticsService:      services.NewStatisticsService(),
	}
}

func getDryChance(dropRate float64, trials int) float64 {
	return math.Pow(1-dropRate, float64(trials))
}

func isCodestone(itemName models.ItemName) bool {
	return slices.Contains(constants.BROWN_CODESTONES, string(itemName)) || slices.Contains(constants.RED_CODESTONES, string(itemName))
}

func (viewer *DataComparisonViewer) generateProfitableItemsTable(data models.NormalisedBattledomeItems, isRealData bool) (*helpers.Table, error) {
	dataCopy := models.NormalisedBattledomeItems{}
	for k, v := range data {
		dataCopy[k] = v.Copy()
	}

	itemPriceCache, err := caches.GetItemPriceCacheInstance()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get item price cache instance")
	}
	defer itemPriceCache.Close()

	headers := helpers.When(isRealData, []string{
		"i",
		"Item Name",
		"Drop Rate",
		// Don't include Dry Chance in real data
		"Price",
		"Expectation",
		"%",
	}, []string{
		"i",
		"Item Name",
		"Drop Rate",
		"Dry Chance",
		"Price",
		"Expectation",
		"%",
	})
	table := helpers.NewNamedTable(helpers.When(isRealData, "Actual", "Predicted"), headers)

	_, err = data.GetMetadata()
	if err != nil {
		// Means there isn't any data to render
		return table, nil
	}

	predictedProfit := helpers.Sum(helpers.Map(helpers.Values(dataCopy), func(item *models.BattledomeItem) float64 {
		return item.GetProfit(itemPriceCache)
	}))
	profitableItems := helpers.OrderByDescending(helpers.Values(dataCopy), func(item *models.BattledomeItem) float64 {
		return item.GetProfit(itemPriceCache)
	})

	runningIndex := 1
	for _, item := range profitableItems {
		if runningIndex > constants.NUMBER_OF_ITEMS_TO_PRINT {
			break
		}

		if item.Name == "nothing" {
			continue
		}

		itemDropRate := item.GetDropRate(data)
		expectedItemProfit := itemDropRate * itemPriceCache.GetPrice(string(item.Name)) * constants.BATTLEDOME_DROPS_PER_DAY
		itemDropRateLeftBound, itemDropRateRightBound, err := viewer.StatisticsService.ClopperPearsonInterval(int(item.Quantity), data.GetTotalItemQuantity(), constants.SIGNIFICANCE_LEVEL)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to generate drop rate confidence interval")
		}
		row := helpers.When(isRealData,
			[]string{
				strconv.Itoa(runningIndex),
				string(item.Name),
				fmt.Sprintf("%s ∈ %s%%", helpers.FormatPercentage(itemDropRate), helpers.FormatPercentageRange("[%s, %s]", itemDropRateLeftBound, itemDropRateRightBound)),
				// Don't include dry chance in real data
				helpers.FormatFloat(itemPriceCache.GetPrice(string(item.Name))) + " NP",
				helpers.FormatFloat(expectedItemProfit) + " NP",
				helpers.FormatPercentage(item.GetProfit(itemPriceCache)/predictedProfit) + "%",
			},
			[]string{
				strconv.Itoa(runningIndex),
				string(item.Name),
				helpers.FormatPercentage(item.GetDropRate(data)) + "%",
				helpers.FormatPercentage(getDryChance(item.GetDropRate(data), 30*constants.BATTLEDOME_DROPS_PER_DAY)) + "%",
				helpers.FormatFloat(itemPriceCache.GetPrice(string(item.Name))) + " NP",
				helpers.FormatFloat(expectedItemProfit) + " NP",
				helpers.FormatPercentage(item.GetProfit(itemPriceCache)/predictedProfit) + "%",
			},
		)
		table.AddRow(row)

		runningIndex++
	}

	return table, nil
}

func (viewer *DataComparisonViewer) generateCodestoneDropRatesTable(realData models.NormalisedBattledomeItems, generatedData models.NormalisedBattledomeItems, codestoneList []string /* Keep as string to prevent circular dependency*/) (*helpers.Table, error) {
	var tableName string
	if slices.Contains(codestoneList, constants.BROWN_CODESTONES[0]) {
		tableName = "Brown Codestone Drop Rates"
	} else if slices.Contains(codestoneList, constants.RED_CODESTONES[0]) {
		tableName = "Red Codestone Drop Rates"
	} else {
		tableName = "Codestone Drop Rates"
	}

	table := helpers.NewNamedTable(tableName, []string{
		"Item Name",
		"Predicted",
		"Real",
	})
	table.IsLastRowDistinct = true

	realCodestones := helpers.ToMap(helpers.Filter(helpers.Values(realData), func(itemProfit *models.BattledomeItem) bool {
		return slices.Contains(codestoneList, string(itemProfit.Name))
	}), func(itemProfit *models.BattledomeItem) models.ItemName {
		return itemProfit.Name
	}, func(itemProfit *models.BattledomeItem) *models.BattledomeItem {
		return itemProfit
	})
	generatedCodestones := helpers.ToMap(helpers.Filter(helpers.Values(generatedData), func(itemProfit *models.BattledomeItem) bool {
		return slices.Contains(codestoneList, string(itemProfit.Name))
	}), func(itemProfit *models.BattledomeItem) models.ItemName {
		return itemProfit.Name
	}, func(itemProfit *models.BattledomeItem) *models.BattledomeItem {
		return itemProfit
	})

	realTotalCodestoneCount := helpers.Sum(helpers.Map(helpers.Values(realCodestones), func(item *models.BattledomeItem) float64 {
		return float64(item.Quantity)
	}))
	generatedTotalCodestoneDropRate := helpers.Sum(helpers.Map(helpers.Values(generatedCodestones), func(profit *models.BattledomeItem) float64 {
		return profit.GetDropRate(generatedData)
	}))

	slices.Sort(codestoneList)
	for _, codestoneName := range codestoneList {
		realCodestones, existsInReal := realCodestones[models.ItemName(codestoneName)]
		generatedCodestones, existsInGenerated := generatedCodestones[models.ItemName(codestoneName)]
		var realDropRate float64 = 0.0
		var realMinDropRate float64 = 0.0
		var realMaxDropRate float64 = 0.0
		var err error

		var generatedCodestoneDropRate float64 = 0.0

		if existsInGenerated {
			generatedCodestoneDropRate = generatedCodestones.GetDropRate(generatedData)
		}

		if existsInReal {
			realDropRate = realCodestones.GetDropRate(realData)
			realMinDropRate, realMaxDropRate, err = viewer.StatisticsService.ClopperPearsonInterval(int(realCodestones.Quantity), realData.GetTotalItemQuantity(), constants.SIGNIFICANCE_LEVEL)
			if err != nil {
				return nil, stacktrace.Propagate(err, "failed to generate drop rate confidence interval")
			}
		}

		table.AddRow([]string{
			codestoneName,
			helpers.FormatPercentage(generatedCodestoneDropRate) + "%",
			fmt.Sprintf("%s ∈ %s%%", helpers.FormatPercentage(realDropRate), helpers.FormatPercentageRange("[%s, %s]", realMinDropRate, realMaxDropRate)),
		})
	}

	realTotalMinDropRate, realTotalMaxDropRate, err := viewer.StatisticsService.ClopperPearsonInterval(int(realTotalCodestoneCount), realData.GetTotalItemQuantity(), constants.SIGNIFICANCE_LEVEL)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to generate drop rate confidence interval")
	}

	table.AddRow([]string{
		"Sum",
		helpers.FormatPercentage(generatedTotalCodestoneDropRate) + "%",
		helpers.When(realTotalMinDropRate == realTotalMaxDropRate, helpers.FormatPercentage(realTotalMinDropRate)+"%", fmt.Sprintf("[%s, %s]%%", helpers.FormatPercentage(realTotalMinDropRate), helpers.FormatPercentage(realTotalMaxDropRate))),
	})

	return table, nil
}

func (viewer *DataComparisonViewer) ViewChallengerComparison(realData models.NormalisedBattledomeItems, generatedData models.NormalisedBattledomeItems) ([]string, error) {
	metadata, err := realData.GetMetadata()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get metadata")
	}

	profitComparisonTable := helpers.NewNamedTable(fmt.Sprintf("%s %s in %s", metadata.Difficulty, metadata.Challenger, metadata.Arena), []string{
		"Type",
		"Value",
	})
	profitComparisonTable.IsLastRowDistinct = true

	generatedMeanProfit, err := generatedData.GetMeanDropsProfit()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get generated mean drops profit")
	}
	generatedProfitStdev, err := generatedData.GetDropsProfitStdev()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get generated drop profit stdev")
	}

	realMeanProfit, err := realData.GetMeanDropsProfit()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get real mean drops profit")
	}
	realProfitStdev, err := realData.GetDropsProfitStdev()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get real drop profit stdev")
	}

	profitComparisonTable.AddRow([]string{
		"Predicted",
		fmt.Sprintf("%s ± %s NP", helpers.FormatFloat(generatedMeanProfit), helpers.FormatFloat(generatedProfitStdev)),
	})
	profitComparisonTable.AddRow([]string{
		"Actual",
		fmt.Sprintf("%s ± %s NP", helpers.FormatFloat(realMeanProfit), helpers.FormatFloat(realProfitStdev)),
	})
	profitComparisonTable.AddRow([]string{
		"Difference",
		fmt.Sprintf("%s NP", helpers.FormatFloat(realMeanProfit-generatedMeanProfit)),
	})

	realProfitableItemsTable, err := viewer.generateProfitableItemsTable(realData, true)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to generate profitable items table for real data")
	}
	generatedProfitableItemsTable, err := viewer.generateProfitableItemsTable(generatedData, false)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to generate profitable items table for generated data")
	}

	brownCodestoneDropRatesTable, err := viewer.generateCodestoneDropRatesTable(realData, generatedData, constants.BROWN_CODESTONES)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to generate brown codestones drop rate table for real and generated data")
	}

	redCodestoneDropRatesTable, err := viewer.generateCodestoneDropRatesTable(realData, generatedData, constants.RED_CODESTONES)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to generate red codestones drop rate table for real and generated data")
	}

	arenaSpecificDropsTable, err := viewer.generateArenaSpecificDropsTable(realData, generatedData)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to generate arena-specific drop rate table for real and generated data")
	}

	challengerSpecificDropsTable, err := viewer.generateChallengerSpecificDropsTable(realData, generatedData)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed ot generate challenger-specific drop rate table for real and generated data")
	}

	tableSeparator := "  "

	lines := []string{}
	lines = append(lines, profitComparisonTable.GetLines()...)
	lines = append(lines, "\n")
	lines = append(lines, fmt.Sprintf("Top %d most profitable items", constants.NUMBER_OF_ITEMS_TO_PRINT))
	lines = append(lines, generatedProfitableItemsTable.GetLinesWith(tableSeparator, realProfitableItemsTable)...)
	lines = append(lines, "\n")
	lines = append(lines, brownCodestoneDropRatesTable.GetLinesWith(tableSeparator, redCodestoneDropRatesTable)...)
	lines = append(lines, "\n")
	lines = append(lines, arenaSpecificDropsTable.GetLinesWith(tableSeparator, challengerSpecificDropsTable)...)

	return lines, nil
}

func isArenaSpecificDrop(itemName models.ItemName, items models.NormalisedBattledomeItems) bool {
	_, exists := items[itemName]
	return exists
}

func (viewer *DataComparisonViewer) generateArenaSpecificDropsTable(realData models.NormalisedBattledomeItems, generatedData models.NormalisedBattledomeItems) (*helpers.Table, error) {
	itemPriceCache, err := caches.GetItemPriceCacheInstance()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get item price cache instance")
	}
	defer itemPriceCache.Close()

	table := helpers.NewNamedTable("Arena-specific drops", []string{
		"i",
		"Item Name",
		"Drop Rate",
		"Price",
		"Expectation",
		"%",
	})
	table.IsLastRowDistinct = true

	realItems := helpers.Filter(helpers.Values(realData), func(item *models.BattledomeItem) bool {
		return isArenaSpecificDrop(item.Name, generatedData)
	})
	orderedRealItems := helpers.OrderByDescending(realItems, func(item *models.BattledomeItem) float64 {
		return float64(item.Quantity) * itemPriceCache.GetPrice(string(item.Name))
	})
	totalExpectation := helpers.Sum(helpers.Map(helpers.Values(realData), func(item *models.BattledomeItem) float64 {
		return item.GetDropRate(realData) * itemPriceCache.GetPrice(string(item.Name)) * constants.BATTLEDOME_DROPS_PER_DAY
	}))
	arenaSpecificDropsExpectation := helpers.Sum(helpers.Map(realItems, func(item *models.BattledomeItem) float64 {
		return item.GetDropRate(realData) * itemPriceCache.GetPrice(string(item.Name)) * constants.BATTLEDOME_DROPS_PER_DAY
	}))

	totalArenaItemCount := helpers.Sum(helpers.Map(realItems, func(item *models.BattledomeItem) int {
		if item.Name == "nothing" {
			return 0
		}

		return int(item.Quantity)
	}))
	totalDropRateLeftBound, totalDropRateRightBound, err := viewer.StatisticsService.ClopperPearsonInterval(totalArenaItemCount, realData.GetTotalItemQuantity(), constants.SIGNIFICANCE_LEVEL)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to generate total drop rate bounds")
	}

	for i, item := range orderedRealItems {
		if i > constants.NUMBER_OF_ITEMS_TO_PRINT-1 {
			break
		}

		itemDropRateLeftBound, itemDropRateRightBound, err := viewer.StatisticsService.ClopperPearsonInterval(int(item.Quantity), realData.GetTotalItemQuantity(), constants.SIGNIFICANCE_LEVEL)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to generate item drop bounds")
		}
		itemDropRate := item.GetDropRate(realData)
		itemPrice := itemPriceCache.GetPrice(string(item.Name))
		itemExpectation := itemDropRate * itemPrice * constants.BATTLEDOME_DROPS_PER_DAY

		table.AddRow([]string{
			strconv.Itoa(i + 1),
			string(item.Name),
			helpers.When(itemDropRateLeftBound == itemDropRateRightBound, fmt.Sprintf("%s%%", helpers.FormatPercentage(itemDropRateLeftBound)), fmt.Sprintf("[%s, %s]%%", helpers.FormatPercentage(itemDropRateLeftBound), helpers.FormatPercentage(itemDropRateRightBound))),
			helpers.FormatFloat(itemPrice) + " NP",
			helpers.FormatFloat(itemExpectation) + " NP",
			helpers.FormatPercentage(itemExpectation/totalExpectation) + "%",
		})
	}

	table.AddRow([]string{
		"",
		"Total",
		helpers.When(totalDropRateLeftBound == totalDropRateRightBound,
			fmt.Sprintf("%s%%", helpers.FormatPercentage(totalDropRateLeftBound)),
			fmt.Sprintf("[%s, %s]%%", helpers.FormatPercentage(totalDropRateLeftBound), helpers.FormatPercentage(totalDropRateRightBound))),
		"",
		helpers.FormatFloat(arenaSpecificDropsExpectation) + " NP",
		helpers.FormatPercentage(arenaSpecificDropsExpectation/totalExpectation) + "%",
	})

	return table, nil
}

func (viewer *DataComparisonViewer) generateChallengerSpecificDropsTable(realData models.NormalisedBattledomeItems, generatedData models.NormalisedBattledomeItems) (*helpers.Table, error) {
	itemPriceCache, err := caches.GetItemPriceCacheInstance()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get item price cache instance")
	}
	defer itemPriceCache.Close()

	table := helpers.NewNamedTable("Challenger-specific drops", []string{
		"i",
		"Item Name",
		"Drop Rate",
		"Price",
		"Expectation",
		"%",
	})
	table.IsLastRowDistinct = true

	realItems := helpers.Filter(helpers.Values(realData), func(item *models.BattledomeItem) bool {
		return !isArenaSpecificDrop(item.Name, generatedData)
	})
	orderedRealItems := helpers.OrderByDescending(realItems, func(item *models.BattledomeItem) float64 {
		return float64(item.Quantity) * itemPriceCache.GetPrice(string(item.Name))
	})
	totalExpectation := helpers.Sum(helpers.Map(helpers.Values(realData), func(item *models.BattledomeItem) float64 {
		return item.GetDropRate(realData) * itemPriceCache.GetPrice(string(item.Name)) * constants.BATTLEDOME_DROPS_PER_DAY
	}))
	challengerSpecificDropsExpectation := helpers.Sum(helpers.Map(realItems, func(item *models.BattledomeItem) float64 {
		return item.GetDropRate(realData) * itemPriceCache.GetPrice(string(item.Name)) * constants.BATTLEDOME_DROPS_PER_DAY
	}))

	totalChallengerItemCount := helpers.Sum(helpers.Map(realItems, func(item *models.BattledomeItem) int {
		if item.Name == "nothing" {
			return 0
		}

		return int(item.Quantity)
	}))
	totalDropRateLeftBound, totalDropRateRightBound, err := viewer.StatisticsService.ClopperPearsonInterval(totalChallengerItemCount, realData.GetTotalItemQuantity(), constants.SIGNIFICANCE_LEVEL)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to generate total drop rate bounds")
	}

	for i, item := range orderedRealItems {
		if i > constants.NUMBER_OF_ITEMS_TO_PRINT-1 {
			break
		}

		itemDropRateLeftBound, itemDropRateRightBound, err := viewer.StatisticsService.ClopperPearsonInterval(int(item.Quantity), realData.GetTotalItemQuantity(), constants.SIGNIFICANCE_LEVEL)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to generate item drop bounds")
		}
		itemDropRate := item.GetDropRate(realData)
		itemPrice := itemPriceCache.GetPrice(string(item.Name))
		itemExpectation := itemDropRate * itemPrice * constants.BATTLEDOME_DROPS_PER_DAY

		table.AddRow([]string{
			strconv.Itoa(i + 1),
			string(item.Name),
			helpers.When(itemDropRateLeftBound == itemDropRateRightBound, fmt.Sprintf("%s%%", helpers.FormatPercentage(itemDropRateLeftBound)), fmt.Sprintf("[%s, %s]%%", helpers.FormatPercentage(itemDropRateLeftBound), helpers.FormatPercentage(itemDropRateRightBound))),
			helpers.FormatFloat(itemPrice) + " NP",
			helpers.FormatFloat(itemExpectation) + " NP",
			helpers.FormatPercentage(itemExpectation/totalExpectation) + "%",
		})
	}

	table.AddRow([]string{
		"",
		"Total",
		helpers.When(totalDropRateLeftBound == totalDropRateRightBound,
			fmt.Sprintf("%s%%", helpers.FormatPercentage(totalDropRateLeftBound)),
			fmt.Sprintf("[%s, %s]%%", helpers.FormatPercentage(totalDropRateLeftBound), helpers.FormatPercentage(totalDropRateRightBound))),
		"",
		helpers.FormatFloat(challengerSpecificDropsExpectation) + " NP",
		helpers.FormatPercentage(challengerSpecificDropsExpectation/totalExpectation) + "%",
	})

	return table, nil
}

func (viewer *DataComparisonViewer) ViewChallengerComparisons(challengerItems []models.NormalisedBattledomeItems) ([]string, error) {
	challengerItems = helpers.OrderByDescending(challengerItems, func(items models.NormalisedBattledomeItems) float64 {
		meanDropsProfit, err := items.GetMeanDropsProfit()
		if err != nil {
			return 0.0
		}
		return meanDropsProfit
	})

	profitComparisonTable := helpers.NewNamedTable("Challenger profit comparison", []string{
		"i",
		"Arena",
		"Challenger",
		"Difficulty",
		"Samples",
		"Actual Profit",
		"Predicted Profit",
	})

	arenaAndChallengerDropRateTable := helpers.NewNamedTable("Arena/challenger-specific drop rate comparison", []string{
		"i",
		"Arena",
		"Challenger",
		"Difficulty",
		"Arena Drop Rate",
		"Challenger Drop Rate",
	})

	for i, items := range challengerItems {
		metadata, err := items.GetMetadata()
		if err != nil {
			return nil, helpers.PropagateWithSerialisedValue(err, "failed to get metadata from %s", "failed to get metadata from items; additionally encountered an error when trying to serialise the items for logging: %s", items)
		}

		actualProfit, err := items.GetMeanDropsProfit()
		if err != nil {
			return nil, helpers.PropagateWithSerialisedValue(err, "failed to get mean drops profit from %s", "failed to get mean drops profit from items; additionally encountered an error when trying to serialise the items for logging: %s", items)
		}
		actualProfitLeftBound, actualProfitRightBound, err := items.GetProfitConfidenceInterval()
		if err != nil {
			return nil, helpers.PropagateWithSerialisedValue(err, "failed to get profit confidence interval from %s", "failed to get profit confidence interval from items; additionally encountered an error when trying to serialise the items for logging: %s", items)
		}

		generatedItems, err := viewer.BattledomeItemsService.GenerateDropsByArena(metadata.Arena)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to generate drops by arena for \"%s\"", metadata.Arena)
		}
		generatedProfit, err := generatedItems.GetMeanDropsProfit()
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to get mean drops profit for generated items")
		}
		generatedProfitLeftBound, generatedProfitRightBound, err := generatedItems.GetProfitConfidenceInterval()
		if err != nil {
			return nil, helpers.PropagateWithSerialisedValue(err, "failed to get profit confidence interval from %s", "failed to get profit confidence interval from items; additionally encountered an error when trying to serialise the items for logging: %s", items)
		}

		profitComparisonTable.AddRow([]string{
			strconv.Itoa(i + 1),
			string(metadata.Arena),
			string(metadata.Challenger),
			string(metadata.Difficulty),
			helpers.FormatInt(items.GetTotalItemQuantity()),
			fmt.Sprintf("%s ∈ %s NP", helpers.FormatFloat(actualProfit), helpers.FormatFloatRange("[%s, %s]", actualProfitLeftBound, actualProfitRightBound)),
			fmt.Sprintf("%s ∈ %s NP", helpers.FormatFloat(generatedProfit), helpers.FormatFloatRange("[%s, %s]", generatedProfitLeftBound, generatedProfitRightBound)),
		})

		arenaDropsCount := helpers.Sum(helpers.Map(
			helpers.Values(items),
			func(item *models.BattledomeItem) int {
				if item.Name == "nothing" {
					return 0
				}

				_, exists := generatedItems[item.Name]
				if !exists {
					return 0
				}

				return int(item.Quantity)
			},
		))

		challengerDropsCount := helpers.Sum(helpers.Map(
			helpers.Values(items),
			func(item *models.BattledomeItem) int {
				if item.Name == "nothing" {
					return 0
				}

				_, exists := generatedItems[item.Name]
				if exists {
					return 0
				}

				return int(item.Quantity)
			},
		))

		arenaAndChallengerDropRateTable.AddRow([]string{
			strconv.Itoa(i + 1),
			string(metadata.Arena),
			string(metadata.Challenger),
			string(metadata.Difficulty),
			helpers.FormatPercentage(float64(arenaDropsCount)/float64(arenaDropsCount+challengerDropsCount)) + "%",
			helpers.FormatPercentage(float64(challengerDropsCount)/float64(arenaDropsCount+challengerDropsCount)) + "%",
		})
	}

	lines := profitComparisonTable.GetLines()
	lines = append(lines, "\n")
	lines = append(lines, arenaAndChallengerDropRateTable.GetLines()...)
	return lines, nil
}

func (viewer *DataComparisonViewer) ViewArenaComparison(realData models.NormalisedBattledomeItems, generatedData models.NormalisedBattledomeItems) ([]string, error) {
	metadata, err := generatedData.GetMetadata()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get metadata")
	}

	profitComparisonTable := helpers.NewNamedTable(fmt.Sprintf("Profit in %s", metadata.Arena), []string{
		"Type",
		"Value",
	})
	profitComparisonTable.IsLastRowDistinct = true

	generatedMeanProfit, err := generatedData.GetMeanDropsProfit()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get mean drops profit")
	}
	generatedProfitLeftBound, generatedProfitRightBound, err := generatedData.GetProfitConfidenceInterval()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get profit confidence interval")
	}

	realDataCopy := models.NormalisedBattledomeItems{}
	for k, v := range realData {
		if _, exists := generatedData[k]; constants.SHOULD_IGNORE_CHALLENGER_DROPS_IN_ARENA_COMPARISON && !exists {
			continue
		}

		realDataCopy[k] = v
	}

	realMeanProfit, err := realDataCopy.GetMeanDropsProfit()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get mean drops profit")
	}
	realProfitLeftBound, realProfitRightBound, err := realDataCopy.GetProfitConfidenceInterval()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get profit confidence interval")
	}

	profitComparisonTable.AddRow([]string{
		"Predicted",
		fmt.Sprintf("%s ∈ %s NP", helpers.FormatFloat(generatedMeanProfit), helpers.FormatFloatRange("[%s, %s]", generatedProfitLeftBound, generatedProfitRightBound)),
	})
	profitComparisonTable.AddRow([]string{
		"Actual",
		fmt.Sprintf("%s ∈ %s NP", helpers.FormatFloat(realMeanProfit), helpers.FormatFloatRange("[%s, %s]", realProfitLeftBound, realProfitRightBound)),
	})
	profitComparisonTable.AddRow([]string{
		"Difference",
		fmt.Sprintf("%s NP", helpers.FormatFloat(realMeanProfit-generatedMeanProfit)),
	})

	realProfitableItemsTable, err := viewer.generateProfitableItemsTable(realDataCopy, true)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to generate profitable items table")
	}
	generatedProfitableItemsTable, err := viewer.generateProfitableItemsTable(generatedData, false)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to generate profitable items table")
	}

	brownCodestoneDropRatesTable, err := viewer.generateCodestoneDropRatesTable(realDataCopy, generatedData, constants.BROWN_CODESTONES)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to generate brown codestones drop rates table")
	}

	redCodestoneDropRatesTable, err := viewer.generateCodestoneDropRatesTable(realDataCopy, generatedData, constants.RED_CODESTONES)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to generate red codestones drop rates table")
	}

	tableSeparator := "\t"

	lines := []string{}
	lines = append(lines, profitComparisonTable.GetLines()...)
	lines = append(lines, "\n")
	lines = append(lines, fmt.Sprintf("Top %d most profitable items in %s", constants.NUMBER_OF_ITEMS_TO_PRINT, metadata.Arena))
	lines = append(lines, generatedProfitableItemsTable.GetLinesWith(tableSeparator, realProfitableItemsTable)...)
	lines = append(lines, "\n")
	lines = append(lines, fmt.Sprintf("Codestone drop rates in %s", metadata.Arena))
	lines = append(lines, brownCodestoneDropRatesTable.GetLinesWith(tableSeparator, redCodestoneDropRatesTable)...)
	return lines, nil
}

func (viewer *DataComparisonViewer) ViewBriefArenaComparisons(realData map[models.Arena]models.NormalisedBattledomeItems, generatedData map[models.Arena]models.NormalisedBattledomeItems) ([]string, error) {
	orderedArenas := helpers.OrderByDescending(constants.ARENAS, func(arena string) float64 {
		normalisedItems, exists := realData[models.Arena(arena)]
		if !exists || normalisedItems.GetTotalItemQuantity() == 0 {
			return 0.0
		}

		profit, err := normalisedItems.GetMeanDropsProfit()
		if err != nil {
			return 0.0
		}

		return profit
	})

	table := helpers.NewNamedTable("Profit", []string{
		"i",
		"Arena",
		"Predicted",
		"Actual",
	})

	for i, arena := range orderedArenas {
		slog.Debug(fmt.Sprintf("Generating comparison for %s", arena))
		var realMeanProfit float64 = 0.0
		var realProfitLeftBound float64 = 0.0
		var realProfitRightBound float64 = 0.0
		var err error

		realArenaData, exists := realData[models.Arena(arena)]
		if exists {
			realProfitLeftBound, realProfitRightBound, err = realArenaData.GetProfitConfidenceInterval()
			if err != nil {
				return nil, stacktrace.Propagate(err, "failed to get real profit confidence interval")
			}
			realMeanProfit, err = realArenaData.GetMeanDropsProfit()
			if err != nil {
				return nil, stacktrace.Propagate(err, "failed to get real mean profit")
			}
		}

		var generatedProfitLeftBound float64 = 0.0
		var generatedProfitRightBound float64 = 0.0
		var generatedMeanProfit float64 = 0.0

		generatedArenaData, exists := generatedData[models.Arena(arena)]
		if exists {
			generatedProfitLeftBound, generatedProfitRightBound, err = generatedArenaData.GetProfitConfidenceInterval()
			if err != nil {
				return nil, stacktrace.Propagate(err, "failed to get generated profit confidence interval")
			}
			generatedMeanProfit, err = generatedArenaData.GetMeanDropsProfit()
			if err != nil {
				return nil, stacktrace.Propagate(err, "failed to get generated mean profit")
			}
		}

		table.AddRow([]string{
			strconv.Itoa(i + 1),
			arena,
			fmt.Sprintf("%s ∈ %s NP", helpers.FormatFloat(generatedMeanProfit), helpers.FormatFloatRange("[%s, %s]", generatedProfitLeftBound, generatedProfitRightBound)),
			fmt.Sprintf("%s ∈ %s NP", helpers.FormatFloat(realMeanProfit), helpers.FormatFloatRange("[%s, %s]", realProfitLeftBound, realProfitRightBound)),
		})
	}

	return table.GetLines(), nil
}
