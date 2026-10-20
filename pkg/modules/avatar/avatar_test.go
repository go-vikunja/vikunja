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

package avatar

import (
	"testing"

	"code.vikunja.io/api/pkg/config"

	"github.com/stretchr/testify/assert"
)

// GHSA-3x3m-2j93-w46j
func TestNormalizeAvatarSize(t *testing.T) {
	oldMax := config.ServiceMaxAvatarSize.GetInt64()
	t.Cleanup(func() { config.ServiceMaxAvatarSize.Set(oldMax) })

	config.ServiceMaxAvatarSize.Set(1024)
	for in, want := range map[int64]int64{
		-1000: 256,
		-1:    256,
		0:     256,
		1:     16,
		16:    16,
		17:    32,
		49:    64,
		250:   256,
		1000:  1024,
		1024:  1024,
		99999: 1024,
	} {
		assert.Equal(t, want, normalizeAvatarSize(in), "size %d", in)
	}

	t.Run("capped at the configured max", func(t *testing.T) {
		config.ServiceMaxAvatarSize.Set(100)
		assert.Equal(t, int64(64), normalizeAvatarSize(64))
		assert.Equal(t, int64(100), normalizeAvatarSize(65))
		assert.Equal(t, int64(100), normalizeAvatarSize(99999))
	})

	t.Run("max above the largest allowed size", func(t *testing.T) {
		config.ServiceMaxAvatarSize.Set(2048)
		assert.Equal(t, int64(1024), normalizeAvatarSize(1000))
		assert.Equal(t, int64(2048), normalizeAvatarSize(1500))
	})
}
