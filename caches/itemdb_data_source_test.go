package caches

import (
	"testing"
)

func TestGetItemDbPrice(t *testing.T) {
	itemName := "Green Apple"
	target := NewItemDbDataSource()
	price := target.GetPrice(itemName)

	if price <= 0 {
		t.Fatalf("failed to retrieve price for \"%s\" from ItemDb! The retrieved price was %f", itemName, price)
	}
}
