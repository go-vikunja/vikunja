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
	"sync"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/cron"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/utils"
)

// ProviderStatus reports whether one configured OpenID Connect provider is
// available for login. A configured but unavailable provider was unreachable
// when Vikunja last initialized its providers and would stay broken until a
// restart (vikunja#3135) — initialization is retried automatically instead,
// see RegisterProviderAvailabilityCron.
type ProviderStatus struct {
	Key       string `json:"key" doc:"The config key of the provider."`
	Available bool   `json:"available" doc:"True when the provider is initialized and offered for login. This reflects the last initialization attempt, not the provider's current reachability. A configured but unavailable provider was unreachable or misconfigured when Vikunja last initialized its providers; initialization is retried automatically with exponential backoff, after at most 15 minutes."`
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

func unavailableProviderKeys() []string {
	var unavailable []string
	for _, p := range GetProvidersStatus() {
		if !p.Available {
			unavailable = append(unavailable, p.Key)
		}
	}
	return unavailable
}

// initializeUnavailableProviders re-runs provider initialization for the
// given configured but unavailable providers. This heals the state from
// vikunja#3135: a provider that was down while Vikunja started stayed
// unusable for login until a manual restart.
func initializeUnavailableProviders(unavailable []string) {
	log.Infof("Openid providers %v are configured but not available, retrying initialization", unavailable)
	CleanupSavedOpenIDProviders()

	available := availableProviderKeys()
	for _, key := range unavailable {
		if available[key] {
			log.Infof("Openid provider %s is now available", key)
		}
	}
}

const (
	providerRetryBaseInterval = time.Minute
	providerRetryMaxInterval  = 15 * time.Minute
)

// providerRetryState paces the initialization retries with capped
// exponential backoff so a long-unavailable provider is not hammered on
// every cron tick.
var providerRetryState struct {
	sync.Mutex
	failures    int
	nextAttempt time.Time
}

// retryUnavailableProviders runs on every cron tick but only attempts
// initialization once the backoff delay has passed. The delay doubles per
// failed attempt from providerRetryBaseInterval up to
// providerRetryMaxInterval, with equal jitter (delay/2 + random(delay/2)) so
// many instances don't retry in lockstep. Any success resets the backoff.
func retryUnavailableProviders() {
	providerRetryState.Lock()
	defer providerRetryState.Unlock()

	unavailable := unavailableProviderKeys()
	if len(unavailable) == 0 {
		providerRetryState.failures = 0
		providerRetryState.nextAttempt = time.Time{}
		return
	}

	now := time.Now()
	if now.Before(providerRetryState.nextAttempt) {
		return
	}

	initializeUnavailableProviders(unavailable)

	unavailable = unavailableProviderKeys()
	if len(unavailable) == 0 {
		providerRetryState.failures = 0
		providerRetryState.nextAttempt = time.Time{}
		return
	}

	backoff := providerRetryBaseInterval << min(providerRetryState.failures, 10)
	if backoff > providerRetryMaxInterval {
		backoff = providerRetryMaxInterval
	}
	providerRetryState.failures++

	delay := backoff/2 + randomJitter(backoff/2)
	providerRetryState.nextAttempt = now.Add(delay)
	log.Debugf("Openid providers %v are still not available, retrying initialization after %s", unavailable, delay)
}

func randomJitter(limit time.Duration) time.Duration {
	n, err := utils.CryptoRandomInt(int64(limit))
	if err != nil {
		return limit / 2
	}
	return time.Duration(n)
}

// RegisterProviderAvailabilityCron periodically retries initializing
// configured openid providers which are not available, typically because
// they were unreachable while Vikunja started. Retries are paced with
// capped exponential backoff.
func RegisterProviderAvailabilityCron() {
	err := cron.Schedule("* * * * *", retryUnavailableProviders)
	if err != nil {
		log.Fatalf("Could not register openid provider availability cron: %s", err)
	}
}
