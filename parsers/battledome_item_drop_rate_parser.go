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

type BattledomeItemDropRateParser struct{}

func NewBattledomeItemDropRateParser() *BattledomeItemDropRateParser {
	return &BattledomeItemDropRateParser{}
}

func (parser *BattledomeItemDropRateParser) Parse(filePath string) ([]models.BattledomeItemDropRate, error) {
	if !helpers.IsFileExists(filePath) {
		return nil, fmt.Errorf("item drop rates file does not exist! file path was \"%s\"", filePath)
	}

	dropRates := []models.BattledomeItemDropRate{}
	file, err := os.OpenFile(filePath, os.O_RDONLY, 0755)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to open file: \"%s\"", filePath)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		tokens := strings.Split(strings.TrimSpace(line), "|")
		dropRate, err := strconv.ParseFloat(strings.TrimSpace(tokens[2]), 64)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to parse \"%s\" as float", strings.TrimSpace(tokens[2]))
		}
		dropRates = append(dropRates, models.BattledomeItemDropRate{
			Metadata: models.BattledomeItemMetadata{
				Arena:      models.Arena(strings.TrimSpace(tokens[0])),
				Challenger: "(parsed)",
				Difficulty: "(parsed)",
			},
			ItemName: models.ItemName(strings.TrimSpace(tokens[1])),
			DropRate: dropRate,
		})
	}
	return dropRates, nil
}

func (parser *BattledomeItemDropRateParser) Save(data []models.BattledomeItemDropRate, filePath string) error {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return stacktrace.Propagate(err, "failed to open file \"%s\"", filePath)
	}
	for _, dropRate := range data {
		data := fmt.Sprintf("%s|%s|%f\n", dropRate.Metadata.Arena, dropRate.ItemName, dropRate.DropRate)
		_, err := file.WriteString(data)
		if err != nil {
			return stacktrace.Propagate(err, "failed to write \"%s\" to \"%s\"", data, filePath)
		}
	}

	return nil
}
