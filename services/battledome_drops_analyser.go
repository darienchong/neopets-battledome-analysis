package services

import (
	"github.com/darienchong/neopets-battledome-analysis/models"
)

type BattledomeDropsAnalyser struct{}

func NewBattledomeDropsAnalyser() *BattledomeDropsAnalyser {
	return &BattledomeDropsAnalyser{}
}

func (analyser *BattledomeDropsAnalyser) Analyse(drops []*models.BattledomeDrops) (*models.BattledomeDropsAnalysis, error) {
	res := new(models.BattledomeDropsAnalysis)
	combinedDrops := new(models.BattledomeDrops)
	res.Metadata = drops[0].Metadata.DropsMetadata
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
