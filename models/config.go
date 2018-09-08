package models

import (
	"crypto/rand"
	"fmt"
	"github.com/spf13/viper"
	"os"
	"strings"
)

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

	// Init checking for environment variables
	viper.SetEnvPrefix("vikunja")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Load the config file
	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	return
}

func random(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return fmt.Sprintf("%X", b), nil
}

// SetConfig initianlises the config and publishes it for other functions to use
func SetConfig() (err error) {

	// File Checks
	if _, err := os.Stat("config.ini"); os.IsNotExist(err) {
		return err
	}

	// Load the config
	//cfg, err := ini.Load("config.ini")
	if err != nil {
		return err
	}

	// Map the config to our struct
	//err = cfg.MapTo(Config)
	if err != nil {
		return err
	}

	// Set default value for interface to listen on
	/*Config.Interface = cfg.Section("General").Key("Interface").String()
	if Config.Interface == "" {
		Config.Interface = ":8080"
	}

	// JWT secret
	Config.JWTLoginSecret = []byte(cfg.Section("General").Key("JWTSecret").String())*/

	return nil
}
