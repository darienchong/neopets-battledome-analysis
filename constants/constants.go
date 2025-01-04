package constants

import (
	"fmt"
	"path/filepath"
	"runtime"
)

const (
	DATA_FOLDER                                                  = "./../data/"
	ITEM_PRICE_CACHE_FILE                                        = "neopets_item_price_cache.txt"
	ITEW_WEIGHTS_FILE                                            = "neopets_battledome_item_weights.txt"
	ITEM_DROP_RATES_FILE_TEMPLATE                                = "neopets_battledome_item_drop_rates_%d.txt"
	GENERATED_DROPS_FILE_TEMPLATE                                = "neopets_battledome_generated_items_%d.txt"
	DATA_EXPIRY_TIME_LAYOUT                                      = "2006-01-02 15:04:05.000000"
	TIME_LAYOUT                                                  = "2006/01/02 15:04:05"
	BATTLEDOME_DROPS_FOLDER                                      = "./../battledome_drop_data/"
	FLOAT_FORMAT_LAYOUT                                          = "#,###."
	PERCENTAGE_FORMAT_LAYOUT                                     = "#,###.##"
	NUMBER_OF_ITEMS_TO_PRINT                                     = 10
	BATTLEDOME_DROPS_PER_DAY                                     = 15
	NUMBER_OF_ITEMS_TO_GENERATE_FOR_ESTIMATING_DROP_RATES        = 100_000_000
	NUMBER_OF_ITEMS_TO_GENERATE_FOR_ESTIMATING_PROFIT_STATISTICS = 100_000_000

	FILTER_ARENA = ""
	ACTION       = "Compare"
)

var (
	BROWN_CODESTONES = []string{
		"Bri Codestone",
		"Eo Codestone",
		"Har Codestone",
		"Lu Codestone",
		"Main Codestone",
		"Mau Codestone",
		"Orn Codestone",
		"Tai-Kai Codestone",
		"Vo Codestone",
		"Zei Codestone",
	}
	RED_CODESTONES = []string{
		"Cui Codestone",
		"Kew Codestone",
		"Mag Codestone",
		"Sho Codestone",
		"Vux Codestone",
		"Zed Codestone",
	}
	ARENAS = []string{
		"Cosmic Dome",
		"Neocola Centre",
		"Central Arena",
		"Dome of the Deep",
		"Rattling Cauldron",
		"Pango Palladium",
		"Frost Arena",
		"Ugga Dome",
	}
)

func combineRelativeFolderAndFilename(folder string, fileName string) string {
	_, b, _, _ := runtime.Caller(0)
	exPath := filepath.Dir(b)
	return filepath.Join(exPath, folder, fileName)
}

func GetItemPriceCacheFilePath() string {
	return combineRelativeFolderAndFilename(DATA_FOLDER, ITEM_PRICE_CACHE_FILE)
}

func GetDropDataFilePath(fileName string) string {
	return combineRelativeFolderAndFilename(BATTLEDOME_DROPS_FOLDER, fileName)
}

func GetItemWeightsFilePath() string {
	return combineRelativeFolderAndFilename(DATA_FOLDER, ITEW_WEIGHTS_FILE)
}

func GetDropRatesFilePath() string {
	return combineRelativeFolderAndFilename(DATA_FOLDER, fmt.Sprintf(ITEM_DROP_RATES_FILE_TEMPLATE, NUMBER_OF_ITEMS_TO_GENERATE_FOR_ESTIMATING_DROP_RATES))
}

func GetGeneratedDropsFilePath() string {
	return combineRelativeFolderAndFilename(DATA_FOLDER, fmt.Sprintf(GENERATED_DROPS_FILE_TEMPLATE, NUMBER_OF_ITEMS_TO_GENERATE_FOR_ESTIMATING_PROFIT_STATISTICS))
}
