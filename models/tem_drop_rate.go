package models

type ItemDropRate struct {
	Arena    string
	ItemName string
	DropRate float64
}

type ItemProfit struct {
	ItemDropRate
	IndividualPrice float64
}

func (profit *ItemProfit) GetProfit() float64 {
	return profit.DropRate * profit.IndividualPrice
}

func (profit *ItemProfit) GetPercentageProfit(totalProfit float64) float64 {
	return profit.GetProfit() / totalProfit
}
