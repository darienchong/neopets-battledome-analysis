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
	"github.com/palantir/stacktrace"
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

func (logger *DataComparisonLogger) BriefCompareAllArenas() error {
	realData := map[models.Arena]models.NormalisedBattledomeItems{}
	generatedData := map[models.Arena]models.NormalisedBattledomeItems{}

	for _, arenaString := range constants.ARENAS {
		arena := models.Arena(arenaString)
		realArenaData, generatedArenaData, err := logger.DataComparisonService.CompareArena(arena)
		if err != nil {
			return stacktrace.Propagate(err, "failed to compare arena \"%s\"", arenaString)
		}
		realData[arena] = realArenaData
		generatedData[arena] = generatedArenaData
	}

	lines, err := logger.DataComparisonViewer.ViewBriefArenaComparisons(realData, generatedData)
	if err != nil {
		return stacktrace.Propagate(err, "failed to generate brief arena comparisons")
	}

	for _, line := range lines {
		slog.Info(line)
	}

	return nil
}

func (logger *DataComparisonLogger) CompareAllArenas() error {
	comparisonData := helpers.OrderByDescending(
		helpers.Map(
			constants.ARENAS,
			func(arena string) *helpers.Tuple {
				realData, generatedData, err := logger.DataComparisonService.CompareArena(models.Arena(arena))
				if err != nil {
					panic(stacktrace.Propagate(err, "failed to compare %s", arena))
				}
				return &helpers.Tuple{Elements: []any{models.Arena(arena), realData, generatedData}}
			},
		),
		func(tuple *helpers.Tuple) float64 {
			realData := tuple.Elements[1].(models.NormalisedBattledomeItems)
			generatedData := tuple.Elements[2].(models.NormalisedBattledomeItems)
			profit, err := helpers.LazyWhenError(
				constants.SHOULD_IGNORE_CHALLENGER_DROPS_IN_ARENA_COMPARISON,
				func() (float64, error) { return realData.GetArenaMeanDropsProfit(generatedData) },
				func() (float64, error) { return realData.GetMeanDropsProfit() },
			)
			if err != nil {
				return 0
			}
			return profit
		},
	)

	for i, comparisonDatum := range comparisonData {
		arena := comparisonDatum.Elements[0].(models.Arena)
		realData := comparisonDatum.Elements[1].(models.NormalisedBattledomeItems)
		generatedData := comparisonDatum.Elements[2].(models.NormalisedBattledomeItems)
		lines, err := logger.DataComparisonViewer.ViewArenaComparison(realData, generatedData)
		if err != nil {
			return stacktrace.Propagate(err, "failed to get arena comparison for \"%s\"", arena)
		}

		slog.Info(fmt.Sprintf("%d. %s (%d samples)", i+1, arena, realData.GetTotalItemQuantity()))
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
		return stacktrace.Propagate(err, "failed to generate metadata comparison for \"%s\"", metadata)
	}
	lines, err := logger.DataComparisonViewer.ViewChallengerComparison(realData, generatedData)
	if err != nil {
		return stacktrace.Propagate(err, "failed to generate challenger comparison for \"%s\"", metadata)
	}

	for _, line := range lines {
		slog.Info(line)
	}
	return nil
}

func (logger *DataComparisonLogger) CompareAllChallengers() error {
	data, err := logger.DataComparisonService.CompareAllChallengers()
	if err != nil {
		return stacktrace.Propagate(err, "failed to compare all challengers")
	}

	lines, err := logger.DataComparisonViewer.ViewChallengerComparisons(data)
	if err != nil {
		return stacktrace.Propagate(err, "failed to generate challenger comparison view")
	}

	for _, line := range lines {
		slog.Info(line)
	}

	return nil
}
