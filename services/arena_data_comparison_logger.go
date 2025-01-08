package services

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/darienchong/neopets-battledome-analysis/constants"
	"github.com/darienchong/neopets-battledome-analysis/helpers"
)

func getPrefix(indentLevel int) string {
	return strings.Repeat("  ", indentLevel)
}

type ArenaDataComparisonLogger struct {
	ArenaDataComparisonService *ArenaDataComparisonService
	ArenaDataComparisonViewer  *ArenaDataComparisonViewer
}

func NewArenaDataComparisonLogger() *ArenaDataComparisonLogger {
	return &ArenaDataComparisonLogger{
		ArenaDataComparisonService: NewArenaDataComparisonService(),
		ArenaDataComparisonViewer:  NewArenaDataComparisonViewer(),
	}
}

func (logger *ArenaDataComparisonLogger) CompareAll() error {
	comparisonData := helpers.OrderByDescending(
		helpers.Map(
			constants.ARENAS,
			func(arena string) *helpers.Tuple {
				realData, generatedData, err := logger.ArenaDataComparisonService.Compare(arena)
				if err != nil {
					panic(err)
				}
				return &helpers.Tuple{Elements: []any{realData, generatedData}}
			},
		),
		func(tuple *helpers.Tuple) float64 {
			realData := tuple.Elements[0].(*ArenaComparisonData)
			return realData.Statistics.GetDropsProfitMean()
		},
	)

	for i, comparisonDatum := range comparisonData {
		realData := comparisonDatum.Elements[0].(*ArenaComparisonData)
		generatedData := comparisonDatum.Elements[1].(*ArenaComparisonData)
		lines, err := logger.ArenaDataComparisonViewer.View(realData, generatedData)
		if err != nil {
			return err
		}

		slog.Info(fmt.Sprintf("%d. %s", i+1, realData.Statistics.Arena))
		for _, line := range lines {
			slog.Info(getPrefix(1) + line)
		}
		slog.Info("\n\n")
	}

	return nil
}
