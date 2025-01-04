package models

import (
	"math"

	"gonum.org/v1/gonum/stat/distuv"
)

type ResultStatistics struct {
	SampleSize              int
	Min                     float64
	Max                     float64
	Mean                    float64
	Median                  float64
	StandardDeviationSample float64
}

func (resultStats *ResultStatistics) GetMean(scaleFactor int) float64 {
	return float64(scaleFactor) * resultStats.Mean
}

func (resultStats *ResultStatistics) GetStandardDeviationSample(scaleFactor int) float64 {
	return math.Sqrt(float64(scaleFactor)) * resultStats.StandardDeviationSample
}

func (resultStats *ResultStatistics) GetCoefficientOfVariation(scaleFactor int) float64 {
	return resultStats.GetStandardDeviationSample(scaleFactor) / resultStats.GetMean(scaleFactor)
}

func (resultStats *ResultStatistics) GetMeanError(scaleFactor int, alpha float64) float64 {
	sampleStdev := resultStats.GetStandardDeviationSample(scaleFactor)
	tDist := distuv.StudentsT{
		Mu:    0,
		Sigma: 1,
		Nu:    float64(resultStats.SampleSize - 1),
	}
	return sampleStdev * tDist.Quantile(1-alpha/2)
}
