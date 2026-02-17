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

package yaegi

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"code.vikunja.io/api/pkg/plugins"
	"code.vikunja.io/api/pkg/yaegi_symbols"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

// LoadedPlugin holds a plugin loaded via Yaegi along with its optional capabilities.
// Because Yaegi wraps interpreted values per return type, sub-interface type assertions
// (e.g. Plugin -> AuthenticatedRouterPlugin) do not work. Instead, plugins must export
// typed factory functions for each capability they implement:
//
//   - NewPlugin() plugins.Plugin                                          (required)
//   - NewAuthenticatedRouterPlugin() plugins.AuthenticatedRouterPlugin    (optional)
//   - NewUnauthenticatedRouterPlugin() plugins.UnauthenticatedRouterPlugin (optional)
type LoadedPlugin struct {
	Plugin          plugins.Plugin
	AuthRouter      plugins.AuthenticatedRouterPlugin
	UnauthRouter    plugins.UnauthenticatedRouterPlugin
}

// LoadPlugin loads a plugin from a directory of Go source files using the Yaegi interpreter.
func LoadPlugin(dir string) (plugins.Plugin, error) {
	loaded, err := LoadPluginFull(dir)
	if err != nil {
		return nil, err
	}
	return loaded.Plugin, nil
}

// LoadPluginFull loads a plugin and all its optional capabilities via typed factory functions.
func LoadPluginFull(dir string) (*LoadedPlugin, error) {
	i := interp.New(interp.Options{})
	i.Use(stdlib.Symbols)
	i.Use(yaegi_symbols.Symbols)

	// Read all .go files in the directory
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading plugin dir %s: %w", dir, err)
	}

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") {
			continue
		}
		src, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", e.Name(), err)
		}
		_, err = i.Eval(string(src))
		if err != nil {
			return nil, fmt.Errorf("evaluating %s: %w", e.Name(), err)
		}
	}

	loaded := &LoadedPlugin{}

	// Required: NewPlugin
	v, err := i.Eval("main.NewPlugin")
	if err != nil {
		return nil, fmt.Errorf("looking up NewPlugin: %w", err)
	}
	newPlugin, ok := v.Interface().(func() plugins.Plugin)
	if !ok {
		return nil, fmt.Errorf("NewPlugin has wrong signature: %T", v.Interface())
	}
	loaded.Plugin = newPlugin()

	// Optional: NewAuthenticatedRouterPlugin
	if v, err := i.Eval("main.NewAuthenticatedRouterPlugin"); err == nil {
		if fn, ok := v.Interface().(func() plugins.AuthenticatedRouterPlugin); ok {
			loaded.AuthRouter = fn()
		}
	}

	// Optional: NewUnauthenticatedRouterPlugin
	if v, err := i.Eval("main.NewUnauthenticatedRouterPlugin"); err == nil {
		if fn, ok := v.Interface().(func() plugins.UnauthenticatedRouterPlugin); ok {
			loaded.UnauthRouter = fn()
		}
	}

	return loaded, nil
}
