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

package oauth2server

import (
	"net/url"
	"strings"

	"code.vikunja.io/api/pkg/models"
)

type Scope string

const (
	ScopeMCPRead  Scope = "mcp:read"
	ScopeMCPWrite Scope = "mcp:write"
)

// Resource is the protected resource scoped grants are issued for.
type Resource interface {
	// ResourceURL is the RFC 8707 resource indicator clients name it by.
	ResourceURL() string
	// ScopePermissions returns the API token permissions a scope grants right now.
	ScopePermissions(scope Scope) models.APIPermissions
}

// parseScope ignores unknown values: clients routinely add ones like offline_access.
func parseScope(raw string) ([]Scope, error) {
	read, write := false, false
	for _, value := range strings.Fields(raw) {
		switch Scope(value) {
		case ScopeMCPRead:
			read = true
		case ScopeMCPWrite:
			write = true
		}
	}
	switch {
	case write:
		return []Scope{
			ScopeMCPRead,
			ScopeMCPWrite,
		}, nil
	case read:
		return []Scope{ScopeMCPRead}, nil
	default:
		return nil, &models.ErrOAuthInvalidScope{}
	}
}

func requestsMCPScope(raw string) bool {
	_, err := parseScope(raw)
	return err == nil
}

func formatScope(scopes []Scope) string {
	values := make([]string, 0, len(scopes))
	for _, s := range scopes {
		values = append(values, string(s))
	}
	return strings.Join(values, " ")
}

// SameResource compares two resource indicators, ignoring the case of scheme and host and a trailing slash.
func SameResource(a, b string) bool {
	ua, ok := parseResource(a)
	if !ok {
		return false
	}
	ub, ok := parseResource(b)
	if !ok {
		return false
	}
	return strings.EqualFold(ua.Scheme, ub.Scheme) &&
		strings.EqualFold(ua.Host, ub.Host) &&
		strings.TrimRight(ua.Path, "/") == strings.TrimRight(ub.Path, "/")
}

// RFC 8707 §2: an absolute URI without a fragment. A query never names this instance's endpoint.
func parseResource(raw string) (*url.URL, bool) {
	u, err := url.Parse(raw)
	if err != nil || !u.IsAbs() || u.Host == "" || u.User != nil || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" || strings.Contains(raw, "#") {
		return nil, false
	}
	return u, true
}
