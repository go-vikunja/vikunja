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

package handler

import (
	"encoding/json"
	"fmt"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/modules/migration"
	"code.vikunja.io/api/pkg/notifications"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/getsentry/sentry-go"
)

func RegisterListeners() {
	events.RegisterListener((&MigrationRequestedEvent{}).Name(), &MigrationListener{})
}

// Only used for sentry
type migrationFailedError struct {
	MigratorKind  string
	OriginalError error
}

func (m *migrationFailedError) Error() string {
	return fmt.Sprintf("migration from %s failed, original error message was: %s", m.MigratorKind, m.OriginalError.Error())
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

	// unmarshalling again to make sure the migrator has the correct type now
	err = json.Unmarshal(msg.Payload, event)
	if err != nil {
		return
	}

	ms := event.Migrator.(migration.Migrator)

	m, err := migrateInListener(ms, event)
	if err != nil {
		log.Errorf("[Migration] Migration %d from %s for user %d failed. Error was: %s", m.ID, event.MigratorKind, event.User.ID, err.Error())

		var nerr error
		if config.SentryEnabled.GetBool() {
			nerr = notifications.Notify(event.User, &MigrationFailedReportedNotification{
				MigratorName: ms.Name(),
			})
			sentry.CaptureException(&migrationFailedError{
				MigratorKind:  event.MigratorKind,
				OriginalError: err,
			})
		} else {
			nerr = notifications.Notify(event.User, &MigrationFailedNotification{
				MigratorName: ms.Name(),
				Error:        err,
			})
		}
		if nerr != nil {
			log.Errorf("[Migration] Could not sent failed migration notification for migration %d to user %d, error was: %s", m.ID, event.User.ID, err.Error())
		}

		// Still need to finish the migration, otherwise restarting will not work
		err = migration.FinishMigration(m)
		if err != nil {
			log.Errorf("[Migration] Could not finish migration %d for user %d, error was: %s", m.ID, event.User.ID, err.Error())
		}
	}

	return nil // We do not want the queue to restart this job as we've already handled the error.
}

func migrateInListener(ms migration.Migrator, event *MigrationRequestedEvent) (m *migration.Status, err error) {
	m, err = migration.StartMigration(ms, event.User)
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
		log.Errorf("[Migration] Could not finish migration %d for user %d, error was: %s", m.ID, event.User.ID, err.Error())
		return
	}

	err = notifications.Notify(event.User, &MigrationDoneNotification{
		MigratorName: ms.Name(),
	})
	if err != nil {
		log.Errorf("[Migration] Could not sent migration success notification for migration %d to user %d, error was: %s", m.ID, event.User.ID, err.Error())
		return
	}

	log.Debugf("[Migration] Successfully done migration %d from %s for user %d", m.ID, event.MigratorKind, event.User.ID)
	return
}
