package main

import (
	"code.vikunja.io/api/models"
	"code.vikunja.io/api/models/mail"
	"code.vikunja.io/api/routes"

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
	models.InitLogger()

	// Init Config
	err := models.InitConfig()
	if err != nil {
		models.Log.Error(err.Error())
		os.Exit(1)
	}

	// Set Engine
	err = models.SetEngine()
	if err != nil {
		models.Log.Error(err.Error())
		os.Exit(1)
	}

	// Start the mail daemon
	mail.StartMailDaemon()

	// Version notification
	models.Log.Infof("Vikunja version %s", Version)

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
	models.Log.Infof("Sutting down...")
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
