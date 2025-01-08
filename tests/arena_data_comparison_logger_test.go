package tests

import (
	"testing"

	"github.com/darienchong/neopets-battledome-analysis/services"
)

func TestArenaDataComparisonLogger(t *testing.T) {
	services.NewArenaDataComparisonLogger().CompareAll()
}
