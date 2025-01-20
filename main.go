package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/darienchong/neopets-battledome-analysis/caches"
	"github.com/darienchong/neopets-battledome-analysis/constants"
	"github.com/darienchong/neopets-battledome-analysis/loggers"
	"github.com/darienchong/neopets-battledome-analysis/models"
	"github.com/palantir/stacktrace"
)

var (
	clear        map[string]func()
	possibleArgs = []string{
		"drops",
		"arenas",
		"challengers",
		"challenger",
	}
)

func init() {
	clear = make(map[string]func()) //Initialize it
	clear["linux"] = func() {
		cmd := exec.Command("clear") //Linux example, its tested
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
	clear["windows"] = func() {
		cmd := exec.Command("cmd", "/c", "cls") //Windows example, its tested
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}

func callClear() {
	value, ok := clear[runtime.GOOS] //runtime.GOOS -> linux, windows, darwin etc.
	if ok {                          //if we defined a clear func for that platform:
		value() //we execute it
	} else { //unsupported platform
		panic("Your platform is unsupported! I can't clear terminal screen :(")
	}
}

func main() {
	callClear()

	var dataSource caches.ItemPriceDataSource
	switch constants.ITEM_PRICE_DATA_SOURCE {
	case constants.JellyNeo:
		dataSource = caches.NewJellyNeoDataSource()
	case constants.ItemDb:
		dataSource = caches.NewItemDbDataSource()
	default:
		panic(stacktrace.NewError("unrecognised item price data source type: %d", constants.ITEM_PRICE_DATA_SOURCE))
	}
	cache, err := caches.GetItemPriceCacheInstance(dataSource)
	if err != nil {
		panic(stacktrace.Propagate(err, "failed to get item price cache instance"))
	}
	defer cache.Close()

	args := os.Args[1:]
	if len(args) == 0 {
		panic(fmt.Errorf("please provide an argument (one of %s)", strings.Join(possibleArgs, ", ")))
	}
	dataFolderPath := strings.Replace(constants.BATTLEDOME_DROPS_FOLDER, "../", "", 1)
	switch args[0] {
	case possibleArgs[0]:
		var numDropsToLog int64
		var err error
		numDropsToLog = constants.NUMBER_OF_DROPS_TO_PRINT
		if len(args) > 1 {
			numDropsToLog, err = strconv.ParseInt(args[1], 0, 64)
			if err != nil {
				panic(err)
			}
		}
		loggers.NewArenaDropsLogger().Log(dataFolderPath, int(numDropsToLog))
	case possibleArgs[1]:
		if len(args) > 1 && args[1] == "brief" {
			err := loggers.NewDataComparisonLogger().BriefCompareAllArenas()
			if err != nil {
				panic(err)
			}
		} else {
			err := loggers.NewDataComparisonLogger().CompareAllArenas()
			if err != nil {
				panic(err)
			}
		}
	case possibleArgs[2]:
		err := loggers.NewDataComparisonLogger().CompareAllChallengers()
		if err != nil {
			panic(err)
		}
	case possibleArgs[3]:
		if len(args) == 1 || args[1] == "" {
			panic(fmt.Errorf("please provide an arena"))
		}
		if len(args) == 2 || args[2] == "" {
			panic(fmt.Errorf("please provide a challenger"))
		}
		if len(args) == 3 || args[3] == "" {
			panic(fmt.Errorf("please provide a difficulty"))
		}

		err := loggers.NewDataComparisonLogger().CompareChallenger(models.BattledomeItemMetadata{
			Arena:      models.Arena(strings.ReplaceAll(args[1], "_", " ")),
			Challenger: models.Challenger(strings.ReplaceAll(args[2], "_", " ")),
			Difficulty: models.Difficulty(strings.ReplaceAll(args[3], "_", " ")),
		})
		if err != nil {
			panic(err)
		}
	default:
		panic(fmt.Errorf("please provide an argument (one of %s)", strings.Join(possibleArgs, ", ")))
	}
}
