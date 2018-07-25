package main

import (
	"code.vikunja.io/api/models"
	"code.vikunja.io/api/routes"

	"context"
	"fmt"
	"os"
	"os/signal"
	"time"
)

// Version sets the version to be printed to the user. Gets overwritten by "make release" or "make build" with last git commit or tag.
var Version = "0.1"

func main() {

	// Init Config
	err := models.SetConfig()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Set Engine
	err = models.SetEngine()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Version notification
	fmt.Println("Vikunja version", Version)

	// Start the webserver
	e := routes.NewEcho()
	routes.RegisterRoutes(e)
	// Start server
	go func() {
		if err := e.Start(models.Config.Interface); err != nil {
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
	fmt.Println("Shutting down...")
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
