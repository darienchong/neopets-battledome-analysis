package tests

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/darienchong/neopets-battledome-analysis/models"
	"github.com/palantir/stacktrace"
)

func first() error {
	return stacktrace.NewError("error that originated from first()")
}

func second() error {
	err := first()
	if err != nil {
		return stacktrace.Propagate(err, "error that originated from second()")
	}

	return nil
}

func third() error {
	err := second()
	if err != nil {
		return stacktrace.Propagate(err, "error that originated from third()")
	}

	return nil
}

func firstWithValue() error {
	items := models.NormalisedBattledomeItems{}
	items[models.ItemName("My Item")] = &models.BattledomeItem{
		Metadata: models.BattledomeItemMetadata{
			Arena:      "My Arena",
			Challenger: "My Challenger",
			Difficulty: "My Difficulty",
		},
		Name:     "My Item",
		Quantity: 123,
	}

	err := stacktrace.NewError("base error that originated from first()")
	serialised, serialisedErr := json.Marshal(items)
	serialisedErr = fmt.Errorf("some error that occurred during serialisation")
	if serialisedErr != nil {
		return stacktrace.Propagate(err, "error that originated from first(); additionally, an error occurred while trying to serialise the input: %s", serialisedErr)
	}
	return stacktrace.Propagate(err, "error that originated from first(): %s", serialised)
}

func secondWithValue() error {
	err := firstWithValue()
	if err != nil {
		return stacktrace.Propagate(err, "error that originated from second()")
	}

	return nil
}

func thirdWithValue() error {
	err := secondWithValue()
	if err != nil {
		return stacktrace.Propagate(err, "error that originated from third()")
	}

	return nil
}

func TestErrorPropagation(t *testing.T) {
	err := third()
	if err != nil {
		t.Fatalf("%+v", err)
	}
}

func TestErrorPropagationWithSerialisation(t *testing.T) {
	err := thirdWithValue()
	if err != nil {
		t.Fatalf("%+v", err)
	}
}
