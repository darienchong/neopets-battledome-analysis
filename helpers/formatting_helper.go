package helpers

import (
	"github.com/darienchong/neopets-battledome-analysis/constants"
	"github.com/dustin/go-humanize"
)

func FormatInt(n int) string {
	return humanize.FormatInteger(constants.FLOAT_FORMAT_LAYOUT, n)
}

func FormatFloat(n float64) string {
	return humanize.FormatFloat(constants.FLOAT_FORMAT_LAYOUT, n)
}

func FormatPercentage(n float64) string {
	return humanize.FormatFloat(constants.PERCENTAGE_FORMAT_LAYOUT, n*100)
}
