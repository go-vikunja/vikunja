package models

import (
	"crypto/rand"
	"fmt"
	"github.com/spf13/viper"
	"strings"
)

// InitConfig initializes the config, sets defaults etc.
func InitConfig() (err error) {

	// Set defaults
	// Service config
	random, err := random(32)
	if err != nil {
		return err
	}

	viper.SetDefault("service.JWTSecret", random)
	viper.SetDefault("service.interface", ":3456")
	// Database
	viper.SetDefault("database.type", "sqlite")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.user", "vikunja")
	viper.SetDefault("database.password", "")
	viper.SetDefault("database.database", "vikunja")
	viper.SetDefault("database.path", "./vikunja.db")
	viper.SetDefault("database.showqueries", false)
	// Cacher
	viper.SetDefault("cache.enabled", false)
	viper.SetDefault("cache.type", "memory")
	viper.SetDefault("cache.maxelementsize", 1000)
	viper.SetDefault("cache.redishost", "localhost:6379")
	viper.SetDefault("cache.redispassword", "")

	// Init checking for environment variables
	viper.SetEnvPrefix("vikunja")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Load the config file
	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	err = viper.ReadInConfig()
	if err != nil {
		fmt.Println(err)
		fmt.Println("Using defaults.")
	}

	return nil
}

func random(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return fmt.Sprintf("%X", b), nil
}
