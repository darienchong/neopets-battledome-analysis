package services

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/darienchong/neopetsbattledomeanalysis/models"
)

type DropDataLineParser interface {
	IsApplicable(line string) bool
	Parse(line string, drops *models.BattledomeDrops) error
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

func (parser *MetadataParser) Parse(line string, drops *models.BattledomeDrops) error {
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

func (parser *CommentParser) Parse(line string, drops *models.BattledomeDrops) error {
	return nil
}

type ItemDataParser struct{}

func (parser *ItemDataParser) IsApplicable(line string) bool {
	return true
}

func (parser *ItemDataParser) Parse(line string, drops *models.BattledomeDrops) error {
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
