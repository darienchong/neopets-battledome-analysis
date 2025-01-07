package tests

import (
	"testing"

	"github.com/darienchong/neopets-battledome-analysis/services"
)

func TestProfitEstimationLogger(t *testing.T) {
	services.NewProfitEstimationLogger().Log()
}
