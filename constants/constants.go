package constants

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
)

type ItemPriceDataSourceType int

const (
	Unknown ItemPriceDataSourceType = iota
	JellyNeo
	ItemDb
)

const (
	ItemPriceDataSource            = JellyNeo
	DataFolder                     = "./../data/"
	ItemDBItemPriceCacheFile       = "neopets_itemdb_item_price_cache.txt"
	JellyNeoItemPriceCacheFile     = "neopets_jellyneo_item_price_cache.txt"
	ItemWeightsFileName            = "neopets_battledome_item_weights.txt"
	ItemDropRatesFileNameTemplate  = "neopets_battledome_item_drop_rates_%s_%d.txt"
	GeneratedDropsFileNameTemplate = "neopets_battledome_generated_items_%s_%d.txt"
	DataExpiryTimeLayout           = "2006-01-02 15:04:05.000000"
	TimeLayout                     = "2006/01/02 15:04:05"
	BattledomeDropsFolder          = "./../battledome_drop_data/"
	FloatFormatLayout              = "#,###."
	PercentageFormatLayout         = "#,###.##"
	NumberOfItemsToPrint           = 15
	BattledomeDropsPerDay          = 15
	NumberOfItemsToGenerate        = 100_000_000
	SignificanceLevel              = 0.05
	NumberOfBootstrapSamples       = 100_000

	FilterArena                                  = ""
	NumberOfDropsToPrint                         = 3
	ShouldIgnoreChallengerDropsInArenaComparison = true
)

var (
	BrownCodestones = []string{
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
	RedCodestones = []string{
		"Cui Codestone",
		"Kew Codestone",
		"Mag Codestone",
		"Sho Codestone",
		"Vux Codestone",
		"Zed Codestone",
	}
	Arenas = []string{
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

func CombineRelativeFolderAndFilename(folder string, fileName string) string {
	_, b, _, _ := runtime.Caller(0)
	exPath := filepath.Dir(b)
	return filepath.Join(exPath, folder, fileName)
}

func DropDataFilePath(fileName string) string {
	return CombineRelativeFolderAndFilename(BattledomeDropsFolder, fileName)
}

func ItemWeightsFilePath() string {
	return CombineRelativeFolderAndFilename(DataFolder, ItemWeightsFileName)
}

func DropRatesFilePath(arena string) string {
	return CombineRelativeFolderAndFilename(DataFolder, fmt.Sprintf(ItemDropRatesFileNameTemplate, strings.ReplaceAll(arena, " ", "_"), NumberOfItemsToGenerate))
}

func GeneratedDropsFilePath(arena string) string {
	return CombineRelativeFolderAndFilename(DataFolder, fmt.Sprintf(GeneratedDropsFileNameTemplate, strings.ReplaceAll(arena, " ", "_"), NumberOfItemsToGenerate))
}
