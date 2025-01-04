package models

import (
	"github.com/darienchong/neopetsbattledomeanalysis/caches"
	"github.com/darienchong/neopetsbattledomeanalysis/constants"
)

type BattledomeDrops struct {
	Metadata DropsMetadata
	Items    map[string]*BattledomeItem
}

func NewBattledomeDrops() *BattledomeDrops {
	instance := new(BattledomeDrops)
	instance.Items = map[string]*BattledomeItem{}
	return instance
}

func (drops *BattledomeDrops) GetTotalItemQuantity() int {
	totalItemQuantity := 0
	for _, item := range drops.Items {
		totalItemQuantity += int(item.Quantity)
	}
	return totalItemQuantity
}

func (drops *BattledomeDrops) Validate() bool {
	if drops.GetTotalItemQuantity() != constants.BATTLEDOME_DROPS_PER_DAY {
		return false
	}

	return true
}

func (drops *BattledomeDrops) AddPrices(cache *caches.ItemPriceCache) {
	for _, v := range drops.Items {
		v.AddPrice(cache)
	}
}

func (first *BattledomeDrops) Append(second *BattledomeDrops) error {
	combinedMetadata, err := first.Metadata.Combine(&second.Metadata)
	if err != nil {
		return err
	}

	first.Metadata = *combinedMetadata
	for _, item := range second.Items {
		_, itemExistsInFirst := first.Items[item.Name]
		if itemExistsInFirst {
			err := first.Items[item.Name].Combine(item)
			if err != nil {
				return err
			}
		} else {
			first.Items[item.Name] = item
		}
	}

	return nil
}

func (first *BattledomeDrops) Union(second *BattledomeDrops) (*BattledomeDrops, error) {
	combinedDrops := NewBattledomeDrops()

	combinedMetadata, err := first.Metadata.Combine(&second.Metadata)
	if err != nil {
		return nil, err
	}
	combinedDrops.Metadata = *combinedMetadata

	for _, item := range first.Items {
		combinedItem, itemExistsInCombined := combinedDrops.Items[item.Name]
		if itemExistsInCombined {
			combinedItem, err := combinedItem.Union(item)
			if err != nil {
				return nil, err
			}
			combinedDrops.Items[item.Name] = combinedItem
		} else {
			combinedDrops.Items[item.Name] = item.Copy()
		}
	}

	for _, item := range second.Items {
		combinedItem, itemExistsInCombined := combinedDrops.Items[item.Name]
		if itemExistsInCombined {
			combinedItem, err := combinedItem.Union(item)
			if err != nil {
				return nil, err
			}
			combinedDrops.Items[item.Name] = combinedItem
		} else {
			combinedDrops.Items[item.Name] = item.Copy()
		}
	}

	return combinedDrops, nil
}
