// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2021 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package cmd

import (
	"context"
	"net"
	"os"
	"os/signal"
	"time"

	"code.vikunja.io/api/pkg/cron"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/initialize"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/routes"
	"code.vikunja.io/api/pkg/swagger"
	"code.vikunja.io/api/pkg/utils"
	"code.vikunja.io/api/pkg/version"
	"github.com/labstack/echo/v4"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(webCmd)
}

func setupUnixSocket(e *echo.Echo) error {
	path := config.ServiceUnixSocket.GetString()

	// Remove old unix socket that may have remained after a crash
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}

	if config.ServiceUnixSocketMode.Get() != nil {
		// Use Umask instead of Chmod to prevent insecure race condition
		// (no-op on Windows)
		mode := config.ServiceUnixSocketMode.GetInt()
		oldmask := utils.Umask(0o777 &^ mode)
		defer utils.Umask(oldmask)
	}

	l, err := net.Listen("unix", path)
	if err != nil {
		return err
	}

	e.Listener = l
	return nil
}

var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Starts the rest api web server",
	PreRun: func(cmd *cobra.Command, args []string) {
		initialize.FullInit()
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
			// Listen unix socket if needed (ServiceInterface will be ignored)
			if config.ServiceUnixSocket.GetString() != "" {
				if err := setupUnixSocket(e); err != nil {
					e.Logger.Fatal(err)
				}
			}
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
		cron.Stop()
	},
}
