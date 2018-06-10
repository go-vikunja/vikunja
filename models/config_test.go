package models

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func TestSetConfig(t *testing.T) {
	// Create test database
	assert.NoError(t, PrepareTestDatabase())

	// This should fail as it is looking for a nonexistent config
	err := SetConfig()
	assert.Error(t, err)

	// Write an invalid config
	configString := `[General
JWTSecret = Supersecret
Interface = ; This should make it automatically to :8080

[Database
Type = sqlite
Path = ./library.db`
	err = ioutil.WriteFile("config.ini", []byte(configString), 0644)
	assert.NoError(t, err)

	// Test setConfig (should fail as we're trying to parse an invalid config)
	err = SetConfig()
	assert.Error(t, err)

	// Delete the invalid file
	err = os.Remove("config.ini")
	assert.NoError(t, err)

	// Write a fake config
	configString = `[General]
JWTSecret = Supersecret
Interface = ; This should make it automatically to :8080

[Database]
Type = sqlite
Path = ./library.db`
	err = ioutil.WriteFile("config.ini", []byte(configString), 0644)
	assert.NoError(t, err)

	// Test setConfig
	err = SetConfig()
	assert.NoError(t, err)

	// Check for the values
	assert.Equal(t, []byte("Supersecret"), Config.JWTLoginSecret)
	assert.Equal(t, string(":8080"), Config.Interface)
	assert.Equal(t, string("sqlite"), Config.Database.Type)
	assert.Equal(t, string("./library.db"), Config.Database.Path)

	// Remove the dummy config
	err = os.Remove("config.ini")
	assert.NoError(t, err)
}
