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
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"
)

// validateTLSFiles reports whether custom TLS files are configured and
// rejects incomplete or conflicting configurations.
func validateTLSFiles(certFile, keyFile string, autoTLSEnabled bool) (bool, error) {
	if certFile == "" && keyFile == "" {
		return false, nil
	}
	if certFile == "" || keyFile == "" {
		return false, errors.New("service.tlscert and service.tlskey must both be set to serve Vikunja over TLS")
	}
	if autoTLSEnabled {
		return false, errors.New("service.tlscert/service.tlskey and autotls.enabled cannot be used together, please use only one of them")
	}
	return true, nil
}

type certReloader struct {
	certFile string
	keyFile  string

	mu   sync.RWMutex
	cert *tls.Certificate
}

func newCertReloader(certFile, keyFile string) (*certReloader, error) {
	r := &certReloader{
		certFile: certFile,
		keyFile:  keyFile,
	}
	if err := r.reload(); err != nil {
		return nil, err
	}
	return r, nil
}

// reload keeps serving the previous certificate if the new one can't be loaded,
// so a half-written renewal doesn't take the server down.
func (r *certReloader) reload() error {
	cert, err := tls.LoadX509KeyPair(r.certFile, r.keyFile)
	if err != nil {
		return fmt.Errorf("could not load tls certificate %s and key %s: %w", r.certFile, r.keyFile, err)
	}
	r.mu.Lock()
	r.cert = &cert
	r.mu.Unlock()
	return nil
}

func (r *certReloader) GetCertificate(_ *tls.ClientHelloInfo) (*tls.Certificate, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.cert, nil
}

func (r *certReloader) reloadOnSIGHUP() {
	hup := make(chan os.Signal, 1)
	signal.Notify(hup, syscall.SIGHUP)
	for range hup {
		if err := r.reload(); err != nil {
			log.Errorf("Could not reload TLS certificate, still using the previous one: %s", err)
			continue
		}
		log.Infof("Reloaded TLS certificate from %s", r.certFile)
	}
}

// setupCustomTLS configures the server to use the certificate from
// service.tlscert and service.tlskey and reports whether TLS is enabled.
func setupCustomTLS(server *http.Server) bool {
	certFile := config.ServiceTLSCert.GetString()
	keyFile := config.ServiceTLSKey.GetString()

	enabled, err := validateTLSFiles(certFile, keyFile, config.AutoTLSEnabled.GetBool())
	if err != nil {
		log.Fatal(err)
	}
	if !enabled {
		return false
	}

	reloader, err := newCertReloader(certFile, keyFile)
	if err != nil {
		log.Fatal(err)
	}

	server.TLSConfig = &tls.Config{
		GetCertificate: reloader.GetCertificate,
		MinVersion:     tls.VersionTLS12,
	}

	go reloader.reloadOnSIGHUP()

	return true
}
