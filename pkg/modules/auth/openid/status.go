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
	"sort"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/cron"
	"code.vikunja.io/api/pkg/log"
)

// ProviderStatus reports whether one configured OpenID Connect provider is
// available for login. A configured but unavailable provider was unreachable
// when Vikunja last initialized its providers and would stay broken until a
// restart (vikunja#3135) — initialization is retried every minute instead,
// see RegisterProviderAvailabilityCron.
type ProviderStatus struct {
	Key       string `json:"key" doc:"The config key of the provider."`
	Available bool   `json:"available" doc:"True when the provider is initialized and can be used to log in. A configured but unavailable provider was unreachable or misconfigured when Vikunja last initialized its providers; initialization is retried every minute."`
}

func availableProviderKeys() map[string]bool {
	providers, err := GetAllProviders()
	if err != nil {
		log.Errorf("Could not get available openid providers: %s", err)
		return nil
	}
	keys := make(map[string]bool, len(providers))
	for _, p := range providers {
		keys[p.Key] = true
	}
	return keys
}

// GetProvidersStatus returns the availability of every configured OpenID
// Connect provider. It returns nil when OpenID Connect auth is disabled or
// no providers are configured.
//
// It enumerates providers from the raw config instead of GetAllProviders
// because the latter silently drops providers whose initialization failed —
// exactly the ones a healthcheck needs to report.
func GetProvidersStatus() []ProviderStatus {
	if !config.AuthOpenIDEnabled.GetBool() {
		return nil
	}

	configured := rawProviderConfigs()
	if len(configured) == 0 {
		return nil
	}

	available := availableProviderKeys()
	statuses := make([]ProviderStatus, 0, len(configured))
	for key := range configured {
		statuses = append(statuses, ProviderStatus{
			Key:       key,
			Available: available[key],
		})
	}
	sort.Slice(statuses, func(i, j int) bool { return statuses[i].Key < statuses[j].Key })
	return statuses
}

// initializeUnavailableProviders re-runs provider initialization when a
// configured provider is missing from the available set. This heals the
// state from vikunja#3135: a provider that was down while Vikunja started
// stayed unusable for login until a manual restart.
func initializeUnavailableProviders() {
	var unavailable []string
	for _, p := range GetProvidersStatus() {
		if !p.Available {
			unavailable = append(unavailable, p.Key)
		}
	}
	if len(unavailable) == 0 {
		return
	}

	log.Infof("Openid providers %v are configured but not available, retrying initialization", unavailable)
	CleanupSavedOpenIDProviders()

	available := availableProviderKeys()
	for _, key := range unavailable {
		if available[key] {
			log.Infof("Openid provider %s is now available", key)
		}
	}
}

// RegisterProviderAvailabilityCron periodically retries initializing
// configured openid providers which are not available, typically because
// they were unreachable while Vikunja started.
func RegisterProviderAvailabilityCron() {
	err := cron.Schedule("* * * * *", initializeUnavailableProviders)
	if err != nil {
		log.Fatalf("Could not register openid provider availability cron: %s", err)
	}
}
