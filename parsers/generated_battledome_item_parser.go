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

type GeneratedBattledomeItemParser struct{}

func NewGeneratedBattledomeItemParser() *GeneratedBattledomeItemParser {
	return &GeneratedBattledomeItemParser{}
}

func (p *GeneratedBattledomeItemParser) Save(items models.NormalisedBattledomeItems, filePath string) error {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return stacktrace.Propagate(err, "failed to open file: %q", filePath)
	}
	defer file.Close()
	for _, item := range items {
		file.WriteString(fmt.Sprintf("%s|%s|%d\n", item.Metadata.Arena, item.Name, item.Quantity))
	}
	return nil
}

func (p *GeneratedBattledomeItemParser) Parse(filePath string) (models.NormalisedBattledomeItems, error) {
	if !helpers.IsFileExists(filePath) {
		return nil, fmt.Errorf("generated drop file does not exist! supplied file path was %q", filePath)
	}

	file, err := os.OpenFile(filePath, os.O_RDONLY, 0755)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to open file: %q", filePath)
	}
	defer file.Close()

	items := models.NormalisedBattledomeItems{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		tokens := strings.Split(line, "|")
		arena := models.Arena(strings.TrimSpace(tokens[0]))
		metadata := *models.GeneratedMetadata(arena)
		itemName := models.ItemName(strings.TrimSpace(tokens[1]))
		itemQuantity, err := strconv.ParseInt(strings.TrimSpace(tokens[2]), 0, 32)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to parse %s as int", strings.TrimSpace(tokens[2]))
		}

		_, exists := items[itemName]
		if !exists {
			items[itemName] = &models.BattledomeItem{
				Metadata: metadata.BattledomeItemMetadata,
				Name:     itemName,
				Quantity: int32(itemQuantity),
			}
		} else {
			items[itemName].Quantity += int32(itemQuantity)
		}

	}

	return items, nil
}
