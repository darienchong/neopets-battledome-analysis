package tests

import (
	"testing"

	"github.com/darienchong/neopets-battledome-analysis/services"
)

func TestProfitEstimationLogger(t *testing.T) {
	new(services.ProfitEstimationLogger).Log()
}
