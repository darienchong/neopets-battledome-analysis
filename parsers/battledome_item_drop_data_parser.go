package parsers

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

type BattledomeItemDropDataParser struct{}

func NewBattledomeItemDropDataParser() *BattledomeItemDropDataParser {
	return &BattledomeItemDropDataParser{}
}

var (
	parsers = []DropDataLineParser{
		new(MetadataParser),
		new(CommentParser),
		new(ItemDataParser),
	}
)

func (parser *BattledomeItemDropDataParser) Parse(filePath string) (*models.BattledomeItemsDto, error) {
	if !helpers.IsFileExists(filePath) {
		return nil, fmt.Errorf("file at \"%s\" does not exist", filePath)
	}

	file, err := os.OpenFile(filePath, os.O_RDONLY, 0755)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	dto := &models.BattledomeItemsDto{
		Metadata: models.DropsMetadataWithSource{},
		Items:    models.BattledomeItems{},
	}
	dto.Metadata.Source = filepath.Base(filePath)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		for _, parser := range parsers {
			if !parser.IsApplicable(line) {
				continue
			}

			parser.Parse(line, dto)
			break
		}
	}

	itemCount := 0
	for _, item := range dto.Items {
		itemCount += int(item.Quantity)
	}
	if itemCount != constants.BATTLEDOME_DROPS_PER_DAY {
		slog.Error(fmt.Sprintf("WARNING! The drop data in \"%s\" does not contain %d drops; %d drops were detected.", dto.Metadata.Source, constants.BATTLEDOME_DROPS_PER_DAY, itemCount))
	}
	return dto, nil
}

type DropDataLineParser interface {
	IsApplicable(line string) bool
	Parse(line string, items *models.BattledomeItemsDto) error
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

func (parser *MetadataParser) Parse(line string, dto *models.BattledomeItemsDto) error {
	tokens := strings.Split(line, ":")
	metadata := dto.Metadata.Copy()
	metadataKey := strings.ToLower(strings.TrimSpace(tokens[0]))
	metadataValue := strings.TrimSpace(tokens[1])

	switch key := metadataKey; key {
	case ARENA_KEY:
		slog.Debug(fmt.Sprintf("Set Arena to \"%s\"", metadataValue))
		metadata.Arena = models.Arena(metadataValue)
	case CHALLENGER_KEY:
		slog.Debug(fmt.Sprintf("Set Challenger to \"%s\"", metadataValue))
		metadata.Challenger = models.Challenger(metadataValue)
	case DIFFICULTY_KEY:
		slog.Debug(fmt.Sprintf("Set Difficulty to \"%s\"", metadataValue))
		metadata.Difficulty = models.Difficulty(metadataValue)
	default:
		slog.Warn(fmt.Sprintf("Encountered an unrecognised metadata key while parsing drop data; the unrecognised key was \"%s\"", metadataKey))
	}

	dto.Metadata = *metadata
	return nil
}

type CommentParser struct{}

func (parser *CommentParser) IsApplicable(line string) bool {
	return strings.HasPrefix(line, "#")
}

func (parser *CommentParser) Parse(line string, dto *models.BattledomeItemsDto) error {
	return nil
}

type ItemDataParser struct{}

func (parser *ItemDataParser) IsApplicable(line string) bool {
	return true
}

func (parser *ItemDataParser) Parse(line string, dto *models.BattledomeItemsDto) error {
	tokens := strings.Split(line, "|")
	itemName := models.ItemName(strings.TrimSpace(tokens[0]))
	itemQuantity, err := strconv.ParseInt(strings.TrimSpace(tokens[1]), 0, 32)
	if err != nil {
		return err
	}

	dto.Items = append(dto.Items, &models.BattledomeItem{
		Metadata: dto.Metadata.BattledomeItemMetadata,
		Name:     itemName,
		Quantity: int32(itemQuantity),
	})

	return nil
}
