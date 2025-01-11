package loggers

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/darienchong/neopets-battledome-analysis/constants"
	"github.com/darienchong/neopets-battledome-analysis/helpers"
	"github.com/darienchong/neopets-battledome-analysis/models"
	"github.com/darienchong/neopets-battledome-analysis/services"
	"github.com/darienchong/neopets-battledome-analysis/viewers"
)

func getPrefix(indentLevel int) string {
	return strings.Repeat("  ", indentLevel)
}

type DataComparisonLogger struct {
	DataComparisonService *services.DataComparisonService
	DataComparisonViewer  *viewers.DataComparisonViewer
}

func NewDataComparisonLogger() *DataComparisonLogger {
	return &DataComparisonLogger{
		DataComparisonService: services.NewDataComparisonService(),
		DataComparisonViewer:  viewers.NewDataComparisonViewer(),
	}
}

func (logger *DataComparisonLogger) CompareAllArenas() error {
	comparisonData := helpers.OrderByDescending(
		helpers.Map(
			constants.ARENAS,
			func(arena string) *helpers.Tuple {
				realData, generatedData, err := logger.DataComparisonService.CompareArena(models.Arena(arena))
				if err != nil {
					panic(err)
				}
				return &helpers.Tuple{Elements: []any{realData, generatedData}}
			},
		),
		func(tuple *helpers.Tuple) float64 {
			realData := tuple.Elements[0].(models.NormalisedBattledomeItems)
			profit, err := realData.GetMeanDropsProfit()
			if err != nil {
				return 0
			}
			return profit
		},
	)

	for i, comparisonDatum := range comparisonData {
		realData := comparisonDatum.Elements[0].(models.NormalisedBattledomeItems)
		generatedData := comparisonDatum.Elements[1].(models.NormalisedBattledomeItems)
		lines, err := logger.DataComparisonViewer.ViewArenaComparison(realData, generatedData)
		if err != nil {
			return err
		}

		realItemCount := helpers.Sum(helpers.Map(helpers.Values(realData), func(item *models.BattledomeItem) int32 {
			return helpers.When(item.Name == "nothing", 0, item.Quantity)
		}))

		metadata, err := realData.GetMetadata()
		if err != nil {
			return err
		}
		slog.Info(fmt.Sprintf("%d. %s (%d samples)", i+1, metadata.Arena, realItemCount))
		for _, line := range lines {
			slog.Info(getPrefix(1) + line)
		}
		slog.Info("\n\n")
	}

	return nil
}

func (logger *DataComparisonLogger) CompareChallenger(metadata models.BattledomeItemMetadata) error {
	realData, generatedData, err := logger.DataComparisonService.CompareByMetadata(metadata)
	if err != nil {
		return err
	}
	lines, err := logger.DataComparisonViewer.ViewChallengerComparison(realData, generatedData)
	if err != nil {
		return err
	}

	for _, line := range lines {
		slog.Info(line)
	}
	return nil
}

func (logger *DataComparisonLogger) CompareAllChallengers() error {
	data, err := logger.DataComparisonService.CompareAllChallengers()
	if err != nil {
		return err
	}

	lines, err := logger.DataComparisonViewer.ViewChallengerComparisons(data)
	if err != nil {
		return err
	}

	for _, line := range lines {
		slog.Info(line)
	}

	return nil
}
