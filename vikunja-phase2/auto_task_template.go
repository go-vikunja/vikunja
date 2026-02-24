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
	"time"

	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"xorm.io/xorm"
)

// AutoTaskTemplate defines a recurring task template that auto-generates task
// instances when they become due. Only one active instance can exist at a time —
// if the previous task hasn't been completed, it simply goes overdue.
// The next instance is scheduled based on when the user completes the previous one.
type AutoTaskTemplate struct {
	// The unique, numeric id of this template.
	ID int64 `xorm:"autoincr not null unique pk" json:"id" param:"autotask"`
	// The user who owns this template.
	OwnerID int64      `xorm:"bigint not null INDEX" json:"-"`
	Owner   *user.User `xorm:"-" json:"owner" valid:"-"`
	// The project to create tasks in. 0 = user's default project.
	ProjectID int64 `xorm:"bigint null" json:"project_id"`
	// The task title.
	Title string `xorm:"varchar(250) not null" json:"title" valid:"required,runelength(1|250)" minLength:"1" maxLength:"250"`
	// Optional description for the generated task.
	Description string `xorm:"longtext null" json:"description"`
	// Task priority (0-5).
	Priority int64 `xorm:"bigint null" json:"priority"`
	// Task color in hex.
	HexColor string `xorm:"varchar(7) null" json:"hex_color" valid:"runelength(0|7)" maxLength:"7"`
	// Label IDs to apply to generated tasks.
	LabelIDs []int64 `xorm:"json null" json:"label_ids"`
	// User IDs to assign to generated tasks.
	AssigneeIDs []int64 `xorm:"json null" json:"assignee_ids"`

	// -- Scheduling --
	// The numeric interval value (e.g. 1, 7, 14).
	IntervalValue int `xorm:"int not null default 1" json:"interval_value"`
	// The unit for the interval: hours, days, weeks, months.
	IntervalUnit string `xorm:"varchar(10) not null default 'days'" json:"interval_unit"`
	// When this template first becomes active.
	StartDate time.Time `xorm:"datetime not null" json:"start_date"`
	// Optional end date — stop generating after this.
	EndDate *time.Time `xorm:"datetime null" json:"end_date"`
	// Whether this template is actively generating tasks. False = paused.
	Active bool `xorm:"bool not null default true" json:"active"`

	// -- Tracking --
	// When the last task instance was auto-created.
	LastCreatedAt *time.Time `xorm:"datetime null" json:"last_created_at"`
	// When the user last completed a generated task (drives next scheduling).
	LastCompletedAt *time.Time `xorm:"datetime null" json:"last_completed_at"`
	// Pre-computed: when the next instance should be created.
	NextDueAt *time.Time `xorm:"datetime null" json:"next_due_at"`

	// Generation log entries (loaded separately, most recent first).
	Log []*AutoTaskLog `xorm:"-" json:"log"`

	// Timestamps
	Created time.Time `xorm:"created not null" json:"created"`
	Updated time.Time `xorm:"updated not null" json:"updated"`

	web.CRUDable    `xorm:"-" json:"-"`
	web.Permissions `xorm:"-" json:"-"`
}

// TableName returns the table name for xorm.
func (a *AutoTaskTemplate) TableName() string {
	return "auto_task_templates"
}

// AutoTaskLog records each time a task is generated from a template.
type AutoTaskLog struct {
	ID         int64     `xorm:"bigint autoincr not null unique pk" json:"id"`
	TemplateID int64     `xorm:"bigint not null INDEX" json:"template_id"`
	TaskID     int64     `xorm:"bigint not null" json:"task_id"`
	// "system" for cron/auto-check, "manual" for user-triggered, "cron" for background
	TriggerType string    `xorm:"varchar(20) not null" json:"trigger_type"`
	// The user who triggered (null for system cron)
	TriggeredByID int64  `xorm:"bigint null" json:"triggered_by_id"`
	Created       time.Time `xorm:"created not null" json:"created"`

	// Enrichment fields (populated after load, not in DB)
	TaskTitle    string     `xorm:"-" json:"task_title"`
	TaskDone     bool       `xorm:"-" json:"task_done"`
	TaskDoneAt   *time.Time `xorm:"-" json:"task_done_at"`
	TaskUpdated  *time.Time `xorm:"-" json:"task_updated"`
	CommentCount int64      `xorm:"-" json:"comment_count"`
}

// TableName returns the table name for xorm.
func (a *AutoTaskLog) TableName() string {
	return "auto_task_log"
}

// AutoTaskTemplateAttachment represents a file attached to an auto-task template.
type AutoTaskTemplateAttachment struct {
	ID          int64     `xorm:"bigint autoincr not null unique pk" json:"id"`
	TemplateID  int64     `xorm:"bigint not null INDEX" json:"template_id"`
	FileID      int64     `xorm:"bigint not null" json:"file_id"`
	FileName    string    `xorm:"varchar(250) not null" json:"file_name"`
	CreatedByID int64     `xorm:"bigint not null" json:"-"`
	Created     time.Time `xorm:"created" json:"created"`
}

// TableName returns the table name for xorm.
func (a *AutoTaskTemplateAttachment) TableName() string {
	return "auto_task_template_attachments"
}

// --------- CRUD ---------

// ReadAll returns all auto-task templates for the current user.
func (a *AutoTaskTemplate) ReadAll(s *xorm.Session, auth web.Auth, _ string, _ int, _ int) (interface{}, int, int64, error) {
	u := auth.(*user.User)
	templates := []*AutoTaskTemplate{}
	err := s.Where("owner_id = ?", u.ID).OrderBy("title").Find(&templates)
	if err != nil {
		return nil, 0, 0, err
	}

	// Load log entries for each template (last 10) and enrich
	for _, tmpl := range templates {
		tmpl.Owner = u
		logs := []*AutoTaskLog{}
		_ = s.Where("template_id = ?", tmpl.ID).OrderBy("created DESC").Limit(10).Find(&logs)
		enrichAutoTaskLogs(s, logs)
		tmpl.Log = logs
	}

	return templates, len(templates), 0, nil
}

// ReadOne loads a single auto-task template with its log.
func (a *AutoTaskTemplate) ReadOne(s *xorm.Session, _ web.Auth) error {
	tmpl := &AutoTaskTemplate{}
	has, err := s.ID(a.ID).Get(tmpl)
	if err != nil {
		return err
	}
	if !has {
		return ErrAutoTaskTemplateNotFound{ID: a.ID}
	}

	// Load log entries
	logs := []*AutoTaskLog{}
	_ = s.Where("template_id = ?", a.ID).OrderBy("created DESC").Limit(20).Find(&logs)
	enrichAutoTaskLogs(s, logs)
	tmpl.Log = logs

	*a = *tmpl
	return nil
}

// Create creates a new auto-task template.
func (a *AutoTaskTemplate) Create(s *xorm.Session, auth web.Auth) error {
	u := auth.(*user.User)
	a.OwnerID = u.ID
	a.Owner = u

	// Compute first next_due_at from start_date
	nextDue := a.StartDate
	a.NextDueAt = &nextDue

	_, err := s.Insert(a)
	return err
}

// Update saves changes to an auto-task template.
func (a *AutoTaskTemplate) Update(s *xorm.Session, _ web.Auth) error {
	_, err := s.ID(a.ID).Cols(
		"title", "description", "project_id", "priority", "hex_color",
		"label_ids", "assignee_ids",
		"interval_value", "interval_unit", "start_date", "end_date",
		"active",
	).Update(a)
	return err
}

// Delete removes an auto-task template and its log entries.
func (a *AutoTaskTemplate) Delete(s *xorm.Session, _ web.Auth) error {
	_, _ = s.Where("template_id = ?", a.ID).Delete(&AutoTaskLog{})
	_, _ = s.Where("template_id = ?", a.ID).Delete(&AutoTaskTemplateAttachment{})
	_, err := s.ID(a.ID).Delete(&AutoTaskTemplate{})
	return err
}

// --------- Permissions ---------

func (a *AutoTaskTemplate) CanRead(s *xorm.Session, auth web.Auth) (bool, int, error) {
	return a.isOwner(s, auth)
}

func (a *AutoTaskTemplate) CanCreate(_ *xorm.Session, _ web.Auth) (bool, error) {
	return true, nil
}

func (a *AutoTaskTemplate) CanUpdate(s *xorm.Session, auth web.Auth) (bool, error) {
	can, _, err := a.isOwner(s, auth)
	return can, err
}

func (a *AutoTaskTemplate) CanDelete(s *xorm.Session, auth web.Auth) (bool, error) {
	can, _, err := a.isOwner(s, auth)
	return can, err
}

func (a *AutoTaskTemplate) isOwner(s *xorm.Session, auth web.Auth) (bool, int, error) {
	u := auth.(*user.User)
	tmpl := &AutoTaskTemplate{}
	has, err := s.Where("id = ? AND owner_id = ?", a.ID, u.ID).Get(tmpl)
	if err != nil {
		return false, 0, err
	}
	return has, int(PermissionAdmin), nil
}

// --------- Errors ---------

type ErrAutoTaskTemplateNotFound struct {
	ID int64
}

func (e ErrAutoTaskTemplateNotFound) Error() string {
	return "Auto-task template not found"
}

func (e ErrAutoTaskTemplateNotFound) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: 404,
		Code:     15001,
		Message:  "The auto-task template was not found.",
	}
}

// enrichAutoTaskLogs populates task metadata on log entries by batch-querying the tasks table.
func enrichAutoTaskLogs(s *xorm.Session, logs []*AutoTaskLog) {
	if len(logs) == 0 {
		return
	}

	// Collect unique task IDs
	taskIDs := make([]int64, 0, len(logs))
	seen := make(map[int64]bool)
	for _, l := range logs {
		if !seen[l.TaskID] {
			taskIDs = append(taskIDs, l.TaskID)
			seen[l.TaskID] = true
		}
	}

	if len(taskIDs) == 0 {
		return
	}

	ids := inClause(taskIDs)

	// Batch load task info via raw SQL
	type taskInfo struct {
		ID      int64     `xorm:"'id'"`
		Title   string    `xorm:"'title'"`
		Done    int       `xorm:"'done'"`
		DoneAt  time.Time `xorm:"'done_at'"`
		Updated time.Time `xorm:"'updated'"`
	}
	tasks := make([]*taskInfo, 0)
	_ = s.SQL("SELECT id, title, done, done_at, updated FROM tasks WHERE id IN (" + ids + ")").Find(&tasks)

	taskMap := make(map[int64]*taskInfo)
	for _, t := range tasks {
		taskMap[t.ID] = t
	}

	// Batch load comment counts
	type commentCount struct {
		TaskID int64 `xorm:"'task_id'"`
		Count  int64 `xorm:"'count'"`
	}
	counts := make([]*commentCount, 0)
	_ = s.SQL(
		"SELECT task_id, COUNT(*) as count FROM task_comments WHERE task_id IN (" + ids + ") GROUP BY task_id",
	).Find(&counts)

	countMap := make(map[int64]int64)
	for _, c := range counts {
		countMap[c.TaskID] = c.Count
	}

	// Enrich each log entry
	for _, l := range logs {
		if t, ok := taskMap[l.TaskID]; ok {
			l.TaskTitle = t.Title
			l.TaskDone = t.Done != 0
			if !t.DoneAt.IsZero() {
				doneAt := t.DoneAt
				l.TaskDoneAt = &doneAt
			}
			if !t.Updated.IsZero() {
				updated := t.Updated
				l.TaskUpdated = &updated
			}
		}
		if c, ok := countMap[l.TaskID]; ok {
			l.CommentCount = c
		}
	}
}

// inClause generates a comma-separated list of int64s for SQL IN clauses.
func inClause(ids []int64) string {
	s := ""
	for i, id := range ids {
		if i > 0 {
			s += ","
		}
		s += fmt.Sprintf("%d", id)
	}
	return s
}
