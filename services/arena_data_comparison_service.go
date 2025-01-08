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

type ArenaComparisonData struct {
	Statistics *models.DropsStatistics
	Analysis   *models.DropsAnalysis
	Profit     map[string]*models.ItemProfit
}

type ArenaDataComparisonService struct {
	GeneratedDropsService *GeneratedDropsService
	EmpiricalDropsService *EmpiricalDropsService
	DropStatisticsService *DropStatisticsService
	DropsAnalysisService  *DropsAnalysisService
	DropRateService       *DropRateService
}

func NewArenaDataComparisonService() *ArenaDataComparisonService {
	return &ArenaDataComparisonService{
		GeneratedDropsService: NewGeneratedDropsService(),
		EmpiricalDropsService: NewEmpiricalDropsService(),
		DropStatisticsService: NewDropStatisticsService(),
		DropsAnalysisService:  NewDropsAnalysisService(),
		DropRateService:       NewDropRateService(),
	}
}

func (service *ArenaDataComparisonService) Compare(arena string) (*ArenaComparisonData, *ArenaComparisonData, error) {
	itemPriceCache, err := caches.GetItemPriceCacheInstance()
	if err != nil {
		return nil, nil, err
	}
	defer itemPriceCache.Close()

	realDrops, err := service.EmpiricalDropsService.GetDrops(arena)
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
		combinedRealDrops.Metadata = models.DropsMetadata{
			Arena: arena,
		}
	}

	combinedGeneratedDrops, err := service.GeneratedDropsService.GenerateDrops(arena)
	if err != nil {
		return nil, nil, err
	}

	realStats, err := service.DropStatisticsService.Estimate(combinedRealDrops)
	if err != nil {
		return nil, nil, err
	}
	generatedStats, err := service.DropStatisticsService.Estimate(combinedGeneratedDrops)
	if err != nil {
		return nil, nil, err
	}

	realDropRates, err := service.DropRateService.CalculateDropRates(realDrops)
	if err != nil {
		return nil, nil, err
	}
	realProfits := map[string]*models.ItemProfit{}
	for _, itemDropRate := range realDropRates[arena] {
		realProfits[itemDropRate.ItemName] = &models.ItemProfit{
			ItemDropRate:    *itemDropRate,
			IndividualPrice: itemPriceCache.GetPrice(itemDropRate.ItemName),
		}
	}
	generatedDropRates, err := service.DropRateService.CalculateDropRates([]*models.BattledomeDrops{combinedGeneratedDrops})
	if err != nil {
		return nil, nil, err
	}
	generatedProfits := map[string]*models.ItemProfit{}
	for _, itemDropRate := range generatedDropRates[arena] {
		generatedProfits[itemDropRate.ItemName] = &models.ItemProfit{
			ItemDropRate:    *itemDropRate,
			IndividualPrice: itemPriceCache.GetPrice(itemDropRate.ItemName),
		}
	}

	realAnalysisResult := service.DropsAnalysisService.Analyse(combinedRealDrops)
	generatedAnalysisResult := service.DropsAnalysisService.Analyse(combinedGeneratedDrops)

	return &ArenaComparisonData{
			Statistics: realStats,
			Analysis:   realAnalysisResult,
			Profit:     realProfits,
		}, &ArenaComparisonData{
			Statistics: generatedStats,
			Analysis:   generatedAnalysisResult,
			Profit:     generatedProfits,
		}, nil
}

type ArenaDataComparisonViewer struct{}

func NewArenaDataComparisonViewer() *ArenaDataComparisonViewer {
	return &ArenaDataComparisonViewer{}
}

func getDryChance(dropRate float64, trials int) float64 {
	return math.Pow(1-dropRate, float64(trials))
}

func generateProfitableItemsTable(data *ArenaComparisonData, isRealData bool) *helpers.Table {
	headers := helpers.When(isRealData, []string{
		"i",
		"ItemName",
		"Drop Rate",
		// Don't include Dry Chance in real data
		"Price",
		// "Expectation",
		"%",
	}, []string{
		"i",
		"ItemName",
		"Drop Rate",
		"Dry Chance",
		"Price",
		// "Expectation",
		"%",
	})
	table := helpers.NewNamedTable(helpers.When(isRealData, "Actual", "Predicted"), headers)

	predictedProfit := helpers.Sum(helpers.Map(helpers.Values(data.Profit), func(profit *models.ItemProfit) float64 { return profit.GetProfit() })) * constants.BATTLEDOME_DROPS_PER_DAY
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
				// helpers.FormatFloat(itemProfit.GetProfit()) + " NP",
				helpers.FormatPercentage(itemProfit.GetProfit()*constants.BATTLEDOME_DROPS_PER_DAY/predictedProfit) + "%",
			},
			[]string{
				strconv.Itoa(i + 1),
				itemProfit.ItemName,
				helpers.FormatPercentage(itemProfit.DropRate) + "%",
				helpers.FormatPercentage(getDryChance(itemProfit.DropRate, 30*constants.NUMBER_OF_ITEMS_TO_PRINT)) + "%",
				helpers.FormatFloat(itemProfit.IndividualPrice) + " NP",
				// helpers.FormatFloat(itemProfit.GetProfit()) + " NP",
				helpers.FormatPercentage(itemProfit.GetProfit()*constants.BATTLEDOME_DROPS_PER_DAY/predictedProfit) + "%",
			})

		table.AddRow(row)
	}

	return table
}

func generateCodestoneDropRatesTable(realData *ArenaComparisonData, generatedData *ArenaComparisonData, codestoneList []string) *helpers.Table {
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

func (viewer *ArenaDataComparisonViewer) View(realData *ArenaComparisonData, generatedData *ArenaComparisonData) ([]string, error) {
	arena := realData.Statistics.Arena

	profitComparisonTable := helpers.NewNamedTable(fmt.Sprintf("Profit in %s", arena), []string{
		"Type",
		"Value",
	})
	profitComparisonTable.IsLastRowDistinct = true
	profitComparisonTable.AddRow([]string{
		"Predicted",
		fmt.Sprintf("%s ± %s NP", helpers.FormatFloat(generatedData.Statistics.GetDropsProfitMean()), helpers.FormatFloat(generatedData.Statistics.GetDropsProfitStandardDeviation())),
	})
	profitComparisonTable.AddRow([]string{
		"Actual",
		fmt.Sprintf("%s ± %s NP", helpers.FormatFloat(realData.Statistics.GetDropsProfitMean()), helpers.FormatFloat(realData.Statistics.GetDropsProfitStandardDeviation())),
	})
	profitComparisonTable.AddRow([]string{
		"Difference",
		fmt.Sprintf("%s NP", helpers.FormatFloat(realData.Statistics.GetDropsProfitMean()-generatedData.Statistics.GetDropsProfitMean())),
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
