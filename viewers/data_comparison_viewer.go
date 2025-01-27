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

func NewDataComparisonViewer(battledomeItemsService *services.BattledomeItemsService, dataComparisonService *services.DataComparisonService, statisticsService *services.StatisticsService) *DataComparisonViewer {
	return &DataComparisonViewer{
		BattledomeItemsService: battledomeItemsService,
		DataComparisonService:  dataComparisonService,
		StatisticsService:      statisticsService,
	}
}

func dryChance(dropRate float64, trials int) float64 {
	return math.Pow(1-dropRate, float64(trials))
}

func (v *DataComparisonViewer) generateProfitableItemsTable(itemPriceCache caches.ItemPriceCache, data models.NormalisedBattledomeItems, isRealData bool) (*helpers.Table, error) {
	dataCopy := models.NormalisedBattledomeItems{}
	for k, v := range data {
		dataCopy[k] = v.Copy()
	}

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

	_, err := data.Metadata()
	if err != nil {
		// Means there isn't any data to render
		return table, nil
	}

	predictedProfit := helpers.Sum(helpers.Map(helpers.Values(dataCopy), func(item *models.BattledomeItem) float64 {
		return item.Profit(itemPriceCache)
	}))
	profitableItems := helpers.OrderByDescending(helpers.Values(dataCopy), func(item *models.BattledomeItem) float64 {
		return item.Profit(itemPriceCache)
	})

	runningIndex := 1
	for _, item := range profitableItems {
		if runningIndex > constants.NumberOfItemsToPrint {
			break
		}

		if item.Name == "nothing" {
			continue
		}

		itemDropRate := item.DropRate(data)
		expectedItemProfit := itemDropRate * itemPriceCache.Price(string(item.Name)) * constants.BattledomeDropsPerDay
		itemDropRateLeftBound, itemDropRateRightBound, err := v.StatisticsService.ClopperPearsonInterval(int(item.Quantity), data.TotalItemQuantity(), constants.SignificanceLevel)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to generate drop rate confidence interval")
		}
		row := helpers.When(isRealData,
			[]string{
				strconv.Itoa(runningIndex),
				string(item.Name),
				fmt.Sprintf("%s ∈ %s%%", helpers.FormatPercentage(itemDropRate), helpers.FormatPercentageRange("[%s, %s]", itemDropRateLeftBound, itemDropRateRightBound)),
				// Don't include dry chance in real data
				helpers.FormatFloat(itemPriceCache.Price(string(item.Name))) + " NP",
				helpers.FormatFloat(expectedItemProfit) + " NP",
				helpers.FormatPercentage(item.Profit(itemPriceCache)/predictedProfit) + "%",
			},
			[]string{
				strconv.Itoa(runningIndex),
				string(item.Name),
				helpers.FormatPercentage(item.DropRate(data)) + "%",
				helpers.FormatPercentage(dryChance(item.DropRate(data), 30*constants.BattledomeDropsPerDay)) + "%",
				helpers.FormatFloat(itemPriceCache.Price(string(item.Name))) + " NP",
				helpers.FormatFloat(expectedItemProfit) + " NP",
				helpers.FormatPercentage(item.Profit(itemPriceCache)/predictedProfit) + "%",
			},
		)
		table.AddRow(row)

		runningIndex++
	}

	return table, nil
}

func (v *DataComparisonViewer) generateArenaProfitableItemsTable(itemPriceCache caches.ItemPriceCache, data models.NormalisedBattledomeItems, generatedItems models.NormalisedBattledomeItems, isRealData bool) (*helpers.Table, error) {
	dataCopy := models.NormalisedBattledomeItems{}
	for k, v := range data {
		dataCopy[k] = v.Copy()
	}

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

	_, err := data.Metadata()
	if err != nil {
		// Means there isn't any data to render
		return table, nil
	}

	predictedProfit := helpers.Sum(helpers.Map(helpers.Values(dataCopy), func(item *models.BattledomeItem) float64 {
		return item.Profit(itemPriceCache)
	}))
	arenaSpecificItems := helpers.Filter(helpers.Values(dataCopy), func(item *models.BattledomeItem) bool {
		_, exists := generatedItems[item.Name]
		return exists
	})
	profitableItems := helpers.OrderByDescending(arenaSpecificItems, func(item *models.BattledomeItem) float64 {
		return item.Profit(itemPriceCache)
	})

	runningIndex := 1
	for _, item := range profitableItems {
		if runningIndex > constants.NumberOfItemsToPrint {
			break
		}

		if item.Name == "nothing" {
			continue
		}

		itemDropRate := item.DropRate(data)
		expectedItemProfit := itemDropRate * itemPriceCache.Price(string(item.Name)) * constants.BattledomeDropsPerDay
		itemDropRateLeftBound, itemDropRateRightBound, err := v.StatisticsService.ClopperPearsonInterval(int(item.Quantity), data.TotalItemQuantity(), constants.SignificanceLevel)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to generate drop rate confidence interval")
		}
		row := helpers.When(isRealData,
			[]string{
				strconv.Itoa(runningIndex),
				string(item.Name),
				fmt.Sprintf("%s ∈ %s%%", helpers.FormatPercentage(itemDropRate), helpers.FormatPercentageRange("[%s, %s]", itemDropRateLeftBound, itemDropRateRightBound)),
				// Don't include dry chance in real data
				helpers.FormatFloat(itemPriceCache.Price(string(item.Name))) + " NP",
				helpers.FormatFloat(expectedItemProfit) + " NP",
				helpers.FormatPercentage(item.Profit(itemPriceCache)/predictedProfit) + "%",
			},
			[]string{
				strconv.Itoa(runningIndex),
				string(item.Name),
				helpers.FormatPercentage(item.DropRate(data)) + "%",
				helpers.FormatPercentage(dryChance(item.DropRate(data), 30*constants.BattledomeDropsPerDay)) + "%",
				helpers.FormatFloat(itemPriceCache.Price(string(item.Name))) + " NP",
				helpers.FormatFloat(expectedItemProfit) + " NP",
				helpers.FormatPercentage(item.Profit(itemPriceCache)/predictedProfit) + "%",
			},
		)
		table.AddRow(row)

		runningIndex++
	}

	return table, nil
}

func (v *DataComparisonViewer) generateCodestoneDropRatesTable(realData models.NormalisedBattledomeItems, generatedData models.NormalisedBattledomeItems, codestoneList []string /* Keep as string to prevent circular dependency*/) (*helpers.Table, error) {
	var tableName string
	if slices.Contains(codestoneList, constants.BrownCodestones[0]) {
		tableName = "Brown Codestone Drop Rates"
	} else if slices.Contains(codestoneList, constants.RedCodestones[0]) {
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
		return profit.DropRate(generatedData)
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
			generatedCodestoneDropRate = generatedCodestones.DropRate(generatedData)
		}

		if existsInReal {
			realDropRate = realCodestones.DropRate(realData)
			realMinDropRate, realMaxDropRate, err = v.StatisticsService.ClopperPearsonInterval(int(realCodestones.Quantity), realData.TotalItemQuantity(), constants.SignificanceLevel)
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

	realTotalMinDropRate, realTotalMaxDropRate, err := v.StatisticsService.ClopperPearsonInterval(int(realTotalCodestoneCount), realData.TotalItemQuantity(), constants.SignificanceLevel)
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

func (v *DataComparisonViewer) ViewChallengerComparison(itemPriceCache caches.ItemPriceCache, realData models.NormalisedBattledomeItems, generatedData models.NormalisedBattledomeItems) ([]string, error) {
	metadata, err := realData.Metadata()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get metadata")
	}

	profitComparisonTable := helpers.NewNamedTable(fmt.Sprintf("%s %s in %s", metadata.Difficulty, metadata.Challenger, metadata.Arena), []string{
		"Type",
		"Value",
	})
	profitComparisonTable.IsLastRowDistinct = true

	generatedMeanProfit, err := generatedData.MeanDropsProfit(itemPriceCache)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get generated mean drops profit")
	}
	generatedProfitStdev, err := generatedData.DropsProfitStdev(itemPriceCache)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get generated drop profit stdev")
	}

	realMeanProfit, err := realData.MeanDropsProfit(itemPriceCache)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get real mean drops profit")
	}
	realProfitStdev, err := realData.DropsProfitStdev(itemPriceCache)
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

	realProfitableItemsTable, err := v.generateProfitableItemsTable(itemPriceCache, realData, true)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to generate profitable items table for real data")
	}
	generatedProfitableItemsTable, err := v.generateProfitableItemsTable(itemPriceCache, generatedData, false)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to generate profitable items table for generated data")
	}

	brownCodestoneDropRatesTable, err := v.generateCodestoneDropRatesTable(realData, generatedData, constants.BrownCodestones)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to generate brown codestones drop rate table for real and generated data")
	}

	redCodestoneDropRatesTable, err := v.generateCodestoneDropRatesTable(realData, generatedData, constants.RedCodestones)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to generate red codestones drop rate table for real and generated data")
	}

	arenaSpecificDropsTable, err := v.generateArenaSpecificDropsTable(itemPriceCache, realData, generatedData)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to generate arena-specific drop rate table for real and generated data")
	}

	challengerSpecificDropsTable, err := v.generateChallengerSpecificDropsTable(itemPriceCache, realData, generatedData)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed ot generate challenger-specific drop rate table for real and generated data")
	}

	tableSeparator := "  "

	lines := []string{}
	lines = append(lines, profitComparisonTable.Lines()...)
	lines = append(lines, "\n")
	lines = append(lines, fmt.Sprintf("Top %d most profitable items", constants.NumberOfItemsToPrint))
	lines = append(lines, generatedProfitableItemsTable.LinesWith(tableSeparator, realProfitableItemsTable)...)
	lines = append(lines, "\n")
	lines = append(lines, brownCodestoneDropRatesTable.LinesWith(tableSeparator, redCodestoneDropRatesTable)...)
	lines = append(lines, "\n")
	lines = append(lines, arenaSpecificDropsTable.LinesWith(tableSeparator, challengerSpecificDropsTable)...)

	return lines, nil
}

func isArenaSpecificDrop(itemName models.ItemName, items models.NormalisedBattledomeItems) bool {
	_, exists := items[itemName]
	return exists
}

func (v *DataComparisonViewer) generateArenaSpecificDropsTable(itemPriceCache caches.ItemPriceCache, realData models.NormalisedBattledomeItems, generatedData models.NormalisedBattledomeItems) (*helpers.Table, error) {
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
		return float64(item.Quantity) * itemPriceCache.Price(string(item.Name))
	})
	totalExpectation := helpers.Sum(helpers.Map(helpers.Values(realData), func(item *models.BattledomeItem) float64 {
		return item.DropRate(realData) * itemPriceCache.Price(string(item.Name)) * constants.BattledomeDropsPerDay
	}))
	arenaSpecificDropsExpectation := helpers.Sum(helpers.Map(realItems, func(item *models.BattledomeItem) float64 {
		return item.DropRate(realData) * itemPriceCache.Price(string(item.Name)) * constants.BattledomeDropsPerDay
	}))

	totalArenaItemCount := helpers.Sum(helpers.Map(realItems, func(item *models.BattledomeItem) int {
		if item.Name == "nothing" {
			return 0
		}

		return int(item.Quantity)
	}))
	totalDropRateLeftBound, totalDropRateRightBound, err := v.StatisticsService.ClopperPearsonInterval(totalArenaItemCount, realData.TotalItemQuantity(), constants.SignificanceLevel)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to generate total drop rate bounds")
	}

	for i, item := range orderedRealItems {
		if i > constants.NumberOfItemsToPrint-1 {
			break
		}

		itemDropRateLeftBound, itemDropRateRightBound, err := v.StatisticsService.ClopperPearsonInterval(int(item.Quantity), realData.TotalItemQuantity(), constants.SignificanceLevel)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to generate item drop bounds")
		}
		itemDropRate := item.DropRate(realData)
		itemPrice := itemPriceCache.Price(string(item.Name))
		itemExpectation := itemDropRate * itemPrice * constants.BattledomeDropsPerDay

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

func (v *DataComparisonViewer) generateChallengerSpecificDropsTable(itemPriceCache caches.ItemPriceCache, realData models.NormalisedBattledomeItems, generatedData models.NormalisedBattledomeItems) (*helpers.Table, error) {
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
		return float64(item.Quantity) * itemPriceCache.Price(string(item.Name))
	})
	totalExpectation := helpers.Sum(helpers.Map(helpers.Values(realData), func(item *models.BattledomeItem) float64 {
		return item.DropRate(realData) * itemPriceCache.Price(string(item.Name)) * constants.BattledomeDropsPerDay
	}))
	challengerSpecificDropsExpectation := helpers.Sum(helpers.Map(realItems, func(item *models.BattledomeItem) float64 {
		return item.DropRate(realData) * itemPriceCache.Price(string(item.Name)) * constants.BattledomeDropsPerDay
	}))

	totalChallengerItemCount := helpers.Sum(helpers.Map(realItems, func(item *models.BattledomeItem) int {
		if item.Name == "nothing" {
			return 0
		}

		return int(item.Quantity)
	}))
	totalDropRateLeftBound, totalDropRateRightBound, err := v.StatisticsService.ClopperPearsonInterval(totalChallengerItemCount, realData.TotalItemQuantity(), constants.SignificanceLevel)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to generate total drop rate bounds")
	}

	for i, item := range orderedRealItems {
		if i > constants.NumberOfItemsToPrint-1 {
			break
		}

		itemDropRateLeftBound, itemDropRateRightBound, err := v.StatisticsService.ClopperPearsonInterval(int(item.Quantity), realData.TotalItemQuantity(), constants.SignificanceLevel)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to generate item drop bounds")
		}
		itemDropRate := item.DropRate(realData)
		itemPrice := itemPriceCache.Price(string(item.Name))
		itemExpectation := itemDropRate * itemPrice * constants.BattledomeDropsPerDay

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

func (v *DataComparisonViewer) ViewChallengerComparisons(itemPriceCache caches.ItemPriceCache, challengerItems []models.NormalisedBattledomeItems) ([]string, error) {
	challengerItems = helpers.OrderByDescending(challengerItems, func(items models.NormalisedBattledomeItems) float64 {
		meanDropsProfit, err := items.MeanDropsProfit(itemPriceCache)
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
		"Brown Codestone Drop Rate",
		"Red Codestone Drop Rate",
		"Challenger Drop Rate",
	})

	for i, items := range challengerItems {
		metadata, err := items.Metadata()
		if err != nil {
			return nil, helpers.PropagateWithSerialisedValue(err, "failed to get metadata from %s", "failed to get metadata from items; additionally encountered an error when trying to serialise the items for logging: %s", items)
		}

		actualProfit, err := items.MeanDropsProfit(itemPriceCache)
		if err != nil {
			return nil, helpers.PropagateWithSerialisedValue(err, "failed to get mean drops profit from %s", "failed to get mean drops profit from items; additionally encountered an error when trying to serialise the items for logging: %s", items)
		}

		actualProfitLeftBound, actualProfitRightBound, err := items.ProfitConfidenceInterval(itemPriceCache)
		if err != nil {
			return nil, helpers.PropagateWithSerialisedValue(err, "failed to get profit confidence interval from %s", "failed to get profit confidence interval from items; additionally encountered an error when trying to serialise the items for logging: %s", items)
		}

		totalRealBrownCodestoneQuantity := helpers.Sum(
			helpers.Map(
				helpers.Filter(
					helpers.Values(items),
					func(item *models.BattledomeItem) bool {
						return slices.Contains(constants.BrownCodestones, string(item.Name))
					}),
				func(item *models.BattledomeItem) int {
					return int(item.Quantity)
				},
			),
		)
		totalRealRedCodestoneQuantity := helpers.Sum(
			helpers.Map(
				helpers.Filter(
					helpers.Values(items),
					func(item *models.BattledomeItem) bool {
						return slices.Contains(constants.RedCodestones, string(item.Name))
					}),
				func(item *models.BattledomeItem) int {
					return int(item.Quantity)
				},
			),
		)
		totalRealItemQuantity := items.TotalItemQuantity()
		brownCodestoneDropRateLeftBound, brownCodestoneDropRateRightBound, err := v.StatisticsService.ClopperPearsonInterval(totalRealBrownCodestoneQuantity, totalRealItemQuantity, constants.SignificanceLevel)
		if err != nil {
			return nil, stacktrace.Propagate(err, "faiiled to generate confidence interval for brown codestone drop rates for %q", metadata.Arena)
		}
		redCodestoneDropRateLeftBound, redCodestoneDropRateRightBound, err := v.StatisticsService.ClopperPearsonInterval(totalRealRedCodestoneQuantity, totalRealItemQuantity, constants.SignificanceLevel)
		if err != nil {
			return nil, stacktrace.Propagate(err, "faiiled to generate confidence interval for red codestone drop rates for %q", metadata.Arena)
		}

		generatedItems, err := v.BattledomeItemsService.GeneratedDropsByArena(metadata.Arena)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to generate drops by arena for %q", metadata.Arena)
		}

		generatedProfit, err := generatedItems.MeanDropsProfit(itemPriceCache)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to get mean drops profit for generated items")
		}

		generatedProfitLeftBound, generatedProfitRightBound, err := generatedItems.ProfitConfidenceInterval(itemPriceCache)
		if err != nil {
			return nil, helpers.PropagateWithSerialisedValue(err, "failed to get profit confidence interval from %s", "failed to get profit confidence interval from items; additionally encountered an error when trying to serialise the items for logging: %s", items)
		}

		profitComparisonTable.AddRow([]string{
			strconv.Itoa(i + 1),
			string(metadata.Arena),
			string(metadata.Challenger),
			string(metadata.Difficulty),
			helpers.FormatInt(items.TotalItemQuantity()),
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
			fmt.Sprintf("%s ∈ [%s, %s]%%", helpers.FormatPercentage(float64(totalRealBrownCodestoneQuantity)/float64(totalRealItemQuantity)), helpers.FormatPercentage(float64(brownCodestoneDropRateLeftBound)), helpers.FormatPercentage(float64(brownCodestoneDropRateRightBound))),
			fmt.Sprintf("%s ∈ [%s, %s]%%", helpers.FormatPercentage(float64(totalRealRedCodestoneQuantity)/float64(totalRealItemQuantity)), helpers.FormatPercentage(float64(redCodestoneDropRateLeftBound)), helpers.FormatPercentage(float64(redCodestoneDropRateRightBound))),
			helpers.FormatPercentage(float64(challengerDropsCount)/float64(arenaDropsCount+challengerDropsCount)) + "%",
		})
	}

	lines := profitComparisonTable.Lines()
	lines = append(lines, "\n")
	lines = append(lines, arenaAndChallengerDropRateTable.Lines()...)
	return lines, nil
}

func (viewer *DataComparisonViewer) ViewArenaComparison(itemPriceCache caches.ItemPriceCache, realData models.NormalisedBattledomeItems, generatedData models.NormalisedBattledomeItems) ([]string, error) {
	metadata, err := generatedData.Metadata()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get metadata")
	}

	var tableName string
	if constants.ShouldIgnoreChallengerDropsInArenaComparison {
		tableName = fmt.Sprintf("Profit in %s (exc. challenger drops)", metadata.Arena)
	} else {
		tableName = fmt.Sprintf("Profit in %s (inc. challenger drops)", metadata.Arena)
	}

	profitComparisonTable := helpers.NewNamedTable(tableName, []string{
		"Type",
		"Value",
	})
	profitComparisonTable.IsLastRowDistinct = true

	generatedMeanProfit, err := generatedData.MeanDropsProfit(itemPriceCache)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get mean drops profit")
	}
	generatedProfitLeftBound, generatedProfitRightBound, err := generatedData.ProfitConfidenceInterval(itemPriceCache)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get profit confidence interval")
	}

	var realMeanProfit float64 = 0.0
	if constants.ShouldIgnoreChallengerDropsInArenaComparison {
		realMeanProfit, err = realData.ArenaMeanDropsProfit(itemPriceCache, generatedData)
	} else {
		realMeanProfit, err = realData.MeanDropsProfit(itemPriceCache)
	}
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get mean drops profit")
	}

	var realProfitLeftBound float64 = 0.0
	var realProfitRightBound float64 = 0.0
	if constants.ShouldIgnoreChallengerDropsInArenaComparison {
		realProfitLeftBound, realProfitRightBound, err = realData.ArenaProfitConfidenceInterval(itemPriceCache, generatedData)
	} else {
		realProfitLeftBound, realProfitRightBound, err = realData.ProfitConfidenceInterval(itemPriceCache)
	}
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

	var realProfitableItemsTable *helpers.Table
	if constants.ShouldIgnoreChallengerDropsInArenaComparison {
		realProfitableItemsTable, err = viewer.generateArenaProfitableItemsTable(itemPriceCache, realData, generatedData, true)
	} else {
		realProfitableItemsTable, err = viewer.generateProfitableItemsTable(itemPriceCache, realData, true)
	}
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to generate profitable items table")
	}

	var generatedProfitableItemsTable *helpers.Table
	if constants.ShouldIgnoreChallengerDropsInArenaComparison {
		generatedProfitableItemsTable, err = viewer.generateArenaProfitableItemsTable(itemPriceCache, generatedData, generatedData, false)
	} else {
		generatedProfitableItemsTable, err = viewer.generateProfitableItemsTable(itemPriceCache, generatedData, false)
	}
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to generate profitable items table")
	}

	var brownCodestoneDropRatesTable *helpers.Table
	brownCodestoneDropRatesTable, err = viewer.generateCodestoneDropRatesTable(realData, generatedData, constants.BrownCodestones)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to generate brown codestones drop rates table")
	}

	var redCodestoneDropRatesTable *helpers.Table
	redCodestoneDropRatesTable, err = viewer.generateCodestoneDropRatesTable(realData, generatedData, constants.RedCodestones)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to generate red codestones drop rates table")
	}

	tableSeparator := "\t"

	lines := []string{}
	lines = append(lines, profitComparisonTable.Lines()...)
	lines = append(lines, "\n")
	lines = append(lines, fmt.Sprintf("Top %d most profitable items in %s", constants.NumberOfItemsToPrint, metadata.Arena))
	lines = append(lines, generatedProfitableItemsTable.LinesWith(tableSeparator, realProfitableItemsTable)...)
	lines = append(lines, "\n")
	lines = append(lines, fmt.Sprintf("Codestone drop rates in %s", metadata.Arena))
	lines = append(lines, brownCodestoneDropRatesTable.LinesWith(tableSeparator, redCodestoneDropRatesTable)...)
	return lines, nil
}

func (viewer *DataComparisonViewer) ViewBriefArenaComparisons(itemPriceCache caches.ItemPriceCache, realData map[models.Arena]models.NormalisedBattledomeItems, generatedData map[models.Arena]models.NormalisedBattledomeItems) ([]string, error) {
	orderedArenas := helpers.OrderByDescending(constants.Arenas, func(arena string) float64 {
		normalisedItems, exists := realData[models.Arena(arena)]
		if !exists || normalisedItems.TotalItemQuantity() == 0 {
			return 0.0
		}

		var profit float64 = 0.0
		var err error
		if constants.ShouldIgnoreChallengerDropsInArenaComparison {
			profit, err = normalisedItems.ArenaMeanDropsProfit(itemPriceCache, generatedData[models.Arena(arena)])
		} else {
			profit, err = normalisedItems.MeanDropsProfit(itemPriceCache)
		}
		if err != nil {
			return 0.0
		}

		return profit
	})

	var tableName string
	if constants.ShouldIgnoreChallengerDropsInArenaComparison {
		tableName = "Profit (excluding challenger-specific drops)"
	} else {
		tableName = "Profit (including challenger-specific drops)"
	}
	table := helpers.NewNamedTable(tableName, []string{
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
			if constants.ShouldIgnoreChallengerDropsInArenaComparison {
				realProfitLeftBound, realProfitRightBound, err = realArenaData.ArenaProfitConfidenceInterval(itemPriceCache, generatedData[models.Arena(arena)])
			} else {
				realProfitLeftBound, realProfitRightBound, err = realArenaData.ProfitConfidenceInterval(itemPriceCache)
			}
			if err != nil {
				return nil, stacktrace.Propagate(err, "failed to get real profit confidence interval")
			}

			if constants.ShouldIgnoreChallengerDropsInArenaComparison {
				realMeanProfit, err = realArenaData.ArenaMeanDropsProfit(itemPriceCache, generatedData[models.Arena(arena)])
			} else {
				realMeanProfit, err = realArenaData.MeanDropsProfit(itemPriceCache)
			}
			if err != nil {
				return nil, stacktrace.Propagate(err, "failed to get real mean profit")
			}
		}

		var generatedProfitLeftBound float64 = 0.0
		var generatedProfitRightBound float64 = 0.0
		var generatedMeanProfit float64 = 0.0

		generatedArenaData, exists := generatedData[models.Arena(arena)]
		if exists {
			generatedProfitLeftBound, generatedProfitRightBound, err = generatedArenaData.ProfitConfidenceInterval(itemPriceCache)
			if err != nil {
				return nil, stacktrace.Propagate(err, "failed to get generated profit confidence interval")
			}
			generatedMeanProfit, err = generatedArenaData.MeanDropsProfit(itemPriceCache)
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

	return table.Lines(), nil
}
