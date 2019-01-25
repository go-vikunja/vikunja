//  Vikunja is a todo-list application to facilitate your life.
//  Copyright 2018 Vikunja and contributors. All rights reserved.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU General Public License as published by
//  the Free Software Foundation, either version 3 of the License, or
//  (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU General Public License for more details.
//
//  You should have received a copy of the GNU General Public License
//  along with this program.  If not, see <https://www.gnu.org/licenses/>.

package config

import (
	"code.vikunja.io/api/pkg/log"
	"crypto/rand"
	"fmt"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"strings"
)

// InitConfig initializes the config, sets defaults etc.
func init() {

	// Set defaults
	// Service config
	random, err := random(32)
	if err != nil {
		log.Log.Fatal(err.Error())
	}

	// Service
	viper.SetDefault("service.JWTSecret", random)
	viper.SetDefault("service.interface", ":3456")
	viper.SetDefault("service.frontendurl", "")

	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)
	viper.SetDefault("service.rootpath", exPath)
	viper.SetDefault("service.pagecount", 50)
	viper.SetDefault("service.enablemetrics", false)
	// Database
	viper.SetDefault("database.type", "sqlite")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.user", "vikunja")
	viper.SetDefault("database.password", "")
	viper.SetDefault("database.database", "vikunja")
	viper.SetDefault("database.path", "./vikunja.db")
	viper.SetDefault("database.openconnections", 100)
	// Cacher
	viper.SetDefault("cache.enabled", false)
	viper.SetDefault("cache.type", "memory")
	viper.SetDefault("cache.maxelementsize", 1000)
	// Mailer
	viper.SetDefault("mailer.enabled", false)
	viper.SetDefault("mailer.host", "")
	viper.SetDefault("mailer.port", "587")
	viper.SetDefault("mailer.user", "user")
	viper.SetDefault("mailer.password", "")
	viper.SetDefault("mailer.skiptlsverify", false)
	viper.SetDefault("mailer.fromemail", "mail@vikunja")
	viper.SetDefault("mailer.queuelength", 100)
	viper.SetDefault("mailer.queuetimeout", 30)
	// Redis
	viper.SetDefault("redis.enabled", false)
	viper.SetDefault("redis.host", "localhost:6379")
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)
	// Logger
	viper.SetDefault("log.enabled", true)
	viper.SetDefault("log.errors", "stdout")
	viper.SetDefault("log.standard", "stdout")
	viper.SetDefault("log.database", "off")
	viper.SetDefault("log.http", "stdout")
	viper.SetDefault("log.echo", "off")
	viper.SetDefault("log.path", viper.GetString("service.rootpath")+"/logs")

	// Init checking for environment variables
	viper.SetEnvPrefix("vikunja")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Load the config file
	viper.AddConfigPath(viper.GetString("service.rootpath"))
	viper.AddConfigPath("/etc/vikunja/")
	viper.AddConfigPath("~/.config/vikunja")
	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	err = viper.ReadInConfig()
	if err != nil {
		log.Log.Info(err)
		log.Log.Info("Using defaults.")
	}
}

func random(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return fmt.Sprintf("%X", b), nil
}
