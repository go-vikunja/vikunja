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
	"fmt"
	"net"
	"net/http"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/modules/auth/ldap"
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
	// Initialize Redis
	red.InitRedis()

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

	protocol := "ldap"
	if useTLS {
		protocol = "ldaps"
	}

	// Use the actual LDAP connection function which tests bind credentials
	l, err := ldap.ConnectAndBindToLDAPDirectory()
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
	defer l.Close()

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
	// Parse raw config to get all providers (including ones that fail to connect)
	rawProviders := config.AuthOpenIDProviders.Get()
	if rawProviders == nil {
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

	// Convert to map[string]interface{}
	var providerMap map[string]interface{}
	switch p := rawProviders.(type) {
	case map[string]interface{}:
		providerMap = p
	case map[interface{}]interface{}:
		providerMap = make(map[string]interface{}, len(p))
		for k, v := range p {
			if key, ok := k.(string); ok {
				providerMap[key] = v
			}
		}
	default:
		return CheckGroup{
			Name: "OpenID Connect",
			Results: []CheckResult{
				{
					Name:   "Configuration",
					Passed: false,
					Error:  "invalid provider configuration format",
				},
			},
		}
	}

	if len(providerMap) == 0 {
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
	for key, p := range providerMap {
		result := checkOpenIDProvider(key, p)
		results = append(results, result)
	}

	return CheckGroup{
		Name:    "OpenID Connect",
		Results: results,
	}
}

func checkOpenIDProvider(key string, rawProvider interface{}) CheckResult {
	// Extract provider config
	var pi map[string]interface{}
	switch p := rawProvider.(type) {
	case map[string]interface{}:
		pi = p
	case map[interface{}]interface{}:
		pi = make(map[string]interface{}, len(p))
		for k, v := range p {
			if kStr, ok := k.(string); ok {
				pi[kStr] = v
			}
		}
	default:
		return CheckResult{
			Name:   fmt.Sprintf("Provider: %s", key),
			Passed: false,
			Error:  "invalid configuration format",
		}
	}

	// Get provider name
	name := key
	if n, ok := pi["name"].(string); ok {
		name = n
	}

	// Get auth URL
	authURL, ok := pi["authurl"].(string)
	if !ok || authURL == "" {
		return CheckResult{
			Name:   fmt.Sprintf("Provider: %s", name),
			Passed: false,
			Error:  "authurl not configured",
		}
	}

	// Check if the provider's discovery endpoint is reachable
	// OpenID Connect discovery is at /.well-known/openid-configuration
	discoveryURL := authURL
	if authURL[len(authURL)-1] != '/' {
		discoveryURL += "/"
	}
	discoveryURL += ".well-known/openid-configuration"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, discoveryURL, nil)
	if err != nil {
		return CheckResult{
			Name:   fmt.Sprintf("Provider: %s", name),
			Passed: false,
			Error:  err.Error(),
		}
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return CheckResult{
			Name:   fmt.Sprintf("Provider: %s", name),
			Passed: false,
			Error:  err.Error(),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return CheckResult{
			Name:   fmt.Sprintf("Provider: %s", name),
			Passed: false,
			Error:  fmt.Sprintf("discovery endpoint returned status %d", resp.StatusCode),
		}
	}

	return CheckResult{
		Name:   fmt.Sprintf("Provider: %s", name),
		Passed: true,
		Value:  "OK",
	}
}
