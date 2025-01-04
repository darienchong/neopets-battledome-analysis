package tests

import (
	"testing"

	"github.com/darienchong/neopetsbattledomeanalysis/constants"
	"github.com/darienchong/neopetsbattledomeanalysis/services"
)

func TestArenaDataComparisonLogger(t *testing.T) {
	services.NewArenaDataComparisonLogger().LogComparison(constants.BATTLEDOME_DROPS_FOLDER)
}
