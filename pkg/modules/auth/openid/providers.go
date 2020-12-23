// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2020 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package openid

import (
	"context"
	"regexp"
	"strconv"
	"strings"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/modules/keyvalue"
	kerr "code.vikunja.io/api/pkg/modules/keyvalue/error"
	"github.com/coreos/go-oidc"
	"golang.org/x/oauth2"
)

// GetAllProviders returns all configured providers
func GetAllProviders() (providers []*Provider, err error) {
	ps, err := keyvalue.Get("openid_providers")
	if err != nil && kerr.IsErrValueNotFoundForKey(err) {
		rawProviders := config.AuthOpenIDProviders.Get()
		if rawProviders == nil {
			return nil, nil
		}

		rawProvider := rawProviders.([]interface{})

		for _, p := range rawProvider {
			pi := p.(map[interface{}]interface{})

			provider, err := getProviderFromMap(pi)
			if err != nil {
				return nil, err
			}

			providers = append(providers, provider)

			k := getKeyFromName(pi["name"].(string))
			err = keyvalue.Put("openid_provider_"+k, provider)
			if err != nil {
				return nil, err
			}
		}
		err = keyvalue.Put("openid_providers", providers)
	}

	if ps != nil {
		return ps.([]*Provider), nil
	}

	return
}

// GetProvider retrieves a provider from keyvalue
func GetProvider(key string) (provider *Provider, err error) {
	var p interface{}
	p, err = keyvalue.Get("openid_provider_" + key)
	if err != nil && kerr.IsErrValueNotFoundForKey(err) {
		_, err = GetAllProviders() // This will put all providers in cache
		if err != nil {
			return nil, err
		}

		p, err = keyvalue.Get("openid_provider_" + key)
	}

	if p != nil {
		return p.(*Provider), nil
	}

	return nil, err
}

func getKeyFromName(name string) string {
	reg := regexp.MustCompile("[^a-z0-9]+")
	return reg.ReplaceAllString(strings.ToLower(name), "")
}

func getProviderFromMap(pi map[interface{}]interface{}) (*Provider, error) {
	k := getKeyFromName(pi["name"].(string))

	provider := &Provider{
		Name:         pi["name"].(string),
		Key:          k,
		AuthURL:      pi["authurl"].(string),
		ClientSecret: pi["clientsecret"].(string),
	}

	cl, is := pi["clientid"].(int)
	if is {
		provider.ClientID = strconv.Itoa(cl)
	} else {
		provider.ClientID = pi["clientid"].(string)
	}

	var err error
	provider.OpenIDProvider, err = oidc.NewProvider(context.Background(), provider.AuthURL)
	if err != nil {
		return nil, err
	}

	provider.Oauth2Config = &oauth2.Config{
		ClientID:     provider.ClientID,
		ClientSecret: provider.ClientSecret,
		RedirectURL:  config.AuthOpenIDRedirectURL.GetString() + k,

		// Discovery returns the OAuth2 endpoints.
		Endpoint: provider.OpenIDProvider.Endpoint(),

		// "openid" is a required scope for OpenID Connect flows.
		Scopes: []string{oidc.ScopeOpenID, "profile", "email"},
	}

	provider.AuthURL = provider.Oauth2Config.Endpoint.AuthURL

	return provider, nil
}
