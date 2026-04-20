package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAuthFromContext_NoEchoContext(t *testing.T) {
	_, err := GetAuthFromContext(context.Background())
	assert.Error(t, err, "should fail when echo.Context isn't stashed on ctx")
}

// NOTE: A full positive test requires a valid JWT and DB fixtures.
// That path is exercised by the Label integration test in Phase E.
// Here we only prove the helper returns an error (not a panic) on an
// unwrapped context.
