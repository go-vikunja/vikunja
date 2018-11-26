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

package main

import (
	"code.vikunja.io/api/docs"
	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/mail"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/routes"

	"context"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"time"
)

// Version sets the version to be printed to the user. Gets overwritten by "make release" or "make build" with last git commit or tag.
var Version = "0.1"

func main() {

	// Init logging
	log.InitLogger()

	// Init Config
	err := config.InitConfig()
	if err != nil {
		log.Log.Error(err.Error())
		os.Exit(1)
	}

	// Set Engine
	err = models.SetEngine()
	if err != nil {
		log.Log.Error(err.Error())
		os.Exit(1)
	}

	// Start the mail daemon
	mail.StartMailDaemon()

	// Version notification
	log.Log.Infof("Vikunja version %s", Version)

	// Additional swagger information
	docs.SwaggerInfo.Version = Version

	// Start the webserver
	e := routes.NewEcho()
	routes.RegisterRoutes(e)
	// Start server
	go func() {
		if err := e.Start(viper.GetString("service.interface")); err != nil {
			e.Logger.Info("shutting down...")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 10 seconds.
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	log.Log.Infof("Shutting down...")
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
