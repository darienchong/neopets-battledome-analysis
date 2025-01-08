package services

import (
	"fmt"
	"math"
	"slices"
	"strconv"

	"github.com/darienchong/neopets-battledome-analysis/caches"
	"github.com/darienchong/neopets-battledome-analysis/constants"
	"github.com/darienchong/neopets-battledome-analysis/helpers"
	"github.com/darienchong/neopets-battledome-analysis/models"
)

type DataComparisonService struct {
	GeneratedDropsService *GeneratedDropsService
	EmpiricalDropsService *EmpiricalDropsService
	DropsAnalysisService  *DropsAnalysisService
	DropRateService       *DropRateService
}

func NewDataComparisonService() *DataComparisonService {
	return &DataComparisonService{
		GeneratedDropsService: NewGeneratedDropsService(),
		EmpiricalDropsService: NewEmpiricalDropsService(),
		DropsAnalysisService:  NewDropsAnalysisService(),
		DropRateService:       NewDropRateService(),
	}
}

func (service *DataComparisonService) ToComparisonResult(drop *models.BattledomeDrops) (*models.ComparisonResult, error) {
	itemPriceCache, err := caches.GetItemPriceCacheInstance()
	if err != nil {
		return nil, err
	}
	defer itemPriceCache.Close()

	analysis := service.DropsAnalysisService.Analyse(drop)
	dropRates, err := service.DropRateService.CalculateDropRates(&models.BattledomeDrops{
		Metadata: models.DropsMetadataWithSource{
			Source:        "(multiple sources)",
			DropsMetadata: analysis.Metadata,
		},
		Items: helpers.ToPointerMap(
			analysis.GetItemsOrderedByProfit(),
			func(item *models.BattledomeItem) string {
				return item.Name
			},
			func(item *models.BattledomeItem) *models.BattledomeItem {
				return item
			}),
	})
	if err != nil {
		return nil, err
	}

	profits := map[string]*models.ItemProfit{}
	for _, itemDropRate := range dropRates[analysis.Metadata.Arena] {
		profits[itemDropRate.ItemName] = &models.ItemProfit{
			ItemDropRate:    *itemDropRate,
			IndividualPrice: itemPriceCache.GetPrice(itemDropRate.ItemName),
		}
	}

	return &models.ComparisonResult{
		Analysis: analysis,
		Profit:   profits,
	}, nil
}

func (service *DataComparisonService) CompareByMetadata(metadata models.DropsMetadata) (*models.ComparisonResult, *models.ComparisonResult, error) {
	itemPriceCache, err := caches.GetItemPriceCacheInstance()
	if err != nil {
		return nil, nil, err
	}
	defer itemPriceCache.Close()

	realDrops, err := service.EmpiricalDropsService.GetDropsByMetadata(metadata)
	if err != nil {
		return nil, nil, err
	}
	for _, drop := range realDrops {
		for _, item := range drop.Items {
			item.IndividualPrice = itemPriceCache.GetPrice(item.Name)
		}
	}

	var combinedRealDrops *models.BattledomeDrops
	if len(realDrops) > 0 {
		combinedRealDrops = helpers.Reduce(realDrops, func(first *models.BattledomeDrops, second *models.BattledomeDrops) *models.BattledomeDrops {
			combined, err := first.Union(second)
			if err != nil {
				panic(err)
			}
			return combined
		})
		combinedRealDrops.Metadata = realDrops[0].Metadata
		combinedRealDrops.Metadata.Source = "(multiple sources)"
	} else {
		combinedRealDrops = models.NewBattledomeDrops()
		combinedRealDrops.Metadata = models.DropsMetadataWithSource{
			Source:        "(none)",
			DropsMetadata: metadata,
		}
	}

	combinedGeneratedDrops, err := service.GeneratedDropsService.GenerateDrops(metadata.Arena)
	if err != nil {
		return nil, nil, err
	}

	realComparisonResult, err := service.ToComparisonResult(combinedRealDrops)
	if err != nil {
		return nil, nil, err
	}
	generatedComparisonResult, err := service.ToComparisonResult(combinedGeneratedDrops)
	if err != nil {
		return nil, nil, err
	}

	return realComparisonResult, generatedComparisonResult, nil
}

func (service *DataComparisonService) CompareArena(arena string) (*models.ComparisonResult, *models.ComparisonResult, error) {
	itemPriceCache, err := caches.GetItemPriceCacheInstance()
	if err != nil {
		return nil, nil, err
	}
	defer itemPriceCache.Close()

	realDrops, err := service.EmpiricalDropsService.GetDropsByArena(arena)
	if err != nil {
		return nil, nil, err
	}
	for _, drop := range realDrops {
		for _, item := range drop.Items {
			item.IndividualPrice = itemPriceCache.GetPrice(item.Name)
		}
	}

	var combinedRealDrops *models.BattledomeDrops
	if len(realDrops) > 0 {
		combinedRealDrops = helpers.Reduce(realDrops, func(first *models.BattledomeDrops, second *models.BattledomeDrops) *models.BattledomeDrops {
			combined, err := first.Union(second)
			if err != nil {
				panic(err)
			}
			return combined
		})
	} else {
		combinedRealDrops = models.NewBattledomeDrops()
		combinedRealDrops.Metadata = models.DropsMetadataWithSource{
			Source: "(none)",
			DropsMetadata: models.DropsMetadata{
				Arena:      arena,
				Challenger: "(none)",
				Difficulty: "(none)",
			},
		}
	}

	combinedGeneratedDrops, err := service.GeneratedDropsService.GenerateDrops(arena)
	if err != nil {
		return nil, nil, err
	}

	realComparisonResult, err := service.ToComparisonResult(combinedRealDrops)
	if err != nil {
		return nil, nil, err
	}
	generatedComparisonResult, err := service.ToComparisonResult(combinedGeneratedDrops)
	if err != nil {
		return nil, nil, err
	}

	return realComparisonResult, generatedComparisonResult, nil
}

type DataComparisonViewer struct{}

func NewDataComparisonViewer() *DataComparisonViewer {
	return &DataComparisonViewer{}
}

func getDryChance(dropRate float64, trials int) float64 {
	return math.Pow(1-dropRate, float64(trials))
}

func generateProfitableItemsTable(data *models.ComparisonResult, isRealData bool) *helpers.Table {
	headers := helpers.When(isRealData, []string{
		"i",
		"ItemName",
		"Drop Rate",
		// Don't include Dry Chance in real data
		"Price",
		"Expectation",
		"%",
	}, []string{
		"i",
		"ItemName",
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

	tableSeparator := "\t"

	lines := []string{}
	lines = append(lines, profitComparisonTable.GetLines()...)
	lines = append(lines, "\n")
	lines = append(lines, fmt.Sprintf("Top %d most profitable items", constants.NUMBER_OF_ITEMS_TO_PRINT))
	lines = append(lines, generatedProfitableItemsTable.GetLinesWith(tableSeparator, realProfitableItemsTable)...)
	lines = append(lines, "\n")
	lines = append(lines, fmt.Sprintf("Codestone drop rates"))
	lines = append(lines, brownCodestoneDropRatesTable.GetLinesWith(tableSeparator, redCodestoneDropRatesTable)...)
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
