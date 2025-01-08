package tests

import (
	"log/slog"
	"testing"

	"github.com/darienchong/neopets-battledome-analysis/models"
	"github.com/darienchong/neopets-battledome-analysis/services"
)

func TestArenaView(t *testing.T) {
	svc := services.NewDataComparisonService()
	target := services.NewDataComparisonViewer()

	realData, generatedData, err := svc.CompareArena("Neocola Centre")
	if err != nil {
		panic(err)
	}
	lines, err := target.ViewArenaComparison(realData, generatedData)
	if err != nil {
		panic(err)
	}

	for _, line := range lines {
		slog.Info(line)
	}
}

func TestChallengerView(t *testing.T) {
	svc := services.NewDataComparisonService()
	target := services.NewDataComparisonViewer()

	metadata := models.DropsMetadata{
		Arena:      "Central Arena",
		Challenger: "Flaming Meerca",
		Difficulty: "Mighty",
	}
	realData, generatedData, err := svc.CompareByMetadata(metadata)

	if realData.Analysis.Metadata.Arena != metadata.Arena || realData.Analysis.Metadata.Challenger != metadata.Challenger || realData.Analysis.Metadata.Difficulty != metadata.Difficulty {
		t.Fatalf("real data's metadata did not match expected\nExpected: %s\nGot: %s", &metadata, &realData.Analysis.Metadata)
	}

	if err != nil {
		panic(err)
	}
	lines, err := target.ViewChallengerComparison(realData, generatedData)
	if err != nil {
		panic(err)
	}

	for _, line := range lines {
		slog.Info(line)
	}
}
