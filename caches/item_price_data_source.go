package caches

import "strings"

type ItemPriceDataSource interface {
	GetPrice(itemName string) float64
	GetFilePath() string
}

func getNormalisedItemName(itemName string) string {
	itemName = strings.ToLower(itemName)
	itemName = strings.ReplaceAll(itemName, " ", "-")
	itemName = strings.ReplaceAll(itemName, ":", "")
	itemName = strings.ReplaceAll(itemName, "!", "")
	return itemName
}
