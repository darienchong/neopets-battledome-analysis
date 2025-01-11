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
	GeneratedDropsService *services.GeneratedDropsService
	DataComparisonService *services.DataComparisonService
}

func NewDataComparisonViewer() *DataComparisonViewer {
	return &DataComparisonViewer{
		GeneratedDropsService: services.NewGeneratedDropsService(),
		DataComparisonService: services.NewDataComparisonService(),
	}
}

func getDryChance(dropRate float64, trials int) float64 {
	return math.Pow(1-dropRate, float64(trials))
}

func generateProfitableItemsTable(data *models.ComparisonResult, isRealData bool) *helpers.Table {
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

	predictedProfit := helpers.Sum(helpers.Map(helpers.Values(data.Profit), func(profit *models.ItemProfit) float64 { return profit.GetProfit() }))
	profitableItems := helpers.OrderByDescending(helpers.Values(data.Profit), func(profit *models.ItemProfit) float64 {
		return profit.GetProfit()
	})
	for i, itemProfit := range profitableItems {
		if i > constants.NUMBER_OF_ITEMS_TO_PRINT-1 {
			break
		}

		row := helpers.When(isRealData,
			[]string{
				strconv.Itoa(i + 1),
				itemProfit.ItemName,
				helpers.FormatPercentage(itemProfit.DropRate) + "%",
				// Don't include dry chance in real data
				helpers.FormatFloat(itemProfit.IndividualPrice) + " NP",
				helpers.FormatFloat(itemProfit.GetProfit()*constants.BATTLEDOME_DROPS_PER_DAY) + " NP",
				helpers.FormatPercentage(itemProfit.GetProfit()/predictedProfit) + "%",
			},
			[]string{
				strconv.Itoa(i + 1),
				itemProfit.ItemName,
				helpers.FormatPercentage(itemProfit.DropRate) + "%",
				helpers.FormatPercentage(getDryChance(itemProfit.DropRate, 30*constants.NUMBER_OF_ITEMS_TO_PRINT)) + "%",
				helpers.FormatFloat(itemProfit.IndividualPrice) + " NP",
				helpers.FormatFloat(itemProfit.GetProfit()*constants.BATTLEDOME_DROPS_PER_DAY) + " NP",
				helpers.FormatPercentage(itemProfit.GetProfit()/predictedProfit) + "%",
			})

		table.AddRow(row)
	}

	return table
}

func generateCodestoneDropRatesTable(realData *models.ComparisonResult, generatedData *models.ComparisonResult, codestoneList []string) *helpers.Table {
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

	realCodestoneProfits := helpers.ToMap(helpers.Filter(helpers.Values(realData.Profit), func(itemProfit *models.ItemProfit) bool {
		return slices.Contains(codestoneList, itemProfit.ItemName)
	}), func(itemProfit *models.ItemProfit) string {
		return itemProfit.ItemName
	}, func(itemProfit *models.ItemProfit) *models.ItemProfit {
		return itemProfit
	})
	generatedCodestoneProfits := helpers.ToMap(helpers.Filter(helpers.Values(generatedData.Profit), func(itemProfit *models.ItemProfit) bool {
		return slices.Contains(codestoneList, itemProfit.ItemName)
	}), func(itemProfit *models.ItemProfit) string {
		return itemProfit.ItemName
	}, func(itemProfit *models.ItemProfit) *models.ItemProfit {
		return itemProfit
	})

	realTotalCodestoneDropRate := helpers.Sum(helpers.Map(helpers.Values(realCodestoneProfits), func(profit *models.ItemProfit) float64 {
		return profit.DropRate
	}))
	generatedTotalCodestoneDropRate := helpers.Sum(helpers.Map(helpers.Values(generatedCodestoneProfits), func(profit *models.ItemProfit) float64 {
		return profit.DropRate
	}))

	slices.Sort(codestoneList)
	for _, codestoneName := range codestoneList {
		realCodestoneProfit, existsInReal := realCodestoneProfits[codestoneName]
		generatedCodestoneProfit, existsInGenerated := generatedCodestoneProfits[codestoneName]
		var realCodestoneDropRate float64 = 0.0
		var generatedCodestoneDropRate float64 = 0.0

		if existsInGenerated {
			generatedCodestoneDropRate = generatedCodestoneProfit.DropRate
		}
		if existsInReal {
			realCodestoneDropRate = realCodestoneProfit.DropRate
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

func (viewer *DataComparisonViewer) ViewChallengerComparison(realData *models.ComparisonResult, generatedData *models.ComparisonResult) ([]string, error) {
	profitComparisonTable := helpers.NewNamedTable(fmt.Sprintf("%s %s in %s", realData.Analysis.Metadata.Difficulty, realData.Analysis.Metadata.Challenger, realData.Analysis.Metadata.Arena), []string{
		"Type",
		"Value",
	})
	profitComparisonTable.IsLastRowDistinct = true

	generatedMeanProfit, err := generatedData.Analysis.GetMeanDropsProfit()
	if err != nil {
		return nil, err
	}
	generatedProfitStdev, err := generatedData.Analysis.GetDropsProfitStdev()
	if err != nil {
		return nil, err
	}

	realMeanProfit, err := realData.Analysis.GetMeanDropsProfit()
	if err != nil {
		return nil, err
	}
	realProfitStdev, err := realData.Analysis.GetDropsProfitStdev()
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

	realProfitableItemsTable := generateProfitableItemsTable(realData, true)
	generatedProfitableItemsTable := generateProfitableItemsTable(generatedData, false)

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

func isArenaSpecificDrop(itemName string, arenaItems map[string]*models.BattledomeItem) bool {
	_, exists := arenaItems[itemName]
	return exists
}

func generateArenaSpecificDropsTable(realData *models.ComparisonResult, generatedData *models.ComparisonResult) (*helpers.Table, error) {
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

	realItems := helpers.Filter(helpers.Values(realData.Analysis.Items), func(item *models.BattledomeItem) bool {
		return isArenaSpecificDrop(item.Name, generatedData.Analysis.Items)
	})
	orderedRealItems := helpers.OrderByDescending(realItems, func(item *models.BattledomeItem) float64 {
		return float64(item.Quantity) * itemPriceCache.GetPrice(item.Name)
	})
	totalExpectation := helpers.Sum(helpers.Map(helpers.Values(realData.Analysis.Items), func(item *models.BattledomeItem) float64 {
		return item.GetDropRate(realData.Analysis) * itemPriceCache.GetPrice(item.Name) * constants.BATTLEDOME_DROPS_PER_DAY
	}))
	arenaSpecificDropsExpectation := helpers.Sum(helpers.Map(realItems, func(item *models.BattledomeItem) float64 {
		return item.GetDropRate(realData.Analysis) * itemPriceCache.GetPrice(item.Name) * constants.BATTLEDOME_DROPS_PER_DAY
	}))
	totalDropRate := helpers.Sum(helpers.Map(realItems, func(item *models.BattledomeItem) float64 {
		return item.GetDropRate(realData.Analysis)
	}))

	for i, item := range orderedRealItems {
		if i > constants.NUMBER_OF_ITEMS_TO_PRINT-1 {
			break
		}

		itemDropRate := item.GetDropRate(realData.Analysis)
		itemPrice := itemPriceCache.GetPrice(item.Name)
		itemExpectation := itemDropRate * itemPrice * constants.BATTLEDOME_DROPS_PER_DAY

		table.AddRow([]string{
			strconv.Itoa(i + 1),
			item.Name,
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

func generateChallengerSpecificDropsTable(realData *models.ComparisonResult, generatedData *models.ComparisonResult) (*helpers.Table, error) {
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

	realItems := helpers.Filter(helpers.Values(realData.Analysis.Items), func(item *models.BattledomeItem) bool {
		return !isArenaSpecificDrop(item.Name, generatedData.Analysis.Items)
	})
	orderedRealItems := helpers.OrderByDescending(realItems, func(item *models.BattledomeItem) float64 {
		return float64(item.Quantity) * itemPriceCache.GetPrice(item.Name)
	})
	totalExpectation := helpers.Sum(helpers.Map(helpers.Values(realData.Analysis.Items), func(item *models.BattledomeItem) float64 {
		return item.GetDropRate(realData.Analysis) * itemPriceCache.GetPrice(item.Name) * constants.BATTLEDOME_DROPS_PER_DAY
	}))
	challengerSpecificDropsExpectation := helpers.Sum(helpers.Map(realItems, func(item *models.BattledomeItem) float64 {
		return item.GetDropRate(realData.Analysis) * itemPriceCache.GetPrice(item.Name) * constants.BATTLEDOME_DROPS_PER_DAY
	}))
	totalDropRate := helpers.Sum(helpers.Map(realItems, func(item *models.BattledomeItem) float64 {
		return item.GetDropRate(realData.Analysis)
	}))

	for i, item := range orderedRealItems {
		if i > constants.NUMBER_OF_ITEMS_TO_PRINT-1 {
			break
		}

		itemDropRate := item.GetDropRate(realData.Analysis)
		itemPrice := itemPriceCache.GetPrice(item.Name)
		itemExpectation := itemDropRate * itemPrice * constants.BATTLEDOME_DROPS_PER_DAY

		table.AddRow([]string{
			strconv.Itoa(i + 1),
			item.Name,
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

func (viewer *DataComparisonViewer) ViewChallengerComparisons(comparisonResults []*models.ComparisonResult) ([]string, error) {
	comparisonResults = helpers.OrderByDescending(comparisonResults, func(res *models.ComparisonResult) float64 {
		meanDropsProfit, err := res.Analysis.GetMeanDropsProfit()
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

	for i, result := range comparisonResults {
		actualProfit, err := result.Analysis.GetMeanDropsProfit()
		if err != nil {
			return nil, err
		}
		actualStdev, err := result.Analysis.GetDropsProfitStdev()
		if err != nil {
			return nil, err
		}

		generatedDrops, err := viewer.GeneratedDropsService.GenerateDrops(result.Analysis.Metadata.Arena)
		if err != nil {
			return nil, err
		}
		generatedResult, err := viewer.DataComparisonService.ToComparisonResult(generatedDrops)
		if err != nil {
			return nil, err
		}
		generatedProfit, err := generatedResult.Analysis.GetMeanDropsProfit()
		if err != nil {
			return nil, err
		}
		generatedStdev, err := generatedResult.Analysis.GetDropsProfitStdev()
		if err != nil {
			return nil, err
		}

		profitComparisonTable.AddRow([]string{
			strconv.Itoa(i + 1),
			result.Analysis.Metadata.Arena,
			result.Analysis.Metadata.Challenger,
			result.Analysis.Metadata.Difficulty,
			helpers.FormatInt(result.Analysis.GetTotalItemQuantity()),
			helpers.FormatFloat(actualProfit) + " ± " + helpers.FormatFloat(actualStdev) + " NP",
			helpers.FormatFloat(generatedProfit) + " ± " + helpers.FormatFloat(generatedStdev) + " NP",
		})

		arenaDropsCount := helpers.Sum(helpers.Map(
			helpers.Values(result.Analysis.Items),
			func(item *models.BattledomeItem) int {
				if item.Name == "nothing" {
					return 0
				}

				_, exists := generatedDrops.Items[item.Name]
				if !exists {
					return 0
				}

				return int(item.Quantity)
			},
		))

		challengerDropsCount := helpers.Sum(helpers.Map(
			helpers.Values(result.Analysis.Items),
			func(item *models.BattledomeItem) int {
				if item.Name == "nothing" {
					return 0
				}

				_, exists := generatedDrops.Items[item.Name]
				if exists {
					return 0
				}

				return int(item.Quantity)
			},
		))

		arenaAndChallengerDropRateTable.AddRow([]string{
			strconv.Itoa(i + 1),
			result.Analysis.Metadata.Arena,
			result.Analysis.Metadata.Challenger,
			result.Analysis.Metadata.Difficulty,
			helpers.FormatPercentage(float64(arenaDropsCount)/float64(arenaDropsCount+challengerDropsCount)) + "%",
			helpers.FormatPercentage(float64(challengerDropsCount)/float64(arenaDropsCount+challengerDropsCount)) + "%",
		})
	}

	lines := profitComparisonTable.GetLines()
	lines = append(lines, "\n")
	lines = append(lines, arenaAndChallengerDropRateTable.GetLines()...)
	return lines, nil
}

func (viewer *DataComparisonViewer) ViewArenaComparison(realData *models.ComparisonResult, generatedData *models.ComparisonResult) ([]string, error) {
	arena := realData.Analysis.Metadata.Arena

	profitComparisonTable := helpers.NewNamedTable(fmt.Sprintf("Profit in %s", arena), []string{
		"Type",
		"Value",
	})
	profitComparisonTable.IsLastRowDistinct = true

	generatedMeanProfit, err := generatedData.Analysis.GetMeanDropsProfit()
	if err != nil {
		return nil, err
	}
	generatedProfitStdev, err := generatedData.Analysis.GetDropsProfitStdev()
	if err != nil {
		return nil, err
	}

	realMeanProfit, err := realData.Analysis.GetMeanDropsProfit()
	if err != nil {
		return nil, err
	}
	realProfitStdev, err := realData.Analysis.GetDropsProfitStdev()
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

	realProfitableItemsTable := generateProfitableItemsTable(realData, true)
	generatedProfitableItemsTable := generateProfitableItemsTable(generatedData, false)

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
