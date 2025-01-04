package tests

import (
	"testing"

	"github.com/darienchong/neopets-battledome-analysis/constants"
	"github.com/darienchong/neopets-battledome-analysis/services"
)

func TestArenaDropsLogger(t *testing.T) {
	new(services.ArenaDropsLogger).Log(constants.BATTLEDOME_DROPS_FOLDER)
}
