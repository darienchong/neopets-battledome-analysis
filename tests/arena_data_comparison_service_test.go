package tests

import (
	"log/slog"
	"testing"

	"github.com/darienchong/neopets-battledome-analysis/services"
)

func TestView(t *testing.T) {
	svc := services.NewArenaDataComparisonService()
	target := services.NewArenaDataComparisonViewer()

	realData, generatedData, err := svc.Compare("Neocola Centre")
	if err != nil {
		panic(err)
	}
	lines, err := target.View(realData, generatedData)
	if err != nil {
		panic(err)
	}

	for _, line := range lines {
		slog.Info(line)
	}
}
