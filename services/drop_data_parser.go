package services

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"

	"github.com/darienchong/neopets-battledome-analysis/constants"
	"github.com/darienchong/neopets-battledome-analysis/helpers"
	"github.com/darienchong/neopets-battledome-analysis/models"
)

type DropDataParser struct{}

var (
	parsers = []DropDataLineParser{
		new(MetadataParser),
		new(CommentParser),
		new(ItemDataParser),
	}
)

func (parser *DropDataParser) Parse(filePath string) (*models.BattledomeDrops, error) {
	if !helpers.IsFileExists(filePath) {
		return &models.BattledomeDrops{}, fmt.Errorf("file at \"%s\" does not exist", filePath)
	}

	file, err := os.OpenFile(filePath, os.O_RDONLY, 0755)
	if err != nil {
		return &models.BattledomeDrops{}, err
	}
	defer file.Close()

	dropsMetadata := new(models.DropsMetadata)
	drops := models.NewBattledomeDrops()
	drops.Metadata = *dropsMetadata
	drops.Metadata.Source = filepath.Base(filePath)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		for _, parser := range parsers {
			if !parser.IsApplicable(line) {
				slog.Debug(fmt.Sprintf("%s was not applicable for \"%s\"", reflect.TypeOf(parser), line))
				continue
			}

			slog.Debug(fmt.Sprintf("Applying %s for \"%s\"", reflect.TypeOf(parser), line))
			parser.Parse(line, drops)
			break
		}
	}

	itemCount := 0
	for _, item := range drops.Items {
		itemCount += int(item.Quantity)
	}
	if itemCount != constants.BATTLEDOME_DROPS_PER_DAY {
		slog.Error(fmt.Sprintf("WARNING! The drop data in \"%s\" does not contain %d drops; %d drops were detected.", drops.Metadata.Source, constants.BATTLEDOME_DROPS_PER_DAY, itemCount))
	}
	return drops, nil
}
