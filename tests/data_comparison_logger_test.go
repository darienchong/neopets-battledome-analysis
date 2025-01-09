package tests

import (
	"testing"

	"github.com/darienchong/neopets-battledome-analysis/loggers"
	"github.com/darienchong/neopets-battledome-analysis/models"
)

func TestArenaDataComparison(t *testing.T) {
	loggers.NewDataComparisonLogger().CompareAllArenas()
}

func TestChallengerDataComparison(t *testing.T) {
	loggers.NewDataComparisonLogger().CompareChallenger(models.DropsMetadata{
		Arena:      "Dome of the Deep",
		Challenger: "Koi Warrior",
		Difficulty: "Average",
	})
}

func TestChallengersDataComparison(t *testing.T) {
	loggers.NewDataComparisonLogger().CompareAllChallengers()
}
