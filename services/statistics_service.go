package services

import (
	"math"

	"gonum.org/v1/gonum/stat/distuv"
)

type StatisticsService struct{}

func NewStatisticsService() *StatisticsService {
	return &StatisticsService{}
}

// Calculates a confidence interval with significance level {alpha}
// Useful for when \cap{p} ~= 0
func (s *StatisticsService) ClopperPearsonInterval(x int, n int, alpha float64) (float64, float64, error) {
	if n == 0 {
		return 0, 0, nil
	}

	if x == 0 {
		return 0, 1 - math.Pow(alpha/2, 1/float64(n)), nil
	}

	if x == n {
		return math.Pow(alpha/2, 1/float64(n)), 1, nil
	}

	betaLower := distuv.Beta{
		Alpha: float64(x),
		Beta:  float64(n - x + 1),
	}
	betaUpper := distuv.Beta{
		Alpha: float64(x + 1),
		Beta:  float64(n - x),
	}

	return betaLower.Quantile(alpha / 2), betaUpper.Quantile(1 - alpha/2), nil
}
