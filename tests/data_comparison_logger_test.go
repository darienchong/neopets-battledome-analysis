//go:build ignore

package tests

import (
	"testing"

	"github.com/darienchong/neopets-battledome-analysis/loggers"
	"github.com/darienchong/neopets-battledome-analysis/models"
)

func TestArenaDataComparison(t *testing.T) {
	loggers.NewDataComparisonLogger().CompareAllArenas()
}

func TestBriefArenaDataComparison(t *testing.T) {
	loggers.NewDataComparisonLogger().BriefCompareAllArenas()
}

func TestChallengerDataComparison(t *testing.T) {
	loggers.NewDataComparisonLogger().CompareChallenger(models.BattledomeItemMetadata{
		Arena:      "Dome of the Deep",
		Challenger: "Koi Warrior",
		Difficulty: "Average",
	})
}

func TestChallengersDataComparison(t *testing.T) {
	loggers.NewDataComparisonLogger().CompareAllChallengers()
}
