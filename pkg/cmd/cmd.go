// Vikunja is a todo-list application to facilitate your life.
// Copyright 2019 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package cmd

import (
	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/mail"
	"code.vikunja.io/api/pkg/migration"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/red"
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

func init() {
	cobra.OnInitialize(initialize)
}

var rootCmd = &cobra.Command{
	Use:   "vikunja",
	Short: "Vikunja is the to-do app to organize your life.",
	Long: `Vikunja (/vɪˈkuːnjə/)
The to-do app to organize your life.

Also one of the two wild South American camelids which live in the high
alpine areas of the Andes and a relative of the llama.

Vikunja is a self-hosted To-Do list application with a web app and mobile apps for all platforms. It is licensed under the GPLv3.

Find more info at vikunja.io.`,
	Run: webCmd.Run,
}

// Execute starts the application
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// Initializes all kinds of things in the right order
func initialize() {
	// Init the config
	config.InitConfig()

	// Init redis
	red.InitRedis()

	// Set logger
	log.InitLogger()

	// Run the migrations
	migration.Migrate(nil)

	// Set Engine
	err := models.SetEngine()
	if err != nil {
		log.Fatal(err.Error())
	}
	err = files.SetEngine()
	if err != nil {
		log.Fatal(err.Error())
	}

	// Initialize the files handler
	files.InitFileHandler()

	// Start the mail daemon
	mail.StartMailDaemon()
}
