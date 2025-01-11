package parsers

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/darienchong/neopets-battledome-analysis/helpers"
	"github.com/darienchong/neopets-battledome-analysis/models"
)

type DropRateParser struct{}

func NewDropRateParser() *DropRateParser {
	return &DropRateParser{}
}

func (parser *DropRateParser) Parse(filePath string) ([]models.ItemDropRate, error) {
	if !helpers.IsFileExists(filePath) {
		return nil, fmt.Errorf("item drop rates file does not exist! file path was \"%s\"", filePath)
	}

	dropRates := []models.ItemDropRate{}
	file, err := os.OpenFile(filePath, os.O_RDONLY, 0755)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		tokens := strings.Split(strings.TrimSpace(line), "|")
		dropRate, err := strconv.ParseFloat(strings.TrimSpace(tokens[2]), 64)
		if err != nil {
			return nil, err
		}
		dropRates = append(dropRates, models.ItemDropRate{
			Arena:    strings.TrimSpace(tokens[0]),
			ItemName: strings.TrimSpace(tokens[1]),
			DropRate: dropRate,
		})
	}
	return dropRates, nil
}

func (parser *DropRateParser) Save(data []models.ItemDropRate, filePath string) error {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	for _, dropRate := range data {
		_, err := file.WriteString(fmt.Sprintf("%s|%s|%f\n", dropRate.Arena, dropRate.ItemName, dropRate.DropRate))
		if err != nil {
			return err
		}
	}

	return nil
}
