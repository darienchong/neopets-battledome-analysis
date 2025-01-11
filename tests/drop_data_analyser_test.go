package tests

import (
	"log/slog"
	"testing"

	"github.com/darienchong/neopets-battledome-analysis/constants"
	"github.com/darienchong/neopets-battledome-analysis/models"
	"github.com/darienchong/neopets-battledome-analysis/parsers"
	"github.com/darienchong/neopets-battledome-analysis/services"
)

func TestAnalyser(t *testing.T) {
	target := services.NewDropsAnalysisService()
	parser := parsers.NewDropDataParser()
	dropsDto, err := parser.Parse(constants.GetDropDataFilePath("2024_12_20.txt"))
	if err != nil {
		slog.Any("error", err)
		t.Fatalf("Failed to parse drop file!")
	}
	res := target.Analyse(dropsDto.ToBattledomeDrops())
	prev := &models.BattledomeItem{}
	for idx, item := range res.GetItemsOrderedByPrice() {
		if idx != 0 {
			if prev.IndividualPrice < item.IndividualPrice {
				t.Fatalf("GetItemsOrderedByPrice did not return a list of items ordered by price in descending order: \"%s\" came before \"%s\".", prev, item)
			}
		}
		prev = item
	}
	for idx, item := range res.GetItemsOrderedByProfit() {
		if idx != 0 {
			prevProfit, prevErr := prev.GetProfit()
			currProfit, currErr := item.GetProfit()
			if prevErr != nil || currErr != nil {
				if prevErr != nil {
					slog.Any("error", prevErr)
				} else {
					slog.Any("error", currErr)
				}

				t.Fatalf("Failed to get profit for either the previous element (%s), or the current element (%s).", prev, item)
			}
			if prevProfit < currProfit {
				t.Fatalf("GetItemsOrderedByProfit did not return a list of items ordered by price in descending order: \"%s\" came before \"%s\"", prev, item)
			}
		}
		prev = item
	}
}
