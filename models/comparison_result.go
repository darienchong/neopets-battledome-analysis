package models

type ComparisonResult struct {
	Analysis *BattledomeDropsAnalysis
	Profit   map[string]*ItemProfit
}
