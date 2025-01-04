package tests

import (
	"testing"

	"github.com/darienchong/neopetsbattledomeanalysis/constants"
	"github.com/darienchong/neopetsbattledomeanalysis/helpers"
	"github.com/darienchong/neopetsbattledomeanalysis/models"
	"github.com/darienchong/neopetsbattledomeanalysis/services"
)

func TestItemWeightsParser(t *testing.T) {
	target := new(services.ItemWeightParser)
	itemWeights, err := target.Parse(constants.GetItemWeightsFilePath())
	if err != nil {
		panic(err)
	}
	diamondSnowballWeight := helpers.Filter(itemWeights, func(itemWeight models.ItemWeight) bool {
		return itemWeight.Name == "Diamond Snowball"
	})[0].Weight
	if diamondSnowballWeight != 0.02 {
		t.Fatalf("Diamond Snowball's weight was not correctly parsed!\nExpected: 0.02\nReceived:%f", diamondSnowballWeight)
	}
	weakBottledEarthFaerieWeight := helpers.Filter(itemWeights, func(itemWeight models.ItemWeight) bool {
		return itemWeight.Name == "Weak Bottled Earth Faerie"
	})[0].Weight
	if weakBottledEarthFaerieWeight != 0.015 {
		t.Fatalf("Weak Bottled Earth Faerie's weight was not correctly parsed!\nExpected: 0.015\nReceived:%f", weakBottledEarthFaerieWeight)
	}
}
