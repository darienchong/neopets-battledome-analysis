package models

import "fmt"

type ArenaProfitStatistics struct {
	Arena             string
	Mean              float64
	Median            float64
	StandardDeviation float64
}

func (stats *ArenaProfitStatistics) String() string {
	return fmt.Sprintf("%s|%f|%f|%f\n", stats.Arena, stats.Mean, stats.Median, stats.StandardDeviation)
}
