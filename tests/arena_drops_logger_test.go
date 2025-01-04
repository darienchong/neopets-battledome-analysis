package tests

import (
	"testing"

	"github.com/darienchong/neopetsbattledomeanalysis/constants"
	"github.com/darienchong/neopetsbattledomeanalysis/services"
)

func TestArenaDropsLogger(t *testing.T) {
	new(services.ArenaDropsLogger).Log(constants.BATTLEDOME_DROPS_FOLDER)
}
