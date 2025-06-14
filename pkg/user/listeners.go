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

package user

import (
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/metrics"
	"code.vikunja.io/api/pkg/modules/keyvalue"
	"github.com/ThreeDotsLabs/watermill/message"
)

func RegisterListeners() {
	events.RegisterListener((&CreatedEvent{}).Name(), &IncreaseUserCounter{})
}

///////
// User Events

// IncreaseUserCounter  represents a listener
type IncreaseUserCounter struct {
}

// Name defines the name for the IncreaseUserCounter listener
func (s *IncreaseUserCounter) Name() string {
	return "increase.user.counter"
}

// Handle is executed when the event IncreaseUserCounter listens on is fired
func (s *IncreaseUserCounter) Handle(_ *message.Message) (err error) {
	return keyvalue.IncrBy(metrics.UserCountKey, 1)
}
