package tests

import (
	"log/slog"
	"testing"

	"github.com/darienchong/neopets-battledome-analysis/services"
)

func TestChallengerComparison(t *testing.T) {
	svc := services.NewChallengerComparisonService()
	target := services.NewChallengerComparisonViewer()

	data, err := svc.CompareAll()
	if err != nil {
		panic(err)
	}

	lines, err := target.View(data)
	if err != nil {
		panic(err)
	}

	for _, line := range lines {
		slog.Info(line)
	}
}
