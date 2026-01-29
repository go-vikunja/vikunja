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
	"os"
	"path/filepath"
	"testing"
)

const examplePluginDir = "../../../examples/plugins/example"

func TestLoadPlugin(t *testing.T) {
	pluginDir := filepath.Join(examplePluginDir)

	mainGo := filepath.Join(pluginDir, "main.go")
	if _, err := os.Stat(mainGo); err != nil {
		t.Fatalf("example plugin source not found at %s: %v", mainGo, err)
	}

	p, err := LoadPlugin(pluginDir)
	if err != nil {
		t.Fatalf("LoadPlugin failed: %v", err)
	}

	if p.Name() != "example" {
		t.Errorf("expected plugin name 'example', got %q", p.Name())
	}
	if p.Version() != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %q", p.Version())
	}
}

func TestLoadPluginFull(t *testing.T) {
	loaded, err := LoadPluginFull(examplePluginDir)
	if err != nil {
		t.Fatalf("LoadPluginFull failed: %v", err)
	}

	if loaded.Plugin == nil {
		t.Fatal("Plugin is nil")
	}
	if loaded.Plugin.Name() != "example" {
		t.Errorf("expected plugin name 'example', got %q", loaded.Plugin.Name())
	}

	if loaded.AuthRouter == nil {
		t.Fatal("AuthRouter is nil — typed factory NewAuthenticatedRouterPlugin not found")
	}
	t.Logf("AuthRouter type: %T, name: %s", loaded.AuthRouter, loaded.AuthRouter.Name())

	if loaded.UnauthRouter == nil {
		t.Fatal("UnauthRouter is nil — typed factory NewUnauthenticatedRouterPlugin not found")
	}
	t.Logf("UnauthRouter type: %T, name: %s", loaded.UnauthRouter, loaded.UnauthRouter.Name())
}
