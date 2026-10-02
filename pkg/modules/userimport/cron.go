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
	"context"
	"errors"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/cron"
	"code.vikunja.io/api/pkg/log"
)

const logPrefix = "[User Import Cron] "

// RegisterCron schedules the user import when userimport.enabled is set. The schedule is a cron
// expression in the server's local time, by default every day at 07:30.
func RegisterCron() {
	if !config.UserImportEnabled.GetBool() {
		return
	}

	schedule := config.UserImportSchedule.GetString()
	if err := cron.Schedule(schedule, runScheduled); err != nil {
		log.Fatalf(logPrefix+"Could not register the user import with schedule %q: %s", schedule, err)
	}

	log.Infof(logPrefix+"Scheduled the user import with %q (dry run: %t)", schedule, config.UserImportDryRun.GetBool())
}

func runScheduled() {
	// The cron library does not recover panics, and one must not take the server down.
	defer func() {
		if r := recover(); r != nil {
			log.Errorf(logPrefix+"The user import panicked: %v", r)
		}
	}()

	result, err := Run(context.Background())

	var stale *ErrStaleFile
	switch {
	case errors.Is(err, ErrAlreadyRunning):
		// Nothing ran, the run that is going records its own result.
		log.Warningf(logPrefix+"Skipped: %s", err)
		return
	case errors.As(err, &stale), errors.Is(err, ErrFileChanging):
		recordRun(TriggerSchedule, "", result, err)
		log.Warningf(logPrefix+"Skipped: %s", err)
		return
	case err != nil:
		recordRun(TriggerSchedule, "", result, err)
		log.Errorf(logPrefix+"The user import failed: %s", err)
		return
	}
	recordRun(TriggerSchedule, "", result, nil)

	log.Infof(logPrefix+"Done (dry run: %t): created %d, failed to create %d, updated %d, re-enabled %d, deactivated %d, unchanged %d, exempt %d, skipped invalid %d, skipped duplicate %d",
		result.DryRun, result.Created, result.Failed, result.Updated, result.Reenabled, result.Disabled,
		result.Unchanged, result.Exempt, result.SkippedInvalid, result.SkippedDuplicate)
}
