package helpers

import (
	"fmt"

	"github.com/darienchong/neopets-battledome-analysis/constants"
	"github.com/dustin/go-humanize"
)

func FormatInt(n int) string {
	return humanize.FormatInteger(constants.FloatFormatLayout, n)
}

func FormatFloat(n float64) string {
	return humanize.FormatFloat(constants.FloatFormatLayout, n)
}

func FormatPercentage(n float64) string {
	return humanize.FormatFloat(constants.PercentageFormatLayout, n*100)
}

func FormatFloatRange(template string, leftBound float64, rightBound float64) string {
	if leftBound == rightBound {
		return FormatFloat(leftBound)
	}

	return fmt.Sprintf(template, FormatFloat(leftBound), FormatFloat(rightBound))
}

func FormatPercentageRange(template string, leftBound float64, rightBound float64) string {
	if leftBound == rightBound {
		return FormatPercentage(leftBound)
	}

	return fmt.Sprintf(template, FormatPercentage(leftBound), FormatPercentage(rightBound))
}

func CrossFormat(formatter func(float64) string, template string, values ...float64) string {
	return fmt.Sprintf(template, Map(values, formatter))
}
