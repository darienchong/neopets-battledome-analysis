package constants

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
)

type ItemPriceDataSourceType int

func (i ItemPriceDataSourceType) String() string {
	switch i {
	case JellyNeo:
		return "JellyNeo"
	case ItemDB:
		return "ItemDB"
	default:
		return "?"
	}
}

const (
	Unknown ItemPriceDataSourceType = iota
	JellyNeo
	ItemDB
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
	// Additional prizes not covered by the initial drop estimations
	// e.g. Central Arena had Nerkmids added to their drop pool
	AdditionalArenaSpecificDrops = map[string]map[string]any{
		"Cosmic Dome":    {},
		"Neocola Centre": {},
		"Central Arena": {
			"Nimmo Battle Cry":        nil,
			"Aluminium Nerkmid":       nil,
			"Basic Golden Nerkmid":    nil,
			"Copper Nerkmid":          nil,
			"Golden Nerkmid X":        nil,
			"Golden Nerkmid XX":       nil,
			"Good Nerkmid":            nil,
			"Lesser Nerkmid":          nil,
			"Magical Golden Nerkmid":  nil,
			"Normal Golden Nerkmid":   nil,
			"Normal Platinum Nerkmid": nil,
			"Platinum Nerkmid X":      nil,
			"Platinum Nerkmid XX":     nil,
			"Ultimate Nerkmid":        nil,
			"Ultra Golden Nerkmid":    nil,
		},
		"Dome of the Deep":  {},
		"Rattling Cauldron": {},
		"Pango Palladium":   {},
		"Frost Arena":       {},
		"Ugga Dome":         {},
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
