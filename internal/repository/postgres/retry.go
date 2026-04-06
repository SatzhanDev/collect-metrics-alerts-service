package postgres

import (
	"context"
	"time"
)

func withRetry(ctx context.Context, fn func() error) error {
	delays := []time.Duration{
		1 * time.Second,
		3 * time.Second,
		5 * time.Second,
	}

	var err error

	for i := 0; i <= len(delays); i++ {
		err = fn()
		if err == nil {
			return nil
		}

		if classifyPGError(err) == NonRetriable {
			return err
		}

		if i == len(delays) {
			break
		}

		time.Sleep(delays[i])
	}

	return err
}
