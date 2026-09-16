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

package client

import (
	"context"
	"errors"
	"net/url"
	"strings"

	"code.vikunja.io/veans/internal/output"
)

// defaultAPIPort is what `VIKUNJA_SERVICE_INTERFACE` ships with — handy
// when the user types just `myhost.example.com` for a default install
// running on an unusual port.
const defaultAPIPort = "3456"

// DiscoverServer normalizes `input` and probes a small set of plausible
// URLs for /api/v2/info, returning the canonical base URL (without the
// /api/v2 suffix — that's what client.New expects) and the parsed Info.
//
// Probing /api/v2/info doubles as the "is this server new enough" check: a
// Vikunja without /api/v2 fails discovery cleanly rather than limping along
// against endpoints veans needs.
//
// Mirrors the discovery the Vikunja web frontend does in
// helpers/checkAndSetApiUrl.ts: try the URL as-given, with the API path
// appended, and with the default :3456 port — across http / https. The
// first response that parses as Info wins.
func DiscoverServer(ctx context.Context, input string) (string, *Info, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", nil, output.New(output.CodeValidation, "server URL is required")
	}

	candidates, err := serverCandidates(input)
	if err != nil {
		return "", nil, output.Wrap(output.CodeValidation, err,
			"can't parse server URL %q: %v", input, err)
	}

	var attempts []string
	var lastErr error
	for _, base := range candidates {
		attempts = append(attempts, base+"/api/v2/info")
		info, err := New(base, "").Info(ctx)
		if err == nil && info != nil {
			return base, info, nil
		}
		lastErr = err
	}

	return "", nil, output.New(output.CodeValidation,
		"couldn't find a Vikunja instance reachable from %q — tried:\n  - %s\nlast error: %v",
		input, strings.Join(attempts, "\n  - "), lastErr)
}

// serverCandidates expands `input` into the ordered list of base URLs
// to probe for /api/v2/info. A "base URL" here is what client.New wants:
// the origin + the path that should sit BEFORE /api/v2 (typically empty
// or a reverse-proxy prefix). The probe itself adds /api/v2/info.
func serverCandidates(input string) ([]string, error) {
	// Strip a trailing /api/v1 or /api/v2[/] the user might have copied
	// from a curl example. We add the API path back in the probe, and
	// otherwise we'd end up calling /api/v2/api/v2/info.
	trimmed := strings.TrimRight(input, "/")
	trimmed = strings.TrimSuffix(trimmed, "/api/v1")
	trimmed = strings.TrimSuffix(trimmed, "/api/v2")
	trimmed = strings.TrimRight(trimmed, "/")

	withScheme := trimmed
	if !strings.HasPrefix(withScheme, "http://") && !strings.HasPrefix(withScheme, "https://") {
		withScheme = defaultScheme(trimmed) + "://" + trimmed
	}

	u, err := url.Parse(withScheme)
	if err != nil {
		return nil, err
	}
	if u.Host == "" {
		return nil, errors.New("missing host")
	}

	// Build the candidate set, dedup-preserving-order. The order here
	// is the search policy: as-given, with default port, then the
	// opposite scheme for each. Stops on the first one that responds
	// with a parseable Info.
	var bases []string
	add := func(scheme, host, path string) {
		base := scheme + "://" + host + strings.TrimRight(path, "/")
		base = strings.TrimRight(base, "/")
		for _, existing := range bases {
			if existing == base {
				return
			}
		}
		bases = append(bases, base)
	}

	hosts := []string{u.Host}
	if u.Port() == "" {
		hosts = append(hosts, u.Hostname()+":"+defaultAPIPort)
	}
	schemes := []string{u.Scheme}
	if u.Scheme == "https" {
		schemes = append(schemes, "http")
	} else {
		schemes = append(schemes, "https")
	}
	for _, s := range schemes {
		for _, h := range hosts {
			add(s, h, u.Path)
		}
	}
	return bases, nil
}

// defaultScheme picks http for loopback hosts and https for everything
// else — matches the heuristic most CLIs use when a scheme isn't typed.
func defaultScheme(input string) string {
	host := input
	if i := strings.IndexByte(host, '/'); i >= 0 {
		host = host[:i]
	}
	if i := strings.IndexByte(host, ':'); i >= 0 {
		host = host[:i]
	}
	switch host {
	case "localhost", "127.0.0.1", "[::1]", "::1":
		return "http"
	}
	return "https"
}
