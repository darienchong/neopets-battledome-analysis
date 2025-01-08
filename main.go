package main

import (
	"strings"

	"github.com/darienchong/neopets-battledome-analysis/constants"
	"github.com/darienchong/neopets-battledome-analysis/services"
)

func main() {
	dataFolderPath := strings.Replace(constants.BATTLEDOME_DROPS_FOLDER, "../", "", 1)
	switch constants.ACTION {
	case "AnalyseDrops":
		services.NewArenaDropsLogger().Log(dataFolderPath)
	case "Compare":
		services.NewDataComparisonLogger().CompareAllArenas()
	}
}
