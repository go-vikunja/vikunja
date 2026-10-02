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

package userimport

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"
)

// maxDetailEntries caps every list of Details, so a preview of a huge change stays small.
const maxDetailEntries = 200

// MaxUploadBytes is the size limit of an uploaded user list.
const MaxUploadBytes = 10 * 1024 * 1024

// Entry is one person in a Details list.
type Entry struct {
	Username string
	Name     string
	Email    string
}

// Details lists who a run creates, updates and deactivates. Every list is capped; the totals are
// in the Result.
type Details struct {
	Create    []Entry
	Update    []Entry
	Disable   []Entry
	Truncated bool
}

func (p *Plan) details() *Details {
	d := &Details{}
	add := func(list *[]Entry, e Entry) {
		if len(*list) >= maxDetailEntries {
			d.Truncated = true
			return
		}
		*list = append(*list, e)
	}

	for _, person := range p.Create {
		add(&d.Create, Entry{Username: importUsername(person), Name: person.DisplayName, Email: person.Mail})
	}
	for _, change := range p.Update {
		add(&d.Update, Entry{Username: change.User.Username, Name: change.User.Name, Email: change.User.Email})
	}
	for _, u := range p.Disable {
		add(&d.Disable, Entry{Username: u.Username, Name: u.Name, Email: u.Email})
	}
	return d
}

// Trigger says what started a run.
type Trigger string

// The ways a run can be started.
const (
	TriggerSchedule Trigger = "schedule"
	TriggerManual   Trigger = "manual"
)

// LastRun is the record of the most recent finished run. It lives in memory only and is lost on
// restart.
type LastRun struct {
	Trigger    Trigger
	Actor      string
	FinishedAt time.Time
	Result     *Result
	Error      string
}

// Status describes the import for the admin page.
type Status struct {
	Enabled      bool
	Schedule     string
	File         string
	DryRun       bool
	TenantSet    bool
	FileExists   bool
	FileSize     int64
	FileModified time.Time
	Running      bool
	LastRun      *LastRun
}

var (
	lastRunMu sync.Mutex
	lastRun   *LastRun
	running   bool
)

func recordRun(trigger Trigger, actor string, result *Result, err error) {
	rec := &LastRun{Trigger: trigger, Actor: actor, FinishedAt: time.Now(), Result: result}
	if err != nil {
		rec.Error = err.Error()
	}
	if rec.Result != nil {
		// The record is kept for the status page; the person lists belong to the preview only.
		copied := *rec.Result
		copied.Details = nil
		rec.Result = &copied
	}

	lastRunMu.Lock()
	lastRun = rec
	lastRunMu.Unlock()
}

func setRunning(v bool) {
	lastRunMu.Lock()
	running = v
	lastRunMu.Unlock()
}

// CurrentStatus returns the configuration, the state of the file and the last run.
func CurrentStatus() Status {
	st := Status{
		Enabled:   config.UserImportEnabled.GetBool(),
		Schedule:  config.UserImportSchedule.GetString(),
		File:      config.UserImportFile.GetString(),
		DryRun:    config.UserImportDryRun.GetBool(),
		TenantSet: config.UserImportTenantID.GetString() != "",
	}

	if info, err := os.Stat(st.File); err == nil && !info.IsDir() {
		st.FileExists = true
		st.FileSize = info.Size()
		st.FileModified = info.ModTime()
	}

	lastRunMu.Lock()
	st.Running = running
	if lastRun != nil {
		copied := *lastRun
		st.LastRun = &copied
	}
	lastRunMu.Unlock()
	return st
}

// manualOptions are the options of a run an admin starts or previews. The admin chose the file, so
// the freshness checks do not apply; the percentage safety limit does.
func manualOptions() Options {
	opts := optionsFromConfig()
	opts.MaxFileAge = 0
	opts.SettleTime = 0
	// userimport.dryrun only governs the scheduled job. An admin who presses "run" wants changes,
	// and previews first (Preview forces a dry run itself).
	opts.DryRun = false
	return opts
}

// Preview computes what a run would do, regardless of the dry run setting, and writes nothing.
func Preview(ctx context.Context) (*Result, error) {
	opts := manualOptions()
	opts.DryRun = true
	opts.Detail = true
	return RunWithOptions(ctx, opts)
}

// StartAsync starts a real run in the background with the manual options. It returns
// ErrAlreadyRunning when a run is going. The outcome is available through CurrentStatus.
func StartAsync(actor string) error {
	// The lock is taken inside RunWithOptions; check early so the caller gets a 409 instead of a
	// silently dropped request.
	if !runMu.TryLock() {
		return ErrAlreadyRunning
	}
	runMu.Unlock()

	setRunning(true)
	go func() {
		defer setRunning(false)
		defer func() {
			if r := recover(); r != nil {
				log.Errorf("[User Import] The manual import panicked: %v", r)
				recordRun(TriggerManual, actor, nil, fmt.Errorf("the import panicked: %v", r))
			}
		}()

		// Detached from the request: the run must outlive it.
		result, err := RunWithOptions(context.Background(), manualOptions())
		recordRun(TriggerManual, actor, result, err)
		if err != nil {
			log.Errorf("[User Import] The manual import failed: %s", err)
		}
	}()
	return nil
}

// ErrInvalidUpload is returned when an uploaded list cannot be used.
var ErrInvalidUpload = errors.New("invalid user list")

// ReplaceFile validates an uploaded user list and puts it in place of the configured file. The
// list must parse and contain at least one valid row. It is written to a temporary file in the
// same directory and renamed over the target, so a reader never sees a half-written file, and the
// previous file is kept as <file>.previous.
func ReplaceFile(data []byte) (*ParseResult, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("%w: the file is empty", ErrInvalidUpload)
	}
	parsed, err := ParseCSV(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidUpload, err)
	}
	if len(parsed.People) == 0 {
		return nil, fmt.Errorf("%w: it contains no valid rows", ErrInvalidUpload)
	}

	target := config.UserImportFile.GetString()
	dir := filepath.Dir(target)
	err = os.MkdirAll(dir, 0o750)
	if err != nil {
		return nil, fmt.Errorf("could not create the directory of the user list: %w", err)
	}

	tmp, err := os.CreateTemp(dir, ".user_list-*.tmp")
	if err != nil {
		return nil, fmt.Errorf("could not write the user list, is the directory writable? %w", err)
	}
	tmpName := tmp.Name()
	cleanup := func() { _ = os.Remove(tmpName) }

	_, err = tmp.Write(data)
	if err != nil {
		_ = tmp.Close()
		cleanup()
		return nil, fmt.Errorf("could not write the user list: %w", err)
	}
	err = tmp.Close()
	if err != nil {
		cleanup()
		return nil, fmt.Errorf("could not write the user list: %w", err)
	}

	// Best effort: keeping a copy of the previous list must not block the upload.
	existing, readErr := os.ReadFile(target)
	if readErr == nil {
		writeErr := os.WriteFile(target+".previous", existing, 0o600)
		if writeErr != nil {
			log.Warningf("[User Import] Could not keep the previous user list: %s", writeErr)
		}
	}

	err = os.Rename(tmpName, target)
	if err != nil {
		cleanup()
		return nil, fmt.Errorf("could not replace the user list: %w", err)
	}
	return parsed, nil
}
