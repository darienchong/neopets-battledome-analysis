package caches

type ItemPriceDataSource interface {
	GetPrice(itemName string) float64
	GetFilePath() string
}
