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

package vikunjafile

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
)

// The repeat modes of export files written before tasks stored an RRULE.
const (
	legacyRepeatModeDefault         = 0
	legacyRepeatModeMonth           = 1
	legacyRepeatModeFromCurrentDate = 2
)

// convertLegacyRepeat converts the repeat_after/repeat_mode pair of a pre-RRULE export to an
// RFC 5545 rule. It is the importer's own converter: migration 20261022220000 keeps a frozen
// copy, and both are tested against pkg/migration/testdata/legacy_repeat_cases.json. A monthly
// task due on the 29th to 31st keeps landing on the last day of shorter months, as the legacy
// month arithmetic did; the due day is read in loc, the importing user's zone or else the server
// zone.
func convertLegacyRepeat(repeatAfter int64, repeatMode int, dueDate time.Time, loc *time.Location) (rule string, fromCurrentDate bool) {
	switch repeatMode {
	case legacyRepeatModeMonth:
		if !dueDate.IsZero() {
			if day := dueDate.In(loc).Day(); day >= 29 {
				days := make([]string, 0, day-27)
				for d := 28; d <= day; d++ {
					days = append(days, strconv.Itoa(d))
				}
				return "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=" + strings.Join(days, ",") + ";BYSETPOS=-1", false
			}
		}
		return "FREQ=MONTHLY;INTERVAL=1", false
	case legacyRepeatModeDefault, legacyRepeatModeFromCurrentDate:
		if repeatAfter <= 0 {
			return "", false
		}
		return legacySecondsToRRule(repeatAfter), repeatMode == legacyRepeatModeFromCurrentDate
	}
	return "", false
}

// legacySecondsToRRule picks the largest unit that divides the interval evenly.
func legacySecondsToRRule(seconds int64) string {
	const (
		minute = int64(time.Minute / time.Second)
		hour   = int64(time.Hour / time.Second)
		day    = int64((24 * time.Hour) / time.Second)
		week   = int64((7 * 24 * time.Hour) / time.Second)
	)

	switch {
	case seconds%week == 0:
		return fmt.Sprintf("FREQ=WEEKLY;INTERVAL=%d", seconds/week)
	case seconds%day == 0:
		return fmt.Sprintf("FREQ=DAILY;INTERVAL=%d", seconds/day)
	case seconds%hour == 0:
		return fmt.Sprintf("FREQ=HOURLY;INTERVAL=%d", seconds/hour)
	case seconds%minute == 0:
		return fmt.Sprintf("FREQ=MINUTELY;INTERVAL=%d", seconds/minute)
	default:
		return fmt.Sprintf("FREQ=SECONDLY;INTERVAL=%d", seconds)
	}
}

// legacyProject and legacyTask read only the keys a pre-RRULE export used for recurrence.
type legacyProject struct {
	Tasks         []*legacyTask    `json:"tasks"`
	ChildProjects []*legacyProject `json:"child_projects"`
}

type legacyTask struct {
	RepeatAfter int64     `json:"repeat_after"`
	RepeatMode  int       `json:"repeat_mode"`
	DueDate     time.Time `json:"due_date"`
}

// applyImportedRecurrence sets each imported task's rule from the export:
//   - repeats_rrule, the exact rule text, wins and the structured repeat is dropped, so the
//     create path stores the text verbatim;
//   - otherwise a structured repeat is kept, and the create path serialises it;
//   - otherwise the legacy repeat_after/repeat_mode keys of a pre-RRULE export are converted.
//
// Whichever source it came from, a rule the create path would refuse is then dropped with a log
// line and the task imported without it (models.DropUnimportableRepeat).
//
// data is the raw data.json the projects were decoded from. It is decoded a second time into the
// legacy shape. Both decodes of the same bytes have the same shape, so the two trees are walked in
// step and each task is matched to its own entry in the file, never by an id that could repeat
// across projects.
//
// loc is the zone a legacy monthly task's due day is read in: the importing user's, who owns the
// projects being created, else the server's.
func applyImportedRecurrence(projects []*models.ProjectWithTasksAndBuckets, data []byte, loc *time.Location) error {
	var legacy []*legacyProject
	if err := json.Unmarshal(data, &legacy); err != nil {
		return err
	}
	applyImportedRecurrenceToProjects(projects, legacy, loc)
	return nil
}

func applyImportedRecurrenceToProjects(projects []*models.ProjectWithTasksAndBuckets, legacy []*legacyProject, loc *time.Location) {
	for i, p := range projects {
		if p == nil {
			continue
		}
		var lp *legacyProject
		if i < len(legacy) {
			lp = legacy[i]
		}

		for j, t := range p.Tasks {
			if t == nil {
				continue
			}
			var lt *legacyTask
			if lp != nil && j < len(lp.Tasks) {
				lt = lp.Tasks[j]
			}
			applyImportedRecurrenceToTask(t, lt, loc)
			// A rule the create path would refuse is dropped, not fatal: the rest of the
			// import goes ahead. Ids are the export file's, never titles.
			if err := models.DropUnimportableRepeat(&t.Task); err != nil {
				log.Warningf(logPrefix+"Importing task %d of project %d without its recurrence rule: %v", t.ID, p.ID, err)
			}
		}

		var children []*legacyProject
		if lp != nil {
			children = lp.ChildProjects
		}
		applyImportedRecurrenceToProjects(p.ChildProjects, children, loc)
	}
}

func applyImportedRecurrenceToTask(t *models.TaskWithComments, legacy *legacyTask, loc *time.Location) {
	switch {
	case t.RepeatsRRule != "":
		t.Repeats = t.RepeatsRRule
		t.Repeat = nil
	case t.Repeat != nil:
		// The create path serialises the structured repeat.
	case legacy != nil:
		if rule, fromCurrentDate := convertLegacyRepeat(legacy.RepeatAfter, legacy.RepeatMode, legacy.DueDate, loc); rule != "" {
			t.Repeats = rule
			t.RepeatsFromCurrentDate = fromCurrentDate
		}
	}
}
