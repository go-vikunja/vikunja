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
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/errorreport"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/modules/migration"
	"code.vikunja.io/api/pkg/notifications"
	user2 "code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/getsentry/sentry-go"
)

func RegisterListeners() {
	events.RegisterListener((&MigrationRequestedEvent{}).Name(), &MigrationListener{})
	events.RegisterListener((&FileMigrationRequestedEvent{}).Name(), &FileMigrationListener{})

	// An instance that died mid-import left its upload behind.
	migration.CleanupSpooledUploads()
}

// Only used for sentry
type migrationFailedError struct {
	MigratorKind  string
	OriginalError error
}

func (m *migrationFailedError) Error() string {
	return fmt.Sprintf("migration from %s failed, original error message was: %s", m.MigratorKind, m.OriginalError.Error())
}

// shouldReportMigrationError filters out failures we cannot fix: a 4xx from the service we migrate from,
// or a domain error that maps to a 4xx, means the user's account, token, url or data is the problem, not
// Vikunja. The user still gets notified, with the actual message instead of the "we have been notified" one.
func shouldReportMigrationError(err error) bool {
	// ErrUpstreamRequestFailed maps to 502 no matter what the upstream said, so it needs its own check.
	var upstreamErr *migration.ErrUpstreamRequestFailed
	if errors.As(err, &upstreamErr) {
		return !upstreamErr.IsClientError()
	}

	var httpErr web.HTTPErrorProcessor
	if errors.As(err, &httpErr) {
		code := httpErr.HTTPError().HTTPCode
		return code < http.StatusBadRequest || code >= http.StatusInternalServerError
	}

	return true
}

// migrationFingerprint keeps failures apart by migrator and cause: every migration error reaches
// Sentry wrapped in the same migrationFailedError from the same call site.
func migrationFingerprint(migratorKind string, err error) []string {
	fingerprint := []string{"migration_failed", migratorKind}

	var upstreamErr *migration.ErrUpstreamRequestFailed
	if errors.As(err, &upstreamErr) {
		// The upstream body is user data, the status is what we can act on.
		return append(fingerprint, "upstream", strconv.Itoa(upstreamErr.StatusCode))
	}

	return append(fingerprint, errorreport.Fingerprint(err)...)
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

	mstr, has := registeredMigrators[event.MigratorKind]
	if !has {
		log.Errorf("[Migration] No migrator registered for kind %s, discarding event", event.MigratorKind)
		return nil
	}
	event.Migrator = mstr.MigrationStruct()

	// unmarshalling again to make sure the migrator has the correct type now
	err = json.Unmarshal(msg.Payload, event)
	if err != nil {
		return
	}

	ms := event.Migrator.(migration.Migrator)

	m, err := migrateInListener(ms, event)
	if err != nil {
		reportMigrationFailure(event.User, event.MigratorKind, ms, m, err)
	}

	return nil // We do not want the queue to restart this job as we've already handled the error.
}

// reportMigrationFailure notifies the user, reports what we can act on to
// Sentry and releases the claim so a retry is possible.
func reportMigrationFailure(u *user2.User, migratorKind string, ms migration.MigratorName, m *migration.Status, err error) {
	migrationID := int64(0)
	if m != nil {
		migrationID = m.ID
	}
	log.Errorf("[Migration] Migration %d from %s for user %d failed. Error was: %s", migrationID, migratorKind, u.ID, err.Error())

	var nerr error
	if config.SentryEnabled.GetBool() && shouldReportMigrationError(err) {
		nerr = notifications.Notify(u, &MigrationFailedReportedNotification{
			MigratorName: ms.Name(),
		})
		failure := &migrationFailedError{
			MigratorKind:  migratorKind,
			OriginalError: err,
		}
		sentry.WithScope(func(scope *sentry.Scope) {
			errorreport.ApplyFingerprint(scope, err, migrationFingerprint(migratorKind, err)...)
			sentry.CaptureException(failure)
		})
	} else {
		nerr = notifications.Notify(u, &MigrationFailedNotification{
			MigratorName: ms.Name(),
			Error:        err,
		})
	}
	if nerr != nil {
		log.Errorf("[Migration] Could not send failed migration notification for migration %d to user %d, error was: %s", migrationID, u.ID, nerr.Error())
	}

	// Still need to finish the migration, otherwise restarting will not work
	if m != nil {
		if ferr := migration.FinishMigration(m); ferr != nil {
			log.Errorf("[Migration] Could not finish migration %d for user %d, error was: %s", m.ID, u.ID, ferr.Error())
		}
	}
}

func migrateInListener(ms migration.Migrator, event *MigrationRequestedEvent) (m *migration.Status, err error) {
	if event.MigrationStatusID == 0 {
		// Events queued before claim support must acquire one during an upgrade.
		m, err = migration.ClaimMigration(ms, event.User)
		if err != nil {
			return
		}
	} else {
		m, err = migration.GetMigrationStatusByID(event.MigrationStatusID)
		if err != nil {
			return
		}
		if !m.FinishedAt.IsZero() {
			log.Debugf("[Migration] Skipping stale migration event for status %d of user %d", m.ID, event.User.ID)
			return
		}
	}

	// Convert panics to errors so the caller releases the claim.
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("[Migration] Migration %d from %s for user %d panicked: %v", m.ID, event.MigratorKind, event.User.ID, r)
			err = fmt.Errorf("migration panicked: %v", r)
		}
	}()

	log.Infof("[Migration] Starting migration %d from %s for user %d", m.ID, event.MigratorKind, event.User.ID)
	err = ms.Migrate(event.User)
	if err != nil {
		return
	}

	err = migration.FinishMigration(m)
	if err != nil {
		log.Errorf("[Migration] Could not finish migration %d for user %d, error was: %s", m.ID, event.User.ID, err.Error())
		return
	}

	if nerr := notifications.Notify(event.User, &MigrationDoneNotification{
		MigratorName: ms.Name(),
	}); nerr != nil {
		log.Errorf("[Migration] Could not send migration success notification for migration %d to user %d, error was: %s", m.ID, event.User.ID, nerr.Error())
	}

	log.Infof("[Migration] Successfully done migration %d from %s for user %d", m.ID, event.MigratorKind, event.User.ID)
	return m, nil
}

// FileMigrationListener runs queued file imports.
type FileMigrationListener struct {
}

// Name defines the name for the FileMigrationListener listener
func (s *FileMigrationListener) Name() string {
	return "migration.file.listener"
}

// Handle is executed when the event FileMigrationListener listens on is fired
func (s *FileMigrationListener) Handle(msg *message.Message) (err error) {
	event := &FileMigrationRequestedEvent{}
	if err = json.Unmarshal(msg.Payload, event); err != nil {
		log.Errorf("[Migration] Could not unmarshal file migration event, discarding it. Error was: %s", err.Error())
		return nil
	}

	defer migration.RemoveSpooledUpload(event.UploadName)

	factory, has := registeredFileMigrators[event.MigratorKind]
	if !has {
		log.Errorf("[Migration] No file migrator registered for kind %s, discarding event", event.MigratorKind)
		if ferr := migration.FinishMigration(&migration.Status{ID: event.MigrationStatusID}); ferr != nil {
			log.Errorf("[Migration] Could not finish migration %d for user %d, error was: %s", event.MigrationStatusID, event.User.ID, ferr.Error())
		}
		return nil
	}
	ms := factory()

	m, err := importInListener(ms, event)
	if err != nil {
		reportMigrationFailure(event.User, event.MigratorKind, ms, m, err)
	}

	return nil // A retry would re-run a partially applied import, so the error is handled here.
}

func importInListener(ms migration.FileMigrator, event *FileMigrationRequestedEvent) (m *migration.Status, err error) {
	m, err = migration.GetMigrationStatusByID(event.MigrationStatusID)
	if err != nil {
		return &migration.Status{ID: event.MigrationStatusID}, fmt.Errorf("could not get migration status %d: %w", event.MigrationStatusID, err)
	}
	if m.UserID != event.User.ID {
		return nil, fmt.Errorf("migration status %d does not belong to user %d", m.ID, event.User.ID)
	}
	if !m.FinishedAt.IsZero() {
		log.Debugf("[Migration] Skipping stale file migration event for status %d of user %d", m.ID, event.User.ID)
		return nil, nil
	}

	// Convert panics to errors so the caller releases the claim.
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("[Migration] Import %d from %s for user %d panicked: %v", m.ID, event.MigratorKind, event.User.ID, r)
			err = fmt.Errorf("migration panicked: %v", r)
		}
	}()

	if err := applyMigratorOptions(ms, event.Options); err != nil {
		return m, err
	}

	file, err := migration.OpenSpooledUpload(event.UploadName)
	if err != nil {
		// The wrapped error contains the server's file path, which must not reach the user's inbox.
		log.Errorf("[Migration] Could not open the spooled import file for migration %d: %s", m.ID, err.Error())
		return m, errors.New("the uploaded import file is no longer available, please upload it again")
	}
	defer file.Close()

	log.Infof("[Migration] Starting import %d from %s for user %d", m.ID, event.MigratorKind, event.User.ID)
	if err := ms.Migrate(event.User, file, event.UploadSize); err != nil {
		return m, asImportFileError(err)
	}

	if err := migration.FinishMigration(m); err != nil {
		return m, err
	}

	log.Infof("[Migration] Finished import %d from %s for user %d", m.ID, event.MigratorKind, event.User.ID)

	if err := notifications.Notify(event.User, &MigrationDoneNotification{MigratorName: ms.Name()}); err != nil {
		log.Errorf("[Migration] Could not send migration success notification for migration %d to user %d, error was: %s", m.ID, event.User.ID, err.Error())
	}

	return m, nil
}
