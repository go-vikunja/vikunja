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

package notifications

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	isUnderTest           bool
	sentTestNotifications []Notification
)

func Fake() {
	isUnderTest = true
	sentTestNotifications = nil
}

// AssertSent asserts a notification has been sent
func AssertSent(t *testing.T, n Notification) {
	var found bool
	for _, testNotification := range sentTestNotifications {
		if n.Name() == testNotification.Name() {
			found = true
			break
		}
	}

	assert.True(t, found, "Failed to assert "+n.Name()+" has been sent.")
}
