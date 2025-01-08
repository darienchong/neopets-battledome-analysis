package tests

import (
	"path/filepath"
	"testing"

	"github.com/darienchong/neopets-battledome-analysis/constants"
	"github.com/darienchong/neopets-battledome-analysis/models"
	"github.com/darienchong/neopets-battledome-analysis/services"
)

func shouldHaveItemAndQuantity(dropsDto *models.BattledomeDropsDto, t *testing.T, itemName string, itemQuantity int32) {
	battledomeItem, isInItems := dropsDto.Items[itemName]
	if !isInItems {
		t.Fatalf("Expected \"%s\" to be in items, but it was not.", itemName)
	}

	if battledomeItem.Quantity != itemQuantity {
		t.Fatalf("Expected \"%s\" to have quantity \"%d\", but it was \"%d\".", itemName, itemQuantity, battledomeItem.Quantity)
	}
}

func TestDropDataParser(t *testing.T) {
	target := new(services.DropDataParser)
	drops, err := target.Parse(constants.GetDropDataFilePath("2024_12_20.txt"))
	if err != nil {
		t.Fatalf("Failed to parse file")
		panic(err)
	}

	expectedMetadata := new(models.DropsMetadata)
	expectedMetadata.Source = filepath.Base(constants.GetDropDataFilePath("2024_12_20.txt"))
	expectedMetadata.Arena = "Central Arena"
	expectedMetadata.Challenger = "Flaming Meerca"
	expectedMetadata.Difficulty = "Mighty"
	if drops.Metadata != *expectedMetadata {
		t.Fatalf("Expected metadata and actual metadata did not match:\n\tExpected: \"%s\"\n\tReceived: \"%s\"", expectedMetadata, drops.Metadata)
	}

	shouldHaveItemAndQuantity(drops, t, "Ridiculously Heavy Battle Hammer", 1)
	shouldHaveItemAndQuantity(drops, t, "Cursed Wand of Shadow", 1)
	shouldHaveItemAndQuantity(drops, t, "Chocolate Creme Pie", 1)
	shouldHaveItemAndQuantity(drops, t, "Can of Neocola", 1)
	shouldHaveItemAndQuantity(drops, t, "Unidentifiable Weak Bottled Faerie", 1)
	shouldHaveItemAndQuantity(drops, t, "Har Codestone", 2)
	shouldHaveItemAndQuantity(drops, t, "Orn Codestone", 1)
	shouldHaveItemAndQuantity(drops, t, "Bri Codestone", 1)
	shouldHaveItemAndQuantity(drops, t, "Main Codestone", 1)
	shouldHaveItemAndQuantity(drops, t, "Tai-Kai Codestone", 1)
	shouldHaveItemAndQuantity(drops, t, "Marzipan Sugared Slorg", 1)
	shouldHaveItemAndQuantity(drops, t, "Kew Codestone", 1)
	shouldHaveItemAndQuantity(drops, t, "Eo Codestone", 1)
	shouldHaveItemAndQuantity(drops, t, "Robot Muffin", 1)
}
