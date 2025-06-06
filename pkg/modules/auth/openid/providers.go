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
	"fmt"
	"strconv"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/modules/keyvalue"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// GetAllProviders returns all configured providers
func GetAllProviders() (providers []*Provider, err error) {
	if !config.AuthOpenIDEnabled.GetBool() {
		return nil, nil
	}

	providers = []*Provider{}
	exists, err := keyvalue.GetWithValue("openid_providers", &providers)
	if !exists {
		rawProviders := config.AuthOpenIDProviders.Get()
		if rawProviders == nil {
			return nil, nil
		}

		rawProvider, is := rawProviders.(map[string]interface{})
		if !is {
			log.Criticalf("It looks like your openid configuration is in the wrong format. Please check the docs for the correct format.")
			return
		}

		for key, p := range rawProvider {
			var pi map[string]interface{}
			var is bool
			pi, is = p.(map[string]interface{})
			// JSON config is a map[string]interface{}, other providers are not. Under the hood they are all strings so
			// it is safe to cast.
			if !is {
				pis := p.(map[interface{}]interface{})
				pi = make(map[string]interface{}, len(pis))
				for i, s := range pis {
					pi[i.(string)] = s
				}
			}

			provider, err := getProviderFromMap(pi, key)

			if err != nil {
				log.Errorf("Error while getting openid provider %s: %s", key, err)
				continue
			}

			if provider == nil {
				log.Errorf("Could not openid provider %s, please check your config", key)
				continue
			}

			providers = append(providers, provider)

			err = keyvalue.Put("openid_provider_"+key, provider)
			if err != nil {
				return nil, err
			}
		}
		err = keyvalue.Put("openid_providers", providers)
	}

	return
}

// GetProvider retrieves a provider from keyvalue
func GetProvider(key string) (provider *Provider, err error) {
	provider = &Provider{}
	exists, err := keyvalue.GetWithValue("openid_provider_"+key, provider)
	if err != nil {
		return nil, err
	}
	if !exists {
		_, err = GetAllProviders() // This will put all providers in cache
		if err != nil {
			return nil, err
		}

		_, err = keyvalue.GetWithValue("openid_provider_"+key, provider)
		if err != nil {
			return nil, err
		}
	}

	err = provider.setOicdProvider()
	return
}

func getProviderFromMap(pi map[string]interface{}, key string) (provider *Provider, err error) {

	requiredKeys := []string{
		"name",
		"authurl",
		"clientsecret",
		"clientid",
	}

	allKeys := append(
		[]string{
			"logouturl",
			"scope",
			"emailfallback",
			"usernamefallback",
			"forceuserinfo",
		},
		requiredKeys...,
	)

	for _, configKey := range allKeys {
		valueFromFile := config.GetConfigValueFromFile("auth.openid.providers." + key + "." + configKey)
		if valueFromFile != "" {
			pi[configKey] = valueFromFile
		}
	}

	for _, key := range requiredKeys {
		if _, exists := pi[key]; !exists {
			return nil, fmt.Errorf("required key '%s' is missing in the provider configuration", key)
		}
	}

	name, is := pi["name"].(string)
	if !is {
		return nil, nil
	}

	var logoutURL string
	logoutValue, exists := pi["logouturl"]
	if exists {
		url, ok := logoutValue.(string)
		if ok {
			logoutURL = url
		}
	}

	var scope string
	if scopeValue, exists := pi["scope"]; exists {
		scope = scopeValue.(string)
	}
	if scope == "" {
		scope = "openid profile email"
	}

	var emailFallback = false
	emailFallbackValue, exists := pi["emailfallback"]
	if exists {
		emailFallbackTypedValue, ok := emailFallbackValue.(bool)
		if ok {
			emailFallback = emailFallbackTypedValue
		}
	}
	var usernameFallback = false
	usernameFallbackValue, exists := pi["usernamefallback"]
	if exists {
		usernameFallbackTypedValue, ok := usernameFallbackValue.(bool)
		if ok {
			usernameFallback = usernameFallbackTypedValue
		}
	}

	var forceUserInfo = false
	forceUserInfoValue, exists := pi["forceuserinfo"]
	if exists {
		forceUserInfoTypedValue, ok := forceUserInfoValue.(bool)
		if ok {
			forceUserInfo = forceUserInfoTypedValue
		} else {
			log.Errorf("forceuserinfo is not a boolean for provider %s, value: %v", key, forceUserInfoValue)
		}
	}

	provider = &Provider{
		Name:             name,
		Key:              key,
		AuthURL:          pi["authurl"].(string),
		OriginalAuthURL:  pi["authurl"].(string),
		ClientSecret:     pi["clientsecret"].(string),
		LogoutURL:        logoutURL,
		Scope:            scope,
		EmailFallback:    emailFallback,
		UsernameFallback: usernameFallback,
		ForceUserInfo:    forceUserInfo,
	}

	cl, is := pi["clientid"].(int)
	if is {
		provider.ClientID = strconv.Itoa(cl)
	} else {
		provider.ClientID = pi["clientid"].(string)
	}

	err = provider.setOicdProvider()
	if err != nil {
		return
	}

	provider.Oauth2Config = &oauth2.Config{
		ClientID:     provider.ClientID,
		ClientSecret: provider.ClientSecret,
		// Discovery returns the OAuth2 endpoints.
		Endpoint: provider.openIDProvider.Endpoint(),

		// "openid" is a required scope for OpenID Connect flows.
		Scopes: []string{oidc.ScopeOpenID, "profile", "email"},
	}

	provider.AuthURL = provider.Oauth2Config.Endpoint.AuthURL

	return
}

func CleanupSavedOpenIDProviders() {
	_ = keyvalue.Del("openid_providers")
}
