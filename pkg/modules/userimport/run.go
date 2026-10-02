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

// Package userimport refreshes the users table from a CSV file on a schedule.
//
// People in the file are created if missing, and their name, job title and department are
// updated by matching the file's mail column against the user's email. Every other user that
// signs in through Entra is deactivated. Admins, bots and accounts of other login methods are
// never touched, and nothing is ever deleted.
//
// The job is meant to run on a single instance, like the other cron jobs and the in-process event
// bus. The lock that keeps two runs apart is per process.
package userimport

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/user"

	"xorm.io/builder"
	"xorm.io/xorm"
)

// batchSize is the number of users changed per transaction. One transaction for a whole first run
// would hold the write lock for minutes, which on SQLite blocks every other writer.
const batchSize = 100

// ErrAlreadyRunning is returned when a run starts while another one is still going.
var ErrAlreadyRunning = errors.New("a user import is already running")

// ErrStaleFile is returned when the file was not modified recently enough. A stale file would
// deactivate everybody who joined since it was written.
type ErrStaleFile struct {
	Age    time.Duration
	MaxAge time.Duration
}

func (e *ErrStaleFile) Error() string {
	return fmt.Sprintf("the user list was last modified %s ago, more than the allowed %s", e.Age.Round(time.Minute), e.MaxAge)
}

// ErrFileChanging is returned when the file is, or looks like it is still being written. The
// exporter should write to a temporary name and rename it into place.
var ErrFileChanging = errors.New("the user list is still being written")

// fileSettleTime is how long the file must have been left alone before a scheduled run reads it.
const fileSettleTime = time.Minute

var runMu sync.Mutex

// Options configure a single run.
type Options struct {
	File string
	// MaxDisablePercent is the most active users, in percent, one run may deactivate.
	MaxDisablePercent int
	// MaxFileAge skips the run when the file is older. Zero disables the check.
	MaxFileAge time.Duration
	// SettleTime skips the run when the file was modified more recently. Zero disables the check.
	SettleTime time.Duration
	// DryRun only computes the plan and changes nothing.
	DryRun bool
	// Detail also collects who would be created, updated and deactivated (capped, see maxDetailEntries).
	Detail bool
}

// Result summarises a run. It carries counts only, never names or emails. For a dry run and for a
// run that was refused, the counts are what the run planned. After a real run, Created and Failed
// are what actually happened.
type Result struct {
	Created          int
	Updated          int
	Reenabled        int
	Disabled         int
	Unchanged        int
	Exempt           int
	SkippedInvalid   int
	SkippedDuplicate int
	// Failed is the number of people that could not be created. They are skipped so one bad row
	// cannot block everybody else, and are tried again on the next run.
	Failed int
	DryRun bool
	// Details is only set when Options.Detail was requested. It carries personal data, so it is
	// meant for admins and is never logged.
	Details *Details
}

func optionsFromConfig() Options {
	return Options{
		File:              config.UserImportFile.GetString(),
		MaxDisablePercent: config.UserImportMaxDisablePercent.GetInt(),
		MaxFileAge:        time.Duration(config.UserImportMaxFileAgeHours.GetInt()) * time.Hour,
		SettleTime:        fileSettleTime,
		DryRun:            config.UserImportDryRun.GetBool(),
	}
}

// Run executes one import with the configured options.
func Run(ctx context.Context) (*Result, error) {
	return RunWithOptions(ctx, optionsFromConfig())
}

// RunWithOptions executes one import. Everything is decided in memory first and the safety limit
// is checked on that plan, before anything is written. The writes then happen in bounded
// transactions: deactivations first, because they are what keeps former staff out, then updates,
// then one transaction per new person. A failure in a batch of deactivations or updates stops the
// run; the plan is rebuilt from the database on the next run, and every step is idempotent.
func RunWithOptions(ctx context.Context, opts Options) (*Result, error) {
	if !runMu.TryLock() {
		return nil, ErrAlreadyRunning
	}
	defer runMu.Unlock()

	data, err := readList(opts)
	if err != nil {
		return nil, err
	}

	parsed, err := ParseCSV(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	if len(parsed.People) == 0 {
		return nil, errors.New("the user list contains no valid rows, nothing was changed")
	}

	users, err := loadUsers()
	if err != nil {
		return nil, err
	}

	plan := buildPlan(parsed.People, users)
	result := &Result{
		Created:          len(plan.Create),
		Updated:          len(plan.Update),
		Reenabled:        plan.Reenabled,
		Disabled:         len(plan.Disable),
		Unchanged:        plan.Unchanged,
		Exempt:           plan.Exempt,
		SkippedInvalid:   parsed.SkippedInvalid,
		SkippedDuplicate: parsed.SkippedDuplicate,
		DryRun:           opts.DryRun,
	}
	if opts.Detail {
		result.Details = plan.details()
	}

	err = plan.checkLimit(opts.MaxDisablePercent)
	if err != nil {
		return result, fmt.Errorf("aborted, nothing was changed: %w", err)
	}

	if opts.DryRun {
		return result, nil
	}

	return result, apply(ctx, plan, result)
}

// readList reads the whole file through one open handle. A list that is stale, that was modified
// moments ago, or that changed while it was being read is refused: a half-written file parses as a
// valid, shorter list, and everybody missing from it would be deactivated.
func readList(opts Options) ([]byte, error) {
	f, err := os.Open(opts.File)
	if err != nil {
		return nil, fmt.Errorf("could not open the user list: %w", err)
	}
	defer f.Close()

	before, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("could not read the user list: %w", err)
	}
	if before.IsDir() {
		return nil, fmt.Errorf("the user list path %q is a directory", opts.File)
	}

	age := time.Since(before.ModTime())
	if opts.MaxFileAge > 0 && age > opts.MaxFileAge {
		return nil, &ErrStaleFile{Age: age, MaxAge: opts.MaxFileAge}
	}
	if opts.SettleTime > 0 && age < opts.SettleTime {
		return nil, fmt.Errorf("%w: it was modified %s ago and must be untouched for %s", ErrFileChanging, age.Round(time.Second), opts.SettleTime)
	}

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("could not read the user list: %w", err)
	}

	after, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("could not read the user list: %w", err)
	}
	if int64(len(data)) != before.Size() || after.Size() != before.Size() || !after.ModTime().Equal(before.ModTime()) {
		return nil, fmt.Errorf("%w: it changed while it was being read", ErrFileChanging)
	}

	return data, nil
}

// loadUsers returns every user except bots. Bots are owned by another user and never part of the list.
func loadUsers() ([]*user.User, error) {
	s := db.NewReadSession()
	defer s.Close()

	users := []*user.User{}
	err := s.
		Where(builder.Or(builder.IsNull{"bot_owner_id"}, builder.Eq{"bot_owner_id": 0})).
		OrderBy("id").
		Find(&users)
	return users, err
}

// inTransaction runs fn in its own transaction. The events fn queued are dispatched after the
// commit, and discarded when fn fails.
func inTransaction(ctx context.Context, fn func(s *xorm.Session) error) error {
	s := db.NewSession()
	defer s.Close()
	// Discards queued events on failure; a no-op once DispatchPending has run.
	defer events.CleanupPending(s)

	err := fn(s)
	if err != nil {
		_ = s.Rollback()
		return err
	}
	err = s.Commit()
	if err != nil {
		return fmt.Errorf("could not commit the user import: %w", err)
	}
	events.DispatchPending(ctx, s)
	return nil
}

// chunk splits items into slices of at most size elements.
func chunk[T any](items []T, size int) [][]T {
	var chunks [][]T
	for len(items) > 0 {
		n := min(size, len(items))
		chunks = append(chunks, items[:n:n])
		items = items[n:]
	}
	return chunks
}

func apply(ctx context.Context, plan *Plan, result *Result) error {
	for _, batch := range chunk(plan.Disable, batchSize) {
		err := inTransaction(ctx, func(s *xorm.Session) error { return applyDisables(s, batch) })
		if err != nil {
			return fmt.Errorf("could not deactivate users: %w", err)
		}
	}

	for _, batch := range chunk(plan.Update, batchSize) {
		err := inTransaction(ctx, func(s *xorm.Session) error { return applyUpdates(s, batch) })
		if err != nil {
			return fmt.Errorf("could not update users: %w", err)
		}
	}

	// One transaction per person: creating a user is several statements (user, Inbox project, views)
	// that must stay together, and a person that cannot be created must not block the others.
	created, failed := 0, 0
	for _, p := range plan.Create {
		var made bool
		err := inTransaction(ctx, func(s *xorm.Session) (err error) {
			made, err = createImportedUser(s, p)
			return err
		})
		if err != nil {
			failed++
			log.Errorf("[User Import] Could not create the user for id %s: %s", p.ID, err)
			continue
		}
		if made {
			created++
		}
	}
	result.Created = created
	result.Failed = failed

	return nil
}

// applyDisables deactivates the users and ends their sessions, and queues the same status-changed
// event the admin path queues, so the change shows up in the audit log. There is no acting user,
// the audit log records it as the system.
func applyDisables(s *xorm.Session, users []*user.User) error {
	for _, u := range users {
		// Admins are exempt so this never triggers; it stays as a backstop.
		err := user.GuardLastAdmin(s, u)
		if err != nil {
			return err
		}

		oldStatus := u.Status
		err = user.SetUserStatus(s, u, user.StatusDisabled)
		if err != nil {
			return fmt.Errorf("could not deactivate user %d: %w", u.ID, err)
		}
		u.Status = user.StatusDisabled

		err = models.DeleteAllUserSessions(s, u.ID)
		if err != nil {
			return fmt.Errorf("could not end the sessions of user %d: %w", u.ID, err)
		}

		events.DispatchOnCommit(s, &models.AdminUserStatusChangedEvent{
			User:      u,
			OldStatus: oldStatus,
			NewStatus: user.StatusDisabled,
		})
	}
	return nil
}

// applyUpdates writes the changed columns, and queues the status-changed event for users that are
// re-enabled.
func applyUpdates(s *xorm.Session, changes []userChange) error {
	for _, change := range changes {
		_, err := s.ID(change.User.ID).Cols(change.Cols...).Update(change.User)
		if err != nil {
			return fmt.Errorf("could not update user %d: %w", change.User.ID, err)
		}

		if change.Reenabled {
			events.DispatchOnCommit(s, &models.AdminUserStatusChangedEvent{
				User:      change.User,
				OldStatus: user.StatusDisabled,
				NewStatus: user.StatusActive,
			})
		}
	}
	return nil
}

// importUsername picks the username of a new user: the local part of the userPrincipalName, else
// of the mail. An empty result lets CreateUserWithRandomUsername generate one, which also handles
// names that are already taken.
func importUsername(p *Person) string {
	for _, login := range []string{p.UserPrincipalName, p.Mail} {
		name := truncateRunes(user.UsernameFromLogin(login), maxFieldLength)
		if name != "" && !user.IsReservedUsername(name) {
			return name
		}
	}
	return ""
}

// createImportedUser creates the user through the same path as a first Entra login, so it gets the
// Inbox project too. It has no password and the placeholder issuer, which the first real login
// replaces (see the openid package). It returns false when nothing had to be created.
func createImportedUser(s *xorm.Session, p *Person) (bool, error) {
	// The plan was built from a snapshot. A login may have given this person an import id since,
	// and creating them again would leave two accounts for one person.
	taken, err := s.Where("import_id = ?", p.ID).Exist(&user.User{})
	if err != nil {
		return false, err
	}
	if taken {
		return false, nil
	}

	u, err := auth.CreateUserWithRandomUsername(s, &user.User{
		Username:   importUsername(p),
		Email:      p.Mail,
		Name:       p.DisplayName,
		Issuer:     user.IssuerImport,
		Subject:    p.ID,
		ImportID:   p.ID,
		JobTitle:   p.JobTitle,
		Department: p.Department,
	})
	if err != nil {
		return false, fmt.Errorf("could not create the user: %w", err)
	}

	// CreateUser overwrites these with the configured defaults. Imported people must be
	// findable, otherwise they still cannot be shared with or assigned.
	_, err = s.
		ID(u.ID).
		Cols("discoverable_by_name", "discoverable_by_email").
		Update(&user.User{DiscoverableByName: true, DiscoverableByEmail: true})
	if err != nil {
		return false, fmt.Errorf("could not make user %d discoverable: %w", u.ID, err)
	}
	return true, nil
}
