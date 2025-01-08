package services

import (
	"fmt"

	"github.com/darienchong/neopets-battledome-analysis/helpers"
	"github.com/darienchong/neopets-battledome-analysis/models"
)

type ArenaDataAnalyser struct{}

func NewArenaDataAnalyser() *ArenaDataAnalyser {
	return &ArenaDataAnalyser{}
}

func validateDropsAreAllFromSameArena(drops []*models.BattledomeDrops) bool {
	if len(drops) == 0 {
		return true
	}

	arena := drops[0].Metadata.Arena
	return helpers.Count(drops, func(drop *models.BattledomeDrops) bool {
		return drop.Metadata.Arena == arena
	}) == len(drops)
}

func (analyser *ArenaDataAnalyser) Analyse(drops []*models.BattledomeDrops) (*models.DropsAnalysis, error) {
	if !validateDropsAreAllFromSameArena(drops) {
		return nil, fmt.Errorf("not all the drops provided were from the same arena")
	}

	res := new(models.DropsAnalysis)
	combinedDrops := new(models.BattledomeDrops)
	res.Metadata = drops[0].Metadata.Copy()
	res.Items = combinedDrops.Items

	for _, drop := range drops {
		combinedMetadata, err := res.Metadata.Combine(&drop.Metadata)
		if err != nil {
			return nil, err
		}
		res.Metadata = combinedMetadata

		err = combinedDrops.Append(drop)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}
