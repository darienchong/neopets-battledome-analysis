package services

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/darienchong/neopets-battledome-analysis/constants"
	"github.com/darienchong/neopets-battledome-analysis/helpers"
	"github.com/darienchong/neopets-battledome-analysis/models"
)

type DropDataParser struct{}

func NewDropDataParser() *DropDataParser {
	return &DropDataParser{}
}

var (
	parsers = []DropDataLineParser{
		new(MetadataParser),
		new(CommentParser),
		new(ItemDataParser),
	}
)

func (parser *DropDataParser) Parse(filePath string) (*models.BattledomeDropsDto, error) {
	if !helpers.IsFileExists(filePath) {
		return &models.BattledomeDropsDto{}, fmt.Errorf("file at \"%s\" does not exist", filePath)
	}

	file, err := os.OpenFile(filePath, os.O_RDONLY, 0755)
	if err != nil {
		return &models.BattledomeDropsDto{}, err
	}
	defer file.Close()

	dropsDto := &models.BattledomeDropsDto{
		Metadata: models.DropsMetadataWithSource{},
		Items:    map[string]*models.BattledomeItem{},
	}
	dropsDto.Metadata.Source = filepath.Base(filePath)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		for _, parser := range parsers {
			if !parser.IsApplicable(line) {
				continue
			}

			parser.Parse(line, dropsDto)
			break
		}
	}

	itemCount := 0
	for _, item := range dropsDto.Items {
		itemCount += int(item.Quantity)
	}
	if itemCount != constants.BATTLEDOME_DROPS_PER_DAY {
		slog.Error(fmt.Sprintf("WARNING! The drop data in \"%s\" does not contain %d drops; %d drops were detected.", dropsDto.Metadata.Source, constants.BATTLEDOME_DROPS_PER_DAY, itemCount))
	}
	return dropsDto, nil
}

type DropDataLineParser interface {
	IsApplicable(line string) bool
	Parse(line string, drops *models.BattledomeDropsDto) error
}

const (
	ARENA_KEY      = "$arena"
	CHALLENGER_KEY = "$challenger"
	DIFFICULTY_KEY = "$difficulty"
)

type MetadataParser struct{}

func (parser *MetadataParser) IsApplicable(line string) bool {
	return strings.HasPrefix(line, "$")
}

func (parser *MetadataParser) Parse(line string, drops *models.BattledomeDropsDto) error {
	tokens := strings.Split(line, ":")
	metadata := &drops.Metadata
	metadataKey := strings.ToLower(strings.TrimSpace(tokens[0]))
	metadataValue := strings.TrimSpace(tokens[1])

	switch key := metadataKey; key {
	case ARENA_KEY:
		slog.Debug(fmt.Sprintf("Set Arena to \"%s\"", metadataValue))
		metadata.Arena = metadataValue
	case CHALLENGER_KEY:
		slog.Debug(fmt.Sprintf("Set Challenger to \"%s\"", metadataValue))
		metadata.Challenger = metadataValue
	case DIFFICULTY_KEY:
		slog.Debug(fmt.Sprintf("Set Difficulty to \"%s\"", metadataValue))
		metadata.Difficulty = metadataValue
	default:
		slog.Warn(fmt.Sprintf("Encountered an unrecognised metadata key while parsing drop data; the unrecognised key was \"%s\"", metadataKey))
	}

	return nil
}

type CommentParser struct{}

func (parser *CommentParser) IsApplicable(line string) bool {
	return strings.HasPrefix(line, "#")
}

func (parser *CommentParser) Parse(line string, drops *models.BattledomeDropsDto) error {
	return nil
}

type ItemDataParser struct{}

func (parser *ItemDataParser) IsApplicable(line string) bool {
	return true
}

func (parser *ItemDataParser) Parse(line string, drops *models.BattledomeDropsDto) error {
	tokens := strings.Split(line, "|")
	itemName := strings.TrimSpace(tokens[0])
	itemQuantity, err := strconv.ParseInt(strings.TrimSpace(tokens[1]), 0, 32)
	if err != nil {
		slog.Any("error", err)
		return err
	}

	_, isInItems := drops.Items[itemName]
	if isInItems {
		drops.Items[itemName].Quantity += int32(itemQuantity)
	} else {
		drops.Items[itemName] = &models.BattledomeItem{
			Name:            itemName,
			Quantity:        int32(itemQuantity),
			IndividualPrice: 0,
		}
	}

	return nil
}
