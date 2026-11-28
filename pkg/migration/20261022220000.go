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

package migration

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

// legacyTaskOwnerZone20261022220000 is one row of a conversion page: the legacy repeat
// columns of a task and the stored zone of its project's owner. It is read by a joined
// query and never synced. Like the plain read it replaces, it has no deleted tag, so
// soft-deleted rows are converted too.
type legacyTaskOwnerZone20261022220000 struct {
	ID          int64     `xorm:"'id'"`
	RepeatAfter int64     `xorm:"'repeat_after'"`
	RepeatMode  int       `xorm:"'repeat_mode'"`
	Repeats     string    `xorm:"'repeats'"`
	DueDate     time.Time `xorm:"DATETIME null 'due_date'"`
	ProjectID   int64     `xorm:"'owner_project_id'"` // 0 when the task's project row is missing
	OwnerID     int64     `xorm:"'owner_id'"`         // 0 when the project's owner row is missing
	Timezone    string    `xorm:"'timezone'"`         // owner's users.timezone; "" when NULL or unjoined
}

// taskRRule20261022220000 declares only the columns this migration adds, with the
// tags of pkg/models/tasks.go, so partialSync adds them and drops nothing.
type taskRRule20261022220000 struct {
	ID                     int64  `xorm:"bigint autoincr not null unique pk"`
	Repeats                string `xorm:"varchar(500) null"`
	RepeatsFromCurrentDate bool   `xorm:"null default false"`
}

func (taskRRule20261022220000) TableName() string {
	return "tasks"
}

// taskRepeatBackup20261022220000 preserves the legacy repeat columns before they
// are dropped, so a mis-conversion can be recovered by hand. External post-upgrade
// verification reads it by name.
type taskRepeatBackup20261022220000 struct {
	ID          int64 `xorm:"bigint not null pk"`
	RepeatAfter int64 `xorm:"bigint null"`
	RepeatMode  int   `xorm:"not null default 0"`
}

func (taskRepeatBackup20261022220000) TableName() string {
	return "task_repeat_legacy_backup"
}

const (
	repeatModeDefault20261022220000         = 0
	repeatModeMonth20261022220000           = 1
	repeatModeFromCurrentDate20261022220000 = 2

	rruleBatchSize20261022220000 = 500
	rruleTmpTable20261022220000  = "tasks_rrule_tmp"
)

// testFailAfter20261022220000 names a point at which the migration returns an error,
// so tests can interrupt it at each destructive boundary. Production never sets it;
// the zero value disables every point.
var testFailAfter20261022220000 string

func failpoint20261022220000(point string) error {
	if testFailAfter20261022220000 != "" && testFailAfter20261022220000 == point {
		return fmt.Errorf("RRULE migration: injected failure at %s", point)
	}
	return nil
}

// Frozen copy of the legacy repeat conversion — migrations must not follow later model changes.
// convertLegacyRepeat20261022220000 converts legacy repeat_after/repeat_mode to an
// RFC 5545 rule. A monthly task due on the 29th to 31st keeps landing on the last
// day of shorter months, as the legacy month arithmetic did. The due day is read in
// loc: the zone of the task's project owner, else the server zone.
func convertLegacyRepeat20261022220000(repeatAfter int64, repeatMode int, dueDate time.Time, loc *time.Location) (rule string, fromCurrentDate bool) {
	switch repeatMode {
	case repeatModeMonth20261022220000:
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
	case repeatModeDefault20261022220000, repeatModeFromCurrentDate20261022220000:
		if repeatAfter <= 0 {
			return "", false
		}
		return secondsToRRule20261022220000(repeatAfter), repeatMode == repeatModeFromCurrentDate20261022220000
	}
	return "", false
}

// secondsToRRule20261022220000 picks the largest unit that divides the interval evenly.
func secondsToRRule20261022220000(seconds int64) string {
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

// ownerZones20261022220000 resolves the zone a converted row's due day is read in, once
// per distinct stored name. It treats names as pkg/models' task-zone resolver does: an
// empty name and "Local" (the process zone) are unset, and a name time.LoadLocation
// rejects is invalid; either falls back to the server zone.
type ownerZones20261022220000 struct {
	loaded map[string]*time.Location // nil for a name that is not usable
}

// zone returns the location for the row and, when it fell back to the server zone, why.
func (z *ownerZones20261022220000) zone(row *legacyTaskOwnerZone20261022220000) (loc *time.Location, fallback string) {
	switch {
	case row.ProjectID == 0:
		return config.GetTimeZone(), "no project"
	case row.OwnerID == 0:
		return config.GetTimeZone(), "no owner"
	case row.Timezone == "":
		return config.GetTimeZone(), "empty zone"
	}

	loc, seen := z.loaded[row.Timezone]
	if !seen {
		if row.Timezone != "Local" {
			loc, _ = time.LoadLocation(row.Timezone) // nil on error
		}
		if loc == nil {
			log.Warningf("RRULE migration: owner time zone %q is not a valid IANA name; using the server zone", row.Timezone)
		}
		z.loaded[row.Timezone] = loc
	}
	if loc == nil {
		return config.GetTimeZone(), "invalid zone"
	}
	return loc, ""
}

func migrateRRule20261022220000(tx *xorm.Engine) error {
	hasRepeatAfter, err := columnExists(tx, "tasks", "repeat_after")
	if err != nil {
		return fmt.Errorf("could not check tasks.repeat_after: %w", err)
	}
	hasRepeatMode, err := columnExists(tx, "tasks", "repeat_mode")
	if err != nil {
		return fmt.Errorf("could not check tasks.repeat_mode: %w", err)
	}
	hasRepeats, err := columnExists(tx, "tasks", "repeats")
	if err != nil {
		return fmt.Errorf("could not check tasks.repeats: %w", err)
	}

	var legacy []string
	if hasRepeatAfter {
		legacy = append(legacy, "repeat_after")
	}
	if hasRepeatMode {
		legacy = append(legacy, "repeat_mode")
	}

	switch {
	case hasRepeatAfter && hasRepeatMode:
		log.Info("RRULE migration: legacy repeat columns present, running the full conversion")
		if err := partialSync(tx, taskRRule20261022220000{}); err != nil {
			return fmt.Errorf("could not add the RRULE columns to tasks: %w", err)
		}
		if err := convertLegacyRepeats20261022220000(tx); err != nil {
			return err
		}
		if err := dropLegacyRepeatColumns20261022220000(tx, legacy); err != nil {
			return err
		}
	case !hasRepeats:
		return fmt.Errorf("RRULE migration: unexpected tasks schema (repeat_after: %t, repeat_mode: %t, repeats: %t); expected the legacy repeat columns or the repeats column", hasRepeatAfter, hasRepeatMode, hasRepeats)
	case len(legacy) > 0:
		// An earlier run converted every row and was interrupted between the drops.
		log.Infof("RRULE migration: finishing an interrupted run, dropping the remaining legacy column %s", legacy[0])
		if err := dropLegacyRepeatColumns20261022220000(tx, legacy); err != nil {
			return err
		}
	default:
		log.Info("RRULE migration: already converted, nothing to do")
	}

	// repeats is nullable and was added NULL for every row; filters compare it with ''.
	if _, err := tx.Exec("UPDATE tasks SET repeats = '' WHERE repeats IS NULL"); err != nil {
		return fmt.Errorf("could not normalise NULL repeats: %w", err)
	}
	return nil
}

// convertLegacyRepeats20261022220000 converts every legacy-repeat row in id-ordered
// pages, backing each page up before it is changed. It is safe to re-run: backup
// rows are replaced and rows that already have a rule are skipped.
func convertLegacyRepeats20261022220000(tx *xorm.Engine) error {
	if err := tx.Sync2(taskRepeatBackup20261022220000{}); err != nil { //nolint:forbidigo // brand-new table, nothing to drop
		return fmt.Errorf("could not create task_repeat_legacy_backup: %w", err)
	}

	// The due day is read in the zone of the project owner. The join is safe on
	// every install that runs this migration: users.timezone exists since 20220112211537
	// (which also back-filled every then-existing user with the service zone) and
	// projects.owner_id since 20221113170740 (lists renamed to projects, owner_id kept),
	// both far below this one; a fresh install never runs it (initSchema creates the
	// current schema directly). LEFT JOINs, so an orphaned task or a missing owner still
	// converts, in the server zone. due_date is decoded through the struct mapping, as the
	// plain read did, so the engine's time zone handling of stored instants is unchanged.
	query := "SELECT t.id, t.repeat_after, t.repeat_mode, t.repeats, t.due_date, " +
		"COALESCE(p.id, 0) AS owner_project_id, COALESCE(u.id, 0) AS owner_id, COALESCE(u.timezone, '') AS timezone " +
		"FROM tasks t LEFT JOIN projects p ON p.id = t.project_id LEFT JOIN users u ON u.id = p.owner_id " +
		"WHERE (t.repeat_after > 0 OR t.repeat_mode > 0) AND t.id > ? ORDER BY t.id ASC LIMIT " + strconv.Itoa(rruleBatchSize20261022220000)

	var (
		lastID                                   int64
		total, converted, skipped, unconvertible int
		serverZone                               int
		zones                                    = ownerZones20261022220000{loaded: map[string]*time.Location{}}
	)
	for {
		var tasks []legacyTaskOwnerZone20261022220000
		err := tx.SQL(query, lastID).Find(&tasks)
		if err != nil {
			return fmt.Errorf("could not read legacy repeats after task %d: %w", lastID, err)
		}
		if len(tasks) == 0 {
			break
		}

		ids := make([]int64, 0, len(tasks))
		backup := make([]taskRepeatBackup20261022220000, 0, len(tasks))
		for _, task := range tasks {
			ids = append(ids, task.ID)
			backup = append(backup, taskRepeatBackup20261022220000{
				ID:          task.ID,
				RepeatAfter: task.RepeatAfter,
				RepeatMode:  task.RepeatMode,
			})
		}
		// A retry after an interrupted run finds this page's backup rows already
		// there, so replace them instead of hitting a PK conflict.
		if _, err := tx.In("id", ids).Delete(&taskRepeatBackup20261022220000{}); err != nil {
			return fmt.Errorf("could not clear the backup page: %w", err)
		}
		if _, err := tx.Insert(&backup); err != nil {
			return fmt.Errorf("could not back up legacy repeats: %w", err)
		}
		if err := failpoint20261022220000("after-backup-page"); err != nil {
			return err
		}

		for _, task := range tasks {
			lastID = task.ID
			total++

			// Defensive: never overwrite a row that already has a rule.
			if task.Repeats != "" {
				skipped++
				continue
			}

			loc, fallback := zones.zone(&task)
			if fallback != "" {
				serverZone++
				log.Debugf("RRULE migration: task %d is read in the server zone (%s)", task.ID, fallback)
			}

			rule, fromCurrentDate := convertLegacyRepeat20261022220000(task.RepeatAfter, task.RepeatMode, task.DueDate, loc)
			if rule == "" {
				// Unexpected mode or a non-convertible value. Leave repeats empty;
				// the original is preserved in the backup table for recovery.
				unconvertible++
				log.Warningf("RRULE migration: task %d has legacy repeat (after=%d, mode=%d) that did not convert; left empty (original preserved in task_repeat_legacy_backup)", task.ID, task.RepeatAfter, task.RepeatMode)
				continue
			}

			if _, err := tx.Exec("UPDATE tasks SET repeats = ?, repeats_from_current_date = ? WHERE id = ?", rule, fromCurrentDate, task.ID); err != nil {
				return fmt.Errorf("could not store the rule of task %d: %w", task.ID, err)
			}
			converted++
		}
	}

	log.Infof("RRULE migration: %d legacy-repeat task(s) processed — %d converted, %d already had an RRULE, %d unconvertible (originals backed up to task_repeat_legacy_backup); %d row(s) used the server zone", total, converted, skipped, unconvertible, serverZone)
	return nil
}

func dropLegacyRepeatColumns20261022220000(tx *xorm.Engine, columns []string) error {
	if tx.Dialect().URI().DBType == schemas.SQLITE {
		return rebuildTasksWithout20261022220000(tx, columns)
	}

	// MySQL and PostgreSQL drop an index on a dropped column with it, and leave
	// every other index alone.
	for i, column := range columns {
		if i > 0 {
			if err := failpoint20261022220000("between-drops"); err != nil {
				return err
			}
		}
		exists, err := columnExists(tx, "tasks", column)
		if err != nil {
			return fmt.Errorf("could not check tasks.%s: %w", column, err)
		}
		if !exists {
			continue
		}
		if err := dropTableColum(tx, "tasks", column); err != nil {
			return fmt.Errorf("could not drop tasks.%s: %w", column, err)
		}
	}
	return nil
}

var legacyColumnRef20261022220000 = regexp.MustCompile(`(?i)\brepeat_(after|mode)\b`)

// rebuildTasksWithout20261022220000 drops columns from tasks on SQLite, which needs a
// table rebuild for that. The new table is derived from the original CREATE TABLE
// text, so implicit UNIQUE/PRIMARY KEY constraints carry over, and every index and
// trigger is recreated from its own stored SQL, minus those on the dropped columns.
// Everything runs in one transaction on one connection: an interruption leaves the
// original table untouched.
func rebuildTasksWithout20261022220000(tx *xorm.Engine, columns []string) (err error) {
	drop := make(map[string]bool, len(columns))
	for _, c := range columns {
		drop[c] = true
	}

	s := tx.NewSession()
	defer s.Close()
	if err := s.Begin(); err != nil {
		return err
	}
	defer func() {
		if err != nil {
			if rbErr := s.Rollback(); rbErr != nil {
				err = errors.Join(err, rbErr)
			}
		}
	}()

	if err := refuseSQLiteForeignKeys20261022220000(s); err != nil {
		return err
	}
	schema, err := readSQLiteTasks20261022220000(s, drop)
	if err != nil {
		return err
	}
	newDDL, err := rebuildTableDDL20261022220000(schema.createSQL, drop)
	if err != nil {
		return err
	}
	columnList := strings.Join(schema.columns, ", ")

	steps := []struct{ sql, after string }{
		{"DROP TABLE IF EXISTS " + rruleTmpTable20261022220000, ""},
		{newDDL, ""},
		{"INSERT INTO " + rruleTmpTable20261022220000 + " (" + columnList + ") SELECT " + columnList + " FROM tasks", "after-copy"},
		{"DROP TABLE tasks", "after-drop"},
		{"ALTER TABLE " + rruleTmpTable20261022220000 + " RENAME TO tasks", ""},
	}
	for _, step := range steps {
		if _, err := s.Exec(step.sql); err != nil {
			return fmt.Errorf("could not rebuild tasks (%s): %w", step.sql, err)
		}
		if step.after != "" {
			if err := failpoint20261022220000(step.after); err != nil {
				return err
			}
		}
	}

	if err := checkSQLiteTasksColumns20261022220000(s, columnList); err != nil {
		return err
	}
	if err := recreateSQLiteTasksDependents20261022220000(s, schema.dependents); err != nil {
		return err
	}
	// Keep the AUTOINCREMENT high-water mark, so ids of deleted tasks are not reused.
	if schema.hasSequence {
		if _, err := s.Exec("UPDATE sqlite_sequence SET seq = MAX(seq, ?) WHERE name = 'tasks'", schema.sequence); err != nil {
			return fmt.Errorf("could not restore the tasks sequence: %w", err)
		}
	}

	return s.Commit()
}

// refuseSQLiteForeignKeys20261022220000 fails when foreign key enforcement is on.
// PRAGMA foreign_keys cannot change inside a transaction, and xorm gives no way to
// set it on the transaction's connection beforehand. Vikunja's DSNs never enable
// it; if a custom one does, refuse before changing anything.
func refuseSQLiteForeignKeys20261022220000(s *xorm.Session) error {
	fk, err := s.QueryString("PRAGMA foreign_keys")
	if err != nil {
		return fmt.Errorf("could not read PRAGMA foreign_keys: %w", err)
	}
	if len(fk) > 0 && fk[0]["foreign_keys"] == "1" {
		return errors.New("RRULE migration: SQLite foreign_keys is ON for this connection, but rebuilding tasks needs it OFF; remove _foreign_keys/_fk from the database DSN and run the migration again")
	}
	return nil
}

// sqliteTasks20261022220000 is what the rebuild captures before it changes anything.
type sqliteTasks20261022220000 struct {
	createSQL   string
	dependents  []map[string]string // indexes and triggers: type, name, sql
	columns     []string            // the surviving columns, quoted, in table order
	sequence    int64
	hasSequence bool
}

func readSQLiteTasks20261022220000(s *xorm.Session, drop map[string]bool) (*sqliteTasks20261022220000, error) {
	tableSQL, err := s.QueryString("SELECT sql FROM sqlite_master WHERE type = 'table' AND name = 'tasks'")
	if err != nil {
		return nil, fmt.Errorf("could not read the tasks DDL: %w", err)
	}
	if len(tableSQL) != 1 {
		return nil, errors.New("could not read the tasks DDL: table not found")
	}
	dependents, err := s.QueryString("SELECT type, name, sql FROM sqlite_master WHERE type IN ('index', 'trigger') AND tbl_name = 'tasks' AND sql IS NOT NULL ORDER BY type, name")
	if err != nil {
		return nil, fmt.Errorf("could not read the tasks indexes: %w", err)
	}
	columnInfo, err := s.QueryString("PRAGMA table_info(tasks)")
	if err != nil {
		return nil, fmt.Errorf("could not read the tasks columns: %w", err)
	}

	t := &sqliteTasks20261022220000{createSQL: tableSQL[0]["sql"], dependents: dependents}
	for _, c := range columnInfo {
		if !drop[c["name"]] {
			t.columns = append(t.columns, quoteSQLiteIdent20261022220000(c["name"]))
		}
	}

	sequence, err := s.QueryString("SELECT seq FROM sqlite_sequence WHERE name = 'tasks'")
	if err != nil && !strings.Contains(err.Error(), "no such table") {
		return nil, fmt.Errorf("could not read the tasks sequence: %w", err)
	}
	if len(sequence) == 1 {
		t.sequence, err = strconv.ParseInt(sequence[0]["seq"], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("could not parse the tasks sequence %q: %w", sequence[0]["seq"], err)
		}
		t.hasSequence = true
	}
	return t, nil
}

// checkSQLiteTasksColumns20261022220000 verifies the rebuilt table has exactly the
// surviving columns, so a DDL the splitter mishandled fails instead of committing.
func checkSQLiteTasksColumns20261022220000(s *xorm.Session, want string) error {
	rebuilt, err := s.QueryString("PRAGMA table_info(tasks)")
	if err != nil {
		return fmt.Errorf("could not read the rebuilt tasks columns: %w", err)
	}
	got := make([]string, 0, len(rebuilt))
	for _, c := range rebuilt {
		got = append(got, quoteSQLiteIdent20261022220000(c["name"]))
	}
	if strings.Join(got, ", ") != want {
		return fmt.Errorf("could not rebuild tasks: rebuilt columns (%s) differ from the expected (%s)", strings.Join(got, ", "), want)
	}
	return nil
}

func recreateSQLiteTasksDependents20261022220000(s *xorm.Session, dependents []map[string]string) error {
	recreated := 0
	for _, d := range dependents {
		if legacyColumnRef20261022220000.MatchString(d["sql"]) {
			log.Infof("RRULE migration: dropped %s %s, which covers a legacy repeat column", d["type"], d["name"])
			continue
		}
		if _, err := s.Exec(d["sql"]); err != nil {
			return fmt.Errorf("could not recreate %s %s: %w", d["type"], d["name"], err)
		}
		recreated++
		if recreated == 1 {
			if err := failpoint20261022220000("during-index-recreate"); err != nil {
				return err
			}
		}
	}
	return nil
}

func quoteSQLiteIdent20261022220000(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

// rebuildTableDDL20261022220000 rewrites a SQLite CREATE TABLE statement for the
// rebuild's temp table, without the given columns. It fails when a table
// constraint refers to a dropped column, rather than guess how to rewrite it.
func rebuildTableDDL20261022220000(createSQL string, drop map[string]bool) (string, error) {
	open := strings.IndexByte(createSQL, '(')
	if open < 0 {
		return "", fmt.Errorf("could not parse the tasks DDL: no column list in %q", createSQL)
	}
	parts, closeAt, err := splitSQLiteList20261022220000(createSQL, open)
	if err != nil {
		return "", err
	}

	kept := make([]string, 0, len(parts))
	for _, part := range parts {
		name, isConstraint := sqliteDefName20261022220000(part)
		if isConstraint {
			if legacyColumnRef20261022220000.MatchString(part) {
				return "", fmt.Errorf("could not rebuild tasks: table constraint %q refers to a legacy repeat column", part)
			}
			kept = append(kept, part)
			continue
		}
		if drop[strings.ToLower(name)] {
			continue
		}
		kept = append(kept, part)
	}

	return "CREATE TABLE " + quoteSQLiteIdent20261022220000(rruleTmpTable20261022220000) + " (" + strings.Join(kept, ", ") + ")" + createSQL[closeAt+1:], nil
}

// splitSQLiteList20261022220000 splits the parenthesised list starting at open into
// its top-level, comma-separated elements, honouring quotes, brackets, nested
// parentheses and comments. Comments are dropped. It returns the index of the
// closing parenthesis.
func splitSQLiteList20261022220000(sql string, open int) (parts []string, closeAt int, err error) {
	var current strings.Builder
	flush := func() {
		if p := strings.TrimSpace(current.String()); p != "" {
			parts = append(parts, p)
		}
		current.Reset()
	}

	depth := 0
	for i := open + 1; i < len(sql); i++ {
		c := sql[i]
		switch {
		case c == '\'' || c == '"' || c == '`' || c == '[':
			end := c
			if c == '[' {
				end = ']'
			}
			j := i + 1
			for ; j < len(sql); j++ {
				if sql[j] == end {
					// A doubled quote is an escaped quote, except inside brackets.
					if end != ']' && j+1 < len(sql) && sql[j+1] == end {
						j++
						continue
					}
					break
				}
			}
			if j >= len(sql) {
				return nil, 0, fmt.Errorf("could not parse the tasks DDL: unterminated %c", c)
			}
			current.WriteString(sql[i : j+1])
			i = j
		case c == '-' && i+1 < len(sql) && sql[i+1] == '-':
			for i < len(sql) && sql[i] != '\n' {
				i++
			}
			current.WriteByte(' ')
		case c == '/' && i+1 < len(sql) && sql[i+1] == '*':
			end := strings.Index(sql[i+2:], "*/")
			if end < 0 {
				return nil, 0, errors.New("could not parse the tasks DDL: unterminated comment")
			}
			i += end + 3
			current.WriteByte(' ')
		case c == '(':
			depth++
			current.WriteByte(c)
		case c == ')' && depth == 0:
			flush()
			return parts, i, nil
		case c == ')':
			depth--
			current.WriteByte(c)
		case c == ',' && depth == 0:
			flush()
		default:
			current.WriteByte(c)
		}
	}
	return nil, 0, errors.New("could not parse the tasks DDL: unbalanced parentheses")
}

// sqliteDefName20261022220000 returns the column name of a column definition, or
// reports that the element is a table constraint.
func sqliteDefName20261022220000(def string) (name string, isConstraint bool) {
	switch def[0] {
	case '"', '`', '\'':
		end := strings.IndexByte(def[1:], def[0])
		if end < 0 {
			return def, false
		}
		return def[1 : end+1], false
	case '[':
		end := strings.IndexByte(def, ']')
		if end < 0 {
			return def, false
		}
		return def[1:end], false
	}
	word := def
	if i := strings.IndexAny(def, " \t\r\n("); i >= 0 {
		word = def[:i]
	}
	switch strings.ToUpper(word) {
	case "CONSTRAINT", "PRIMARY", "UNIQUE", "CHECK", "FOREIGN":
		return "", true
	}
	return word, false
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20261022220000",
		Description: "Convert legacy task repeats to RFC 5545 recurrence rules",
		Migrate:     migrateRRule20261022220000,
		Rollback: func(tx *xorm.Engine) error {
			// Irreversible by design; task_repeat_legacy_backup is the manual way back.
			return nil
		},
	})
}
