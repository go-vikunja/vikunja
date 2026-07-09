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
// registered and thus usable for login. A configured but unregistered
// provider was unreachable when Vikunja last initialized its providers and
// would stay broken until a restart (vikunja#3135) — registration is retried
// every minute instead, see RegisterProviderRegistrationCron.
type ProviderStatus struct {
	Key        string `json:"key" doc:"The config key of the provider."`
	Name       string `json:"name" doc:"The human-readable name of the provider."`
	Registered bool   `json:"registered" doc:"True when the provider is registered and can be used to log in. A configured but unregistered provider was unreachable when Vikunja last initialized its providers; registration is retried every minute."`
}

type configuredProvider struct {
	key, name string
}

func configuredProviders() (providers []configuredProvider) {
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
		providers = append(providers, configuredProvider{key: key, name: name})
	}
	sort.Slice(providers, func(i, j int) bool { return providers[i].key < providers[j].key })
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

// GetProvidersStatus returns the registration state of every configured
// OpenID Connect provider. It returns nil when OpenID Connect auth is
// disabled or no providers are configured.
//
// It enumerates providers from the raw config instead of GetAllProviders
// because the latter silently drops providers whose registration failed —
// exactly the ones a healthcheck needs to report.
func GetProvidersStatus() []ProviderStatus {
	if !config.AuthOpenIDEnabled.GetBool() {
		return nil
	}

	configured := configuredProviders()
	if len(configured) == 0 {
		return nil
	}

	registered := registeredProviderKeys()
	statuses := make([]ProviderStatus, 0, len(configured))
	for _, p := range configured {
		statuses = append(statuses, ProviderStatus{
			Key:        p.key,
			Name:       p.name,
			Registered: registered[p.key],
		})
	}
	return statuses
}

// registerMissingProviders re-runs provider registration when a configured
// provider is missing from the registered set. This heals the state from
// vikunja#3135: a provider that was down while Vikunja started stayed
// unusable for login until a manual restart.
func registerMissingProviders() {
	if !config.AuthOpenIDEnabled.GetBool() {
		return
	}

	configured := configuredProviders()
	if len(configured) == 0 {
		return
	}

	registered := registeredProviderKeys()
	var missing []string
	for _, p := range configured {
		if !registered[p.key] {
			missing = append(missing, p.key)
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

// RegisterProviderRegistrationCron periodically re-registers configured
// openid providers which could not be registered so far, typically because
// they were unreachable while Vikunja started.
func RegisterProviderRegistrationCron() {
	err := cron.Schedule("* * * * *", registerMissingProviders)
	if err != nil {
		log.Fatalf("Could not register openid provider registration cron: %s", err)
	}
}
