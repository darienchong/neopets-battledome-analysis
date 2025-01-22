package caches

import (
	"testing"
)

func TestItemDBPrice(t *testing.T) {
	itemName := "Green Apple"
	target := NewItemDBDataSource()
	price := target.Price(itemName)

	if price <= 0 {
		t.Fatalf("failed to retrieve price for \"%s\" from ItemDb! The retrieved price was %f", itemName, price)
	}
}
