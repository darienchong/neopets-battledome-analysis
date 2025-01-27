package parsers

import (
	"path/filepath"
	"testing"

	"github.com/darienchong/neopets-battledome-analysis/constants"
	"github.com/darienchong/neopets-battledome-analysis/models"
)

func shouldHaveItemAndQuantity(normalisedItems models.NormalisedBattledomeItems, t *testing.T, itemName string, itemQuantity int32) {
	battledomeItem, isInItems := normalisedItems[models.ItemName(itemName)]
	if !isInItems {
		t.Fatalf("Expected %q to be in items, but it was not.", itemName)
	}

	if battledomeItem.Quantity != itemQuantity {
		t.Fatalf("Expected %q to have quantity \"%d\", but it was \"%d\".", itemName, itemQuantity, battledomeItem.Quantity)
	}
}

func TestDropDataParser(t *testing.T) {
	target := NewBattledomeItemDropDataParser()
	dto, err := target.Parse(constants.DropDataFilePath("2024_12_20.txt"))
	if err != nil {
		t.Fatalf("Failed to parse file: %s", err)
	}

	expectedMetadata := new(models.DropsMetadataWithSource)
	expectedMetadata.Source = filepath.Base(constants.DropDataFilePath("2024_12_20.txt"))
	expectedMetadata.Arena = models.Arena("Central Arena")
	expectedMetadata.Challenger = models.Challenger("Flaming Meerca")
	expectedMetadata.Difficulty = models.Difficulty("Mighty")
	if dto.Metadata != *expectedMetadata {
		t.Fatalf("Expected metadata and actual metadata did not match:\n\tExpected: %q\n\tReceived: %q", expectedMetadata, dto.Metadata)
	}

	normalisedItems, err := dto.Items.Normalise()
	if err != nil {
		t.Fatalf("%s", err)
	}

	shouldHaveItemAndQuantity(normalisedItems, t, "Ridiculously Heavy Battle Hammer", 1)
	shouldHaveItemAndQuantity(normalisedItems, t, "Cursed Wand of Shadow", 1)
	shouldHaveItemAndQuantity(normalisedItems, t, "Chocolate Creme Pie", 1)
	shouldHaveItemAndQuantity(normalisedItems, t, "Can of Neocola", 1)
	shouldHaveItemAndQuantity(normalisedItems, t, "Unidentifiable Weak Bottled Faerie", 1)
	shouldHaveItemAndQuantity(normalisedItems, t, "Har Codestone", 2)
	shouldHaveItemAndQuantity(normalisedItems, t, "Orn Codestone", 1)
	shouldHaveItemAndQuantity(normalisedItems, t, "Bri Codestone", 1)
	shouldHaveItemAndQuantity(normalisedItems, t, "Main Codestone", 1)
	shouldHaveItemAndQuantity(normalisedItems, t, "Tai-Kai Codestone", 1)
	shouldHaveItemAndQuantity(normalisedItems, t, "Marzipan Sugared Slorg", 1)
	shouldHaveItemAndQuantity(normalisedItems, t, "Kew Codestone", 1)
	shouldHaveItemAndQuantity(normalisedItems, t, "Eo Codestone", 1)
	shouldHaveItemAndQuantity(normalisedItems, t, "Robot Muffin", 1)
}
