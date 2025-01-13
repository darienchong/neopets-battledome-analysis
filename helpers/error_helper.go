package helpers

import (
	"encoding/json"

	"github.com/palantir/stacktrace"
)

func PropagateWithSerialisedValue[T any](err error, template string, secondaryTemplate string, val T) error {
	serialised, serialisedErr := json.Marshal(val)
	if serialisedErr != nil {
		return stacktrace.Propagate(err, secondaryTemplate, serialisedErr)
	}
	return stacktrace.Propagate(err, template, serialised)
}
