package tests

import (
	"math"
	"testing"

	"github.com/darienchong/neopets-battledome-analysis/constants"
	"github.com/darienchong/neopets-battledome-analysis/helpers"
	"github.com/darienchong/neopets-battledome-analysis/parsers"
	"github.com/darienchong/neopets-battledome-analysis/services"
)

func TestDropRateCalculation(t *testing.T) {
	acceptableDiff := math.Pow(10, -5)
	parser := parsers.NewGeneratedDropsParser()
	target := services.NewDropRateService()

	arena := "Neocola Centre"
	rawData, err := parser.Parse(constants.GetGeneratedDropsFilePath(arena))
	if err != nil {
		t.Fatalf("%s", err)
	}
	generatedDrops := rawData[arena]
	totalItemCount := float64(generatedDrops.GetTotalItemQuantity())

	predictedDropRates, err := target.GetPredictedDropRates(arena)
	if err != nil {
		t.Fatalf("%s", err)
	}

	for _, predictedDropRate := range predictedDropRates {
		actualDropRate := float64(generatedDrops.Items[predictedDropRate.ItemName].Quantity) / totalItemCount
		if math.Abs(actualDropRate-predictedDropRate.DropRate) > acceptableDiff {
			t.Fatalf("%s's drop rate was miscalculated:\nExpected: %s%%\nActual: %s%%", predictedDropRate.ItemName, helpers.FormatPercentage(actualDropRate), helpers.FormatPercentage(predictedDropRate.DropRate))
		}
	}
}
