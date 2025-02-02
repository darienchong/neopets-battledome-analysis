package helpers

import (
	"fmt"
	"log/slog"
	"time"
)

type RetryPolicy[TResult any] struct {
	Backoff  func(int) int
	MaxTries int
}

func (rp RetryPolicy[T]) Execute(action func() (T, error)) (T, error) {
	var outcome T
	var err error
	for i := 1; i <= rp.MaxTries; i++ {
		outcome, err = action()
		if err == nil {
			return outcome, err
		}

		backoff := rp.Backoff(i)
		slog.Error(fmt.Sprintf("Failed to execute action; retrying in %d ms...", backoff))
		time.Sleep(time.Duration(backoff) * time.Millisecond)
	}

	return outcome, err
}
