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
	"errors"
	"fmt"
	"math"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/utils"
	"code.vikunja.io/api/pkg/web"

	"dario.cat/mergo"
	"github.com/google/uuid"
	clone "github.com/huandu/go-clone/generic"
	"github.com/jinzhu/copier"
	"github.com/teambition/rrule-go"
	"xorm.io/builder"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

// validateTaskForCreation holds the model-level rules; the `valid:` tags are enforced at the API boundary only, sparing internal callers which just copy data.
func validateTaskForCreation(t *Task) error {
	if t.Title == "" {
		return ErrTaskCannotBeEmpty{}
	}

	return nil
}

// Task represents a task in a project
type Task struct {
	// The unique, numeric id of this task.
	ID int64 `xorm:"bigint autoincr not null unique pk" json:"id" param:"projecttask" readOnly:"true" doc:"The unique, numeric id of this task."`
	// The task text. This is what you'll see in the project.
	Title string `xorm:"TEXT not null" json:"title" valid:"minstringlength(1)" minLength:"1" doc:"The task title. This is what you'll see in the project."`
	// The task description.
	Description string `xorm:"longtext null" json:"description"`
	// The project this task belongs to.
	// Must precede done/due_date: xorm orders composite index columns by struct field order.
	ProjectID int64 `xorm:"bigint INDEX not null unique(tasks_project_index) index(project_done_due_date)" json:"project_id" param:"project" doc:"The id of the project this task belongs to. On create it is taken from the URL; on update, setting it to a different project moves the task (requires write access to the target project)."`
	// Whether a task is done or not.
	Done bool `xorm:"INDEX null index(project_done_due_date)" json:"done"`
	// The time when a task was marked as done. This field is system-controlled and cannot be set via API.
	DoneAt time.Time `xorm:"INDEX null 'done_at'" json:"done_at" readOnly:"true" doc:"When the task was marked as done. Set by the server; ignored on write."`
	// The time when the task is due.
	DueDate time.Time `xorm:"DATETIME INDEX null index(project_done_due_date) 'due_date'" json:"due_date"`
	// An array of reminders that are associated with this task.
	Reminders []*TaskReminder `xorm:"-" json:"reminders"`
	// An RFC 5545 RRULE string defining the recurrence pattern, stored in the database.
	// Not exposed via JSON; use the structured Repeat field instead.
	Repeats string `xorm:"varchar(500) null" json:"-"`
	// Structured recurrence pattern for the API. Mirrors RFC 5545 RRULE fields.
	// Converted to/from the Repeats RRULE string at the API boundary.
	Repeat *TaskRepeat `xorm:"-" json:"repeat,omitempty" doc:"The recurrence, as RFC 5545 RRULE parts. When the task has a due date, it is the first occurrence. Completing the task moves its dates to the next occurrence and reopens it. Send null to remove the recurrence. A partial update (PATCH) leaves an omitted repeat unchanged; a full update (PUT) replaces every field, so an omitted repeat removes it too. The whole rule, written out as RRULE text, must fit in 500 characters."`
	// If true, the next occurrence is calculated from the completion date rather than the original due date.
	RepeatsFromCurrentDate bool `xorm:"null default false" json:"repeats_from_current_date" doc:"When true, the next occurrence is counted from the moment the task is completed instead of from its previous due date."`
	// The task priority. Can be anything you want, it is possible to sort by this later.
	Priority int64 `xorm:"bigint null" json:"priority"`
	// When this task starts.
	StartDate time.Time `xorm:"DATETIME INDEX null 'start_date'" json:"start_date" query:"-"`
	// When this task ends.
	EndDate time.Time `xorm:"DATETIME INDEX null 'end_date'" json:"end_date" query:"-"`
	// An array of users who are assigned to this task
	Assignees []*user.User `xorm:"-" json:"assignees" readOnly:"true" doc:"The users assigned to this task. Read-only here; use the task-assignee endpoints to change assignments."`
	// An array of labels which are associated with this task. This property is read-only, you must use the separate endpoint to add labels to a task.
	Labels []*Label `xorm:"-" json:"labels" readOnly:"true" doc:"The labels on this task. Read-only here; use the label-task endpoints to add or remove labels."`
	// The task color in hex
	HexColor string `xorm:"varchar(6) null" json:"hex_color" valid:"hexcolor" maxLength:"7" doc:"The task color as a hex string without the leading '#'."`
	// Determines how far a task is left from being done
	PercentDone float64 `xorm:"DOUBLE null" json:"percent_done" doc:"How far the task is from done, between 0 and 1."`

	// The task identifier, based on the project identifier and the task's index
	Identifier string `xorm:"-" json:"identifier" readOnly:"true" doc:"The textual task identifier, derived from the project identifier and the task index (e.g. \"PROJ-12\")."`
	// The task index, calculated per project
	Index int64 `xorm:"bigint not null default 0 unique(tasks_project_index)" json:"index" param:"index" readOnly:"true" doc:"The per-project task index, assigned by the server."`

	// The UID is currently not used for anything other than CalDAV, which is why we don't expose it over json
	UID string `xorm:"varchar(250) null" json:"-"`

	// All related tasks, grouped by their relation kind
	RelatedTasks RelatedTaskMap `xorm:"-" json:"related_tasks" readOnly:"true" doc:"Related tasks grouped by relation kind. Read-only here; use the task-relation endpoints to change relations."`

	// All attachments this task has. This property is read-onlym, you must use the separate endpoint to add attachments to a task.
	Attachments []*TaskAttachment `xorm:"-" json:"attachments" readOnly:"true" doc:"The task's attachments. Read-only here; use the attachment endpoints to add or remove them."`

	// If this task has a cover image, the field will return the id of the attachment that is the cover image.
	CoverImageAttachmentID int64 `xorm:"bigint default 0" json:"cover_image_attachment_id" doc:"The id of the attachment used as this task's cover image, or 0 for none."`

	// True if a task is a favorite task. Favorite tasks show up in a separate "Important" project. This value depends on the user making the call to the api.
	IsFavorite bool `xorm:"-" json:"is_favorite" doc:"Whether the requesting user has favorited this task. Per-user, so it differs between callers."`

	IsUnread *bool `xorm:"-" json:"is_unread,omitempty" readOnly:"true" doc:"Whether the task is unread for the requesting user. Only present when requested via the is_unread expand option."`

	// The subscription status for the user reading this task. You can only read this property, use the subscription endpoints to modify it.
	// Will only returned when retrieving one task.
	Subscription *Subscription `xorm:"-" json:"subscription,omitempty" readOnly:"true" doc:"The requesting user's subscription to this task. Read-only here; use the subscription endpoints to change it. Only present when reading a single task."`

	// A timestamp when this task was created. You cannot change this value.
	Created time.Time `xorm:"created not null" json:"created" readOnly:"true" doc:"When this task was created. Set by the server; ignored on write."`
	// A timestamp when this task was last updated. You cannot change this value.
	Updated time.Time `xorm:"updated not null" json:"updated" readOnly:"true" doc:"When this task was last updated. Set by the server; ignored on write."`
	// A timestamp when this task was deleted. Soft-deleted tasks are kept for 30 days before they are removed permanently.
	// omitzero keeps the field out of the JSON of regular tasks — it only ever appears on soft-deleted ones (the later trash listing).
	DeletedAt time.Time `xorm:"deleted datetime null INDEX 'deleted_at'" json:"deleted_at,omitzero" readOnly:"true" doc:"When this task was soft-deleted. Soft-deleted tasks are kept for 30 days before they are removed permanently."`

	// The bucket id. Will only be populated when the task is accessed via a view with buckets.
	// Can be used to move a task between buckets. In that case, the new bucket must be in the same view as the old one.
	BucketID int64 `xorm:"-" json:"bucket_id" doc:"The bucket the task is in. Only populated when the task is accessed via a view with buckets. To move a task between buckets, the new bucket must be in the same view as the old one."`

	// All buckets across all views this task is part of. Only present when fetching tasks with the `expand` parameter set to `buckets`.
	Buckets []*Bucket `xorm:"-" json:"buckets,omitempty" readOnly:"true" doc:"The task's buckets across all views. Only present when requested via the buckets expand option."`

	// All comments of this task. Only present when fetching tasks with the `expand` parameter set to `comments`.
	Comments []*TaskComment `xorm:"-" json:"comments,omitempty" readOnly:"true" doc:"The task's first 50 comments. Only present when requested via the comments expand option."`

	// Comment count of this task. Only present when fetching tasks with the `expand` parameter set to `comment_count`.
	CommentCount *int64 `xorm:"-" json:"comment_count,omitempty" readOnly:"true" doc:"The number of comments on this task. Only present when requested via the comment_count expand option."`

	// Time entry count of this task. Only present when fetching tasks with the `expand` parameter set to `time_entries_count`.
	TimeEntriesCount *int64 `xorm:"-" json:"time_entries_count,omitempty" readOnly:"true" doc:"The number of time entries on this task. Only present when requested via the time_entries_count expand option."`

	// Behaves exactly the same as with the TaskCollection.Expand parameter
	Expand []TaskCollectionExpandable `xorm:"-" json:"-" query:"expand"`

	// The position of the task - any task project can be sorted as usual by this parameter.
	// When accessing tasks via views with buckets, this is primarily used to sort them based on a range.
	// Positions are always saved per view. They will automatically be set if you request the tasks through a view
	// endpoint, otherwise they will always be 0. To update them, take a look at the Task Position endpoint.
	Position float64 `xorm:"-" json:"position" readOnly:"true" doc:"The task's position, saved per view. Only non-zero when the task is fetched through a view endpoint; use the task-position endpoint to change it."`

	Positions []*TaskPosition `xorm:"-" json:"positions,omitempty" readOnly:"true" doc:"The task's stored position in each view of its project. Only present in create and duplicate responses."`

	// Reactions on that task.
	Reactions ReactionMap `xorm:"-" json:"reactions" readOnly:"true" doc:"Reactions on this task. Only present when requested via the reactions expand option."`

	// The user who initially created the task.
	CreatedBy   *user.User `xorm:"-" json:"created_by" valid:"-" readOnly:"true" doc:"The user who created this task. Set by the server."`
	CreatedByID int64      `xorm:"bigint not null" json:"-"` // ID of the user who put that task on the project

	web.CRUDable    `xorm:"-" json:"-"`
	web.Permissions `xorm:"-" json:"-"`

	// copiesStoredRule marks a task whose Repeats is copied from a rule already stored on
	// another task (duplication), so creation checks it for syntax and the safety bound only,
	// never against the creation policy. Unexported: no API payload, import file or column can
	// set it.
	copiesStoredRule bool
}

type TaskWithComments struct {
	Task
	Comments []*TaskComment `xorm:"-" json:"comments"`
	// RepeatsRRule is export/import-only: it carries the exact stored rule text, which the
	// structured repeat cannot always express. TaskWithComments is never
	// an API response, so this changes no API shape.
	RepeatsRRule string `xorm:"-" json:"repeats_rrule,omitempty"`
}

// TableName returns the table name for tasks
func (*Task) TableName() string {
	return "tasks"
}

// taskNotDeletedCond filters out soft-deleted tasks where the xorm deleted tag
// does not apply: raw SQL, Table("tasks") with non-Task destinations, builder
// subqueries and joins from other beans. IS NULL is enough because deleted_at
// is only ever set on soft delete; restore must set it back to NULL.
func taskNotDeletedCond(tableName string) builder.Cond {
	return builder.IsNull{tableName + ".deleted_at"}
}

// GetFullIdentifier returns the task identifier if the task has one and the index prefixed with # otherwise.
func (t *Task) GetFullIdentifier() string {
	if t.Identifier != "" {
		if strings.HasPrefix(t.Identifier, "-") {
			return "#" + strings.TrimPrefix(t.Identifier, "-")
		}
		return t.Identifier
	}

	return "#" + strconv.FormatInt(t.Index, 10)
}

func (t *Task) GetFrontendURL() string {
	return config.ServicePublicURL.GetString() + "tasks/" + strconv.FormatInt(t.ID, 10)
}

func (t *Task) isRepeating() bool {
	return t.Repeats != ""
}

type taskFilterConcatinator string

const (
	filterConcatAnd taskFilterConcatinator = "and"
	filterConcatOr  taskFilterConcatinator = "or"
)

type taskSearchOptions struct {
	search             string
	page               int
	perPage            int
	sortby             []*sortParam
	parsedFilters      []*taskFilter
	filterIncludeNulls bool
	filter             string
	filterTimezone     string
	isSavedFilter      bool
	projectIDs         []int64
	expand             []TaskCollectionExpandable
	projectViewID      int64

	// userProvidedSort distinguishes an explicit sort_by from the id/position
	// defaults appended later, so relevance ordering only replaces the default sort.
	userProvidedSort bool
}

// ReadAll is a dummy function to still have that endpoint documented
// @Summary Get tasks
// @Description Returns all tasks on any project the user has access to.
// @tags task
// @Accept json
// @Produce json
// @Param page query int false "The page number. Used for pagination. If not provided, the first page of results is returned."
// @Param per_page query int false "The maximum number of items per page. Note this parameter is limited by the configured maximum of items per page."
// @Param s query string false "Search tasks by task text."
// @Param sort_by query string false "The sorting parameter. You can pass this multiple times to get the tasks ordered by multiple different parametes, along with `order_by`. Possible values to sort by are `id`, `title`, `description`, `done`, `done_at`, `due_date`, `created_by_id`, `project_id`, `repeats`, `priority`, `start_date`, `end_date`, `hex_color`, `percent_done`, `uid`, `created`, `updated`, `relevance`. `relevance` sorts by search relevance (most relevant first, requires `s`; ignored when the database cannot score the query). Default is `id`."
// @Param order_by query string false "The ordering parameter. Possible values to order by are `asc` or `desc`. Default is `asc`."
// @Param filter query string false "The filter query to match tasks by. Check out https://vikunja.io/docs/filters for a full explanation of the feature."
// @Param filter_timezone query string false "The time zone which should be used for date match (statements like "now" resolve to different actual times)"
// @Param filter_include_nulls query string false "If set to true the result will include filtered fields whose value is set to `null`. Available values are `true` or `false`. Defaults to `false`."
// @Param expand query string false "If set to `subtasks`, Vikunja will fetch only tasks which do not have subtasks and then in a second step, will fetch all of these subtasks. This may result in more tasks than the pagination limit being returned, but all subtasks will be present in the response. If set to `buckets`, the buckets of each task will be present in the response. If set to `reactions`, the reactions of each task will be present in the response. If set to `comments`, the first 50 comments of each task will be present in the response. You can set this multiple times with different values."
// @Security JWTKeyAuth
// @Success 200 {array} models.Task "The tasks"
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks [get]
func (t *Task) ReadAll(_ *xorm.Session, _ web.Auth, _ string, _ int, _ int) (result interface{}, resultCount int, totalItems int64, err error) {
	return nil, 0, 0, nil
}

func getFilterCond(f *taskFilter, includeNulls bool) (cond builder.Cond, err error) {
	field := f.field

	switch f.comparator {
	case taskFilterComparatorEquals:
		cond = &builder.Eq{field: f.value}
	case taskFilterComparatorNotEquals:
		cond = &builder.Neq{field: f.value}
	case taskFilterComparatorGreater:
		cond = &builder.Gt{field: f.value}
	case taskFilterComparatorGreateEquals:
		cond = &builder.Gte{field: f.value}
	case taskFilterComparatorLess:
		cond = &builder.Lt{field: f.value}
	case taskFilterComparatorLessEquals:
		cond = &builder.Lte{field: f.value}
	case taskFilterComparatorLike:
		val, is := f.value.(string)
		if !is {
			return nil, ErrInvalidTaskFilterValue{Field: field, Value: f.value}
		}
		cond = &builder.Like{field, "%" + val + "%"}
	case taskFilterComparatorIn:
		cond = builder.In(field, f.value)
	case taskFilterComparatorNotIn:
		cond = builder.NotIn(field, f.value)
	case taskFilterComparatorInvalid:
		// Nothing to do
	}

	if includeNulls {
		cond = builder.Or(cond, &builder.IsNull{field})
		if f.isNumeric {
			cond = builder.Or(cond, &builder.IsNull{field}, &builder.Eq{field: 0})
		}
	}

	return
}

func getTaskIndexFromSearchString(s string) (index int64) {
	re := regexp.MustCompile("#([0-9]+)")
	in := re.FindString(s)

	stringIndex := strings.ReplaceAll(in, "#", "")
	index, _ = strconv.ParseInt(stringIndex, 10, 64)
	return
}

func getProjectIDsFromProjects(projects []*Project) (projectIDs []int64, hasFavoritesProject bool) {
	projectIDs = []int64{}
	for _, p := range projects {
		if p.ID == FavoritesPseudoProject.ID {
			hasFavoritesProject = true
			continue
		}
		projectIDs = append(projectIDs, p.ID)
	}
	return
}

func getRawTasksForProjects(s *xorm.Session, projects []*Project, a web.Auth, opts *taskSearchOptions) (tasks []*Task, resultCount int, totalItems int64, err error) {

	// If the user does not have any projects, don't try to get any tasks
	if len(projects) == 0 {
		return nil, 0, 0, nil
	}

	// Get all project IDs and get the tasks
	var hasFavoritesProject bool
	opts.projectIDs, hasFavoritesProject = getProjectIDsFromProjects(projects)

	// Add the id parameter as the last parameter to sortby by default, but only if it is not already passed as the last parameter.
	if len(opts.sortby) == 0 ||
		len(opts.sortby) > 0 && opts.sortby[len(opts.sortby)-1].sortBy != taskPropertyID {
		opts.sortby = append(opts.sortby, &sortParam{
			sortBy:  taskPropertyID,
			orderBy: orderAscending,
		})
	}

	opts.search = strings.TrimSpace(opts.search)

	var dbSearcher taskSearcher = &dbTaskSearcher{
		s:                   s,
		a:                   a,
		hasFavoritesProject: hasFavoritesProject,
	}
	tasks, totalItems, err = dbSearcher.Search(opts)

	return tasks, len(tasks), totalItems, err
}

func getTasksForProjects(s *xorm.Session, projects []*Project, a web.Auth, opts *taskSearchOptions, view *ProjectView) (tasks []*Task, resultCount int, totalItems int64, err error) {
	tasks, resultCount, totalItems, err = getRawTasksForProjects(s, projects, a, opts)
	if err != nil {
		return nil, 0, 0, err
	}

	taskMap := make(map[int64]*Task, len(tasks))
	for _, t := range tasks {
		taskMap[t.ID] = t
	}

	err = addMoreInfoToTasks(s, taskMap, a, view, opts.expand)
	if err != nil {
		return nil, 0, 0, err
	}

	return tasks, resultCount, totalItems, err
}

func taskMemoKey(id int64) string { return "task-" + strconv.FormatInt(id, 10) }

// lockingSession adds a row lock to the next read. xorm rejects FOR UPDATE on SQLite, which serializes writers anyway.
func lockingSession(s *xorm.Session) *xorm.Session {
	if dbType := db.Type(); dbType == schemas.MYSQL || dbType == schemas.POSTGRES {
		return s.ForUpdate()
	}
	return s
}

// GetTaskByIDSimple returns a raw task without extra data by the task ID
func GetTaskByIDSimple(s *xorm.Session, taskID int64) (task Task, err error) {
	if taskID < 1 {
		return Task{}, ErrTaskDoesNotExist{taskID}
	}

	return db.Remember(s, taskMemoKey(taskID), func() (Task, error) {
		return GetTaskSimple(s, &Task{ID: taskID})
	})
}

// GetTaskByProjectAndIndex returns a task by its per-project index.
// Returns ErrTaskDoesNotExist if nothing matches.
func GetTaskByProjectAndIndex(s *xorm.Session, projectID, index int64) (task Task, err error) {
	if projectID < 1 || index < 1 {
		return Task{}, ErrTaskDoesNotExist{}
	}

	has, err := s.
		Where("project_id = ? AND `index` = ?", projectID, index).
		Get(&task)
	if err != nil {
		return Task{}, err
	}
	if !has {
		return Task{}, ErrTaskDoesNotExist{}
	}

	return task, nil
}

// resolveIDFromProjectAndIndex populates t.ID from (ProjectID, Index) for the
// by-index route, which binds project+index from the URL but not id. No-op
// when id is already set.
func (t *Task) resolveIDFromProjectAndIndex(s *xorm.Session) error {
	if t.ID != 0 || t.ProjectID < 1 || t.Index < 1 {
		return nil
	}
	resolved, err := GetTaskByProjectAndIndex(s, t.ProjectID, t.Index)
	if err != nil {
		return err
	}
	t.ID = resolved.ID
	return nil
}

// GetTaskSimple returns a raw task without extra data
func GetTaskSimple(s *xorm.Session, t *Task) (task Task, err error) {
	task = *t
	exists, err := s.Get(&task)
	if err != nil {
		return Task{}, err
	}

	if !exists {
		return Task{}, ErrTaskDoesNotExist{t.ID}
	}
	return
}

func GetTasksSimpleByIDs(s *xorm.Session, ids []int64) (tasks []*Task, err error) {
	err = s.In("id", ids).Find(&tasks)
	return
}

func GetTaskSimpleByUUID(s *xorm.Session, uid string) (task *Task, err error) {
	var has bool
	task = &Task{}

	has, err = s.In("uid", uid).Get(task)
	if !has || err != nil {
		return &Task{}, ErrTaskDoesNotExist{}
	}

	return
}

// GetTasksByUIDs gets all tasks from a bunch of uids, filtering out any
// task whose project the provided auth does not have access to.
func GetTasksByUIDs(s *xorm.Session, uids []string, a web.Auth) (tasks []*Task, err error) {
	tasks = []*Task{}
	accessible, err := accessibleProjectIDsCond(s, a, "`tasks`.`project_id`")
	if err != nil {
		return nil, err
	}
	err = s.
		In("uid", uids).
		And(accessible).
		Find(&tasks)
	if err != nil {
		return
	}

	taskMap := make(map[int64]*Task, len(tasks))
	for _, t := range tasks {
		taskMap[t.ID] = t
	}

	err = addMoreInfoToTasks(s, taskMap, a, nil, nil)
	return
}

func getRemindersForTasks(s *xorm.Session, taskIDs []int64) (reminders []*TaskReminder, err error) {
	reminders = []*TaskReminder{}
	err = s.In("task_id", taskIDs).
		OrderBy("reminder asc").
		Find(&reminders)
	return
}

func (t *Task) setIdentifier(project *Project) {
	if project == nil || project.Identifier == "" {
		t.Identifier = "#" + strconv.FormatInt(t.Index, 10)
		return
	}

	t.Identifier = project.Identifier + "-" + strconv.FormatInt(t.Index, 10)
}

func addIsUnreadToTasks(s *xorm.Session, taskIDs []int64, taskMap map[int64]*Task, a web.Auth) (err error) {
	if len(taskIDs) == 0 {
		return nil
	}

	caller, isUser := a.(*user.User)
	if !isUser {
		return nil
	}

	unreadStatuses := []*TaskUnreadStatus{}
	err = s.In("task_id", taskIDs).
		Where("user_id = ?", caller.ID).
		Find(&unreadStatuses)
	if err != nil {
		return err
	}

	b := true
	for _, status := range unreadStatuses {
		if task, exists := taskMap[status.TaskID]; exists {
			task.IsUnread = &b
		}
	}

	return nil
}

// Get all assignees
func addAssigneesToTasks(s *xorm.Session, taskIDs []int64, taskMap map[int64]*Task) (err error) {
	taskAssignees, err := getRawTaskAssigneesForTasks(s, taskIDs)
	if err != nil {
		return
	}
	// Put the assignees in the task map
	for i, a := range taskAssignees {
		if a != nil {
			a.Email = "" // Obfuscate the email
			taskMap[a.TaskID].Assignees = append(taskMap[a.TaskID].Assignees, &taskAssignees[i].User)
		}
	}

	return
}

// Get all labels for all the tasks
func addLabelsToTasks(s *xorm.Session, taskIDs []int64, taskMap map[int64]*Task) (err error) {
	labels, _, _, err := GetLabelsByTaskIDs(s, taskIDs, nil, -1)
	if err != nil {
		return
	}
	for i, l := range labels {
		if l != nil {
			taskMap[l.TaskID].Labels = append(taskMap[l.TaskID].Labels, &labels[i].Label)
		}
	}

	return
}

// Get task attachments
func addAttachmentsToTasks(s *xorm.Session, taskIDs []int64, taskMap map[int64]*Task) (err error) {
	attachments, err := getTaskAttachmentsByTaskIDs(s, taskIDs)
	if err != nil {
		return
	}

	for _, a := range attachments {
		taskMap[a.TaskID].Attachments = append(taskMap[a.TaskID].Attachments, a)
	}
	return
}

func getTaskReminderMap(s *xorm.Session, taskIDs []int64) (taskReminders map[int64][]*TaskReminder, err error) {
	taskReminders = make(map[int64][]*TaskReminder)

	// Get all reminders and put them in a map to have it easier later
	reminders, err := getRemindersForTasks(s, taskIDs)
	if err != nil {
		return
	}

	for _, r := range reminders {
		taskReminders[r.TaskID] = append(taskReminders[r.TaskID], r)
	}

	return
}

func addRelatedTasksToTasks(s *xorm.Session, taskIDs []int64, taskMap map[int64]*Task, a web.Auth) (err error) {
	relatedTasks := []*TaskRelation{}
	err = s.In("task_id", taskIDs).Find(&relatedTasks)
	if err != nil {
		return
	}

	// Collect all related task IDs, so we can get all related task headers in one go
	var relatedTaskIDs []int64
	for _, rt := range relatedTasks {
		relatedTaskIDs = append(relatedTaskIDs, rt.OtherTaskID)
	}

	if len(relatedTaskIDs) == 0 {
		return
	}

	accessible, err := accessibleProjectIDsCond(s, a, "`tasks`.`project_id`")
	if err != nil {
		return err
	}
	fullRelatedTasks := make(map[int64]*Task)
	err = s.In("id", relatedTaskIDs).
		And(accessible).
		Find(&fullRelatedTasks)
	if err != nil {
		return
	}

	taskFavorites, err := getFavorites(s, relatedTaskIDs, a, FavoriteKindTask)
	if err != nil {
		return err
	}

	// NOTE: while it certainly be possible to run this function on	fullRelatedTasks again, we don't do this for performance reasons.

	// Go through all task relations and put them into the task objects
	for _, rt := range relatedTasks {
		_, has := fullRelatedTasks[rt.OtherTaskID]
		if !has {
			log.Debugf("Related task not found for task relation: taskID=%d, otherTaskID=%d, relationKind=%v", rt.TaskID, rt.OtherTaskID, rt.RelationKind)
			continue
		}
		fullRelatedTasks[rt.OtherTaskID].IsFavorite = taskFavorites[rt.OtherTaskID]

		// We're duplicating the other task to avoid cycles as these can't be represented properly in json
		// and would thus fail with an error.
		otherTask := &Task{}
		err = copier.Copy(otherTask, fullRelatedTasks[rt.OtherTaskID])
		if err != nil {
			log.Errorf("Could not duplicate task object: %v", err)
			continue
		}
		otherTask.RelatedTasks = nil
		taskMap[rt.TaskID].RelatedTasks[rt.RelationKind] = append(taskMap[rt.TaskID].RelatedTasks[rt.RelationKind], otherTask)
	}

	return
}

func addBucketsToTasks(s *xorm.Session, a web.Auth, taskIDs []int64, taskMap map[int64]*Task) (err error) {
	if len(taskIDs) == 0 {
		return nil
	}

	taskBuckets := []*TaskBucket{}
	err = s.
		In("task_id", taskIDs).
		Find(&taskBuckets)
	if err != nil {
		return err
	}

	accessible, err := accessibleProjectIDsCond(s, a, "project_views.project_id")
	if err != nil {
		return err
	}
	buckets := make(map[int64]*Bucket)
	err = s.
		Where(builder.In("id", builder.Select("bucket_id").
			From("task_buckets").
			Where(builder.In("task_id", taskIDs)))).
		And(builder.In("project_view_id", builder.Select("id").
			From("project_views").
			Where(accessible))).
		Find(&buckets)
	if err != nil {
		return err
	}

	for _, tb := range taskBuckets {
		if taskMap[tb.TaskID].Buckets == nil {
			taskMap[tb.TaskID].Buckets = []*Bucket{}
		}
		if bucket, exists := buckets[tb.BucketID]; exists {
			taskMap[tb.TaskID].Buckets = append(taskMap[tb.TaskID].Buckets, bucket)
		}
	}

	return nil
}

// This function takes a map with pointers and returns a slice with pointers to tasks
// It adds more stuff like assignees/labels/etc to a bunch of tasks
//
//nolint:gocyclo
func addMoreInfoToTasks(s *xorm.Session, taskMap map[int64]*Task, a web.Auth, view *ProjectView, expand []TaskCollectionExpandable) (err error) {

	// No need to iterate over users and stuff if the project doesn't have tasks
	if len(taskMap) == 0 {
		return
	}

	// Get all users & task ids and put them into the array
	var userIDs []int64
	var taskIDs []int64
	var projectIDs []int64
	for _, i := range taskMap {
		taskIDs = append(taskIDs, i.ID)
		if i.CreatedByID != 0 {
			userIDs = append(userIDs, i.CreatedByID)
		}
		projectIDs = append(projectIDs, i.ProjectID)
	}

	err = addAssigneesToTasks(s, taskIDs, taskMap)
	if err != nil {
		return
	}

	err = addLabelsToTasks(s, taskIDs, taskMap)
	if err != nil {
		return
	}

	err = addAttachmentsToTasks(s, taskIDs, taskMap)
	if err != nil {
		return
	}

	users, err := getUsersOrLinkSharesFromIDs(s, userIDs)
	if err != nil {
		return
	}

	taskReminders, err := getTaskReminderMap(s, taskIDs)
	if err != nil {
		return err
	}

	taskFavorites, err := getFavorites(s, taskIDs, a, FavoriteKindTask)
	if err != nil {
		return err
	}

	// Get all identifiers
	projects, err := GetProjectsMapByIDs(s, projectIDs)
	if err != nil {
		return err
	}

	var positionsMap = make(map[int64]*TaskPosition)
	if view != nil {
		positions, err := getPositionsForView(s, view)
		if err != nil {
			return err
		}
		for _, position := range positions {
			positionsMap[position.TaskID] = position
		}
	}

	var reactions map[int64]ReactionMap
	if expand != nil {
		expanded := make(map[TaskCollectionExpandable]bool)
		for _, expandable := range expand {
			if expanded[expandable] {
				continue
			}

			switch expandable {
			case TaskCollectionExpandSubtasks:
				// already dealt with earlier
			case TaskCollectionExpandBuckets:
				err = addBucketsToTasks(s, a, taskIDs, taskMap)
				if err != nil {
					return err
				}
			case TaskCollectionExpandReactions:
				reactions, err = getReactionsForEntityIDs(s, ReactionKindTask, taskIDs)
				if err != nil {
					return
				}
			case TaskCollectionExpandComments:
				err = addCommentsToTasks(s, taskIDs, taskMap)
				if err != nil {
					return err
				}
			case TaskCollectionExpandCommentCount:
				err = addCommentCountToTasks(s, taskIDs, taskMap)
				if err != nil {
					return err
				}
			case TaskCollectionExpandTimeEntriesCount:
				err = addTimeEntriesCountToTasks(s, a, taskIDs, taskMap)
				if err != nil {
					return err
				}
			case TaskCollectionExpandIsUnread:
				err = addIsUnreadToTasks(s, taskIDs, taskMap, a)
				if err != nil {
					return
				}
			}
			expanded[expandable] = true
		}
	}

	// Add all objects to their tasks
	for _, task := range taskMap {

		// Make created by user objects
		if createdBy, has := users[task.CreatedByID]; has {
			task.CreatedBy = createdBy
		}

		// Add the reminders
		task.Reminders = taskReminders[task.ID]

		// Prepare the subtasks
		task.RelatedTasks = make(RelatedTaskMap)

		// Build the task identifier from the project identifier and task index
		task.setIdentifier(projects[task.ProjectID])

		task.IsFavorite = taskFavorites[task.ID]

		// Populate structured repeat from RRULE string
		task.Repeat = taskRepeatFromRRule(task.Repeats)

		if reactions != nil {
			r, has := reactions[task.ID]
			if has {
				task.Reactions = r
			}
		}

		p, has := positionsMap[task.ID]
		if has {
			task.Position = p.Position
		}
	}

	// Get all related tasks
	err = addRelatedTasksToTasks(s, taskIDs, taskMap, a)
	return
}

// checkBucketLimit counts pendingInBatch tasks whose task_buckets rows aren't inserted yet, and returns taskCount even on overflow so callers can tell which task overflows.
func checkBucketLimit(s *xorm.Session, a web.Auth, t *Task, bucket *Bucket, view *ProjectView, pendingInBatch int64) (taskCount int64, err error) {
	if view.ProjectID < 0 || (view.Filter != nil && view.Filter.Filter != "") {
		// For saved filters or views with a filter, the count must be scoped to
		// this bucket *and* the filter: raw task_buckets rows can include tasks
		// that no longer match the filter (#355), while the unscoped filter total
		// counts tasks across all buckets, not just this one (#2672). ReadAll
		// combines the bucket_id condition with the saved-filter / view filter.
		tc := &TaskCollection{
			ProjectID:     view.ProjectID,
			ProjectViewID: bucket.ProjectViewID,
			Filter:        "bucket_id = " + strconv.FormatInt(bucket.ID, 10),
		}

		_, _, taskCount, err = tc.ReadAll(s, a, "", 1, 1)
		if err != nil {
			return 0, err
		}
	} else {
		taskCount, err = s.
			Where("bucket_id = ?", bucket.ID).
			GroupBy("task_id").
			Count(&TaskBucket{})
		if err != nil {
			return 0, err
		}
	}

	if bucket.Limit > 0 && taskCount+pendingInBatch >= bucket.Limit {
		return taskCount, ErrBucketLimitExceeded{TaskID: t.ID, BucketID: bucket.ID, Limit: bucket.Limit}
	}

	return
}

func calculateDefaultPosition(entityID int64, position float64) float64 {
	if position == 0 {
		return float64(entityID) * math.Pow(2, 16)
	}

	return position
}

func setNewTaskIndexes(s *xorm.Session, projectID int64, tasks []*Task) error {
	if len(tasks) == 0 {
		return nil
	}

	reserved := int64(len(tasks))
	lastIndex, err := reserveTaskIndexes(s, projectID, reserved)
	if err != nil {
		return err
	}
	previous := lastIndex - reserved

	// Presets above the old high water mark cannot collide with anything, everything else gets a fresh index.
	used := make(map[int64]bool, len(tasks))
	highest := lastIndex
	for _, t := range tasks {
		// Bounding imported presets keeps the counter far away from int64 overflow.
		if t.Index <= previous || t.Index > math.MaxInt32 || used[t.Index] {
			t.Index = 0
			continue
		}
		used[t.Index] = true
		if t.Index > highest {
			highest = t.Index
		}
	}

	next := previous
	for _, t := range tasks {
		if t.Index != 0 {
			continue
		}
		next++
		for used[next] {
			next++
		}
		t.Index = next
	}

	if highest > lastIndex {
		_, err = s.ID(projectID).
			Cols("last_index").
			Update(&ProjectTaskCounter{LastIndex: highest})
	}

	return err
}

// Create is the implementation to create a project task
// @Summary Create a task
// @Description Inserts a task into a project.
// @tags task
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "Project ID"
// @Param task body models.Task true "The task object"
// @Success 201 {object} models.Task "The created task object."
// @Failure 400 {object} web.HTTPError "Invalid task object provided."
// @Failure 403 {object} web.HTTPError "The user does not have access to the project"
// @Failure 500 {object} models.Message "Internal error"
// @Router /projects/{id}/tasks [put]
func (t *Task) Create(s *xorm.Session, a web.Auth) (err error) {
	return createTask(s, t, a, true, true)
}

func createTask(s *xorm.Session, t *Task, a web.Auth, updateAssignees bool, setBucket bool) (err error) {
	return unwrapBulkCreateError(createTasks(s, t.ProjectID, []*Task{t}, a, updateAssignees, setBucket, false))
}

// CreateTasksForImport preserves preset indexes across the whole imported batch.
func CreateTasksForImport(s *xorm.Session, projectID int64, tasks []*Task, a web.Auth) error {
	return unwrapBulkCreateError(createTasks(s, projectID, tasks, a, true, true, true))
}

// unwrapBulkCreateError returns the raw error type callers of a single logical create expect, not the batch wrapper.
func unwrapBulkCreateError(err error) error {
	var berr ErrInvalidTaskInBulkCreation
	if errors.As(err, &berr) {
		return berr.Err
	}
	return err
}

// resolveProvidedBuckets maps task id → explicitly requested bucket, checking each bucket's limit once with the batch's own members added since their rows are inserted later.
func resolveProvidedBuckets(s *xorm.Session, a web.Auth, projectID int64, tasks []*Task) (map[int64]*Bucket, error) {
	bucketOrder := make([]int64, 0, len(tasks))
	// bucket id → payload indexes of the tasks targeting it, in payload order.
	batches := make(map[int64][]int)
	for i, t := range tasks {
		if t.BucketID == 0 {
			continue
		}
		if _, has := batches[t.BucketID]; !has {
			bucketOrder = append(bucketOrder, t.BucketID)
		}
		batches[t.BucketID] = append(batches[t.BucketID], i)
	}

	buckets := make(map[int64]*Bucket, len(bucketOrder))
	for _, bucketID := range bucketOrder {
		bucket, err := getBucketByID(s, bucketID)
		if err != nil {
			return nil, err
		}

		view, err := GetProjectViewByID(s, bucket.ProjectViewID)
		if err != nil {
			// Deleted views leave orphaned buckets behind; reporting the missing view would disclose they exist.
			if IsErrProjectViewDoesNotExist(err) {
				return nil, ErrBucketDoesNotExist{BucketID: bucketID}
			}
			return nil, err
		}
		if view.ProjectID != projectID {
			return nil, ErrBucketDoesNotExist{BucketID: bucketID}
		}

		members := batches[bucketID]
		existing, err := checkBucketLimit(s, a, tasks[members[0]], bucket, view, int64(len(members))-1)
		if err != nil {
			var limitErr ErrBucketLimitExceeded
			if errors.As(err, &limitErr) {
				// Name the first member which no longer fits, not the batch's first one.
				overflowing := max(min(int(bucket.Limit-existing), len(members)-1), 0)
				limitErr.TaskID = tasks[members[overflowing]].ID
				return nil, ErrInvalidTaskInBulkCreation{Index: members[overflowing], Err: limitErr}
			}
			return nil, err
		}
		buckets[bucketID] = bucket
	}

	taskProvidedBucket := make(map[int64]*Bucket, len(tasks))
	for _, t := range tasks {
		if t.BucketID == 0 {
			continue
		}
		taskProvidedBucket[t.ID] = buckets[t.BucketID]
	}
	return taskProvidedBucket, nil
}

// createTasks inserts row by row because multi-row inserts don't reliably return autoincrement ids on all supported databases.
func createTasks(s *xorm.Session, projectID int64, tasks []*Task, a web.Auth, updateAssignees bool, setBucket, preserveIndexes bool) (err error) {
	if len(tasks) == 0 {
		return nil
	}

	for i, t := range tasks {
		err = validateTaskForCreation(t)
		if err != nil {
			return ErrInvalidTaskInBulkCreation{Index: i, Err: err}
		}

		// A raw rule set by a trusted caller (duplication, importers, CalDAV) is kept verbatim.
		decision, err := resolveCreateRepeat(t)
		if err != nil {
			return ErrInvalidTaskInBulkCreation{Index: i, Err: err}
		}
		t.Repeats = decision.Repeats

		t.ProjectID = projectID
		t.ID = 0
		if !preserveIndexes {
			t.Index = 0
		}
	}

	// Check if the project exists
	p, err := GetProjectSimpleByID(s, projectID)
	if err != nil {
		return err
	}

	createdBy, err := GetUserOrLinkShareUser(s, a)
	if err != nil {
		return err
	}

	err = setNewTaskIndexes(s, projectID, tasks)
	if err != nil {
		return err
	}

	for _, t := range tasks {
		t.CreatedByID = createdBy.ID

		// Generate a uuid if we don't already have one
		if t.UID == "" {
			t.UID = uuid.NewString()
		}

		t.HexColor = utils.NormalizeHex(t.HexColor)

		_, err = s.Insert(t)
		if err != nil {
			return err
		}
	}

	taskProvidedBucket, err := resolveProvidedBuckets(s, a, projectID, tasks)
	if err != nil {
		return err
	}

	views, err := lockProjectViewsForPositionUpdate(s, projectID)
	if err != nil {
		return err
	}

	positions, taskBuckets, err := setTasksInBucketInViews(s, views, tasks, a, setBucket, taskProvidedBucket)
	if err != nil {
		return err
	}

	if len(positions) > 0 {
		positions, err = filterNewTaskPositions(s, positions)
		if err != nil {
			return err
		}
	}

	if len(positions) > 0 {
		err = bulkInsertTaskPositions(s, positions, false)
		if err != nil {
			return
		}

		err = resolvePositionConflictsAfterInsert(s, positions)
		if err != nil {
			return
		}
	}

	// Keep statements well below the parameter limits of all supported databases.
	const taskBucketBatchSize = 100
	for chunk := range slices.Chunk(taskBuckets, taskBucketBatchSize) {
		_, err = s.Insert(chunk)
		if err != nil {
			return
		}
	}

	// Link shares can't have subscriptions
	_, creatorIsUser := a.(*user.User)

	for _, t := range tasks {
		t.CreatedBy = createdBy

		// Update the assignees
		if updateAssignees {
			if err := t.updateTaskAssignees(s, t.Assignees, a); err != nil {
				return err
			}
		}

		// Update the reminders
		if err := t.updateReminders(s, t, reminderZone(s, a, t, nil)); err != nil {
			return err
		}

		t.setIdentifier(p)

		if t.IsFavorite {
			if err := addToFavorites(s, t.ID, createdBy, FavoriteKindTask); err != nil {
				return err
			}
		}

		if creatorIsUser {
			if err := subscribeUserImplicitly(s, SubscriptionEntityTask, t.ID, createdBy); err != nil {
				return err
			}
		}

		events.DispatchOnCommit(s, &TaskCreatedEvent{
			Task: t,
			Doer: createdBy,
		})

		// Populate structured repeat for API response
		t.Repeat = taskRepeatFromRRule(t.Repeats)
	}

	events.DispatchOnCommit(s, &TasksBatchCreatedEvent{
		Tasks: tasks,
		Doer:  createdBy,
	})

	err = updateProjectLastUpdated(s, &Project{ID: projectID})
	return
}

func setTasksInBucketInViews(s *xorm.Session, views []*ProjectView, tasks []*Task, a web.Auth, setBucket bool, providedBuckets map[int64]*Bucket) ([]*TaskPosition, []*TaskBucket, error) {
	if len(tasks) == 0 {
		return nil, nil, nil
	}

	var err error
	taskBuckets := []*TaskBucket{}

	defaultBucketIDs := make(map[int64]int64)
	cachedDefaultBucketID := func(view *ProjectView) (int64, error) {
		if id, has := defaultBucketIDs[view.ID]; has {
			return id, nil
		}
		id, err := getDefaultBucketID(s, view)
		if err != nil {
			return 0, err
		}
		defaultBucketIDs[view.ID] = id
		return id, nil
	}

	for _, t := range tasks {
		var moveToDone bool
		taskBucketsForTask := []*TaskBucket{}
		providedBucket := providedBuckets[t.ID]

		for _, view := range views {
			if setBucket && !moveToDone &&
				view.ViewKind == ProjectViewKindKanban &&
				view.BucketConfigurationMode == BucketConfigurationModeManual {

				bucketID := view.DoneBucketID
				if !t.Done || view.DoneBucketID == 0 {
					if providedBucket != nil && view.ID == providedBucket.ProjectViewID {
						bucketID = providedBucket.ID
					} else {
						bucketID, err = cachedDefaultBucketID(view)
						if err != nil {
							return nil, nil, err
						}
					}
				}

				if view.DoneBucketID != 0 && view.DoneBucketID == t.BucketID && !t.Done {
					t.Done = true
					_, err = s.Where("id = ?", t.ID).
						Cols("done").
						Update(t)
					if err != nil {
						return nil, nil, err
					}

					err = t.moveTaskToDoneBuckets(s, a, views)
					if err != nil {
						return nil, nil, err
					}

					moveToDone = true

					continue
				}

				taskBucketsForTask = append(taskBucketsForTask, &TaskBucket{
					BucketID:      bucketID,
					TaskID:        t.ID,
					ProjectViewID: view.ID,
				})
			}
		}

		if !moveToDone {
			taskBuckets = append(taskBuckets, taskBucketsForTask...)
		}
	}

	positions := []*TaskPosition{}
	for _, view := range views {
		viewPositions, err := calculateNewPositionsForTasks(s, a, tasks, view)
		if err != nil {
			return nil, nil, err
		}
		positions = append(positions, viewPositions...)
	}

	return positions, taskBuckets, nil
}

// Update updates a project task
// @Summary Update a task
// @Description Updates a task. This includes marking it as done. Assignees you pass will be updated, see their individual endpoints for more details on how this is done. To update labels, see the description of the endpoint.
// @tags task
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "The Task ID"
// @Param task body models.Task true "The task object"
// @Success 200 {object} models.Task "The updated task object."
// @Failure 400 {object} web.HTTPError "Invalid task object provided."
// @Failure 403 {object} web.HTTPError "The user does not have access to the task (aka its project)"
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{id} [post]
// Update updates a project task by delegating to the shared bulk helper.
func (t *Task) Update(s *xorm.Session, a web.Auth) (err error) {
	return t.updateSingleTask(s, a, nil)
}

//nolint:gocyclo
func (t *Task) updateSingleTask(s *xorm.Session, a web.Auth, fields []string) (err error) {

	// Row lock so concurrent movers serialize; bypasses the session memo so the read is not the earlier permission-check copy.
	ot, err := GetTaskSimple(lockingSession(s), &Task{ID: t.ID})
	if err != nil {
		return
	}

	if t.ProjectID == 0 {
		t.ProjectID = ot.ProjectID
	}

	// Get the stored reminders
	reminders, err := getRemindersForTasks(s, []int64{t.ID})
	if err != nil {
		return
	}

	// Old task has the stored reminders
	ot.Reminders = reminders

	// All columns to update in a separate variable to be able to add to them
	colsToUpdate := []string{
		"title",
		"description",
		"done",
		"due_date",
		"repeats",
		"repeats_from_current_date",
		"priority",
		"start_date",
		"end_date",
		"hex_color",
		"percent_done",
		"project_id",
		"bucket_id",
		"cover_image_attachment_id",
	}

	// Validate fields if provided
	fieldSet := map[string]bool{}
	if len(fields) > 0 {
		allowed := map[string]bool{}
		for _, c := range colsToUpdate {
			allowed[c] = true
		}
		// Not task columns; written separately below.
		relationFields := map[string]bool{
			"assignees":   true,
			"reminders":   true,
			"is_favorite": true,
		}
		cols := []string{}
		for _, f := range fields {
			// `repeat` is the public field name for the RRULE; the stored column is `repeats`.
			if f == "repeat" {
				f = "repeats"
			}
			if !allowed[f] && !relationFields[f] {
				return ErrInvalidTaskColumn{Column: f}
			}
			if !relationFields[f] {
				cols = append(cols, f)
			}
			fieldSet[f] = true
		}
		colsToUpdate = cols

		if !fieldSet["title"] {
			t.Title = ot.Title
		}
		if !fieldSet["description"] {
			t.Description = ot.Description
		}
		if !fieldSet["done"] {
			t.Done = ot.Done
			t.DoneAt = ot.DoneAt
		}
		if !fieldSet["due_date"] {
			t.DueDate = ot.DueDate
		}
		if !fieldSet["repeats"] {
			t.Repeats = ot.Repeats
		}
		if !fieldSet["repeats_from_current_date"] {
			t.RepeatsFromCurrentDate = ot.RepeatsFromCurrentDate
		}
		if !fieldSet["priority"] {
			t.Priority = ot.Priority
		}
		if !fieldSet["start_date"] {
			t.StartDate = ot.StartDate
		}
		if !fieldSet["end_date"] {
			t.EndDate = ot.EndDate
		}
		if !fieldSet["hex_color"] {
			t.HexColor = ot.HexColor
		}
		if !fieldSet["percent_done"] {
			t.PercentDone = ot.PercentDone
		}
		if !fieldSet["project_id"] {
			t.ProjectID = ot.ProjectID
		}
		if !fieldSet["bucket_id"] {
			t.BucketID = ot.BucketID
		}
		if !fieldSet["cover_image_attachment_id"] {
			t.CoverImageAttachmentID = ot.CoverImageAttachmentID
		}
		if !fieldSet["reminders"] {
			t.Reminders = ot.Reminders
		}
	}

	if len(fields) == 0 || fieldSet["assignees"] {
		if err := ot.updateTaskAssignees(s, t.Assignees, a); err != nil {
			return err
		}
	}

	// Recurrence write rules. A field-masked update whose mask excludes
	// recurrence keeps the stored rule, copied above, and never re-validates it.
	if len(fields) == 0 || fieldSet["repeats"] {
		decision, err := resolveIncomingRepeat(ot.Repeats, t.Repeat, t.Repeats)
		if err != nil {
			return err
		}
		t.Repeats = decision.Repeats
	}

	// If the task is being moved between projects, make sure to move the bucket + index as well
	if t.ProjectID != 0 && ot.ProjectID != t.ProjectID {
		// Another task may already hold this retired address; the latest holder wins.
		alias := &TaskIndexAlias{ProjectID: ot.ProjectID, Index: ot.Index}
		has, err := s.Get(alias)
		if err != nil {
			return err
		}
		switch {
		case !has:
			_, err = s.Insert(&TaskIndexAlias{ProjectID: ot.ProjectID, Index: ot.Index, TaskID: ot.ID})
		case alias.TaskID != ot.ID:
			_, err = s.Where("project_id = ? AND `index` = ?", ot.ProjectID, ot.Index).
				Cols("task_id").
				Update(&TaskIndexAlias{TaskID: ot.ID})
		}
		if err != nil {
			return err
		}

		t.Index = 0
		err = setNewTaskIndexes(s, t.ProjectID, []*Task{t})
		if err != nil {
			return err
		}
		t.BucketID = 0
		colsToUpdate = append(colsToUpdate, "index")
	}

	views := []*ProjectView{}
	if t.Done != ot.Done || t.ProjectID != ot.ProjectID {
		// Locks both projects: a move rewrites task_positions in the old project too.
		_, err = lockProjectViewsForPositionUpdate(s, ot.ProjectID, t.ProjectID)
		if err != nil {
			return
		}

		err = s.
			Where("project_id = ? AND view_kind = ? AND bucket_configuration_mode = ?",
				t.ProjectID, ProjectViewKindKanban, BucketConfigurationModeManual).
			Find(&views)
		if err != nil {
			return
		}
	}

	// When a task was moved between projects, ensure it is in the correct bucket
	if t.ProjectID != ot.ProjectID {
		_, err = s.Where("task_id = ?", t.ID).Delete(&TaskBucket{})
		if err != nil {
			return err
		}
		_, err = s.Where("task_id = ?", t.ID).Delete(&TaskPosition{})
		if err != nil {
			return err
		}

		for _, view := range views {
			var bucketID = view.DoneBucketID
			if bucketID == 0 || !t.Done {
				bucketID, err = getDefaultBucketID(s, view)
				if err != nil {
					return err
				}
			}

			tb := &TaskBucket{
				BucketID:      bucketID,
				TaskID:        t.ID,
				ProjectViewID: view.ID,
				ProjectID:     t.ProjectID,
			}
			err = updateTaskBucket(s, a, tb)
			if err != nil {
				return err
			}

			tp, err := calculateNewPositionForTask(s, a, t, view)
			if err != nil {
				return err
			}

			err = updateTaskPosition(s, a, tp)
			if err != nil {
				return err
			}
		}
	}

	// When a task changed its done status, make sure it is in the correct bucket
	if t.ProjectID == ot.ProjectID && !t.isRepeating() && t.Done != ot.Done {
		err = t.moveTaskToDoneBuckets(s, a, views)
		if err != nil {
			return
		}
	}

	// Repeating tasks don't stay in the done bucket — route them back
	// to the default bucket so the next iteration shows up in the
	// "To-Do" column. See #2573.
	if t.ProjectID == ot.ProjectID && t.isRepeating() && !ot.Done && t.Done {
		err = t.moveTaskToDefaultBuckets(s, a, views)
		if err != nil {
			return
		}
	}

	// Completion steps the rule this update stores. When the same update replaces the rule with a
	// new one together with done: true, the new rule drives the date step and its COUNT is
	// decremented; stepping the old one would write the old rule back over the replacement. A
	// stale echo keeps the stored rule (isStaleCount), so it still completes from the stored COUNT.
	//
	// An update that clears the rule while completing the task still steps the old rule and
	// reopens the task, but the rule stays cleared. v1's full-replace update clears the rule
	// whenever repeat is omitted, so a v1 {"done": true} on a repeating task lands here,
	// and upstream has always reopened it for its next occurrence (webtest "Repeating task
	// update done"); writing the stepped old rule back would undo the clear.
	//
	// An update that adds a rule to a task that had none, together with done: true, leaves the task
	// done with the new rule stored, as upstream did (no step, no COUNT decrement). A replacement
	// also brings its from-current-date flag.
	completedFrom := ot
	clearsRule := t.Repeats == "" && ot.Repeats != ""
	replacesRule := ot.Repeats != "" && t.Repeats != "" && t.Repeats != ot.Repeats
	if replacesRule {
		completedFrom.Repeats = t.Repeats
		completedFrom.RepeatsFromCurrentDate = t.RepeatsFromCurrentDate
	}

	// Captured before updateDone reopens the completed repeating task below.
	repeatingCompleted := completedFrom.isRepeating() && !ot.Done && t.Done
	preRepeatDueDate, preRepeatStartDate, preRepeatEndDate := t.DueDate, t.StartDate, t.EndDate
	preRepeatDescription := t.Description

	// When a repeating task is marked as done, we update all deadlines and reminders and set it as undone.
	// The zone is resolved only for that transition, in the project after this update.
	zone := config.GetTimeZone()
	if repeatingCompleted {
		zone = resolveTaskZone(s, a, t.ProjectID)
	}
	updateDoneAt := updateDone(&completedFrom, t, zone)
	if clearsRule {
		t.Repeats = ""
		t.Repeat = nil
	}
	if updateDoneAt {
		colsToUpdate = append(colsToUpdate, "done_at")
	}

	// updateDone reschedules after colsToUpdate was frozen from the caller's field list,
	// so whatever it rewrote has to be added back or it gets computed and thrown away.
	if len(fields) > 0 {
		if !fieldSet["due_date"] && !t.DueDate.Equal(preRepeatDueDate) {
			colsToUpdate = append(colsToUpdate, "due_date")
		}
		if !fieldSet["start_date"] && !t.StartDate.Equal(preRepeatStartDate) {
			colsToUpdate = append(colsToUpdate, "start_date")
		}
		if !fieldSet["end_date"] && !t.EndDate.Equal(preRepeatEndDate) {
			colsToUpdate = append(colsToUpdate, "end_date")
		}
		if !fieldSet["description"] && t.Description != preRepeatDescription {
			colsToUpdate = append(colsToUpdate, "description")
		}
	}

	// Completing a repeating task also rewrites its rule (a COUNT decrement) and reopens it;
	// persist those too, whatever the caller's field list was.
	if repeatingCompleted {
		for _, c := range []string{"repeats", "done", "done_at"} {
			if !slices.Contains(colsToUpdate, c) {
				colsToUpdate = append(colsToUpdate, c)
			}
		}
	}

	// Update the reminders, in the completion's zone when this save completed a repeating task.
	var completionZone *time.Location
	if repeatingCompleted {
		completionZone = zone
	}
	if err := ot.updateReminders(s, t, reminderZone(s, a, t, completionZone)); err != nil {
		return err
	}

	// If a task attachment is being set as cover image, check if the attachment actually belongs to the task
	if t.CoverImageAttachmentID != 0 {
		is, err := s.Exist(&TaskAttachment{
			TaskID: t.ID,
			ID:     t.CoverImageAttachmentID,
		})
		if err != nil {
			return err
		}
		if !is {
			return &ErrAttachmentDoesNotBelongToTask{
				AttachmentID: t.CoverImageAttachmentID,
				TaskID:       t.ID,
			}
		}
	}

	wasFavorite, err := isFavorite(s, t.ID, a, FavoriteKindTask)
	if err != nil {
		return
	}
	if len(fields) > 0 && !fieldSet["is_favorite"] {
		t.IsFavorite = wasFavorite
	}
	if t.IsFavorite && !wasFavorite {
		if err := addToFavorites(s, t.ID, a, FavoriteKindTask); err != nil {
			return err
		}
	}

	if !t.IsFavorite && wasFavorite {
		if err := removeFromFavorite(s, t.ID, a, FavoriteKindTask); err != nil {
			return err
		}
	}

	// Update the labels
	//
	// Maybe FIXME:
	// I've disabled this for now, because it requires significant changes in the way we do updates (using the
	// Update() function. We need a user object in updateTaskLabels to check if the user has the permission to see
	// the label it is currently adding. To do this, we'll need to update the webhandler to let it pass the current
	// user object (like it's already the case with the create method). However when we change it, that'll break
	// a lot of existing code which we'll then need to refactor.
	// This is why.
	//
	// if err := ot.updateTaskLabels(t.Labels); err != nil {
	// 	return err
	// }
	// set the labels to ot.Labels because our updateTaskLabels function puts the full label objects in it pretty nicely
	// We also set this here to prevent it being overwritten later on.
	// t.Labels = ot.Labels

	// For whatever reason, xorm dont detect if done is updated, so we need to update this every time by hand
	// Which is why we merge the actual task struct with the one we got from the db
	// The user struct overrides values in the actual one.
	// Must happen before the merge, ot is what gets saved.
	t.HexColor = utils.NormalizeHex(t.HexColor)

	if err := mergo.Merge(&ot, t, mergo.WithOverride); err != nil {
		return err
	}

	//////
	// Mergo does ignore nil values. Because of that, we need to check all parameters and set the updated to
	// nil/their nil value in the struct which is inserted.
	////
	// Done
	if !t.Done {
		ot.Done = false
	}
	// Priority
	if t.Priority == 0 {
		ot.Priority = 0
	}
	// Description
	if t.Description == "" {
		ot.Description = ""
	}
	// Due date
	if t.DueDate.IsZero() {
		ot.DueDate = time.Time{}
	}
	// RRULE repeats - the rule decided above (and possibly advanced by updateDone) is
	// authoritative; mergo skips an empty one, so a cleared rule is set explicitly.
	ot.Repeats = t.Repeats
	// Repeats from current date
	if !t.RepeatsFromCurrentDate {
		ot.RepeatsFromCurrentDate = false
	}
	// Start date
	if t.StartDate.IsZero() {
		ot.StartDate = time.Time{}
	}
	// End date
	if t.EndDate.IsZero() {
		ot.EndDate = time.Time{}
	}
	// Color
	if t.HexColor == "" {
		ot.HexColor = ""
	}
	// Percent Done
	if t.PercentDone == 0 {
		ot.PercentDone = 0
	}
	// Is Favorite
	if !t.IsFavorite {
		ot.IsFavorite = false
	}
	// Attachment cover image
	if t.CoverImageAttachmentID == 0 {
		ot.CoverImageAttachmentID = 0
	}

	_, err = s.ID(t.ID).
		Cols(colsToUpdate...).
		Update(&ot)
	*t = ot
	if err != nil {
		return err
	}

	// Get the task updated timestamp in a new struct - if we'd just try to put it into t which we already have, it
	// would still contain the old updated date.
	nt := &Task{}
	_, err = s.ID(t.ID).Get(nt)
	if err != nil {
		return err
	}
	t.Updated = nt.Updated

	// Populate structured repeat for API response
	t.Repeat = taskRepeatFromRRule(t.Repeats)

	events.DispatchOnCommit(s, &TaskUpdatedEvent{
		Task: t,
		Doer: doerFromAuth(s, a),
	})

	return updateProjectLastUpdated(s, &Project{ID: t.ProjectID})
}

// updateTasks updates multiple tasks with the same payload.
// If fields is nil, it updates the default set of columns.
func updateTasks(s *xorm.Session, a web.Auth, t *Task, ids []int64, fields []string) (tasks []*Task, err error) {
	seen := make(map[int64]struct{}, len(ids))
	uniqueIDs := make([]int64, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		uniqueIDs = append(uniqueIDs, id)
	}
	ids = uniqueIDs

	// Ids may span projects; per-task locking would follow caller order, not view order.
	existing := []*Task{}
	err = s.Table(&Task{}).Where(builder.In("id", ids)).Distinct("project_id").Find(&existing)
	if err != nil {
		return nil, fmt.Errorf("could not load projects of tasks to update: %w", err)
	}

	projectIDs := make([]int64, 0, len(existing)+1)
	for _, et := range existing {
		projectIDs = append(projectIDs, et.ProjectID)
	}
	if t.ProjectID != 0 {
		projectIDs = append(projectIDs, t.ProjectID)
	}

	_, err = lockProjectViewsForPositionUpdate(s, projectIDs...)
	if err != nil {
		return nil, fmt.Errorf("could not lock project views for bulk task update: %w", err)
	}

	for _, id := range ids {
		nt := clone.Clone(t)
		nt.ID = id
		if err := nt.updateSingleTask(s, a, fields); err != nil {
			return []*Task{}, err
		}
		tasks = append(tasks, nt)
	}
	return tasks, nil
}

func (t *Task) moveTaskToDoneBuckets(s *xorm.Session, a web.Auth, views []*ProjectView) error {
	for _, view := range views {
		currentTaskBucket := &TaskBucket{}
		_, err := s.Where("task_id = ? AND project_view_id = ?", t.ID, view.ID).
			Get(currentTaskBucket)
		if err != nil {
			return err
		}

		var bucketID = currentTaskBucket.BucketID

		// Task done, but no done bucket? Do nothing
		if t.Done && view.DoneBucketID == 0 {
			continue
		}

		// Task not done, currently not in done bucket? Do nothing
		if !t.Done && bucketID != view.DoneBucketID {
			continue
		}

		// Task done? Done bucket
		if t.Done && view.DoneBucketID != 0 {
			bucketID = view.DoneBucketID
		}

		// Task not done, currently in done bucket? Move to default
		if !t.Done && bucketID == view.DoneBucketID {
			bucketID, err = getDefaultBucketID(s, view)
			if err != nil {
				return err
			}
		}

		tb := &TaskBucket{
			BucketID:      bucketID,
			TaskID:        t.ID,
			ProjectViewID: view.ID,
			ProjectID:     t.ProjectID,
		}
		err = updateTaskBucket(s, a, tb)
		if err != nil {
			return err
		}

		tp := TaskPosition{
			TaskID:        t.ID,
			ProjectViewID: view.ID,
			Position:      calculateDefaultPosition(t.Index, t.Position),
		}
		err = updateTaskPosition(s, a, &tp)
		if err != nil {
			return err
		}
	}
	return nil
}

// moveTaskToDefaultBuckets moves the task to the default bucket of
// every provided view. It's the counterpart to moveTaskToDoneBuckets
// and is used when a repeating task is marked done: repeating tasks
// don't stay in the done bucket, so they should be routed back to
// the default ("To-Do") bucket so the next iteration is visible there.
// When no explicit default bucket is configured, the task stays in its
// current bucket — no update needed.
func (t *Task) moveTaskToDefaultBuckets(s *xorm.Session, a web.Auth, views []*ProjectView) error {
	for _, view := range views {
		if view.DefaultBucketID != 0 {
			tb := &TaskBucket{
				BucketID:      view.DefaultBucketID,
				TaskID:        t.ID,
				ProjectViewID: view.ID,
				ProjectID:     t.ProjectID,
			}
			if err := updateTaskBucket(s, a, tb); err != nil {
				return err
			}
		}

		tp := TaskPosition{
			TaskID:        t.ID,
			ProjectViewID: view.ID,
			Position:      calculateDefaultPosition(t.Index, t.Position),
		}
		if err := updateTaskPosition(s, a, &tp); err != nil {
			return err
		}
	}
	return nil
}

// validateRRule checks that an RRULE string is parseable, within the safety bound and allowed
// by the creation policy. Empty strings are valid (no recurrence). Returns ErrInvalidData
// otherwise. The bound is checked on the parsed options, before rrule-go builds the rule, since
// building it is what allocates (checkRRuleBounds).
func validateRRule(rruleStr string) error {
	opt, err := validateRRuleSyntax(rruleStr)
	if err != nil || opt == nil {
		return err
	}
	if rrulePolicyRejects(opt) {
		return ErrInvalidData{Message: "Invalid RRULE: sub-daily frequencies cannot be combined with BY* components"}
	}
	return nil
}

// validateRRuleSyntax checks that an RRULE string parses and is within the safety bound,
// without the creation policy. It returns the parsed options, nil for "".
func validateRRuleSyntax(rruleStr string) (*rrule.ROption, error) {
	if rruleStr == "" {
		return nil, nil
	}
	if len(rruleStr) > maxRRuleLength {
		return nil, ErrInvalidData{Message: fmt.Sprintf("Invalid RRULE: the rule is %d characters long; at most %d characters are allowed", len(rruleStr), maxRRuleLength)}
	}
	opt, err := rrule.StrToROption(rruleStr)
	if err != nil {
		return nil, ErrInvalidData{Message: "Invalid RRULE: " + err.Error()}
	}
	// rrule-go's serialisation can be longer than the text (a date-only UNTIL gains a time), and
	// completing a COUNT rule stores it serialised, so the serialised form has to fit as well.
	if serialised := opt.RRuleString(); len(serialised) > maxRRuleLength {
		return nil, ErrInvalidData{Message: fmt.Sprintf("Invalid RRULE: the rule is %d characters long once normalised; at most %d characters are allowed", len(serialised), maxRRuleLength)}
	}
	if err := checkRRuleWriteLimits(opt); err != nil {
		return nil, err
	}
	if _, err := rrule.NewRRule(*opt); err != nil {
		return nil, ErrInvalidData{Message: "Invalid RRULE: " + err.Error()}
	}
	return opt, nil
}

// fixedDurationForRRule returns the elapsed step of a plain SECONDLY, MINUTELY or HOURLY rule (no
// BY part, COUNT or UNTIL). DAILY and WEEKLY rules step on the local wall clock instead, and
// months and years were never a fixed number of seconds, so every other rule returns false. A step that does not fit time.Duration is not
// fixed either.
func fixedDurationForRRule(opt *rrule.ROption) (time.Duration, bool) {
	if !plainRule(opt) {
		return 0, false
	}

	interval := opt.Interval
	if interval <= 0 {
		interval = 1
	}

	var base time.Duration
	switch opt.Freq {
	case rrule.SECONDLY:
		base = time.Second
	case rrule.MINUTELY:
		base = time.Minute
	case rrule.HOURLY:
		base = time.Hour
	case rrule.DAILY, rrule.WEEKLY, rrule.MONTHLY, rrule.YEARLY:
		return 0, false
	default:
		return 0, false
	}

	if int64(base) > math.MaxInt64/int64(interval) {
		return 0, false
	}

	return base * time.Duration(interval), true
}

// addRepeatIntervalToTime advances t by whole multiples of duration until
// it is strictly after now.
func addRepeatIntervalToTime(now, t time.Time, duration time.Duration) time.Time {
	if duration <= 0 {
		return t
	}

	if !t.Before(now) {
		return t.Add(duration)
	}

	diff := now.Sub(t)
	if diff == math.MaxInt64 {
		// The gap exceeds time.Duration's range (~292 years). Count whole steps in seconds instead
		// of rebasing on now, so the result stays aligned to t.
		if duration%time.Second == 0 {
			step := int64(duration / time.Second)
			next := time.Unix(t.Unix()+(now.Unix()-t.Unix())/step*step, int64(t.Nanosecond())).In(t.Location())
			for !next.After(now) {
				next = next.Add(duration)
			}
			return next
		}
		return now.Add(duration)
	}

	// The whole steps that fit the gap fit time.Duration, since the gap does; the final step is a
	// separate addition, so the total never overflows and stays aligned to t.
	return t.Add(diff / duration * duration).Add(duration)
}

// clampRRuleAnchor fast-forwards a DAILY/WEEKLY DTSTART that sits more than a year
// before now by a whole number of periods, landing it within one period below the
// one-year cutoff. Whole-period AddDate arithmetic preserves the rule's phase, so
// iterating past the cutoff yields the same occurrences with far fewer iterations.
// Other frequencies (and recent anchors) are returned unchanged. The engine passes
// floating values, on which daysBetween is exact and the day-length
// corrections below never fire; they stay as guards. A rule with COUNT is never
// clamped: COUNT counts from DTSTART.
func clampRRuleAnchor(opt *rrule.ROption, baseDate, now time.Time) time.Time {
	if opt.Freq != rrule.DAILY && opt.Freq != rrule.WEEKLY {
		return baseDate
	}
	cutoff := now.AddDate(-1, 0, 0)
	if !baseDate.Before(cutoff) {
		return baseDate
	}

	periodDays := opt.Interval
	if periodDays <= 0 {
		periodDays = 1
	}
	if opt.Freq == rrule.WEEKLY {
		if periodDays > math.MaxInt/7 {
			return baseDate // overflow-safe: a period this long already ends past the cutoff
		}
		periodDays *= 7
	}

	// Estimate whole periods to skip, then correct for day-length variance (a guard: floating
	// values have no DST).
	gapDays := daysBetween(baseDate, cutoff)
	if periodDays > gapDays {
		// One period from the anchor already passes the cutoff: the anchor is within one period
		// below it. Returning early also keeps AddDate away from huge day counts.
		return baseDate
	}
	if k := gapDays / periodDays; k > 0 {
		baseDate = baseDate.AddDate(0, 0, k*periodDays)
	}
	for baseDate.After(cutoff) {
		baseDate = baseDate.AddDate(0, 0, -periodDays)
	}
	for {
		next := baseDate.AddDate(0, 0, periodDays)
		if next.After(cutoff) {
			break
		}
		baseDate = next
	}
	return baseDate
}

// daysBetween returns the number of calendar days from a's date to b's date, each read in its own
// location. It counts on the civil calendar instead of through a Duration, so unlike b.Sub(a) it
// never saturates (~292 years). The engine calls it on floating values (both UTC), where it is
// exact; a location difference could make it off by one day.
func daysBetween(a, b time.Time) int {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	from := time.Date(ay, am, ad, 0, 0, 0, 0, time.UTC).Unix()
	to := time.Date(by, bm, bd, 0, 0, 0, 0, time.UTC).Unix()
	return int((to - from) / (24 * 60 * 60))
}

// timeNow is the recurrence engine's clock; tests replace it to pin "now".
var timeNow = time.Now

// plainUntilStep returns the fixed step and the UNTIL of a SECONDLY, MINUTELY or HOURLY rule with
// UNTIL but no COUNT and no BY* part, or a zero step for any other rule. Iterating such a rule from
// an old due date generated every occurrence up to now (FREQ=SECONDLY with a 2020 due date: 200M+
// per completion), and only DAILY/WEEKLY anchors are clamped, so its due date is advanced by
// arithmetic instead, exactly as the same rule without UNTIL, and UNTIL is applied to the result.
func plainUntilStep(opt *rrule.ROption) (time.Duration, time.Time) {
	if opt.Until.IsZero() || opt.Count > 0 || (opt.Freq != rrule.SECONDLY && opt.Freq != rrule.MINUTELY && opt.Freq != rrule.HOURLY) {
		return 0, time.Time{}
	}
	withoutUntil := *opt
	withoutUntil.Until = time.Time{}
	step, ok := fixedDurationForRRule(&withoutUntil)
	if !ok {
		return 0, time.Time{}
	}
	return step, opt.Until
}

// absoluteReminders returns the reminders the engine moves. Relative ones (relative_to set) are
// recomputed from their moved anchor date by updateRelativeReminderDates instead.
func absoluteReminders(reminders []*TaskReminder) []*TaskReminder {
	absolute := make([]*TaskReminder, 0, len(reminders))
	for _, rem := range reminders {
		if rem.RelativeTo == "" {
			absolute = append(absolute, rem)
		}
	}
	return absolute
}

// setTaskDatesRRule moves a completed repeating task to its next occurrence and reopens it.
// Completion is one nextOccurrence call: the input is built from the old task, and the result is
// applied to the new one. With a due date, the due date moves to the next occurrence and every
// other date moves with it. Without one, the no-due-date policy still decides which dates move, as
// upstream's legacy code moved them. DAILY and WEEKLY rules step on the zone's wall clock, not by
// 24 hours; sub-daily rules step by elapsed time. When the rule has no occurrence left (COUNT
// exhausted or UNTIL passed), or nextOccurrence refuses it, the task stays done. zone is the
// completion's zone. Spans are computed in civil days or through shiftTime, so dates more than
// ~292 years apart move without overflow.
func setTaskDatesRRule(oldTask, newTask *Task, now time.Time, zone *time.Location) {
	absolute := absoluteReminders(oldTask.Reminders)
	res := nextOccurrence(completionInput(oldTask, absolute, now, zone))
	if res.Reason == reasonCountExhausted {
		// The last remaining occurrence was completed: the task stays done, and the rule is cleared
		// so it stops advertising an active recurrence.
		newTask.Repeats = ""
		newTask.Repeat = nil
		return
	}
	if res.Reason != "" {
		// No next occurrence (or a refused rule): the task stays done with its dates unchanged.
		return
	}
	if !applyNextOccurrence(oldTask, newTask, absolute, res, zone) {
		return
	}

	// Decrement the remaining COUNT so finite (COUNT) recurrences actually
	// terminate. The DTSTART reset re-bases the rule's count window to the
	// anchor date on every completion, so without persisting a decremented
	// COUNT the task would repeat forever. opt.RRuleString() matches the format
	// produced by TaskRepeat.toRRule(), so the stored value stays consistent.
	// An update that echoes the unchanged structured repeat keeps this decremented
	// text byte-identical (resolveIncomingRepeat's equal-echo rule), and so does a stale
	// client echoing the higher COUNT it read before the decrement (isStaleCount).
	// nextOccurrence parsed the same text, so this parse cannot fail.
	if opt, err := rrule.StrToROption(oldTask.Repeats); err == nil && opt.Count > 1 {
		opt.Count--
		newTask.Repeats = opt.RRuleString()
	}

	newTask.Done = false
}

// completionInput is the nextOccurrence input for completing oldTask: its rule and dates, with
// absolute (its absolute reminders, in order) as the reminders that step.
func completionInput(oldTask *Task, absolute []*TaskReminder, now time.Time, zone *time.Location) nextOccurrenceInput {
	reminders := make([]datedValue, len(absolute))
	for i, rem := range absolute {
		reminders[i] = datedValue{At: rem.Reminder}
	}
	return nextOccurrenceInput{
		Rule:           oldTask.Repeats,
		Due:            datedValue{At: oldTask.DueDate},
		Start:          datedValue{At: oldTask.StartDate},
		End:            datedValue{At: oldTask.EndDate},
		Reminders:      reminders,
		Now:            now,
		FromCompletion: oldTask.RepeatsFromCurrentDate,
		Zone:           zone,
		TaskID:         oldTask.ID,
	}
}

// applyNextOccurrence writes a successful nextOccurrence result onto newTask, where completion has
// always assigned: newTask shares oldTask's reminders (updateSingleTask and updateTaskBucket rely
// on the shared slice), each absolute one takes its moved value, and the due, start and end dates
// are written when the old task had them. Every moved date is converted to the server zone, as
// every date loaded from the database is, so responses keep their offset; the instants are the
// helper's. Relative reminders are recomputed from the moved dates
// later in the save path (updateRelativeReminderDates, in zone, the completion's); when one of them
// would fall outside the storable range, everything applied here is put back and it reports false,
// the task stays done.
func applyNextOccurrence(oldTask, newTask *Task, absolute []*TaskReminder, res nextOccurrenceResult, zone *time.Location) bool {
	dueDate, startDate, endDate, reminders := newTask.DueDate, newTask.StartDate, newTask.EndDate, newTask.Reminders
	reminderTimes := make([]time.Time, len(absolute))
	for i, rem := range absolute {
		reminderTimes[i] = rem.Reminder
	}

	server := config.GetTimeZone()
	newTask.Reminders = oldTask.Reminders
	for i, rem := range absolute {
		rem.Reminder = res.Reminders[i].At.In(server)
	}
	if !oldTask.DueDate.IsZero() {
		newTask.DueDate = res.Due.At.In(server)
	}
	// When the next occurrence is the stored due date itself, the stepping moves no companion date,
	// so a start or end the same update edits or clears is kept.
	if dueUnmoved := !oldTask.DueDate.IsZero() && res.Due.At.Equal(oldTask.DueDate); !dueUnmoved {
		if !oldTask.StartDate.IsZero() {
			newTask.StartDate = res.Start.At.In(server)
		}
		if !oldTask.EndDate.IsZero() {
			newTask.EndDate = res.End.At.In(server)
		}
	}

	// A date outside 1000-01-01..9999-12-31 cannot be stored (MySQL's DATETIME range, the
	// strictest supported database): leave the task done with its dates as they were.
	if datesOutsideStorableRange(newTask, zone) {
		log.Errorf("Not advancing task %d: a date of its next occurrence is outside the years 1000 to 9999", oldTask.ID)
		newTask.DueDate, newTask.StartDate, newTask.EndDate, newTask.Reminders = dueDate, startDate, endDate, reminders
		for i, rem := range absolute {
			rem.Reminder = reminderTimes[i]
		}
		return false
	}
	return true
}

// advanceTaskDates evaluates the rule (bounded, the options rrule-go evaluates) past the guards of
// nextOccurrence, in the server zone, and applies the result to newTask as completion does. It
// reports whether the dates moved; false means the rule has no occurrence left or could not be
// evaluated, and the task stays done with its dates and reminders as they were. Completion no
// longer calls it (it calls nextOccurrence); it is kept for TestAdvanceTaskDatesRecoversFromAPanic,
// which drives the evaluation stage with a rule the bounds refuse.
func advanceTaskDates(oldTask, newTask *Task, opt *rrule.ROption, bounded rrule.ROption, now time.Time) (moved bool) {
	absolute := absoluteReminders(oldTask.Reminders)
	zone := config.GetTimeZone()
	res := evaluateNextOccurrence(completionInput(oldTask, absolute, now, zone), opt, bounded)
	return res.Reason == "" && applyNextOccurrence(oldTask, newTask, absolute, res, zone)
}

// logNoOccurrenceInWindow logs that an infinite rule yielded no next occurrence. A variable so
// tests can observe it.
var logNoOccurrenceInWindow = func(taskID int64) {
	log.Errorf("Not advancing task %d: no occurrence found within the recurrence library's search window", taskID)
}

// dueDateBeyondHorizon reports whether the task's due date lies more than maxAnchorAgeYears before
// now. Only the due date is checked: without a due date, each date takes one
// step from itself, which costs nothing however old it is.
func dueDateBeyondHorizon(due, now time.Time) bool {
	// In UTC, so the cut-off does not depend on the location now carries.
	return !due.IsZero() && due.Before(now.UTC().AddDate(-maxAnchorAgeYears, 0, 0))
}

// datesOutsideStorableRange reports whether any of the task's dates lies outside the years 1000
// to 9999 in UTC, the range of MySQL's DATETIME. Relative reminders count as updateRelativeReminderDates
// will recompute them from the moved dates in zone, since that happens after completion.
func datesOutsideStorableRange(t *Task, zone *time.Location) bool {
	for _, d := range storableRangeDates(t, zone) {
		// The database stores GMT (pkg/db), so the year is the UTC one, whatever the service zone.
		if year := d.UTC().Year(); !d.IsZero() && (year < 1000 || year > 9999) {
			return true
		}
	}
	return false
}

// storableRangeDates is every date datesOutsideStorableRange judges: the due, start and end dates,
// then each reminder as it will be stored, a relative one computed as updateRelativeReminderDates
// computes it. A relative reminder without its base date is left out.
func storableRangeDates(t *Task, zone *time.Location) []time.Time {
	dates := []time.Time{t.DueDate, t.StartDate, t.EndDate}
	for _, rem := range t.Reminders {
		if rem.RelativeTo == "" {
			dates = append(dates, rem.Reminder)
			continue
		}
		if base := relativeReminderBase(t, rem.RelativeTo); !base.IsZero() {
			dates = append(dates, relativeReminderTime(base, rem.RelativePeriod, zone))
		}
	}
	return dates
}

// relativeReminderBase is the task's date a reminder relative to relTo counts from; zero when the
// task has no such date or relTo names none.
func relativeReminderBase(t *Task, relTo ReminderRelation) time.Time {
	switch relTo {
	case ReminderRelationDueDate:
		return t.DueDate
	case ReminderRelationStartDate:
		return t.StartDate
	case ReminderRelationEndDate:
		return t.EndDate
	}
	return time.Time{}
}

// shiftTime moves t by the same offset that takes from to to.
// time.Time.Sub saturates at ~292 years, so for larger spans the offset is
// applied as whole years plus a small remainder instead of one Duration.
func shiftTime(t, from, to time.Time) time.Time {
	diff := to.Sub(from)
	if diff != math.MaxInt64 && diff != math.MinInt64 {
		return t.Add(diff)
	}
	years := to.Year() - from.Year()
	rest := to.Sub(from.AddDate(years, 0, 0))
	return t.AddDate(years, 0, 0).Add(rest)
}

var (
	checklistTiptapCheckedRegex = regexp.MustCompile(`(data-checked=")true(")`)
	checklistInputCheckedRegex  = regexp.MustCompile(`(<input[^>]*type=["']checkbox["'][^>]*?)\s+checked(?:=["'][^"']*["'])?`)
)

// resetDescriptionChecklist unchecks every checklist item in a TipTap HTML description
// (descriptions are always stored as HTML, never markdown) without touching other content,
// so a recurring task's next occurrence does not inherit checked items.
func resetDescriptionChecklist(description string) string {
	description = checklistTiptapCheckedRegex.ReplaceAllString(description, "${1}false${2}")
	description = checklistInputCheckedRegex.ReplaceAllString(description, "$1")
	return description
}

// This helper function updates the reminders, doneAt, start, end and due dates of the *old* task
// and saves the new values in the newTask object.
// We make a few assumptions here:
//  1. Everything in oldTask is the truth - we figure out if we update anything at all if oldTask.Repeats has a value
//  2. Because of 1., this functions should not be used to update values other than Done in the same go
//
// Completing a repeating task is one nextOccurrence call (setTaskDatesRRule), with a single "now"
// and the completion's zone.
func updateDone(oldTask *Task, newTask *Task, zone *time.Location) (updateDoneAt bool) {
	// Track if the done status changed before repeat helpers modify it
	doneStatusChanged := oldTask.Done != newTask.Done

	if !oldTask.Done && newTask.Done {
		if oldTask.Repeats != "" {
			setTaskDatesRRule(oldTask, newTask, timeNow(), zone)
		}

		// A recurring task reopens for its next occurrence, so its checklist starts fresh.
		if oldTask.isRepeating() && !newTask.Done {
			newTask.Description = resetDescriptionChecklist(newTask.Description)
		}

		newTask.DoneAt = time.Now()
	}

	// When unmarking a task as done, reset the timestamp
	if oldTask.Done && !newTask.Done {
		newTask.DoneAt = time.Time{}
	}

	return doneStatusChanged
}

// Set the absolute trigger dates for Reminders with relative period, whole days counted on zone's
// calendar (relativeReminderTime). zone may be nil when no reminder has a period of a day
// or more with its base date set (taskNeedsReminderZone).
func updateRelativeReminderDates(task *Task, zone *time.Location) (err error) {
	for _, reminder := range task.Reminders {
		if reminder.RelativeTo != "" {
			reminder.Reminder = time.Time{}
		}
		switch reminder.RelativeTo {
		case ReminderRelationDueDate, ReminderRelationStartDate, ReminderRelationEndDate:
			if base := relativeReminderBase(task, reminder.RelativeTo); !base.IsZero() {
				reminder.Reminder = relativeReminderTime(base, reminder.RelativePeriod, zone)
			}
		default:
			if reminder.RelativePeriod != 0 {
				err = ErrReminderRelativeToMissing{
					TaskID: task.ID,
				}
				return err
			}
		}
	}
	return nil
}

// taskNeedsReminderZone reports whether computing the task's relative reminders reads a zone: some
// reminder is relative by a whole day or more and its base date is set. Any other
// save resolves no zone.
func taskNeedsReminderZone(task *Task) bool {
	for _, r := range task.Reminders {
		wholeDays := r.RelativePeriod <= -24*60*60 || r.RelativePeriod >= 24*60*60
		if r.RelativeTo != "" && wholeDays && !relativeReminderBase(task, r.RelativeTo).IsZero() {
			return true
		}
	}
	return false
}

// resolveReminderTaskZone is resolveTaskZone as reminderZone calls it; a variable so tests can
// count the resolutions.
var resolveReminderTaskZone = resolveTaskZone

// reminderZone is the zone the task's relative reminders are computed in when a saves it:
// the completion's zone when the save completes a repeating task, otherwise the resolved one, but
// only when a reminder needs it. Otherwise nil, which relativeReminderTime never reads.
func reminderZone(s *xorm.Session, a web.Auth, task *Task, completionZone *time.Location) *time.Location {
	if completionZone != nil {
		return completionZone
	}
	if !taskNeedsReminderZone(task) {
		return nil
	}
	return resolveReminderTaskZone(s, a, task.ProjectID)
}

// Removes all old reminders and adds the new ones. This is a lot easier and less buggy than
// trying to figure out which reminders changed and then only re-add those needed. And since it does
// not make a performance difference we'll just do that.
// The parameter is a slice which holds the new reminders; zone is the one relative reminders are
// computed in (reminderZone).
func (t *Task) updateReminders(s *xorm.Session, task *Task, zone *time.Location) (err error) {

	_, err = s.
		Where("task_id = ?", t.ID).
		Delete(&TaskReminder{})
	if err != nil {
		return
	}

	err = updateRelativeReminderDates(task, zone)
	if err != nil {
		return
	}

	// Resolve duplicates and sort them
	reminderMap := make(map[int64]*TaskReminder, len(task.Reminders))
	for _, reminder := range task.Reminders {
		reminderMap[reminder.Reminder.UTC().Unix()] = reminder
	}

	t.Reminders = make([]*TaskReminder, 0, len(reminderMap))

	// Loop through all reminders and add them
	for _, r := range reminderMap {
		taskReminder := &TaskReminder{
			TaskID:         t.ID,
			Reminder:       r.Reminder,
			RelativePeriod: r.RelativePeriod,
			RelativeTo:     r.RelativeTo}
		_, err = s.Insert(taskReminder)
		if err != nil {
			return err
		}
		t.Reminders = append(t.Reminders, taskReminder)
	}

	// sort reminders
	sort.Slice(t.Reminders, func(i, j int) bool {
		return t.Reminders[i].Reminder.Before(t.Reminders[j].Reminder)
	})

	if len(t.Reminders) == 0 {
		t.Reminders = nil
	}

	err = updateProjectLastUpdated(s, &Project{ID: t.ProjectID})
	return
}

func updateTaskLastUpdated(s *xorm.Session, task *Task) error {
	_, err := s.ID(task.ID).Cols("updated").Update(task)
	return err
}

// Delete implements the delete method for a task
// @Summary Delete a task
// @Description Deletes a task from a project. This does not mean "mark it done".
// @tags task
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "Task ID"
// @Success 200 {object} models.Message "The created task object."
// @Failure 400 {object} web.HTTPError "Invalid task ID provided."
// @Failure 403 {object} web.HTTPError "The user does not have access to the project"
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{id} [delete]
func (t *Task) Delete(s *xorm.Session, a web.Auth) (err error) {

	// duplicate the task for the event
	fullTask := &Task{ID: t.ID}
	err = fullTask.ReadOne(s, a)
	if err != nil {
		return err
	}

	_, err = lockProjectViewsForPositionUpdate(s, fullTask.ProjectID)
	if err != nil {
		return
	}

	// Bucket and position rows are removed right away because bucket counts
	// don't join the tasks table and would leak soft-deleted tasks; the heal
	// routines re-create them on restore. All other related data is kept until
	// the cleanup cron permanently deletes the task.
	_, err = s.Where("task_id = ?", t.ID).Delete(&TaskPosition{})
	if err != nil {
		return
	}

	_, err = s.Where("task_id = ?", t.ID).Delete(&TaskBucket{})
	if err != nil {
		return
	}

	// The deleted tag on Task.DeletedAt turns this into an update setting
	// deleted_at. Must be a pointer: xorm tracks the after-delete closure that
	// stamps DeletedAt in a map keyed by the bean, and a Task value is unhashable.
	_, err = s.ID(t.ID).Delete(&Task{})
	if err != nil {
		return err
	}

	events.DispatchOnCommit(s, &TaskDeletedEvent{
		Task: fullTask,
		Doer: doerFromAuth(s, a),
	})

	// fullTask, not t: the receiver only has the id from the route param
	err = updateProjectLastUpdated(s, &Project{ID: fullTask.ProjectID})
	return
}

// hardDeleteTask permanently removes a task and all its related entities.
// It does not dispatch a TaskDeletedEvent — that already happened when the
// task was soft-deleted by the user.
func hardDeleteTask(s *xorm.Session, t *Task) (err error) {

	// Delete assignees
	if _, err = s.Where("task_id = ?", t.ID).Delete(&TaskAssginee{}); err != nil {
		return err
	}

	// Favorites of all users, not just the doer's
	_, err = s.Where("entity_id = ? AND kind = ?", t.ID, FavoriteKindTask).Delete(&Favorite{})
	if err != nil {
		return
	}

	// Delete label associations
	_, err = s.Where("task_id = ?", t.ID).Delete(&LabelTask{})
	if err != nil {
		return
	}

	// Not attachment.Delete: it resolves the (now soft-deleted) task and
	// dispatches per-attachment events.
	attachments, err := getTaskAttachmentsByTaskIDs(s, []int64{t.ID})
	if err != nil {
		return err
	}
	for _, attachment := range attachments {
		if attachment.File == nil {
			continue
		}
		err = attachment.File.Delete(s)
		if err != nil && !files.IsErrFileDoesNotExist(err) {
			return err
		}
	}

	_, err = s.Where("task_id = ?", t.ID).Delete(&TaskAttachment{})
	if err != nil {
		return err
	}

	commentIDs := []int64{}
	err = s.Table("task_comments").Where("task_id = ?", t.ID).Cols("id").Find(&commentIDs)
	if err != nil {
		return err
	}
	if len(commentIDs) > 0 {
		_, err = s.In("entity_id", commentIDs).And("entity_kind = ?", ReactionKindComment).Delete(&Reaction{})
		if err != nil {
			return err
		}
	}

	_, err = s.Where("entity_id = ? AND entity_kind = ?", t.ID, ReactionKindTask).Delete(&Reaction{})
	if err != nil {
		return err
	}

	// Delete all comments
	_, err = s.Where("task_id = ?", t.ID).Delete(&TaskComment{})
	if err != nil {
		return
	}

	// Delete all task unread statuses
	_, err = s.Where("task_id = ?", t.ID).Delete(&TaskUnreadStatus{})
	if err != nil {
		return err
	}

	// Delete all relations
	_, err = s.Where("task_id = ? OR other_task_id = ?", t.ID, t.ID).Delete(&TaskRelation{})
	if err != nil {
		return
	}

	// Delete all reminders
	_, err = s.Where("task_id = ?", t.ID).Delete(&TaskReminder{})
	if err != nil {
		return
	}

	_, err = s.Where("entity_id = ? AND entity_type = ?", t.ID, SubscriptionEntityTask).Delete(&Subscription{})
	if err != nil {
		return
	}

	// Already gone after a soft delete, but project deletion hard-deletes directly
	_, err = s.Where("task_id = ?", t.ID).Delete(&TaskPosition{})
	if err != nil {
		return
	}

	_, err = s.Where("task_id = ?", t.ID).Delete(&TaskBucket{})
	if err != nil {
		return
	}

	_, err = s.Where("task_id = ?", t.ID).Delete(&TaskIndexAlias{})
	if err != nil {
		return
	}

	_, err = s.ID(t.ID).Unscoped().Delete(&Task{})
	return
}

// GetDeletedTasksSince returns a project's soft-deleted tasks for the CalDAV
// sync-collection 404 entries. Inclusive, because sync tokens have second
// granularity. Tasks without a stored UID were never synced and are skipped.
func GetDeletedTasksSince(s *xorm.Session, projectID int64, since time.Time) (tasks []*Task, err error) {
	err = s.Unscoped().
		Where(builder.And(
			builder.Eq{"project_id": projectID},
			builder.NotNull{"deleted_at"},
			builder.Gte{"deleted_at": since.UTC()},
			builder.Neq{"uid": ""},
		)).
		Find(&tasks)
	return
}

// ReadOne gets one task by its ID
// @Summary Get one task
// @Description Returns one task by its ID
// @tags task
// @Accept json
// @Produce json
// @Param id path int true "The task ID"
// @Param expand query string false "If set to `subtasks`, Vikunja will fetch only tasks which do not have subtasks and then in a second step, will fetch all of these subtasks. This may result in more tasks than the pagination limit being returned, but all subtasks will be present in the response. If set to `buckets`, the buckets of each task will be present in the response. If set to `reactions`, the reactions of each task will be present in the response. If set to `comments`, the first 50 comments of each task will be present in the response. You can set this multiple times with different values."
// @Security JWTKeyAuth
// @Success 200 {object} models.Task "The task"
// @Failure 404 {object} models.Message "Task not found"
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{id} [get]
func (t *Task) ReadOne(s *xorm.Session, a web.Auth) (err error) {

	expand := t.Expand
	if err = t.resolveIDFromProjectAndIndex(s); err != nil {
		return
	}
	*t, err = GetTaskByIDSimple(s, t.ID)
	if err != nil {
		return
	}
	taskMap := make(map[int64]*Task, 1)
	taskMap[t.ID] = t

	for _, expandValue := range expand {
		err = expandValue.Validate()
		if err != nil {
			return
		}
	}

	err = addMoreInfoToTasks(s, taskMap, a, nil, expand)
	if err != nil {
		return
	}

	if len(taskMap) == 0 {
		return ErrTaskDoesNotExist{t.ID}
	}

	*t = *taskMap[t.ID]

	subs, err := GetSubscriptionForUser(s, SubscriptionEntityTask, t.ID, a)
	if err != nil && IsErrProjectDoesNotExist(err) {
		return nil
	}
	if subs != nil {
		t.Subscription = &subs.Subscription
	}

	return
}

func triggerTaskUpdatedEventForTaskID(s *xorm.Session, auth web.Auth, taskID int64) error {
	t, err := GetTaskByIDSimple(s, taskID)
	if err != nil {
		return err
	}

	events.DispatchOnCommit(s, &TaskUpdatedEvent{
		Task: &t,
		Doer: doerFromAuth(s, auth),
	})
	return nil
}
