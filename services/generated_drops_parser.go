package services

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/darienchong/neopets-battledome-analysis/caches"
	"github.com/darienchong/neopets-battledome-analysis/helpers"
	"github.com/darienchong/neopets-battledome-analysis/models"
)

type GeneratedDropsParser struct{}

func (parser *GeneratedDropsParser) Save(drops map[string]*models.BattledomeDrops, filePath string) error {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer file.Close()
	for _, drop := range drops {
		for _, item := range drop.Items {
			file.WriteString(fmt.Sprintf("%s|%s|%d\n", drop.Metadata.Arena, item.Name, item.Quantity))
		}
	}
	return nil
}

func (parser *GeneratedDropsParser) Parse(filePath string) (map[string]*models.BattledomeDrops, error) {
	if !helpers.IsFileExists(filePath) {
		return nil, fmt.Errorf("generated drop file does not exist! supplied file path was \"%s\"", filePath)
	}

	file, err := os.OpenFile(filePath, os.O_RDONLY, 0755)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	itemPriceCache, err := caches.GetItemPriceCacheInstance()
	if err != nil {
		return nil, err
	}
	defer itemPriceCache.Close()

	drops := map[string]*models.BattledomeDrops{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		tokens := strings.Split(line, "|")
		arena := strings.TrimSpace(tokens[0])
		itemName := strings.TrimSpace(tokens[1])
		itemQuantity, err := strconv.ParseInt(strings.TrimSpace(tokens[2]), 0, 32)
		if err != nil {
			return nil, err
		}

		_, ok := drops[arena]
		if !ok {
			drops[arena] = models.NewBattledomeDrops()
			drops[arena].Metadata = models.DropsMetadata{
				Source:     "(generated)",
				Arena:      arena,
				Challenger: "(generated)",
				Difficulty: "(generated)",
			}
		}
		_, ok = drops[arena].Items[itemName]
		if !ok {
			drops[arena].Items[itemName] = &models.BattledomeItem{
				Name:            itemName,
				Quantity:        int32(itemQuantity),
				IndividualPrice: itemPriceCache.GetPrice(itemName),
			}
		} else {
			drops[arena].Items[itemName].Quantity += int32(itemQuantity)
		}
	}

	return drops, nil
}
