package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/darienchong/neopets-battledome-analysis/constants"
	"github.com/darienchong/neopets-battledome-analysis/loggers"
	"github.com/darienchong/neopets-battledome-analysis/models"
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

	args := os.Args[1:]
	if len(args) == 0 {
		panic(fmt.Errorf("please provide an argument (one of %s)", strings.Join(possibleArgs, ", ")))
	}
	dataFolderPath := strings.Replace(constants.BATTLEDOME_DROPS_FOLDER, "../", "", 1)
	switch args[0] {
	case possibleArgs[0]:
		loggers.NewArenaDropsLogger().Log(dataFolderPath)
	case possibleArgs[1]:
		loggers.NewDataComparisonLogger().CompareAllArenas()
	case possibleArgs[2]:
		loggers.NewDataComparisonLogger().CompareAllChallengers()
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

		loggers.NewDataComparisonLogger().CompareChallenger(models.DropsMetadata{
			Arena:      strings.ReplaceAll(args[1], "_", " "),
			Challenger: strings.ReplaceAll(args[2], "_", " "),
			Difficulty: strings.ReplaceAll(args[3], "_", " "),
		})
	default:
		panic(fmt.Errorf("please provide an argument (one of %s)", strings.Join(possibleArgs, ", ")))
	}

}
