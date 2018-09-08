package models

import (
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetEngine(t *testing.T) {
	viper.Set("database.path", "file::memory:?cache=shared")
	err := SetEngine()
	assert.NoError(t, err)
}
