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
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/modules/migration"
	"code.vikunja.io/api/pkg/modules/migration/planka"
	vikunjafile "code.vikunja.io/api/pkg/modules/migration/vikunja-file"
	"code.vikunja.io/api/pkg/notifications"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// serverSideDomainError stands in for a domain error which maps to a 5xx: those are ours to fix.
type serverSideDomainError struct{}

func (e *serverSideDomainError) Error() string { return "something broke on our end" }

func (e *serverSideDomainError) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusInternalServerError,
		Code:     9999,
		Message:  "something broke on our end",
	}
}

func TestShouldReportMigrationError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "upstream 404",
			err:  migration.NewErrUpstreamRequestFailed("microsoft graph", 404, `{"error":{"code":"itemNotFound"}}`),
			want: false,
		},
		{
			name: "upstream 401",
			err:  migration.NewErrUpstreamRequestFailed("todoist oauth", 401, "unauthorized"),
			want: false,
		},
		{
			name: "wrapped upstream 404",
			err:  fmt.Errorf("could not get data: %w", migration.NewErrUpstreamRequestFailed("microsoft graph", 404, "")),
			want: false,
		},
		{
			name: "upstream 500",
			err:  migration.NewErrUpstreamRequestFailed("planka", 500, "boom"),
			want: true,
		},
		{
			name: "wrapped upstream 503",
			err:  fmt.Errorf("could not get data: %w", migration.NewErrUpstreamRequestFailed("planka", 503, "")),
			want: true,
		},
		{
			name: "planka invalid credentials",
			err:  &planka.ErrInvalidCredentials{},
			want: false,
		},
		{
			name: "planka no api at url",
			err:  &planka.ErrNoPlankaAtURL{Reason: "404"},
			want: false,
		},
		{
			name: "wrapped 4xx domain error",
			err:  fmt.Errorf("could not migrate: %w", &migration.ErrNotAZipFile{}),
			want: false,
		},
		{
			name: "5xx domain error",
			err:  &serverSideDomainError{},
			want: true,
		},
		{
			name: "arbitrary error",
			err:  errors.New("something went wrong"),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, shouldReportMigrationError(tt.err))
		})
	}
}

type stubFileMigratorState struct {
	runs    int
	content []byte
	options []byte
}

type stubFileMigrator struct {
	name       string
	migrateErr error
	state      *stubFileMigratorState
}

func (m *stubFileMigrator) Name() string { return m.name }

func (m *stubFileMigrator) SetOptions(options []byte) error {
	m.state.options = options
	return nil
}

func (m *stubFileMigrator) Migrate(_ *user.User, file io.ReaderAt, size int64) error {
	m.state.runs++
	content := make([]byte, size)
	if _, err := file.ReadAt(content, 0); err != nil {
		return err
	}
	m.state.content = content
	return m.migrateErr
}

// registerStubFileMigrator returns the state shared by every instance the listener builds from the factory.
func registerStubFileMigrator(name string, migrateErr error) *stubFileMigratorState {
	state := &stubFileMigratorState{}
	RegisterFileMigrator(func() migration.FileMigrator {
		return &stubFileMigrator{name: name, migrateErr: migrateErr, state: state}
	})
	return state
}

func useTestSpoolDir(t *testing.T) {
	t.Helper()
	previous := config.FilesBasePath.GetString()
	config.FilesBasePath.Set(t.TempDir())
	t.Cleanup(func() { config.FilesBasePath.Set(previous) })
}

func spoolTestUpload(t *testing.T, src io.Reader) (name string, size int64) {
	t.Helper()
	name, size, err := migration.SpoolUpload(src)
	require.NoError(t, err)
	return name, size
}

func assertSpoolRemoved(t *testing.T, name string) {
	t.Helper()
	_, err := migration.OpenSpooledUpload(name)
	require.Error(t, err)
	assert.True(t, os.IsNotExist(err), "the spooled upload must be gone, got %v", err)
}

func assertClaimReleased(t *testing.T, statusID int64) {
	t.Helper()
	status, err := migration.GetMigrationStatusByID(statusID)
	require.NoError(t, err)
	assert.False(t, status.FinishedAt.IsZero())
	assert.Nil(t, status.ActiveUserID)
}

func assertMigrationOutcome(t *testing.T, statusID int64, wantErrorMessage string) {
	t.Helper()
	status, err := migration.GetMigrationStatusByID(statusID)
	require.NoError(t, err)
	assert.Equal(t, wantErrorMessage, status.ErrorMessage)
}

func enableSentry(t *testing.T) {
	t.Helper()
	previous := config.SentryEnabled.GetBool()
	config.SentryEnabled.Set(true)
	t.Cleanup(func() { config.SentryEnabled.Set(previous) })
}

func TestFileMigrationListenerImportsSpooledUpload(t *testing.T) {
	clearMigrationStatus(t)
	useTestSpoolDir(t)
	notifications.Fake()
	t.Cleanup(notifications.Unfake)

	RegisterFileMigrator(func() migration.FileMigrator { return &vikunjafile.FileMigrator{} })
	u := getTestUser(t)

	export, err := os.Open("../vikunja-file/export.zip")
	require.NoError(t, err)
	defer export.Close()
	uploadName, uploadSize := spoolTestUpload(t, export)

	status, err := migration.ClaimMigration(&vikunjafile.FileMigrator{}, u)
	require.NoError(t, err)

	events.TestListener(t, &FileMigrationRequestedEvent{
		User:              u,
		MigratorKind:      "vikunja-file",
		MigrationStatusID: status.ID,
		UploadName:        uploadName,
		UploadSize:        uploadSize,
	}, &FileMigrationListener{})

	db.AssertExists(t, "projects", map[string]interface{}{
		"title":    "test project",
		"owner_id": u.ID,
	}, false)
	db.AssertExists(t, "tasks", map[string]interface{}{
		"title":         "some other task",
		"created_by_id": u.ID,
	}, false)
	db.AssertExists(t, "labels", map[string]interface{}{
		"title":         "test",
		"created_by_id": u.ID,
	}, false)

	notifications.AssertSent(t, &MigrationDoneNotification{})
	assertClaimReleased(t, status.ID)
	assertMigrationOutcome(t, status.ID, "")
	assertSpoolRemoved(t, uploadName)
}

func TestFileMigrationListenerFailedImportReleasesClaim(t *testing.T) {
	clearMigrationStatus(t)
	useTestSpoolDir(t)
	notifications.Fake()
	t.Cleanup(notifications.Unfake)

	state := registerStubFileMigrator("failing-file-stub", errors.New("import broke"))
	u := getTestUser(t)
	uploadName, uploadSize := spoolTestUpload(t, strings.NewReader("some export"))

	status, err := migration.ClaimMigration(&stubFileMigrator{name: "failing-file-stub"}, u)
	require.NoError(t, err)

	events.TestListener(t, &FileMigrationRequestedEvent{
		User:              u,
		MigratorKind:      "failing-file-stub",
		MigrationStatusID: status.ID,
		UploadName:        uploadName,
		UploadSize:        uploadSize,
	}, &FileMigrationListener{})

	assert.Equal(t, 1, state.runs)
	assert.Equal(t, "some export", string(state.content))
	notifications.AssertSent(t, &MigrationFailedNotification{})
	assertClaimReleased(t, status.ID)
	assertMigrationOutcome(t, status.ID, "import broke")
	assertSpoolRemoved(t, uploadName)
}

func TestFileMigrationListenerAppliesOptions(t *testing.T) {
	clearMigrationStatus(t)
	useTestSpoolDir(t)

	state := registerStubFileMigrator("options-file-stub", nil)
	u := getTestUser(t)
	uploadName, uploadSize := spoolTestUpload(t, strings.NewReader("some export"))

	status, err := migration.ClaimMigration(&stubFileMigrator{name: "options-file-stub"}, u)
	require.NoError(t, err)

	events.TestListener(t, &FileMigrationRequestedEvent{
		User:              u,
		MigratorKind:      "options-file-stub",
		MigrationStatusID: status.ID,
		UploadName:        uploadName,
		UploadSize:        uploadSize,
		Options:           []byte(`{"delimiter":";"}`),
	}, &FileMigrationListener{})

	assert.JSONEq(t, `{"delimiter":";"}`, string(state.options))
	assert.Equal(t, 1, state.runs)
	assertClaimReleased(t, status.ID)
	assertMigrationOutcome(t, status.ID, "")
}

func TestFileMigrationListenerUnregisteredKindReleasesClaim(t *testing.T) {
	clearMigrationStatus(t)
	useTestSpoolDir(t)

	u := getTestUser(t)
	uploadName, uploadSize := spoolTestUpload(t, strings.NewReader("some export"))

	status, err := migration.ClaimMigration(&stubFileMigrator{name: "unregistered-file-stub"}, u)
	require.NoError(t, err)

	events.TestListener(t, &FileMigrationRequestedEvent{
		User:              u,
		MigratorKind:      "unregistered-file-stub",
		MigrationStatusID: status.ID,
		UploadName:        uploadName,
		UploadSize:        uploadSize,
	}, &FileMigrationListener{})

	assertClaimReleased(t, status.ID)
	assertMigrationOutcome(t, status.ID, migration.GenericFailureMessage)
	assertSpoolRemoved(t, uploadName)
}

func TestFileMigrationListenerMalformedPayloadIsNotRetried(t *testing.T) {
	clearMigrationStatus(t)
	useTestSpoolDir(t)

	state := registerStubFileMigrator("malformed-file-stub", nil)

	err := (&FileMigrationListener{}).Handle(message.NewMessage(watermill.NewUUID(), []byte("{not json")))

	require.NoError(t, err)
	assert.Equal(t, 0, state.runs)
}

func TestFileMigrationListenerStaleEventDoesNothing(t *testing.T) {
	clearMigrationStatus(t)
	useTestSpoolDir(t)

	state := registerStubFileMigrator("stale-file-stub", nil)
	u := getTestUser(t)
	uploadName, uploadSize := spoolTestUpload(t, strings.NewReader("some export"))

	status, err := migration.ClaimMigration(&stubFileMigrator{name: "stale-file-stub"}, u)
	require.NoError(t, err)
	require.NoError(t, migration.FinishMigration(status))
	finished, err := migration.GetMigrationStatusByID(status.ID)
	require.NoError(t, err)

	events.TestListener(t, &FileMigrationRequestedEvent{
		User:              u,
		MigratorKind:      "stale-file-stub",
		MigrationStatusID: status.ID,
		UploadName:        uploadName,
		UploadSize:        uploadSize,
	}, &FileMigrationListener{})

	assert.Equal(t, 0, state.runs)
	after, err := migration.GetMigrationStatusByID(status.ID)
	require.NoError(t, err)
	assert.True(t, after.FinishedAt.Equal(finished.FinishedAt))
	assertSpoolRemoved(t, uploadName)
}

func TestFileMigrationListenerForeignStatusIsNotImported(t *testing.T) {
	clearMigrationStatus(t)
	useTestSpoolDir(t)
	notifications.Fake()
	t.Cleanup(notifications.Unfake)

	state := registerStubFileMigrator("foreign-file-stub", nil)
	owner := getTestUser(t)
	other := getTestUserByID(t, 2)
	uploadName, uploadSize := spoolTestUpload(t, strings.NewReader("some export"))

	status, err := migration.ClaimMigration(&stubFileMigrator{name: "foreign-file-stub"}, owner)
	require.NoError(t, err)

	events.TestListener(t, &FileMigrationRequestedEvent{
		User:              other,
		MigratorKind:      "foreign-file-stub",
		MigrationStatusID: status.ID,
		UploadName:        uploadName,
		UploadSize:        uploadSize,
	}, &FileMigrationListener{})

	assert.Equal(t, 0, state.runs)

	after, err := migration.GetMigrationStatusByID(status.ID)
	require.NoError(t, err)
	assert.True(t, after.FinishedAt.IsZero())
	require.NotNil(t, after.ActiveUserID)
	assert.Equal(t, owner.ID, *after.ActiveUserID)
	assertSpoolRemoved(t, uploadName)
}

func TestFileMigrationListenerReportedFailureStoresGenericMessage(t *testing.T) {
	clearMigrationStatus(t)
	useTestSpoolDir(t)
	enableSentry(t)
	notifications.Fake()
	t.Cleanup(notifications.Unfake)

	// The failure a prior review found in a user's inbox: the spool path must not reach them.
	leaky := &os.PathError{Op: "open", Path: "/var/lib/vikunja/files/migration-spool-4711", Err: os.ErrNotExist}
	registerStubFileMigrator("reported-file-stub", leaky)
	u := getTestUser(t)
	uploadName, uploadSize := spoolTestUpload(t, strings.NewReader("some export"))

	status, err := migration.ClaimMigration(&stubFileMigrator{name: "reported-file-stub"}, u)
	require.NoError(t, err)

	events.TestListener(t, &FileMigrationRequestedEvent{
		User:              u,
		MigratorKind:      "reported-file-stub",
		MigrationStatusID: status.ID,
		UploadName:        uploadName,
		UploadSize:        uploadSize,
	}, &FileMigrationListener{})

	notifications.AssertSent(t, &MigrationFailedReportedNotification{})
	assertClaimReleased(t, status.ID)
	assertMigrationOutcome(t, status.ID, migration.GenericFailureMessage)

	stored, err := migration.GetMigrationStatusByID(status.ID)
	require.NoError(t, err)
	assert.NotContains(t, stored.ErrorMessage, "/var/lib/vikunja")
	assertSpoolRemoved(t, uploadName)
}

func TestFileMigrationListenerUserFacingStatusDistinguishesOutcomes(t *testing.T) {
	clearMigrationStatus(t)
	useTestSpoolDir(t)
	notifications.Fake()
	t.Cleanup(notifications.Unfake)

	registerStubFileMigrator("outcome-ok-stub", nil)
	registerStubFileMigrator("outcome-failed-stub", errors.New("the import file is missing a VERSION entry"))
	u := getTestUser(t)

	runImport := func(t *testing.T, kind string) {
		t.Helper()
		uploadName, uploadSize := spoolTestUpload(t, strings.NewReader("some export"))
		status, err := migration.ClaimMigration(&stubFileMigrator{name: kind}, u)
		require.NoError(t, err)
		events.TestListener(t, &FileMigrationRequestedEvent{
			User:              u,
			MigratorKind:      kind,
			MigrationStatusID: status.ID,
			UploadName:        uploadName,
			UploadSize:        uploadSize,
		}, &FileMigrationListener{})
	}

	runImport(t, "outcome-ok-stub")
	runImport(t, "outcome-failed-stub")

	// What the status endpoint hands the client.
	succeeded, err := migration.GetMigrationStatus(&stubFileMigrator{name: "outcome-ok-stub"}, u)
	require.NoError(t, err)
	assert.False(t, succeeded.FinishedAt.IsZero())
	assert.Empty(t, succeeded.ErrorMessage)

	failed, err := migration.GetMigrationStatus(&stubFileMigrator{name: "outcome-failed-stub"}, u)
	require.NoError(t, err)
	assert.False(t, failed.FinishedAt.IsZero())
	assert.Equal(t, "the import file is missing a VERSION entry", failed.ErrorMessage)

	body, err := json.Marshal(failed)
	require.NoError(t, err)
	var wire map[string]interface{}
	require.NoError(t, json.Unmarshal(body, &wire))
	assert.Equal(t, "the import file is missing a VERSION entry", wire["error_message"])
}
