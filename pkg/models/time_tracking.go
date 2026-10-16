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

package models

import (
	"encoding/json"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/license"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"
	"xorm.io/builder"
	"xorm.io/xorm"
)

// TimeEntry is a single tracked time span attached to a task, a project, or
// nothing — at most one of TaskID / ProjectID is set. Entries without either
// are private to their author. A running live timer is just an entry whose
// EndTime is still null.
//
// v2-only: doc: tags are the schema's source of truth (no v1 swaggo), and it
// implements CRUDable + Permissions because the shared handler.Do* pipeline needs them.
type TimeEntry struct {
	ID int64 `xorm:"bigint autoincr not null unique pk" json:"id" param:"timeentry" readOnly:"true" doc:"The unique, numeric id of this time entry."`

	UserID int64 `xorm:"bigint not null INDEX" json:"user_id" readOnly:"true" doc:"The id of the user who logged this time entry. Set by the server."`

	TaskID    int64 `xorm:"bigint null INDEX" json:"task_id" nullable:"true" doc:"The task this entry is attached to. At most one of task_id / project_id may be set; 0 means none. On update, omitting both keeps the current container; sending either, even as null, replaces it."`
	ProjectID int64 `xorm:"bigint null INDEX" json:"project_id" nullable:"true" doc:"The project this entry is attached to directly. At most one of task_id / project_id may be set; 0 means none. Entries without task and project are only visible to their author."`

	StartTime time.Time  `xorm:"not null INDEX" json:"start_time" doc:"When the tracked time started."`
	EndTime   *time.Time `xorm:"null" json:"end_time" doc:"When the tracked time ended. Null means a live timer is still running."`

	Comment string `xorm:"text null" json:"comment" doc:"An optional comment describing the logged time."`

	InTrash bool `xorm:"-" json:"in_trash" readOnly:"true" doc:"Whether the entry's task is in the trash. Such entries are read-only."`

	Created time.Time `xorm:"created not null" json:"created" readOnly:"true" doc:"A timestamp when this time entry was created. You cannot change this value."`
	Updated time.Time `xorm:"updated not null" json:"updated" readOnly:"true" doc:"A timestamp when this time entry was last updated. You cannot change this value."`

	// Filter-only fields (not persisted): set by the v2 list route, read by ReadAll.
	Filter         string `xorm:"-" json:"-"`
	FilterTimezone string `xorm:"-" json:"-"`

	// Whether the request body named a container at all, so an update can tell "detach" from "keep".
	containerSent bool

	web.CRUDable    `xorm:"-" json:"-"`
	web.Permissions `xorm:"-" json:"-"`
}

func (te *TimeEntry) UnmarshalJSON(data []byte) error {
	type plain TimeEntry
	if err := json.Unmarshal(data, (*plain)(te)); err != nil {
		return err
	}
	keys := map[string]json.RawMessage{}
	if err := json.Unmarshal(data, &keys); err != nil {
		return err
	}
	_, hasTask := keys["task_id"]
	_, hasProject := keys["project_id"]
	te.containerSent = hasTask || hasProject
	return nil
}

// TableName is time_entries, not the xorm-default time_entry.
func (*TimeEntry) TableName() string {
	return "time_entries"
}

// --- CRUDable ---

func (te *TimeEntry) Create(s *xorm.Session, a web.Auth) (err error) {
	te.UserID = a.GetID()

	// Starting a new running timer auto-stops the previous one; a completed
	// manual entry (EndTime set) must leave the running timer alone.
	if te.EndTime == nil {
		if _, err = stopRunningTimerForUser(s, te.UserID); err != nil {
			return err
		}
	}

	if te.StartTime.IsZero() {
		te.StartTime = time.Now()
	}

	if err = te.validateTimes(); err != nil {
		return err
	}

	if _, err = s.Insert(te); err != nil {
		return err
	}

	doer, err := user.GetFromAuth(a)
	if err != nil {
		return err
	}
	events.DispatchOnCommit(s, &TimeEntryCreatedEvent{TimeEntry: te, Doer: doer})
	return nil
}

func (te *TimeEntry) ReadOne(_ *xorm.Session, _ web.Auth) (err error) {
	// entry got already fetched in CanRead, nothing left to do here
	return nil
}

// stopRunningTimerForUser stops the user's active timer (end_time = now) and
// returns it, or nil if no timer is running.
func stopRunningTimerForUser(s *xorm.Session, userID int64) (*TimeEntry, error) {
	running := &TimeEntry{}
	exists, err := s.Where("user_id = ? AND end_time IS NULL", userID).Get(running)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, nil
	}
	if err := running.stop(s); err != nil {
		return nil, err
	}

	doer, err := user.GetUserByID(s, userID)
	if err != nil {
		return nil, err
	}
	events.DispatchOnCommit(s, &TimeEntryUpdatedEvent{TimeEntry: running, Doer: doer})
	return running, nil
}

// Stops every user's running timer on the task, not just the doer's.
func stopRunningTimersOnTask(s *xorm.Session, taskID int64, doer *user.User) error {
	running := []*TimeEntry{}
	if err := s.Where("task_id = ? AND end_time IS NULL", taskID).Find(&running); err != nil {
		return err
	}
	for _, entry := range running {
		if err := entry.stop(s); err != nil {
			return err
		}
		events.DispatchOnCommit(s, &TimeEntryUpdatedEvent{TimeEntry: entry, Doer: doer})
	}
	return nil
}

// StopRunningTimer stops the authenticated user's active timer and returns it,
// or ErrNoRunningTimer when none is running. The stop time is the server's now.
func StopRunningTimer(s *xorm.Session, a web.Auth) (*TimeEntry, error) {
	// Link shares have no time tracking (mirrors the Can* methods). Their id is a
	// share id, not a user id, so without this a share whose id collides with a
	// user's would stop and read that user's running timer.
	if _, isShare := a.(*LinkSharing); isShare {
		return nil, ErrGenericForbidden{}
	}

	running, err := stopRunningTimerForUser(s, a.GetID())
	if err != nil {
		return nil, err
	}
	if running == nil {
		return nil, ErrNoRunningTimer{UserID: a.GetID()}
	}
	return running, nil
}

// readableTimeEntriesCond restricts a query to entries the auth can read: a
// standalone entry on an accessible project, one on a task in such a project,
// or the auth's own entries without any container.
func readableTimeEntriesCond(s *xorm.Session, a web.Auth) (builder.Cond, error) {
	accessible, err := accessibleProjectIDsCond(s, a, "project_id")
	if err != nil {
		return nil, err
	}
	return builder.Or(
		entriesForProjectCond(accessible),
		builder.And(timeEntryWithoutContainerCond(), builder.Eq{"user_id": a.GetID()}),
	), nil
}

func timeEntryWithoutContainerCond() builder.Cond {
	return builder.And(
		builder.Or(builder.Eq{"task_id": 0}, builder.IsNull{"task_id"}),
		builder.Or(builder.Eq{"project_id": 0}, builder.IsNull{"project_id"}),
	)
}

func (te *TimeEntry) ReadAll(s *xorm.Session, a web.Auth, search string, page int, perPage int) (result any, resultCount int, numberOfTotalItems int64, err error) {
	// Link shares have no time-tracking access (mirrors the Can* methods);
	// DoReadAll skips the permission check, so it must be guarded here too.
	if _, isShareAuth := a.(*LinkSharing); isShareAuth {
		return []*TimeEntry{}, 0, 0, nil
	}

	cond, err := readableTimeEntriesCond(s, a)
	if err != nil {
		return nil, 0, 0, err
	}
	if te.TaskID > 0 {
		cond = cond.And(builder.Eq{"task_id": te.TaskID})
	}
	if te.ProjectID > 0 {
		cond = cond.And(entriesForProjectCond(builder.Eq{"project_id": te.ProjectID}))
	}

	filterCond, err := timeEntryFilterCond(te.Filter, te.FilterTimezone)
	if err != nil {
		return nil, 0, 0, err
	}
	if filterCond != nil {
		cond = cond.And(filterCond)
	}

	if search != "" {
		cond = cond.And(db.MultiFieldSearch([]string{"comment"}, search))
	}

	total, err := s.Where(cond).
		Count(&TimeEntry{})
	if err != nil {
		return nil, 0, 0, err
	}

	entries := []*TimeEntry{}
	err = s.Where(cond).
		OrderBy("start_time ASC").
		Limit(getLimitFromPageIndex(page, perPage)).
		Find(&entries)
	if err != nil {
		return nil, 0, 0, err
	}
	if err = markTimeEntriesInTrash(s, entries); err != nil {
		return nil, 0, 0, err
	}
	return entries, len(entries), total, nil
}

func (te *TimeEntry) Update(s *xorm.Session, a web.Auth) (err error) {
	// A completed entry can't be reopened into a running timer via update — that
	// would sidestep Create's single-active-timer rule; start a new one instead.
	existing, err := getTimeEntryByID(s, te.ID)
	if err != nil {
		return err
	}
	if existing.EndTime != nil && te.EndTime == nil {
		return ErrTimeEntryAlreadyEnded{TimeEntryID: te.ID}
	}

	if err = te.validateTimes(); err != nil {
		return err
	}

	// task_id / project_id are listed so a reassignment (and the zero value of
	// the side being cleared) is written; validated in CanUpdate.
	_, err = s.
		Where("id = ?", te.ID).
		Cols("task_id", "project_id", "start_time", "end_time", "comment").
		Update(te)
	if err != nil {
		return err
	}

	// reload: Update wrote only the editable columns
	updated, err := getTimeEntryByID(s, te.ID)
	if err != nil {
		return err
	}
	*te = *updated

	doer, err := user.GetFromAuth(a)
	if err != nil {
		return err
	}
	events.DispatchOnCommit(s, &TimeEntryUpdatedEvent{TimeEntry: te, Doer: doer})
	return nil
}

func (te *TimeEntry) Delete(s *xorm.Session, a web.Auth) (err error) {
	entry, err := getTimeEntryByID(s, te.ID)
	if err != nil {
		return err
	}
	if _, err = s.Where("id = ?", te.ID).Delete(&TimeEntry{}); err != nil {
		return err
	}

	doer, err := user.GetFromAuth(a)
	if err != nil {
		return err
	}
	events.DispatchOnCommit(s, &TimeEntryDeletedEvent{TimeEntry: entry, Doer: doer})
	return nil
}

func getTimeEntryByID(s *xorm.Session, id int64) (*TimeEntry, error) {
	entry := &TimeEntry{}
	exists, err := s.Where("id = ?", id).Get(entry)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrTimeEntryDoesNotExist{TimeEntryID: id}
	}
	return entry, nil
}

func (te *TimeEntry) stop(s *xorm.Session) (err error) {
	now := time.Now()
	te.EndTime = &now
	_, err = s.ID(te.ID).Update(te)
	return err
}

// --- Permissions ---

// Returns the loaded entry rather than mutating te, so Update keeps its payload.
func (te *TimeEntry) canDoTimeEntry(s *xorm.Session, a web.Auth, fetch bool) (*TimeEntry, bool, int, error) {
	entry := &TimeEntry{TaskID: te.TaskID, ProjectID: te.ProjectID}
	if fetch {
		var err error
		entry, err = getTimeEntryByID(s, te.ID)
		if err != nil {
			return nil, false, -1, err
		}
	}

	switch {
	case entry.TaskID != 0:
		task := &Task{}
		has, err := s.Unscoped().Where("id = ?", entry.TaskID).Get(task)
		if err != nil {
			return entry, false, -1, err
		}
		if !has {
			return entry, false, -1, ErrTaskDoesNotExist{ID: entry.TaskID}
		}
		entry.InTrash = !task.DeletedAt.IsZero()
		can, maxPerm, err := (&Project{ID: task.ProjectID}).CanRead(s, a)
		if entry.InTrash {
			maxPerm = int(PermissionRead)
		}
		return entry, can, maxPerm, err
	case entry.ProjectID != 0:
		project, _, err := getProjectSimple(s, builder.Eq{"id": entry.ProjectID})
		if err != nil {
			return entry, false, -1, err
		}
		can, maxPerm, err := project.CanRead(s, a)
		return entry, can, maxPerm, err
	default:
		// Private to the author; an entry about to be created has no author yet.
		return entry, !fetch || entry.UserID == a.GetID(), int(PermissionWrite), nil
	}
}

func (te *TimeEntry) CanRead(s *xorm.Session, a web.Auth) (bool, int, error) {
	if _, isShareAuth := a.(*LinkSharing); isShareAuth {
		return false, 0, nil
	}

	entry, can, maxPerm, err := te.canDoTimeEntry(s, a, true)
	if err != nil {
		return false, maxPerm, err
	}
	*te = *entry // ReadOne is a no-op; populate te here
	return can, maxPerm, nil
}

func (te *TimeEntry) validateContainer() error {
	if te.TaskID != 0 && te.ProjectID != 0 {
		return ErrTimeEntryInvalidContainer{TaskID: te.TaskID, ProjectID: te.ProjectID}
	}
	return nil
}

// validateTimes rejects a completed entry whose end precedes its start (a
// negative interval). A null end is a running timer and is always valid; an end
// equal to the start is allowed (a zero-length entry).
func (te *TimeEntry) validateTimes() error {
	if te.EndTime != nil && te.EndTime.Before(te.StartTime) {
		return ErrTimeEntryEndBeforeStart{TimeEntryID: te.ID}
	}
	return nil
}

func (te *TimeEntry) CanCreate(s *xorm.Session, a web.Auth) (bool, error) {
	if _, isShareAuth := a.(*LinkSharing); isShareAuth {
		return false, nil
	}

	if err := te.validateContainer(); err != nil {
		return false, err
	}

	entry, can, _, err := te.canDoTimeEntry(s, a, false)
	if err != nil {
		return false, err
	}
	return can && !entry.InTrash, nil
}

// CanUpdate allows the author to edit their entry, including moving it between
// task / project: on top of the author check it validates the (possibly new)
// container and requires read access to it, mirroring create.
func (te *TimeEntry) CanUpdate(s *xorm.Session, a web.Auth) (bool, error) {
	if _, isShareAuth := a.(*LinkSharing); isShareAuth {
		return false, nil
	}

	existing, err := getTimeEntryByID(s, te.ID)
	if err != nil {
		return false, err
	}
	if existing.UserID != a.GetID() {
		return false, nil
	}
	existingInTrash, err := isTaskTrashed(s, existing.TaskID)
	if err != nil || existingInTrash {
		return false, err
	}

	if te.TaskID == 0 && te.ProjectID == 0 && !te.containerSent {
		te.TaskID = existing.TaskID
		te.ProjectID = existing.ProjectID
	}
	if err := te.validateContainer(); err != nil {
		return false, err
	}

	container, canReadContainer, _, err := te.canDoTimeEntry(s, a, false)
	if err != nil {
		return false, err
	}
	return canReadContainer && !container.InTrash, nil
}

func (te *TimeEntry) CanDelete(s *xorm.Session, a web.Auth) (bool, error) {
	return te.canModify(s, a)
}

// canModify gates delete: read access to the container plus being the author.
func (te *TimeEntry) canModify(s *xorm.Session, a web.Auth) (bool, error) {
	if _, isShareAuth := a.(*LinkSharing); isShareAuth {
		return false, nil
	}

	entry, canRead, _, err := te.canDoTimeEntry(s, a, true)
	if err != nil {
		return false, err
	}
	if !canRead {
		return false, nil
	}
	return entry.UserID == a.GetID() && !entry.InTrash, nil
}

func isTaskTrashed(s *xorm.Session, taskID int64) (bool, error) {
	if taskID == 0 {
		return false, nil
	}
	return s.Table("tasks").Where(builder.And(builder.Eq{"id": taskID}, builder.NotNull{"deleted_at"})).Exist()
}

func markTimeEntriesInTrash(s *xorm.Session, entries []*TimeEntry) error {
	taskIDs := []int64{}
	for _, e := range entries {
		if e.TaskID != 0 {
			taskIDs = append(taskIDs, e.TaskID)
		}
	}
	if len(taskIDs) == 0 {
		return nil
	}

	trashedIDs := []int64{}
	err := s.Table("tasks").
		Cols("id").
		Where(builder.And(builder.In("id", taskIDs), builder.NotNull{"deleted_at"})).
		Find(&trashedIDs)
	if err != nil {
		return err
	}
	trashed := make(map[int64]bool, len(trashedIDs))
	for _, id := range trashedIDs {
		trashed[id] = true
	}
	for _, e := range entries {
		e.InTrash = trashed[e.TaskID]
	}
	return nil
}

// addTimeEntriesCountToTasks attaches each task's time-entry count for the
// `time_entries_count` expand. Mirrors addCommentCountToTasks, but follows the
// same gates as the time-entry endpoints: the count is left unset (absent) for
// link shares or when the feature is unlicensed, so it can't leak that way.
func addTimeEntriesCountToTasks(s *xorm.Session, a web.Auth, taskIDs []int64, taskMap map[int64]*Task) error {
	if _, isShare := a.(*LinkSharing); isShare {
		return nil
	}
	if !license.IsFeatureEnabled(license.FeatureTimeTracking) {
		return nil
	}
	if len(taskIDs) == 0 {
		return nil
	}

	zero := int64(0)
	for _, taskID := range taskIDs {
		if task, ok := taskMap[taskID]; ok {
			task.TimeEntriesCount = &zero
		}
	}

	type timeEntriesCount struct {
		TaskID int64 `xorm:"task_id"`
		Count  int64 `xorm:"count"`
	}

	counts := []timeEntriesCount{}
	if err := s.
		Select("task_id, COUNT(*) as count").
		Where(builder.In("task_id", taskIDs)).
		GroupBy("task_id").
		Table("time_entries").
		Find(&counts); err != nil {
		return err
	}

	for _, c := range counts {
		if task, ok := taskMap[c.TaskID]; ok {
			task.TimeEntriesCount = &c.Count
		}
	}

	return nil
}
