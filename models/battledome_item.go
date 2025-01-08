package models

import (
	"fmt"

	"github.com/darienchong/neopets-battledome-analysis/caches"
	"github.com/darienchong/neopets-battledome-analysis/constants"
	"github.com/darienchong/neopets-battledome-analysis/helpers"
	"github.com/dustin/go-humanize"
)

type BattledomeItem struct {
	Name            string
	Quantity        int32
	IndividualPrice float64
}

func (first *BattledomeItem) Combine(second *BattledomeItem) error {
	if first.Name != second.Name {
		return fmt.Errorf("tried to combine two items that did not have the same name: %s and %s", first, second)
	}

	first.Quantity += second.Quantity
	return nil
}

func (first *BattledomeItem) Union(second *BattledomeItem) (*BattledomeItem, error) {
	if first.Name != second.Name {
		return nil, fmt.Errorf("tried to union two items that did not have the same name: %s and %s", first, second)
	}

	combined := &BattledomeItem{}
	combined.Name = first.Name
	combined.IndividualPrice = first.IndividualPrice
	combined.Quantity = first.Quantity + second.Quantity
	return combined, nil
}

func (item *BattledomeItem) AddPrice(cache *caches.ItemPriceCache) {
	item.IndividualPrice = cache.GetPrice(item.Name)
}

func (item *BattledomeItem) GetProfit() (float64, error) {
	if item.IndividualPrice <= 0 {
		return 0.0, fmt.Errorf("can't calculate profit as price was not set for \"%s\"", item.Name)
	}

	return float64(item.Quantity) * item.IndividualPrice, nil
}

func (item *BattledomeItem) GetPercentageProfit(res *BattledomeDropsAnalysis) (float64, error) {
	profit, err := item.GetProfit()
	if err != nil {
		return 0.0, err
	}

	return profit / res.GetTotalProfit(), nil
}

func (item *BattledomeItem) GetDropRate(res *BattledomeDropsAnalysis) float64 {
	return float64(item.Quantity) / float64(helpers.Sum(helpers.Map(helpers.ToSlice(res.Items), func(tuple helpers.Tuple) int32 {
		return tuple.Elements[1].(*BattledomeItem).Quantity
	})))
}

func (item *BattledomeItem) String() string {
	return fmt.Sprintf("%s × %d @ %s NP/ea (%s NP total)", item.Name, item.Quantity, humanize.FormatFloat(constants.FLOAT_FORMAT_LAYOUT, item.IndividualPrice), humanize.FormatFloat(constants.FLOAT_FORMAT_LAYOUT, float64(item.Quantity)*item.IndividualPrice))
}

func (item *BattledomeItem) Copy() *BattledomeItem {
	copy := &BattledomeItem{}
	copy.Name = item.Name
	copy.Quantity = item.Quantity
	copy.IndividualPrice = item.IndividualPrice
	return copy
}
