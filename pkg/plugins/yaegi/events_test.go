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
	"testing"

	"code.vikunja.io/api/pkg/log"
)

func TestPluginEventListener(t *testing.T) {
	log.InitLogger()

	loaded, err := LoadPluginFull(examplePluginDir)
	if err != nil {
		t.Fatalf("LoadPluginFull failed: %v", err)
	}

	// Call Init() — this registers the TestListener for TaskCreatedEvent.
	// If the Listener interface boundary is broken, this will panic with a
	// reflection error when calling events.RegisterListener.
	err = loaded.Plugin.Init()
	if err != nil {
		t.Fatalf("plugin Init failed: %v", err)
	}
	t.Log("Init() succeeded — events.RegisterListener accepted the interpreted Listener")

	// Verify Shutdown works too
	err = loaded.Plugin.Shutdown()
	if err != nil {
		t.Fatalf("plugin Shutdown failed: %v", err)
	}
	t.Log("Shutdown() succeeded")
}
