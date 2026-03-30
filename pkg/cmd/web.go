// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package cmd

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/cron"
	"code.vikunja.io/api/pkg/initialize"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/plugins"
	"code.vikunja.io/api/pkg/routes"
	"code.vikunja.io/api/pkg/utils"
	"code.vikunja.io/api/pkg/version"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/acme/autocert"
)

func init() {
	rootCmd.AddCommand(webCmd)
}

func setupUnixSocket() (net.Listener, error) {
	path := config.ServiceUnixSocket.GetString()

	// Remove old unix socket that may have remained after a crash
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	if config.ServiceUnixSocketMode.Get() != nil {
		// Use Umask instead of Chmod to prevent insecure race condition
		// (no-op on Windows)
		mode := config.ServiceUnixSocketMode.GetInt()
		oldmask := utils.Umask(0o777 &^ mode)
		defer utils.Umask(oldmask)
	}

	cfg := net.ListenConfig{}
	return cfg.Listen(context.Background(), "unix", path)
}

func setupAutoTLS(server *http.Server) {
	if config.ServiceUnixSocket.GetString() != "" {
		log.Warning("Auto tls is enabled but listening on a unix socket is enabled as well. The latter will be ignored.")
	}
	if config.ServicePublicURL.GetString() == "" {
		log.Fatal("You must configure a publicurl to use autotls.")
	}
	parsed, err := url.Parse(config.ServicePublicURL.GetString())
	if err != nil {
		log.Fatalf("Could not parse hostname from publicurl: %s", err)
	}
	domain := parsed.Hostname()
	if domain == "" {
		log.Fatalf("The hostname cannot be empty. Please make sure the configured publicurl contains a hostname.")
	}
	if !strings.Contains(domain, ".") {
		log.Fatalf("The hostname must be a valid TLD. Please make sure the configured publicurl contains a valid TLD.")
	}
	renew, err := time.ParseDuration(config.AutoTLSRenewBefore.GetString())
	if err != nil {
		log.Fatalf("autotls.renewbefore must be a valid duration: %s", err)
	}
	if config.AutoTLSEmail.GetString() == "" {
		log.Fatalf("You must provide an email address to use autotls.")
	}

	manager := autocert.Manager{
		Prompt: autocert.AcceptTOS,
		Cache: autocert.DirCache(filepath.Join(
			config.FilesBasePath.GetString(),
			".certs",
		)),
		HostPolicy:  autocert.HostWhitelist(domain),
		RenewBefore: renew,
		Email:       config.AutoTLSEmail.GetString(),
	}

	server.TLSConfig = &tls.Config{
		GetCertificate: manager.GetCertificate,
		NextProtos:     []string{"h2", "http/1.1", "acme-tls/1"},
		MinVersion:     tls.VersionTLS12,
	}

	if config.ServiceInterface.GetString() != ":443" {
		log.Warningf("Vikunja's interface is set to %s, with tls it is recommended to set this to :443", config.ServiceInterface.GetString())
	}

	// Start HTTP server for ACME challenges
	go func() {
		httpServer := &http.Server{
			Addr:              ":http",
			Handler:           manager.HTTPHandler(nil),
			ReadHeaderTimeout: 10 * time.Second,
		}
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Errorf("HTTP server for ACME failed: %v", err)
		}
	}()

	log.Infof("HTTPS server listening on %s", config.ServiceInterface.GetString())
	err = server.ListenAndServeTLS("", "")
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Errorf("Server error: %v", err)
	}
	log.Info("shutting down...")
}

var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Starts the rest api web server",
	PreRun: func(_ *cobra.Command, _ []string) {
		initialize.FullInit()
	},
	Run: func(_ *cobra.Command, _ []string) {

		// Version notification
		log.Infof("Vikunja version %s", version.Version)

		// Start the webserver
		e := routes.NewEcho()
		routes.RegisterRoutes(e)

		// Create HTTP server with Echo as handler
		server := &http.Server{
			Addr:              config.ServiceInterface.GetString(),
			Handler:           e,
			ReadHeaderTimeout: 10 * time.Second,
		}

		// Start server
		go func() {
			if config.AutoTLSEnabled.GetBool() {
				setupAutoTLS(server)
				return
			}

			var err error
			// Listen unix socket if needed (ServiceInterface will be ignored)
			if config.ServiceUnixSocket.GetString() != "" {
				var listener net.Listener
				listener, err = setupUnixSocket()
				if err != nil {
					log.Fatalf("Failed to setup unix socket: %v", err)
				}
				log.Infof("HTTP server listening on unix socket %s", config.ServiceUnixSocket.GetString())
				err = server.Serve(listener)
			} else {
				log.Infof("HTTP server listening on %s", config.ServiceInterface.GetString())
				err = server.ListenAndServe()
			}

			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Errorf("Server error: %v", err)
			}
			log.Info("shutting down...")
		}()

		// Wait for interrupt signal to gracefully shut down the server with
		// a timeout of 10 seconds.
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt)
		<-quit
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		log.Infof("Shutting down...")
		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("Server shutdown failed: %v", err)
		}
		cron.Stop()
		plugins.Shutdown()
	},
}
