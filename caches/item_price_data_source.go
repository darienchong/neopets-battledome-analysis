package caches

type ItemPriceDataSource interface {
	Price(itemName string) float64
	FilePath() string
}
