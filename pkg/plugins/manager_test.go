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
	"path/filepath"
	"testing"

	"code.vikunja.io/api/pkg/config"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestPluginPaths(t *testing.T) {
	setup := func(t *testing.T) {
		t.Helper()
		viper.Reset()
		t.Cleanup(viper.Reset)
		config.InitDefaultConfig()
	}

	t.Run("the default plugin dir follows the rootpath", func(t *testing.T) {
		setup(t)
		root := t.TempDir()
		config.ServiceRootpath.Set(root)

		assert.Equal(t, []string{filepath.Join(root, "plugins")}, pluginPaths())
	})

	t.Run("a relative plugins.dir is resolved against the rootpath", func(t *testing.T) {
		setup(t)
		root := t.TempDir()
		config.ServiceRootpath.Set(root)
		config.PluginsDir.Set("custom")

		assert.Equal(t, []string{filepath.Join(root, "custom")}, pluginPaths())
	})

	t.Run("an absolute plugins.dir ignores the rootpath", func(t *testing.T) {
		setup(t)
		config.ServiceRootpath.Set(t.TempDir())
		config.PluginsDir.Set("/opt/vikunja-plugins")

		assert.Equal(t, []string{"/opt/vikunja-plugins"}, pluginPaths())
	})
}
