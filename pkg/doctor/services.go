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

package doctor

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/modules/auth/openid"
	"code.vikunja.io/api/pkg/red"
)

// CheckOptionalServices returns check groups for all enabled optional services.
func CheckOptionalServices() []CheckGroup {
	var groups []CheckGroup

	if config.RedisEnabled.GetBool() {
		groups = append(groups, checkRedis())
	}

	if config.TypesenseEnabled.GetBool() {
		groups = append(groups, checkTypesense())
	}

	if config.MailerEnabled.GetBool() {
		groups = append(groups, checkMailer())
	}

	if config.AuthLdapEnabled.GetBool() {
		groups = append(groups, checkLDAP())
	}

	if config.AuthOpenIDEnabled.GetBool() {
		groups = append(groups, checkOpenID())
	}

	return groups
}

func checkRedis() CheckGroup {
	r := red.GetRedis()
	if r == nil {
		return CheckGroup{
			Name: "Redis",
			Results: []CheckResult{
				{
					Name:   "Connection",
					Passed: false,
					Error:  "Redis client not initialized",
				},
			},
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := r.Ping(ctx).Err(); err != nil {
		return CheckGroup{
			Name: "Redis",
			Results: []CheckResult{
				{
					Name:   "Connection",
					Passed: false,
					Error:  err.Error(),
				},
			},
		}
	}

	return CheckGroup{
		Name: "Redis",
		Results: []CheckResult{
			{
				Name:   "Connection",
				Passed: true,
				Value:  fmt.Sprintf("OK (%s)", config.RedisHost.GetString()),
			},
		},
	}
}

func checkTypesense() CheckGroup {
	url := config.TypesenseURL.GetString()
	healthURL := url + "/health"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, healthURL, nil)
	if err != nil {
		return CheckGroup{
			Name: "Typesense",
			Results: []CheckResult{
				{
					Name:   "Connection",
					Passed: false,
					Error:  err.Error(),
				},
			},
		}
	}

	req.Header.Set("X-TYPESENSE-API-KEY", config.TypesenseAPIKey.GetString())

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return CheckGroup{
			Name: "Typesense",
			Results: []CheckResult{
				{
					Name:   "Connection",
					Passed: false,
					Error:  err.Error(),
				},
			},
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return CheckGroup{
			Name: "Typesense",
			Results: []CheckResult{
				{
					Name:   "Connection",
					Passed: false,
					Error:  fmt.Sprintf("health check returned status %d", resp.StatusCode),
				},
			},
		}
	}

	return CheckGroup{
		Name: "Typesense",
		Results: []CheckResult{
			{
				Name:   "Connection",
				Passed: true,
				Value:  fmt.Sprintf("OK (%s)", url),
			},
		},
	}
}

func checkMailer() CheckGroup {
	host := config.MailerHost.GetString()
	port := config.MailerPort.GetInt()

	if host == "" {
		return CheckGroup{
			Name: "Mailer",
			Results: []CheckResult{
				{
					Name:   "SMTP connection",
					Passed: false,
					Error:  "mailer host not configured",
				},
			},
		}
	}

	address := net.JoinHostPort(host, fmt.Sprintf("%d", port))

	// Simple TCP dial test with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dialer := &net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return CheckGroup{
			Name: "Mailer",
			Results: []CheckResult{
				{
					Name:   "SMTP connection",
					Passed: false,
					Error:  err.Error(),
				},
			},
		}
	}
	defer conn.Close()

	return CheckGroup{
		Name: "Mailer",
		Results: []CheckResult{
			{
				Name:   "SMTP connection",
				Passed: true,
				Value:  fmt.Sprintf("OK (%s)", address),
			},
		},
	}
}

func checkLDAP() CheckGroup {
	host := config.AuthLdapHost.GetString()
	port := config.AuthLdapPort.GetInt()
	useTLS := config.AuthLdapUseTLS.GetBool()

	address := net.JoinHostPort(host, fmt.Sprintf("%d", port))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var conn net.Conn
	var err error

	if useTLS {
		// #nosec G402
		tlsConfig := &tls.Config{
			InsecureSkipVerify: !config.AuthLdapVerifyTLS.GetBool(),
		}
		tlsDialer := &tls.Dialer{
			NetDialer: &net.Dialer{},
			Config:    tlsConfig,
		}
		conn, err = tlsDialer.DialContext(ctx, "tcp", address)
	} else {
		dialer := &net.Dialer{}
		conn, err = dialer.DialContext(ctx, "tcp", address)
	}

	if err != nil {
		return CheckGroup{
			Name: "LDAP",
			Results: []CheckResult{
				{
					Name:   "Connection",
					Passed: false,
					Error:  err.Error(),
				},
			},
		}
	}
	defer conn.Close()

	protocol := "ldap"
	if useTLS {
		protocol = "ldaps"
	}

	return CheckGroup{
		Name: "LDAP",
		Results: []CheckResult{
			{
				Name:   "Connection",
				Passed: true,
				Value:  fmt.Sprintf("OK (%s://%s)", protocol, address),
			},
		},
	}
}

func checkOpenID() CheckGroup {
	providers, err := openid.GetAllProviders()
	if err != nil {
		return CheckGroup{
			Name: "OpenID Connect",
			Results: []CheckResult{
				{
					Name:   "Providers",
					Passed: false,
					Error:  err.Error(),
				},
			},
		}
	}

	if len(providers) == 0 {
		return CheckGroup{
			Name: "OpenID Connect",
			Results: []CheckResult{
				{
					Name:   "Providers",
					Passed: true,
					Value:  "none configured",
				},
			},
		}
	}

	var results []CheckResult
	for _, provider := range providers {
		// The provider was already validated when loaded, so if it's in the list it's working
		results = append(results, CheckResult{
			Name:   fmt.Sprintf("Provider: %s", provider.Name),
			Passed: true,
			Value:  "OK",
		})
	}

	return CheckGroup{
		Name:    "OpenID Connect",
		Results: results,
	}
}
