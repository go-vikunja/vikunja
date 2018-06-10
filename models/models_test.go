package models

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetEngine(t *testing.T) {
	Config.Database.Path = "file::memory:?cache=shared"
	err := SetEngine()
	assert.NoError(t, err)
}
