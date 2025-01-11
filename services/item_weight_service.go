package services

import (
	"github.com/darienchong/neopets-battledome-analysis/constants"
	"github.com/darienchong/neopets-battledome-analysis/helpers"
	"github.com/darienchong/neopets-battledome-analysis/models"
	"github.com/darienchong/neopets-battledome-analysis/parsers"
)

type ItemWeightService struct {
	ItemWeightParser *parsers.ItemWeightParser
}

func NewItemWeightService() *ItemWeightService {
	return &ItemWeightService{
		ItemWeightParser: parsers.NewItemWeightParser(),
	}
}

func (service *ItemWeightService) GetItemWeights(arena string) ([]models.ItemWeight, error) {
	weights, err := service.ItemWeightParser.Parse(constants.GetItemWeightsFilePath())
	if err != nil {
		return nil, err
	}
	return helpers.Filter(weights, func(weight models.ItemWeight) bool {
		return weight.Arena == arena
	}), nil
}
