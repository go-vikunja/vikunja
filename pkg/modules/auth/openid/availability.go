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
	"code.vikunja.io/api/pkg/cron"
	"code.vikunja.io/api/pkg/log"
)

// ProviderAvailability reports the health of one configured OpenID Connect
// provider.
type ProviderAvailability struct {
	Key  string `json:"key" doc:"The config key of the provider."`
	Name string `json:"name" doc:"The human-readable name of the provider."`
	// A provider can be reachable but not registered: if it was down when
	// Vikunja last built its provider list, it is missing from that list and
	// login through it fails even after it comes back (see vikunja#3135).
	// Registration is retried every minute until it succeeds.
	Registered bool `json:"registered" doc:"True when the provider is registered and can be used to log in. A configured but unregistered provider was unreachable when Vikunja last initialized its providers; registration is retried every minute."`
	Reachable  bool `json:"reachable" doc:"True when the provider's OpenID Connect discovery endpoint currently responds with HTTP 200."`
}

const (
	availabilityCacheTTL     = time.Minute
	availabilityProbeTimeout = 5 * time.Second
)

// The healthcheck endpoint is hit frequently by orchestrator probes, so it
// only ever serves cached results and refreshes them in the background — a
// down provider must not make /health block on the probe timeout.
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

func registeredProviderKeys() map[string]bool {
	providers, err := GetAllProviders()
	if err != nil {
		log.Errorf("Could not get registered openid providers: %s", err)
		return nil
	}
	keys := make(map[string]bool, len(providers))
	for _, p := range providers {
		keys[p.Key] = true
	}
	return keys
}

// CheckProvidersAvailability returns the cached availability of every
// configured OpenID Connect provider, kicking off a background refresh when
// the cache is stale. It returns nil when OpenID Connect auth is disabled, no
// providers are configured, or no probe has completed yet (right after
// startup).
//
// It enumerates providers from the raw config instead of GetAllProviders
// because the latter silently drops providers whose discovery failed at
// startup — exactly the ones a healthcheck needs to report.
func CheckProvidersAvailability(_ context.Context) []ProviderAvailability {
	if !config.AuthOpenIDEnabled.GetBool() {
		return nil
	}

	endpoints := configuredProviderEndpoints()
	if len(endpoints) == 0 {
		return nil
	}

	availabilityCache.Lock()
	defer availabilityCache.Unlock()

	results := append([]ProviderAvailability(nil), availabilityCache.results...)
	stale := time.Since(availabilityCache.fetchedAt) >= availabilityCacheTTL
	if (availabilityCache.fetchedAt.IsZero() || stale) && !availabilityCache.refreshing {
		availabilityCache.refreshing = true
		go func() {
			refreshed := ProbeProvidersAvailability(context.Background())
			availabilityCache.Lock()
			defer availabilityCache.Unlock()
			availabilityCache.results = refreshed
			availabilityCache.fetchedAt = time.Now()
			availabilityCache.refreshing = false
		}()
	}
	if availabilityCache.fetchedAt.IsZero() {
		return nil
	}
	return results
}

// ProbeProvidersAvailability synchronously probes every configured provider.
// Unlike CheckProvidersAvailability it never serves cached results, so it can
// block for several seconds — don't call it from a request handler.
func ProbeProvidersAvailability(ctx context.Context) []ProviderAvailability {
	if !config.AuthOpenIDEnabled.GetBool() {
		return nil
	}
	return probeEndpoints(ctx, configuredProviderEndpoints())
}

func probeEndpoints(ctx context.Context, endpoints []providerEndpoint) []ProviderAvailability {
	if len(endpoints) == 0 {
		return nil
	}

	registered := registeredProviderKeys()

	results := make([]ProviderAvailability, len(endpoints))
	var wg sync.WaitGroup
	for i, e := range endpoints {
		wg.Add(1)
		go func(i int, e providerEndpoint) {
			defer wg.Done()
			results[i] = ProviderAvailability{
				Key:        e.key,
				Name:       e.name,
				Registered: registered[e.key],
				Reachable:  probeProvider(ctx, e.authURL),
			}
		}(i, e)
	}
	wg.Wait()
	return results
}

func invalidateAvailabilityCache() {
	availabilityCache.Lock()
	defer availabilityCache.Unlock()
	availabilityCache.results = nil
	availabilityCache.fetchedAt = time.Time{}
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

// registerMissingProviders re-runs provider registration when a configured
// provider is missing from the registered set. This heals the state from
// vikunja#3135: a provider that was down while Vikunja started stayed
// unusable for login until a manual restart.
func registerMissingProviders() {
	if !config.AuthOpenIDEnabled.GetBool() {
		return
	}

	endpoints := configuredProviderEndpoints()
	if len(endpoints) == 0 {
		return
	}

	registered := registeredProviderKeys()
	var missing []string
	for _, e := range endpoints {
		if !registered[e.key] {
			missing = append(missing, e.key)
		}
	}
	if len(missing) == 0 {
		return
	}

	log.Infof("Openid providers %v are configured but not registered, retrying registration", missing)
	CleanupSavedOpenIDProviders()

	providers, err := GetAllProviders()
	if err != nil {
		log.Errorf("Error while re-registering openid providers: %s", err)
		return
	}
	nowRegistered := make(map[string]bool, len(providers))
	for _, p := range providers {
		nowRegistered[p.Key] = true
	}
	for _, key := range missing {
		if nowRegistered[key] {
			log.Infof("Openid provider %s successfully registered", key)
		}
	}
}

// RegisterProviderAvailabilityCron periodically re-registers configured
// openid providers which could not be registered so far, typically because
// they were unreachable while Vikunja started.
func RegisterProviderAvailabilityCron() {
	err := cron.Schedule("* * * * *", registerMissingProviders)
	if err != nil {
		log.Fatalf("Could not register openid provider availability cron: %s", err)
	}
}
