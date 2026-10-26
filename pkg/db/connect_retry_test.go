package db

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPingWithRetry(t *testing.T) {
	t.Run("single attempt when retry interval is disabled", func(t *testing.T) {
		attempts := 0
		err := pingWithRetry(func() error {
			attempts++
			return errors.New("connection refused")
		}, 0, 0)

		require.Error(t, err)
		assert.Equal(t, 1, attempts)
	})

	t.Run("retries until the database responds", func(t *testing.T) {
		attempts := 0
		start := time.Now()
		err := pingWithRetry(func() error {
			attempts++
			if attempts < 3 {
				return errors.New("connection refused")
			}
			return nil
		}, 1, 50)

		require.NoError(t, err)
		assert.Equal(t, 3, attempts)
		assert.GreaterOrEqual(t, time.Since(start), 2*time.Millisecond)
	})
}
