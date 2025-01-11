package viewers

import (
	"fmt"
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
}

func NewDataComparisonViewer() *DataComparisonViewer {
	return &DataComparisonViewer{
		BattledomeItemsService: services.NewBattledomeItemsService(),
		DataComparisonService:  services.NewDataComparisonService(),
	}
}

func getDryChance(dropRate float64, trials int) float64 {
	return math.Pow(1-dropRate, float64(trials))
}

func generateProfitableItemsTable(data models.NormalisedBattledomeItems, isRealData bool) (*helpers.Table, error) {
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

	predictedProfit := helpers.Sum(helpers.Map(helpers.Values(data), func(profit *models.BattledomeItem) float64 { return profit.GetProfit(itemPriceCache) }))
	profitableItems := helpers.OrderByDescending(helpers.Values(data), func(profit *models.BattledomeItem) float64 {
		return profit.GetProfit(itemPriceCache)
	})
	for i, itemProfit := range profitableItems {
		if i > constants.NUMBER_OF_ITEMS_TO_PRINT-1 {
			break
		}

		row := helpers.When(isRealData,
			[]string{
				strconv.Itoa(i + 1),
				string(itemProfit.Name),
				helpers.FormatPercentage(itemProfit.GetDropRate(data)) + "%",
				// Don't include dry chance in real data
				helpers.FormatFloat(itemPriceCache.GetPrice(string(itemProfit.Name))) + " NP",
				helpers.FormatFloat(itemProfit.GetProfit(itemPriceCache)*constants.BATTLEDOME_DROPS_PER_DAY) + " NP",
				helpers.FormatPercentage(itemProfit.GetProfit(itemPriceCache)/predictedProfit) + "%",
			},
			[]string{
				strconv.Itoa(i + 1),
				string(itemProfit.Name),
				helpers.FormatPercentage(itemProfit.GetDropRate(data)) + "%",
				helpers.FormatPercentage(getDryChance(itemProfit.GetDropRate(data), 30*constants.NUMBER_OF_ITEMS_TO_PRINT)) + "%",
				helpers.FormatFloat(itemPriceCache.GetPrice(string(itemProfit.Name))) + " NP",
				helpers.FormatFloat(itemProfit.GetProfit(itemPriceCache)*constants.BATTLEDOME_DROPS_PER_DAY) + " NP",
				helpers.FormatPercentage(itemProfit.GetProfit(itemPriceCache)/predictedProfit) + "%",
			},
		)
		table.AddRow(row)
	}

	return table, nil
}

func generateCodestoneDropRatesTable(realData models.NormalisedBattledomeItems, generatedData models.NormalisedBattledomeItems, codestoneList []string /* Keep as string to prevent circular dependency*/) *helpers.Table {
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
		"Diff",
	})
	table.IsLastRowDistinct = true

	realCodestoneProfits := helpers.ToMap(helpers.Filter(helpers.Values(realData), func(itemProfit *models.BattledomeItem) bool {
		return slices.Contains(codestoneList, string(itemProfit.Name))
	}), func(itemProfit *models.BattledomeItem) models.ItemName {
		return itemProfit.Name
	}, func(itemProfit *models.BattledomeItem) *models.BattledomeItem {
		return itemProfit
	})
	generatedCodestoneProfits := helpers.ToMap(helpers.Filter(helpers.Values(generatedData), func(itemProfit *models.BattledomeItem) bool {
		return slices.Contains(codestoneList, string(itemProfit.Name))
	}), func(itemProfit *models.BattledomeItem) models.ItemName {
		return itemProfit.Name
	}, func(itemProfit *models.BattledomeItem) *models.BattledomeItem {
		return itemProfit
	})

	realTotalCodestoneDropRate := helpers.Sum(helpers.Map(helpers.Values(realCodestoneProfits), func(profit *models.BattledomeItem) float64 {
		return profit.GetDropRate(realData)
	}))
	generatedTotalCodestoneDropRate := helpers.Sum(helpers.Map(helpers.Values(generatedCodestoneProfits), func(profit *models.BattledomeItem) float64 {
		return profit.GetDropRate(generatedData)
	}))

	slices.Sort(codestoneList)
	for _, codestoneName := range codestoneList {
		realCodestoneProfit, existsInReal := realCodestoneProfits[models.ItemName(codestoneName)]
		generatedCodestoneProfit, existsInGenerated := generatedCodestoneProfits[models.ItemName(codestoneName)]
		var realCodestoneDropRate float64 = 0.0
		var generatedCodestoneDropRate float64 = 0.0

		if existsInGenerated {
			generatedCodestoneDropRate = generatedCodestoneProfit.GetDropRate(generatedData)
		}
		if existsInReal {
			realCodestoneDropRate = realCodestoneProfit.GetDropRate(realData)
		}

		table.AddRow([]string{
			codestoneName,
			helpers.FormatPercentage(generatedCodestoneDropRate) + "%",
			helpers.FormatPercentage(realCodestoneDropRate) + "%",
			helpers.FormatPercentage(realCodestoneDropRate-generatedCodestoneDropRate) + "%",
		})
	}
	table.AddRow([]string{
		"Sum",
		helpers.FormatPercentage(generatedTotalCodestoneDropRate) + "%",
		helpers.FormatPercentage(realTotalCodestoneDropRate) + "%",
		helpers.FormatPercentage(realTotalCodestoneDropRate-generatedTotalCodestoneDropRate) + "%",
	})

	return table
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

	brownCodestoneDropRatesTable := generateCodestoneDropRatesTable(realData, generatedData, constants.BROWN_CODESTONES)
	redCodestoneDropRatesTable := generateCodestoneDropRatesTable(realData, generatedData, constants.RED_CODESTONES)

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
	metadata, err := realData.GetMetadata()
	if err != nil {
		return nil, err
	}
	arena := metadata.Arena

	profitComparisonTable := helpers.NewNamedTable(fmt.Sprintf("Profit in %s", arena), []string{
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

	brownCodestoneDropRatesTable := generateCodestoneDropRatesTable(realData, generatedData, constants.BROWN_CODESTONES)
	redCodestoneDropRatesTable := generateCodestoneDropRatesTable(realData, generatedData, constants.RED_CODESTONES)

	tableSeparator := "\t"

	lines := []string{}
	lines = append(lines, profitComparisonTable.GetLines()...)
	lines = append(lines, "\n")
	lines = append(lines, fmt.Sprintf("Top %d most profitable items in %s", constants.NUMBER_OF_ITEMS_TO_PRINT, arena))
	lines = append(lines, generatedProfitableItemsTable.GetLinesWith(tableSeparator, realProfitableItemsTable)...)
	lines = append(lines, "\n")
	lines = append(lines, fmt.Sprintf("Codestone drop rates in %s", arena))
	lines = append(lines, brownCodestoneDropRatesTable.GetLinesWith(tableSeparator, redCodestoneDropRatesTable)...)
	return lines, nil
}
