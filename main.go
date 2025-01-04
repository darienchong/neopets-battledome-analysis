package main

import (
	"strings"

	"github.com/darienchong/neopetsbattledomeanalysis/constants"
	"github.com/darienchong/neopetsbattledomeanalysis/services"
)

func main() {
	dataFolderPath := strings.Replace(constants.BATTLEDOME_DROPS_FOLDER, "../", "", 1)
	switch constants.ACTION {
	case "AnalyseDrops":
		new(services.ArenaDropsLogger).Log(dataFolderPath)
	case "EstimateProfits":
		new(services.ProfitEstimationLogger).Log()
	case "Compare":
		services.NewArenaDataComparisonLogger().LogComparison(dataFolderPath)
	}
}
