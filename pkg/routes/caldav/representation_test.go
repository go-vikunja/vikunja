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

package caldav

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLastModified(t *testing.T) {
	berlin := time.FixedZone("CEST", 2*60*60)

	assert.Equal(t, "Mon, 06 May 2024 05:08:09 GMT", lastModified(time.Date(2024, 5, 6, 7, 8, 9, 0, berlin)))
	assert.Equal(t, "Mon, 06 May 2024 07:08:09 GMT", lastModified(time.Date(2024, 5, 6, 7, 8, 9, 0, time.UTC)))
	assert.Equal(t, "Mon, 01 Jan 0001 00:00:00 GMT", lastModified(time.Time{}))
}
