package services

import (
	"github.com/darienchong/neopets-battledome-analysis/constants"
	"github.com/darienchong/neopets-battledome-analysis/helpers"
	"github.com/darienchong/neopets-battledome-analysis/models"
	"github.com/palantir/stacktrace"
)

type SavedBattledomeItemWeights interface {
	Parse(filePath string) ([]models.BattledomeItemWeight, error)
}

type BattledomeItemWeightService struct {
	SavedBattledomeItemWeights
}

func NewBattledomeItemWeightService(savedBattledomeItemWeights SavedBattledomeItemWeights) *BattledomeItemWeightService {
	return &BattledomeItemWeightService{
		SavedBattledomeItemWeights: savedBattledomeItemWeights,
	}
}

func (s *BattledomeItemWeightService) ItemWeights(arena string) ([]models.BattledomeItemWeight, error) {
	weights, err := s.SavedBattledomeItemWeights.Parse(constants.ItemWeightsFilePath())
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to parse %q as item weights", constants.ItemWeightsFilePath())
	}
	return helpers.Filter(weights, func(weight models.BattledomeItemWeight) bool {
		return weight.Arena == arena
	}), nil
}
