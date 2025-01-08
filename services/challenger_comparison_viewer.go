package services

import (
	"strconv"

	"github.com/darienchong/neopets-battledome-analysis/helpers"
	"github.com/darienchong/neopets-battledome-analysis/models"
)

type ChallengerComparisonViewer struct {
	GeneratedDropsService *GeneratedDropsService
	DataComparisonService *DataComparisonService
}

func NewChallengerComparisonViewer() *ChallengerComparisonViewer {
	return &ChallengerComparisonViewer{
		GeneratedDropsService: NewGeneratedDropsService(),
		DataComparisonService: NewDataComparisonService(),
	}
}

func (viewer *ChallengerComparisonViewer) View(comparisonResults []*models.ComparisonResult) ([]string, error) {
	comparisonResults = helpers.OrderByDescending(comparisonResults, func(res *models.ComparisonResult) float64 {
		meanDropsProfit, err := res.Analysis.GetMeanDropsProfit()
		if err != nil {
			panic(err)
		}
		return meanDropsProfit
	})

	profitComparisonTable := helpers.NewTable([]string{
		"i",
		"Arena",
		"Challenger",
		"Difficulty",
		"Samples",
		"Actual Profit",
		"Predicted Profit",
		"Actual Stdev",
		"Predicted Stdev",
	})

	arenaAndChallengerDropRateTable := helpers.NewTable([]string{
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
			helpers.FormatFloat(actualProfit) + " NP",
			helpers.FormatFloat(generatedProfit) + " NP",
			helpers.FormatFloat(actualStdev) + " NP",
			helpers.FormatFloat(generatedStdev) + " NP",
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
	lines = append(lines, "\n\n")
	lines = append(lines, "Drop Rate Comparison")
	lines = append(lines, arenaAndChallengerDropRateTable.GetLines()...)
	return lines, nil
}
