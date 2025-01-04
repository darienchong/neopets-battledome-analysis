package tests

import (
	"testing"

	"github.com/darienchong/neopetsbattledomeanalysis/services"
)

func TestProfitEstimationLogger(t *testing.T) {
	new(services.ProfitEstimationLogger).Log()
}
