// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2020 Vikunja and contributors. All rights reserved.
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
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/routes"
	"code.vikunja.io/api/pkg/swagger"
	"code.vikunja.io/api/pkg/version"
	"context"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"time"
)

func init() {
	rootCmd.AddCommand(webCmd)
}

var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Starts the rest api web server",
	PreRun: func(cmd *cobra.Command, args []string) {
		fullInit()
	},
	Run: func(cmd *cobra.Command, args []string) {

		// Version notification
		log.Infof("Vikunja version %s", version.Version)

		// Additional swagger information
		swagger.SwaggerInfo.Version = version.Version

		// Start the webserver
		e := routes.NewEcho()
		routes.RegisterRoutes(e)
		// Start server
		go func() {
			if err := e.Start(config.ServiceInterface.GetString()); err != nil {
				e.Logger.Info("shutting down...")
			}
		}()

		// Wait for interrupt signal to gracefully shutdown the server with
		// a timeout of 10 seconds.
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt)
		<-quit
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		log.Infof("Shutting down...")
		if err := e.Shutdown(ctx); err != nil {
			e.Logger.Fatal(err)
		}
	},
}
