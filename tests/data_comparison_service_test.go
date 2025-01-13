package tests

import (
	"log/slog"
	"testing"

	"github.com/darienchong/neopets-battledome-analysis/models"
	"github.com/darienchong/neopets-battledome-analysis/services"
	"github.com/darienchong/neopets-battledome-analysis/viewers"
)

func TestArenaView(t *testing.T) {
	svc := services.NewDataComparisonService()
	target := viewers.NewDataComparisonViewer()

	realData, generatedData, err := svc.CompareArena("Neocola Centre")
	if err != nil {
		t.Fatalf("%s", err)
	}
	lines, err := target.ViewArenaComparison(realData, generatedData)
	if err != nil {
		t.Fatalf("%s", err)
	}

	for _, line := range lines {
		slog.Info(line)
	}
}

func TestChallengerView(t *testing.T) {
	svc := services.NewDataComparisonService()
	target := viewers.NewDataComparisonViewer()

	metadata := models.BattledomeItemMetadata{
		Arena:      "Central Arena",
		Challenger: "Kasuki Lu",
		Difficulty: "Mighty",
	}
	realData, generatedData, err := svc.CompareByMetadata(metadata)
	if err != nil {
		t.Fatalf("%s", err)
	}

	realMetadata, err := realData.GetMetadata()
	if err != nil {
		t.Fatalf("%s", err)
	}
	if realMetadata.Arena != metadata.Arena || realMetadata.Challenger != metadata.Challenger || realMetadata.Difficulty != metadata.Difficulty {
		t.Fatalf("real data's metadata did not match expected\nExpected: %s\nGot: %s", &metadata, &realMetadata)
	}

	if err != nil {
		t.Fatalf("%s", err)
	}
	lines, err := target.ViewChallengerComparison(realData, generatedData)
	if err != nil {
		t.Fatalf("%s", err)
	}

	for _, line := range lines {
		slog.Info(line)
	}
}

func TestChallengersView(t *testing.T) {
	svc := services.NewDataComparisonService()
	target := viewers.NewDataComparisonViewer()

	challengerData, err := svc.CompareAllChallengers()
	if err != nil {
		t.Fatalf("%s", err)
	}

	lines, err := target.ViewChallengerComparisons(challengerData)
	if err != nil {
		t.Fatalf("%s", err)
	}

	for _, line := range lines {
		slog.Info(line)
	}
}
