package services

import (
	"fmt"
	"log/slog"

	"github.com/darienchong/neopets-battledome-analysis/constants"
	"github.com/darienchong/neopets-battledome-analysis/helpers"
	"github.com/darienchong/neopets-battledome-analysis/models"
)

type DropDataService struct{}

func NewDropDataService() *DropDataService {
	return &DropDataService{}
}

func (service *DropDataService) GetAllDrops(dataFolderPath string) ([]*models.BattledomeDrops, error) {
	parser := NewDropDataParser()
	files, err := helpers.GetFilesInFolder(dataFolderPath)
	if err != nil {
		slog.Error("Failed to get files in folder!")
		panic(err)
	}

	drops := []*models.BattledomeDrops{}
	for _, file := range files {
		drop, err := parser.Parse(constants.GetDropDataFilePath(file))
		if err != nil {
			return nil, fmt.Errorf("DropDataService.GetAllDrops(%s): %w", file, err)
		}
		drops = append(drops, drop)
	}

	return drops, nil
}
