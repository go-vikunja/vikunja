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

package plugins

import (
	"errors"
	"os"
	"path/filepath"
	goplugin "plugin"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/migration"

	"github.com/labstack/echo/v5"
)

// YaegiPluginLoader is a function that loads a plugin from a directory of Go source files.
// It is set by the yaegi package's init() to avoid an import cycle.
var YaegiPluginLoader func(dir string) (*LoadedYaegiPlugin, error)

// LoadedYaegiPlugin holds a plugin loaded via Yaegi along with its optional capabilities.
type LoadedYaegiPlugin struct {
	Plugin       Plugin
	AuthRouter   AuthenticatedRouterPlugin
	UnauthRouter UnauthenticatedRouterPlugin
}

// Manager handles loading and managing plugins.
type Manager struct {
	plugins                    []Plugin
	migrationPlugs             []MigrationPlugin
	authenticatedRouterPlugs   []AuthenticatedRouterPlugin
	unauthenticatedRouterPlugs []UnauthenticatedRouterPlugin
}

var manager = &Manager{}

// ManagerInstance returns the global plugin manager.
func ManagerInstance() *Manager { return manager }

// Initialize loads plugins and runs their migrations and init functions.
func Initialize() {
	if !config.PluginsEnabled.GetBool() {
		return
	}

	paths := []string{config.PluginsDir.GetString()}
	if err := manager.loadPlugins(paths); err != nil {
		log.Fatalf("Loading plugins failed: %v", err)
	}

	// Run plugin migrations after core migrations
	if len(manager.migrationPlugs) > 0 {
		migration.Migrate(nil)
	}

	for _, p := range manager.plugins {
		if err := p.Init(); err != nil {
			log.Errorf("Plugin %s failed to init: %s", p.Name(), err)
		}
	}
}

// Shutdown calls Shutdown on all loaded plugins.
func Shutdown() {
	for _, p := range manager.plugins {
		if err := p.Shutdown(); err != nil {
			log.Errorf("Plugin %s shutdown failed: %s", p.Name(), err)
		}
	}
}

// RegisterPluginRoutes registers routes from all router plugins.
func RegisterPluginRoutes(authenticated *echo.Group, unauthenticated *echo.Group) {
	// Register authenticated routes
	for _, p := range manager.authenticatedRouterPlugs {
		p.RegisterAuthenticatedRoutes(authenticated)
		log.Debugf("Registered authenticated routes for plugin %s", p.Name())
	}

	// Register unauthenticated routes
	for _, p := range manager.unauthenticatedRouterPlugs {
		p.RegisterUnauthenticatedRoutes(unauthenticated)
		log.Debugf("Registered unauthenticated routes for plugin %s", p.Name())
	}
}

func (m *Manager) loadPlugins(paths []string) error {
	loader := config.PluginsLoader.GetString()
	for _, p := range paths {
		entries, err := os.ReadDir(p)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return err
		}
		for _, e := range entries {
			full := filepath.Join(p, e.Name())
			switch loader {
			case "native":
				if filepath.Ext(e.Name()) != ".so" {
					continue
				}
				if err := m.loadNativePlugin(full); err != nil {
					log.Errorf("Failed to load native plugin %s: %s", e.Name(), err)
				}
			case "yaegi":
				if !e.IsDir() {
					continue
				}
				if err := m.loadYaegiPlugin(full); err != nil {
					log.Errorf("Failed to load yaegi plugin %s: %s", e.Name(), err)
				}
			}
		}
	}
	return nil
}

func (m *Manager) loadNativePlugin(path string) error {
	pl, err := goplugin.Open(path)
	if err != nil {
		return err
	}
	sym, err := pl.Lookup("NewPlugin")
	if err != nil {
		return err
	}
	newPlugin, ok := sym.(func() Plugin)
	if !ok {
		return errors.New("invalid plugin entry point")
	}
	p := newPlugin()
	m.registerPlugin(p)

	if mp, ok := p.(MigrationPlugin); ok {
		m.migrationPlugs = append(m.migrationPlugs, mp)
		migration.AddPluginMigrations(mp.Migrations())
	}

	if arp, ok := p.(AuthenticatedRouterPlugin); ok {
		m.authenticatedRouterPlugs = append(m.authenticatedRouterPlugs, arp)
	}

	if urp, ok := p.(UnauthenticatedRouterPlugin); ok {
		m.unauthenticatedRouterPlugs = append(m.unauthenticatedRouterPlugs, urp)
	}

	return nil
}

func (m *Manager) loadYaegiPlugin(dir string) error {
	if YaegiPluginLoader == nil {
		return errors.New("yaegi plugin loader not registered")
	}

	loaded, err := YaegiPluginLoader(dir)
	if err != nil {
		return err
	}

	m.registerPlugin(loaded.Plugin)

	if loaded.AuthRouter != nil {
		m.authenticatedRouterPlugs = append(m.authenticatedRouterPlugs, loaded.AuthRouter)
	}

	if loaded.UnauthRouter != nil {
		m.unauthenticatedRouterPlugs = append(m.unauthenticatedRouterPlugs, loaded.UnauthRouter)
	}

	return nil
}

func (m *Manager) registerPlugin(p Plugin) {
	m.plugins = append(m.plugins, p)
	log.Infof("Loaded plugin %s v%s", p.Name(), p.Version())
}
