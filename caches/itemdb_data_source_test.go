package caches

import (
	"testing"
)

func TestItemDBPrice(t *testing.T) {
	itemName := "Green Apple"
	target := NewItemDBDataSource()
	price, err := target.Price(itemName)

	if err != nil || price <= 0 {
		t.Fatalf("failed to retrieve price for %q from ItemDb! The retrieved price was %f", itemName, price)
	}
}
