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

func generateProfitableItemsTable(data models.NormalisedBattledomeItems, isRealData bool) (*helpers.Table, error) {
	dataCopy := models.NormalisedBattledomeItems{}
	for k, v := range data {
		dataCopy[k] = v.Copy()
	}

	itemPriceCache, err := caches.GetItemPriceCacheInstance()
	if err != nil {
		return nil, err
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
		row := helpers.When(isRealData,
			[]string{
				strconv.Itoa(runningIndex),
				string(item.Name),
				helpers.FormatPercentage(itemDropRate) + "%",
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
		var realMinDropRate float64 = 0.0
		var realMaxDropRate float64 = 0.0

		var generatedCodestoneDropRate float64 = 0.0

		if existsInGenerated {
			generatedCodestoneDropRate = generatedCodestones.GetDropRate(generatedData)
		}

		if existsInReal {
			leftBound, rightBound, err := viewer.StatisticsService.ClopperPearsonInterval(int(realCodestones.Quantity), realData.GetTotalItemQuantity(), constants.SIGNIFICANCE_LEVEL)
			if err != nil {
				return nil, err
			}
			realMinDropRate = leftBound
			realMaxDropRate = rightBound
		}

		table.AddRow([]string{
			codestoneName,
			helpers.FormatPercentage(generatedCodestoneDropRate) + "%",
			helpers.When(realMinDropRate == realMaxDropRate, helpers.FormatPercentage(realMinDropRate)+"%", fmt.Sprintf("[%s, %s]%%", helpers.FormatPercentage(realMinDropRate), helpers.FormatPercentage(realMaxDropRate))),
		})
	}

	realTotalMinDropRate, realTotalMaxDropRate, err := viewer.StatisticsService.ClopperPearsonInterval(int(realTotalCodestoneCount), realData.GetTotalItemQuantity(), constants.SIGNIFICANCE_LEVEL)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	profitComparisonTable := helpers.NewNamedTable(fmt.Sprintf("%s %s in %s", metadata.Difficulty, metadata.Challenger, metadata.Arena), []string{
		"Type",
		"Value",
	})
	profitComparisonTable.IsLastRowDistinct = true

	generatedMeanProfit, err := generatedData.GetMeanDropsProfit()
	if err != nil {
		return nil, err
	}
	generatedProfitStdev, err := generatedData.GetDropsProfitStdev()
	if err != nil {
		return nil, err
	}

	realMeanProfit, err := realData.GetMeanDropsProfit()
	if err != nil {
		return nil, err
	}
	realProfitStdev, err := realData.GetDropsProfitStdev()
	if err != nil {
		return nil, err
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

	realProfitableItemsTable, err := generateProfitableItemsTable(realData, true)
	if err != nil {
		return nil, err
	}
	generatedProfitableItemsTable, err := generateProfitableItemsTable(generatedData, false)
	if err != nil {
		return nil, err
	}

	brownCodestoneDropRatesTable, err := viewer.generateCodestoneDropRatesTable(realData, generatedData, constants.BROWN_CODESTONES)
	if err != nil {
		return nil, err
	}

	redCodestoneDropRatesTable, err := viewer.generateCodestoneDropRatesTable(realData, generatedData, constants.RED_CODESTONES)
	if err != nil {
		return nil, err
	}

	arenaSpecificDropsTable, err := generateArenaSpecificDropsTable(realData, generatedData)
	if err != nil {
		return nil, err
	}

	challengerSpecificDropsTable, err := generateChallengerSpecificDropsTable(realData, generatedData)
	if err != nil {
		return nil, err
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

func generateArenaSpecificDropsTable(realData models.NormalisedBattledomeItems, generatedData models.NormalisedBattledomeItems) (*helpers.Table, error) {
	itemPriceCache, err := caches.GetItemPriceCacheInstance()
	if err != nil {
		return nil, err
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
	totalDropRate := helpers.Sum(helpers.Map(realItems, func(item *models.BattledomeItem) float64 {
		return item.GetDropRate(realData)
	}))

	for i, item := range orderedRealItems {
		if i > constants.NUMBER_OF_ITEMS_TO_PRINT-1 {
			break
		}

		itemDropRate := item.GetDropRate(realData)
		itemPrice := itemPriceCache.GetPrice(string(item.Name))
		itemExpectation := itemDropRate * itemPrice * constants.BATTLEDOME_DROPS_PER_DAY

		table.AddRow([]string{
			strconv.Itoa(i + 1),
			string(item.Name),
			helpers.FormatPercentage(itemDropRate) + "%",
			helpers.FormatFloat(itemPrice) + " NP",
			helpers.FormatFloat(itemExpectation) + " NP",
			helpers.FormatPercentage(itemExpectation/totalExpectation) + "%",
		})
	}

	table.AddRow([]string{
		"",
		"Total",
		helpers.FormatPercentage(totalDropRate) + "%",
		"",
		helpers.FormatFloat(arenaSpecificDropsExpectation) + " NP",
		helpers.FormatPercentage(arenaSpecificDropsExpectation/totalExpectation) + "%",
	})

	return table, nil
}

func generateChallengerSpecificDropsTable(realData models.NormalisedBattledomeItems, generatedData models.NormalisedBattledomeItems) (*helpers.Table, error) {
	itemPriceCache, err := caches.GetItemPriceCacheInstance()
	if err != nil {
		return nil, err
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
	totalDropRate := helpers.Sum(helpers.Map(realItems, func(item *models.BattledomeItem) float64 {
		return item.GetDropRate(realData)
	}))

	for i, item := range orderedRealItems {
		if i > constants.NUMBER_OF_ITEMS_TO_PRINT-1 {
			break
		}

		itemDropRate := item.GetDropRate(realData)
		itemPrice := itemPriceCache.GetPrice(string(item.Name))
		itemExpectation := itemDropRate * itemPrice * constants.BATTLEDOME_DROPS_PER_DAY

		table.AddRow([]string{
			strconv.Itoa(i + 1),
			string(item.Name),
			helpers.FormatPercentage(itemDropRate) + "%",
			helpers.FormatFloat(itemPrice) + " NP",
			helpers.FormatFloat(itemExpectation) + " NP",
			helpers.FormatPercentage(itemExpectation/totalExpectation) + "%",
		})
	}

	table.AddRow([]string{
		"",
		"Total",
		helpers.FormatPercentage(totalDropRate) + "%",
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
			panic(err)
		}
		return meanDropsProfit
	})

	profitComparisonTable := helpers.NewNamedTable("Challenger profit comparison", []string{
		"i",
		"Arena",
		"Challenger",
		"Difficulty",
		"Samples",
		"Actual Profit ± Stdev",
		"Predicted Profit ± Stdev",
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
			return nil, err
		}

		actualProfit, err := items.GetMeanDropsProfit()
		if err != nil {
			return nil, err
		}
		actualStdev, err := items.GetDropsProfitStdev()
		if err != nil {
			return nil, err
		}

		generatedItems, err := viewer.BattledomeItemsService.GenerateDropsByArena(metadata.Arena)
		if err != nil {
			return nil, err
		}
		generatedProfit, err := generatedItems.GetMeanDropsProfit()
		if err != nil {
			return nil, err
		}
		generatedStdev, err := generatedItems.GetDropsProfitStdev()
		if err != nil {
			return nil, err
		}

		profitComparisonTable.AddRow([]string{
			strconv.Itoa(i + 1),
			string(metadata.Arena),
			string(metadata.Challenger),
			string(metadata.Difficulty),
			helpers.FormatInt(items.GetTotalItemQuantity()),
			helpers.FormatFloat(actualProfit) + " ± " + helpers.FormatFloat(actualStdev) + " NP",
			helpers.FormatFloat(generatedProfit) + " ± " + helpers.FormatFloat(generatedStdev) + " NP",
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
		return nil, err
	}

	profitComparisonTable := helpers.NewNamedTable(fmt.Sprintf("Profit in %s", metadata.Arena), []string{
		"Type",
		"Value",
	})
	profitComparisonTable.IsLastRowDistinct = true

	generatedMeanProfit, err := generatedData.GetMeanDropsProfit()
	if err != nil {
		return nil, err
	}
	generatedProfitStdev, err := generatedData.GetDropsProfitStdev()
	if err != nil {
		return nil, err
	}

	realMeanProfit, err := realData.GetMeanDropsProfit()
	if err != nil {
		return nil, err
	}
	realProfitStdev, err := realData.GetDropsProfitStdev()
	if err != nil {
		return nil, err
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

	realProfitableItemsTable, err := generateProfitableItemsTable(realData, true)
	if err != nil {
		return nil, err
	}
	generatedProfitableItemsTable, err := generateProfitableItemsTable(generatedData, false)
	if err != nil {
		return nil, err
	}

	brownCodestoneDropRatesTable, err := viewer.generateCodestoneDropRatesTable(realData, generatedData, constants.BROWN_CODESTONES)
	if err != nil {
		return nil, err
	}

	redCodestoneDropRatesTable, err := viewer.generateCodestoneDropRatesTable(realData, generatedData, constants.RED_CODESTONES)
	if err != nil {
		return nil, err
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
			panic(err)
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
		var realProfitStdev float64 = 0.0
		var err error

		realArenaData, exists := realData[models.Arena(arena)]
		if exists {
			realMeanProfit, err = realArenaData.GetMeanDropsProfit()
			if err != nil {
				return nil, fmt.Errorf("failed to generate real mean drops profit for \"%s\": %w", arena, err)
			}
			realProfitStdev, err = realArenaData.GetDropsProfitStdev()
			if err != nil {
				return nil, fmt.Errorf("failed to generate real drops profit stdev for \"%s\": %w", arena, err)
			}
		}

		var generatedMeanProfit float64 = 0.0
		var generatedProfitStdev float64 = 0.0
		generatedArenaData, exists := generatedData[models.Arena(arena)]
		if exists {
			generatedMeanProfit, err = generatedArenaData.GetMeanDropsProfit()
			if err != nil {
				return nil, fmt.Errorf("failed to generate predicted mean drops profit for \"%s\": %w", arena, err)
			}
			generatedProfitStdev, err = generatedArenaData.GetDropsProfitStdev()
			if err != nil {
				return nil, fmt.Errorf("failed to generate predicted mean drops stdev for \"%s\": %w", arena, err)
			}
		}

		table.AddRow([]string{
			strconv.Itoa(i + 1),
			arena,
			fmt.Sprintf("%s ± %s NP", helpers.FormatFloat(generatedMeanProfit), helpers.FormatFloat(generatedProfitStdev)),
			fmt.Sprintf("%s ± %s NP", helpers.FormatFloat(realMeanProfit), helpers.FormatFloat(realProfitStdev)),
		})
	}

	return table.GetLines(), nil
}
