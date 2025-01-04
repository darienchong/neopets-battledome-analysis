package helpers

import (
	"github.com/darienchong/neopetsbattledomeanalysis/constants"
	"github.com/dustin/go-humanize"
)

func FormatFloat(n float64) string {
	return humanize.FormatFloat(constants.FLOAT_FORMAT_LAYOUT, n)
}

func FormatPercentage(n float64) string {
	return humanize.FormatFloat(constants.PERCENTAGE_FORMAT_LAYOUT, n*100)
}
