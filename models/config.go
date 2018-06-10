package models

import (
	"github.com/go-ini/ini"
	"os"
)

// ConfigStruct holds the config struct
type ConfigStruct struct {
	Database struct {
		Type        string
		Host        string
		User        string
		Password    string
		Database    string
		Path        string
		ShowQueries bool
	}

	JWTLoginSecret []byte
	Interface      string
}

// Config holds the configuration for the program
var Config = new(ConfigStruct)

// SetConfig initianlises the config and publishes it for other functions to use
func SetConfig() (err error) {

	// File Checks
	if _, err := os.Stat("config.ini"); os.IsNotExist(err) {
		return err
	}

	// Load the config
	cfg, err := ini.Load("config.ini")
	if err != nil {
		return err
	}

	// Map the config to our struct
	err = cfg.MapTo(Config)
	if err != nil {
		return err
	}

	// Set default value for interface to listen on
	Config.Interface = cfg.Section("General").Key("Interface").String()
	if Config.Interface == "" {
		Config.Interface = ":8080"
	}

	// JWT secret
	Config.JWTLoginSecret = []byte(cfg.Section("General").Key("JWTSecret").String())

	return nil
}
