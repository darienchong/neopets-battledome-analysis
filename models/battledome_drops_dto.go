package models

import "github.com/darienchong/neopets-battledome-analysis/constants"

type BattledomeDropsDto struct {
	Metadata DropsMetadata
	Items    map[string]*BattledomeItem
}

func (dto *BattledomeDropsDto) ToBattledomeDrops() *BattledomeDrops {
	drops := NewBattledomeDrops()
	drops.Metadata = dto.Metadata
	drops.Items = dto.Items
	return drops
}

func (dto *BattledomeDropsDto) GetTotalItemQuantity() int {
	totalItemQuantity := 0
	for _, item := range dto.Items {
		if item.Name == "nothing" {
			continue
		}
		totalItemQuantity += int(item.Quantity)
	}
	return totalItemQuantity
}

func (dto *BattledomeDropsDto) Validate() bool {
	return dto.GetTotalItemQuantity() == constants.BATTLEDOME_DROPS_PER_DAY
}
