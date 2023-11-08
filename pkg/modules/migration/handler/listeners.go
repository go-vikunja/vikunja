// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package handler

import (
	"encoding/json"

	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/modules/migration"
	"code.vikunja.io/api/pkg/notifications"

	"github.com/ThreeDotsLabs/watermill/message"
)

func RegisterListeners() {
	events.RegisterListener((&MigrationRequestedEvent{}).Name(), &MigrationListener{})
}

// MigrationListener  represents a listener
type MigrationListener struct {
}

// Name defines the name for the MigrationListener listener
func (s *MigrationListener) Name() string {
	return "migration.listener"
}

// Handle is executed when the event MigrationListener listens on is fired
func (s *MigrationListener) Handle(msg *message.Message) (err error) {
	event := &MigrationRequestedEvent{}
	err = json.Unmarshal(msg.Payload, event)
	if err != nil {
		return
	}

	mstr := registeredMigrators[event.MigratorKind]
	event.Migrator = mstr.MigrationStruct()

	// unmarshaling again to make sure the migrator has the correct type now
	err = json.Unmarshal(msg.Payload, event)
	if err != nil {
		return
	}

	ms := event.Migrator.(migration.Migrator)

	m, err := migration.StartMigration(ms, event.User)
	if err != nil {
		return
	}

	log.Debugf("[Migration] Starting migration %d from %s for user %d", m.ID, event.MigratorKind, event.User.ID)

	err = ms.Migrate(event.User)
	if err != nil {
		return
	}

	err = migration.FinishMigration(m)
	if err != nil {
		return
	}

	err = notifications.Notify(event.User, &MigrationDoneNotification{
		MigratorName: ms.Name(),
	})
	if err != nil {
		return
	}

	log.Debugf("[Migration] Successfully done migration %d from %s for user %d", m.ID, event.MigratorKind, event.User.ID)
	return
}
