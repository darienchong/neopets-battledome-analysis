package tests

import (
	"testing"

	"github.com/darienchong/neopets-battledome-analysis/constants"
	"github.com/darienchong/neopets-battledome-analysis/loggers"
)

func TestArenaDropsLogger(t *testing.T) {
	loggers.NewArenaDropsLogger().Log(constants.BATTLEDOME_DROPS_FOLDER)
}
