package parsers

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/darienchong/neopets-battledome-analysis/helpers"
	"github.com/darienchong/neopets-battledome-analysis/models"
	"github.com/palantir/stacktrace"
)

type BattledomeItemWeightParser struct{}

func NewBattledomeItemWeightParser() *BattledomeItemWeightParser {
	return &BattledomeItemWeightParser{}
}

func (p *BattledomeItemWeightParser) Parse(filePath string) ([]models.BattledomeItemWeight, error) {
	if !helpers.IsFileExists(filePath) {
		return nil, fmt.Errorf("item weights file does not exist: %s", filePath)
	}

	currentArena := ""
	weights := []models.BattledomeItemWeight{}
	file, err := os.OpenFile(filePath, os.O_RDONLY, 0755)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to open file: \"%s\"", filePath)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		} else if strings.Contains(line, "-") {
			// It's an item weight
			if currentArena == "" {
				return nil, fmt.Errorf("read an item weight before an arena was read! The offending line was \"%s\"", line)
			}
			tokens := strings.Split(line, " - ")
			itemName := strings.TrimSpace(tokens[0])
			itemWeightString := strings.TrimSpace(strings.ReplaceAll(tokens[1], "%", ""))
			parsedItemWeight, err := strconv.ParseFloat(itemWeightString, 64)
			if err != nil {
				return nil, stacktrace.Propagate(err, "failed to parse \"%s\" as float64", itemWeightString)
			}
			itemWeight := parsedItemWeight / 100
			weights = append(weights, models.BattledomeItemWeight{
				Arena:  currentArena,
				Name:   itemName,
				Weight: itemWeight,
			})
		} else {
			currentArena = strings.TrimSpace(line)
		}
	}
	return weights, nil
}
