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

package openid

import (
	"context"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"code.vikunja.io/api/pkg/config"
)

// ProviderAvailability reports whether one configured OpenID Connect provider
// currently serves its discovery document.
type ProviderAvailability struct {
	Key       string `json:"key" doc:"The config key of the provider."`
	Name      string `json:"name" doc:"The human-readable name of the provider."`
	Reachable bool   `json:"reachable" doc:"True when the provider's OpenID Connect discovery endpoint responded with HTTP 200."`
}

const (
	availabilityCacheTTL     = time.Minute
	availabilityProbeTimeout = 5 * time.Second
)

// The healthcheck endpoint is hit frequently by orchestrator probes, so
// results are cached and refreshed in the background once stale — a down
// provider must not make every /health request block on the probe timeout.
var availabilityCache struct {
	sync.Mutex
	results    []ProviderAvailability
	fetchedAt  time.Time
	refreshing bool
}

type providerEndpoint struct {
	key, name, authURL string
}

func configuredProviderEndpoints() (endpoints []providerEndpoint) {
	for key, pi := range rawProviderConfigs() {
		name, _ := pi["name"].(string)
		authURL, _ := pi["authurl"].(string)
		if v := config.GetConfigValueFromFile("auth.openid.providers." + key + ".name"); v != "" {
			name = v
		}
		if v := config.GetConfigValueFromFile("auth.openid.providers." + key + ".authurl"); v != "" {
			authURL = v
		}
		if authURL == "" {
			continue
		}
		endpoints = append(endpoints, providerEndpoint{key: key, name: name, authURL: authURL})
	}
	sort.Slice(endpoints, func(i, j int) bool { return endpoints[i].key < endpoints[j].key })
	return
}

// CheckProvidersAvailability probes the discovery endpoint of every configured
// OpenID Connect provider and returns their reachability. It returns nil when
// OpenID Connect auth is disabled or no providers are configured.
//
// It enumerates providers from the raw config instead of GetAllProviders
// because the latter silently drops providers whose discovery failed at
// startup — exactly the ones a healthcheck needs to report.
func CheckProvidersAvailability(ctx context.Context) []ProviderAvailability {
	if !config.AuthOpenIDEnabled.GetBool() {
		return nil
	}

	endpoints := configuredProviderEndpoints()
	if len(endpoints) == 0 {
		return nil
	}

	availabilityCache.Lock()
	if !availabilityCache.fetchedAt.IsZero() {
		results := append([]ProviderAvailability(nil), availabilityCache.results...)
		if time.Since(availabilityCache.fetchedAt) >= availabilityCacheTTL && !availabilityCache.refreshing {
			availabilityCache.refreshing = true
			go func() {
				refreshed := probeProviders(context.Background(), endpoints)
				availabilityCache.Lock()
				defer availabilityCache.Unlock()
				availabilityCache.results = refreshed
				availabilityCache.fetchedAt = time.Now()
				availabilityCache.refreshing = false
			}()
		}
		availabilityCache.Unlock()
		return results
	}
	availabilityCache.Unlock()

	results := probeProviders(ctx, endpoints)
	availabilityCache.Lock()
	defer availabilityCache.Unlock()
	availabilityCache.results = results
	availabilityCache.fetchedAt = time.Now()
	return append([]ProviderAvailability(nil), results...)
}

func invalidateAvailabilityCache() {
	availabilityCache.Lock()
	defer availabilityCache.Unlock()
	availabilityCache.results = nil
	availabilityCache.fetchedAt = time.Time{}
}

func probeProviders(ctx context.Context, endpoints []providerEndpoint) []ProviderAvailability {
	results := make([]ProviderAvailability, len(endpoints))
	var wg sync.WaitGroup
	for i, e := range endpoints {
		wg.Add(1)
		go func(i int, e providerEndpoint) {
			defer wg.Done()
			results[i] = ProviderAvailability{
				Key:       e.key,
				Name:      e.name,
				Reachable: probeProvider(ctx, e.authURL),
			}
		}(i, e)
	}
	wg.Wait()
	return results
}

func probeProvider(ctx context.Context, authURL string) bool {
	// Same discovery URL construction as go-oidc's oidc.NewProvider.
	wellKnown := strings.TrimSuffix(authURL, "/") + "/.well-known/openid-configuration"

	ctx, cancel := context.WithTimeout(ctx, availabilityProbeTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, wellKnown, nil)
	if err != nil {
		return false
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}
