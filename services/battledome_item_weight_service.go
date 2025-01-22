package services

import (
	"github.com/darienchong/neopets-battledome-analysis/constants"
	"github.com/darienchong/neopets-battledome-analysis/helpers"
	"github.com/darienchong/neopets-battledome-analysis/models"
	"github.com/darienchong/neopets-battledome-analysis/parsers"
	"github.com/palantir/stacktrace"
)

type BattledomeItemWeightService struct {
	ItemWeightParser *parsers.BattledomeItemWeightParser
}

func NewBattledomeItemWeightService(itemWeightParser *parsers.BattledomeItemWeightParser) *BattledomeItemWeightService {
	return &BattledomeItemWeightService{
		ItemWeightParser: itemWeightParser,
	}
}

func (s *BattledomeItemWeightService) GetItemWeights(arena string) ([]models.BattledomeItemWeight, error) {
	weights, err := s.ItemWeightParser.Parse(constants.ItemWeightsFilePath())
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to parse \"%s\" as item weights", constants.ItemWeightsFilePath())
	}
	return helpers.Filter(weights, func(weight models.BattledomeItemWeight) bool {
		return weight.Arena == arena
	}), nil
}
