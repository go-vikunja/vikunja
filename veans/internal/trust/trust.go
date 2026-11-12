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

// Package trust gates the server from .veans.yml, which is repo-controlled.
package trust

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"unicode"

	"code.vikunja.io/veans/internal/fsutil"
	"code.vikunja.io/veans/internal/output"
)

var ErrUntrusted = errors.New("server not trusted")

func Origin(raw string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return "", err
	}
	// client.New concatenates the raw URL, so https://trusted@evil reaches evil.
	if u.User != nil {
		return "", errors.New("server URL must not contain user info")
	}
	scheme := strings.ToLower(u.Scheme)
	port := u.Port()
	switch scheme {
	case "https":
		port = cmp.Or(port, "443")
	case "http":
		port = cmp.Or(port, "80")
	default:
		return "", fmt.Errorf("unsupported scheme %q", u.Scheme)
	}
	host := strings.ToLower(u.Hostname())
	if host == "" {
		return "", errors.New("missing host")
	}
	for _, r := range host {
		if r > unicode.MaxASCII {
			return "", fmt.Errorf("host %q contains non-ASCII characters; use its punycode (xn--) form", host)
		}
	}
	return scheme + "://" + net.JoinHostPort(host, port), nil
}

// Lives in the user's home so a hostile checkout can't ship a pre-approved server.
type Store struct {
	path string
}

type storeSchema struct {
	Origins []string `json:"origins"`
}

func NewStore(path string) *Store { return &Store{path: path} }

// XDG_CONFIG_HOME is ignored: a hostile .envrc could point it at a repo file.
func DefaultStore() *Store {
	h, err := os.UserHomeDir()
	if err != nil || h == "" {
		return &Store{}
	}
	return &Store{path: filepath.Join(h, ".config", "veans", "trusted.json")}
}

func (s *Store) load() (*storeSchema, error) {
	var sc storeSchema
	if s.path == "" {
		return &sc, nil
	}
	buf, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return &sc, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(buf, &sc); err != nil {
		return nil, fmt.Errorf("parse %s: %w", s.path, err)
	}
	return &sc, nil
}

func (s *Store) has(origin string) (bool, error) {
	sc, err := s.load()
	if err != nil {
		return false, err
	}
	return slices.Contains(sc.Origins, origin), nil
}

func (s *Store) add(origin string) error {
	if s.path == "" {
		return errors.New("no trust store path: $HOME is unset")
	}
	sc, err := s.load()
	if err != nil {
		return err
	}
	if slices.Contains(sc.Origins, origin) {
		return nil
	}
	sc.Origins = append(sc.Origins, origin)
	buf, err := json.MarshalIndent(sc, "", "  ")
	if err != nil {
		return err
	}
	return fsutil.WriteFileAtomic(s.path, buf, 0o600)
}

type Prompter interface {
	ReadLine(prompt string) (string, error)
}

type Gate struct {
	Store    *Store
	Prompter Prompter
}

func Default(p Prompter) *Gate {
	return &Gate{Store: DefaultStore(), Prompter: p}
}

func (g *Gate) Check(server string) error {
	origin, err := Origin(server)
	if err != nil {
		return output.Wrap(output.CodeValidation, err, "invalid server %q: %v", server, err)
	}
	// VEANS_SERVER is how CI declares trust without a prompt.
	if env := os.Getenv("VEANS_SERVER"); env != "" {
		if envOrigin, err := Origin(env); err == nil && envOrigin == origin {
			return nil
		}
	}
	ok, err := g.Store.has(origin)
	if err != nil {
		return output.Wrap(output.CodeUnknown, err, "read trust store: %v", err)
	}
	if ok {
		return nil
	}
	if g.Prompter == nil {
		return untrusted(origin)
	}
	ans, err := g.Prompter.ReadLine(fmt.Sprintf(
		"veans has not been told to trust %s.\n"+
			"It will send your credentials and API requests there. Trust it? [y/N]: ", origin))
	if err != nil {
		return output.Wrap(output.CodeUnknown, err, "read trust confirmation: %v", err)
	}
	switch strings.ToLower(strings.TrimSpace(ans)) {
	case "y", "yes":
		return g.record(origin)
	}
	return untrusted(origin)
}

// Record skips the prompt; only for a server the user typed.
func (g *Gate) Record(server string) error {
	origin, err := Origin(server)
	if err != nil {
		return output.Wrap(output.CodeValidation, err, "invalid server %q: %v", server, err)
	}
	return g.record(origin)
}

func (g *Gate) record(origin string) error {
	if err := g.Store.add(origin); err != nil {
		return output.Wrap(output.CodeUnknown, err, "record trust for %s: %v", origin, err)
	}
	return nil
}

func untrusted(origin string) error {
	return output.Wrap(output.CodeAuth, ErrUntrusted,
		"server %s is not trusted — run any veans command in a terminal to confirm it, or set VEANS_SERVER to it", origin)
}
