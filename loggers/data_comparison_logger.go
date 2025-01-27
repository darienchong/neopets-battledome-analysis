package loggers

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/darienchong/neopets-battledome-analysis/caches"
	"github.com/darienchong/neopets-battledome-analysis/constants"
	"github.com/darienchong/neopets-battledome-analysis/helpers"
	"github.com/darienchong/neopets-battledome-analysis/models"
	"github.com/darienchong/neopets-battledome-analysis/services"
	"github.com/darienchong/neopets-battledome-analysis/viewers"
	"github.com/palantir/stacktrace"
)

func prefix(indentLevel int) string {
	return strings.Repeat("  ", indentLevel)
}

type DataComparisonLogger struct {
	DataComparisonService *services.DataComparisonService
	DataComparisonViewer  *viewers.DataComparisonViewer
}

func NewDataComparisonLogger(dataComparisonService *services.DataComparisonService, dataComparisonViewer *viewers.DataComparisonViewer) *DataComparisonLogger {
	return &DataComparisonLogger{
		DataComparisonService: dataComparisonService,
		DataComparisonViewer:  dataComparisonViewer,
	}
}

func (l *DataComparisonLogger) BriefCompareAllArenas(itemPriceCache caches.ItemPriceCache) error {
	realData := map[models.Arena]models.NormalisedBattledomeItems{}
	generatedData := map[models.Arena]models.NormalisedBattledomeItems{}

	for _, arenaString := range constants.Arenas {
		arena := models.Arena(arenaString)
		realArenaData, generatedArenaData, err := l.DataComparisonService.CompareArena(arena)
		if err != nil {
			return stacktrace.Propagate(err, "failed to compare arena %q", arenaString)
		}
		realData[arena] = realArenaData
		generatedData[arena] = generatedArenaData
	}

	lines, err := l.DataComparisonViewer.ViewBriefArenaComparisons(itemPriceCache, realData, generatedData)
	if err != nil {
		return stacktrace.Propagate(err, "failed to generate brief arena comparisons")
	}

	for _, line := range lines {
		slog.Info(line)
	}

	return nil
}

func (l *DataComparisonLogger) CompareAllArenas(itemPriceCache caches.ItemPriceCache) error {
	comparisonData := helpers.OrderByDescending(
		helpers.Map(
			constants.Arenas,
			func(arena string) *helpers.Tuple {
				realData, generatedData, err := l.DataComparisonService.CompareArena(models.Arena(arena))
				if err != nil {
					panic(stacktrace.Propagate(err, "failed to compare %s", arena))
				}
				return &helpers.Tuple{Elements: []any{models.Arena(arena), realData, generatedData}}
			},
		),
		func(tuple *helpers.Tuple) float64 {
			realData := tuple.Elements[1].(models.NormalisedBattledomeItems)
			generatedData := tuple.Elements[2].(models.NormalisedBattledomeItems)
			var profit float64 = 0.0
			var err error
			if constants.ShouldIgnoreChallengerDropsInArenaComparison {
				profit, err = realData.ArenaMeanDropsProfit(itemPriceCache, generatedData)
			} else {
				profit, err = realData.MeanDropsProfit(itemPriceCache)
			}
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
		lines, err := l.DataComparisonViewer.ViewArenaComparison(itemPriceCache, realData, generatedData)
		if err != nil {
			return stacktrace.Propagate(err, "failed to get arena comparison for %q", arena)
		}

		slog.Info(fmt.Sprintf("%d. %s (%d samples)", i+1, arena, realData.TotalItemQuantity()))
		for _, line := range lines {
			slog.Info(prefix(1) + line)
		}
		slog.Info("\n\n")
	}

	return nil
}

func (l *DataComparisonLogger) CompareChallenger(itemPriceCache caches.ItemPriceCache, metadata models.BattledomeItemMetadata) error {
	realData, generatedData, err := l.DataComparisonService.CompareByMetadata(metadata)
	if err != nil {
		return stacktrace.Propagate(err, "failed to generate metadata comparison for %q", metadata)
	}
	lines, err := l.DataComparisonViewer.ViewChallengerComparison(itemPriceCache, realData, generatedData)
	if err != nil {
		return stacktrace.Propagate(err, "failed to generate challenger comparison for %q", metadata)
	}

	for _, line := range lines {
		slog.Info(line)
	}
	return nil
}

func (l *DataComparisonLogger) CompareAllChallengers(itemPriceCache caches.ItemPriceCache) error {
	data, err := l.DataComparisonService.CompareAllChallengers(itemPriceCache)
	if err != nil {
		return stacktrace.Propagate(err, "failed to compare all challengers")
	}

	lines, err := l.DataComparisonViewer.ViewChallengerComparisons(itemPriceCache, data)
	if err != nil {
		return stacktrace.Propagate(err, "failed to generate challenger comparison view")
	}

	for _, line := range lines {
		slog.Info(line)
	}

	return nil
}
