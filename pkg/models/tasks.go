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
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/utils"
	"code.vikunja.io/api/pkg/web"

	"dario.cat/mergo"
	"github.com/google/uuid"
	clone "github.com/huandu/go-clone/generic"
	"github.com/jinzhu/copier"
	"github.com/typesense/typesense-go/v2/typesense"
	"xorm.io/builder"
	"xorm.io/xorm"
)

type TaskRepeatMode int

const (
	TaskRepeatModeDefault TaskRepeatMode = iota
	TaskRepeatModeMonth
	TaskRepeatModeFromCurrentDate
)

// Task represents a task in a project
type Task struct {
	// The unique, numeric id of this task.
	ID int64 `xorm:"bigint autoincr not null unique pk" json:"id" param:"projecttask"`
	// The task text. This is what you'll see in the project.
	Title string `xorm:"TEXT not null" json:"title" valid:"minstringlength(1)" minLength:"1"`
	// The task description.
	Description string `xorm:"longtext null" json:"description"`
	// Whether a task is done or not.
	Done bool `xorm:"INDEX null" json:"done"`
	// The time when a task was marked as done.
	DoneAt time.Time `xorm:"INDEX null 'done_at'" json:"done_at"`
	// The time when the task is due.
	DueDate time.Time `xorm:"DATETIME INDEX null 'due_date'" json:"due_date"`
	// An array of reminders that are associated with this task.
	Reminders []*TaskReminder `xorm:"-" json:"reminders"`
	// The project this task belongs to.
	ProjectID int64 `xorm:"bigint INDEX not null" json:"project_id" param:"project"`
	// An amount in seconds this task repeats itself. If this is set, when marking the task as done, it will mark itself as "undone" and then increase all remindes and the due date by its amount.
	RepeatAfter int64 `xorm:"bigint INDEX null" json:"repeat_after" valid:"range(0|9223372036854775807)"`
	// Can have three possible values which will trigger when the task is marked as done: 0 = repeats after the amount specified in repeat_after, 1 = repeats all dates each months (ignoring repeat_after), 3 = repeats from the current date rather than the last set date.
	RepeatMode TaskRepeatMode `xorm:"not null default 0" json:"repeat_mode"`
	// The task priority. Can be anything you want, it is possible to sort by this later.
	Priority int64 `xorm:"bigint null" json:"priority"`
	// When this task starts.
	StartDate time.Time `xorm:"DATETIME INDEX null 'start_date'" json:"start_date" query:"-"`
	// When this task ends.
	EndDate time.Time `xorm:"DATETIME INDEX null 'end_date'" json:"end_date" query:"-"`
	// An array of users who are assigned to this task
	Assignees []*user.User `xorm:"-" json:"assignees"`
	// An array of labels which are associated with this task. This property is read-only, you must use the separate endpoint to add labels to a task.
	Labels []*Label `xorm:"-" json:"labels"`
	// The task color in hex
	HexColor string `xorm:"varchar(6) null" json:"hex_color" valid:"runelength(0|7)" maxLength:"7"`
	// Determines how far a task is left from being done
	PercentDone float64 `xorm:"DOUBLE null" json:"percent_done"`

	// The task identifier, based on the project identifier and the task's index
	Identifier string `xorm:"-" json:"identifier"`
	// The task index, calculated per project
	Index int64 `xorm:"bigint not null default 0" json:"index"`

	// The UID is currently not used for anything other than CalDAV, which is why we don't expose it over json
	UID string `xorm:"varchar(250) null" json:"-"`

	// All related tasks, grouped by their relation kind
	RelatedTasks RelatedTaskMap `xorm:"-" json:"related_tasks"`

	// All attachments this task has. This property is read-onlym, you must use the separate endpoint to add attachments to a task.
	Attachments []*TaskAttachment `xorm:"-" json:"attachments"`

	// If this task has a cover image, the field will return the id of the attachment that is the cover image.
	CoverImageAttachmentID int64 `xorm:"bigint default 0" json:"cover_image_attachment_id"`

	// True if a task is a favorite task. Favorite tasks show up in a separate "Important" project. This value depends on the user making the call to the api.
	IsFavorite bool `xorm:"-" json:"is_favorite"`

	// The subscription status for the user reading this task. You can only read this property, use the subscription endpoints to modify it.
	// Will only returned when retrieving one task.
	Subscription *Subscription `xorm:"-" json:"subscription,omitempty"`

	// A timestamp when this task was created. You cannot change this value.
	Created time.Time `xorm:"created not null" json:"created"`
	// A timestamp when this task was last updated. You cannot change this value.
	Updated time.Time `xorm:"updated not null" json:"updated"`

	// The bucket id. Will only be populated when the task is accessed via a view with buckets.
	// Can be used to move a task between buckets. In that case, the new bucket must be in the same view as the old one.
	BucketID int64 `xorm:"-" json:"bucket_id"`

	// All buckets across all views this task is part of. Only present when fetching tasks with the `expand` parameter set to `buckets`.
	Buckets []*Bucket `xorm:"-" json:"buckets,omitempty"`

	// All comments of this task. Only present when fetching tasks with the `expand` parameter set to `comments`.
	Comments []*TaskComment `xorm:"-" json:"comments,omitempty"`

	// Behaves exactly the same as with the TaskCollection.Expand parameter
	Expand []TaskCollectionExpandable `xorm:"-" json:"-" query:"expand"`

	// The position of the task - any task project can be sorted as usual by this parameter.
	// When accessing tasks via views with buckets, this is primarily used to sort them based on a range.
	// Positions are always saved per view. They will automatically be set if you request the tasks through a view
	// endpoint, otherwise they will always be 0. To update them, take a look at the Task Position endpoint.
	Position float64 `xorm:"-" json:"position"`

	// Reactions on that task.
	Reactions ReactionMap `xorm:"-" json:"reactions"`

	// The user who initially created the task.
	CreatedBy   *user.User `xorm:"-" json:"created_by" valid:"-"`
	CreatedByID int64      `xorm:"bigint not null" json:"-"` // ID of the user who put that task on the project

	web.CRUDable    `xorm:"-" json:"-"`
	web.Permissions `xorm:"-" json:"-"`
}

type TaskWithComments struct {
	Task
	Comments []*TaskComment `xorm:"-" json:"comments"`
}

// TableName returns the table name for tasks
func (*Task) TableName() string {
	return "tasks"
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
	return t.RepeatAfter > 0 ||
		t.RepeatMode == TaskRepeatModeMonth
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
// @Param sort_by query string false "The sorting parameter. You can pass this multiple times to get the tasks ordered by multiple different parametes, along with `order_by`. Possible values to sort by are `id`, `title`, `description`, `done`, `done_at`, `due_date`, `created_by_id`, `project_id`, `repeat_after`, `priority`, `start_date`, `end_date`, `hex_color`, `percent_done`, `uid`, `created`, `updated`. Default is `id`."
// @Param order_by query string false "The ordering parameter. Possible values to order by are `asc` or `desc`. Default is `asc`."
// @Param filter query string false "The filter query to match tasks by. Check out https://vikunja.io/docs/filters for a full explanation of the feature."
// @Param filter_timezone query string false "The time zone which should be used for date match (statements like "now" resolve to different actual times)"
// @Param filter_include_nulls query string false "If set to true the result will include filtered fields whose value is set to `null`. Available values are `true` or `false`. Defaults to `false`."
// @Param expand query []string false "If set to `subtasks`, Vikunja will fetch only tasks which do not have subtasks and then in a second step, will fetch all of these subtasks. This may result in more tasks than the pagination limit being returned, but all subtasks will be present in the response. If set to `buckets`, the buckets of each task will be present in the response. If set to `reactions`, the reactions of each task will be present in the response. If set to `comments`, the first 50 comments of each task will be present in the response. You can set this multiple times with different values."
// @Security JWTKeyAuth
// @Success 200 {array} models.Task "The tasks"
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/all [get]
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

func getRawTasksForProjects(s *xorm.Session, projects []*Project, a web.Auth, opts *taskSearchOptions) (tasks []*Task, resultCount int, totalItems int64, err error) {

	// If the user does not have any projects, don't try to get any tasks
	if len(projects) == 0 {
		return nil, 0, 0, nil
	}

	// Get all project IDs and get the tasks
	opts.projectIDs = []int64{}
	var hasFavoritesProject bool
	for _, p := range projects {
		if p.ID == FavoritesPseudoProject.ID {
			hasFavoritesProject = true
			continue
		}
		opts.projectIDs = append(opts.projectIDs, p.ID)
	}

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
	if config.TypesenseEnabled.GetBool() {
		var tsSearcher taskSearcher = &typesenseTaskSearcher{
			s: s,
		}
		origOpts := clone.Clone(opts)
		tasks, totalItems, err = tsSearcher.Search(opts)
		// It is possible that project views are not yet in Typesense's index. This causes the query here to fail.
		// To avoid crashing everything, we fall back to the db search in that case.
		var tsErr = &typesense.HTTPError{}
		if err != nil && errors.As(err, &tsErr) && tsErr.Status == 404 {
			log.Warningf("Unable to fetch tasks from Typesense, error was '%v'. Falling back to db.", err)
			tasks, totalItems, err = dbSearcher.Search(origOpts)
		}
	} else {
		tasks, totalItems, err = dbSearcher.Search(opts)
	}

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

// GetTaskByIDSimple returns a raw task without extra data by the task ID
func GetTaskByIDSimple(s *xorm.Session, taskID int64) (task Task, err error) {
	if taskID < 1 {
		return Task{}, ErrTaskDoesNotExist{taskID}
	}

	return GetTaskSimple(s, &Task{ID: taskID})
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

// GetTasksByIDs returns all tasks for a project of ids
func (bt *BulkTask) GetTasksByIDs(s *xorm.Session) (err error) {
	for _, id := range bt.IDs {
		if id < 1 {
			return ErrTaskDoesNotExist{id}
		}
	}

	err = s.In("id", bt.IDs).Find(&bt.Tasks)
	if err != nil {
		return
	}

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

// GetTasksByUIDs gets all tasks from a bunch of uids
func GetTasksByUIDs(s *xorm.Session, uids []string, a web.Auth) (tasks []*Task, err error) {
	tasks = []*Task{}
	err = s.In("uid", uids).Find(&tasks)
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
	if project == nil || (project != nil && project.Identifier == "") {
		t.Identifier = "#" + strconv.FormatInt(t.Index, 10)
		return
	}

	t.Identifier = project.Identifier + "-" + strconv.FormatInt(t.Index, 10)
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
	labels, _, _, err := GetLabelsByTaskIDs(s, &LabelByTaskIDsOptions{
		TaskIDs: taskIDs,
		Page:    -1,
	})
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

	fullRelatedTasks := make(map[int64]*Task)
	err = s.In("id", relatedTaskIDs).Find(&fullRelatedTasks)
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

	// We need to fetch all projects for that user to make sure they only
	// get to see buckets that they have permission to see.
	projectIDs := []int64{}
	allProjects, _, _, err := getAllRawProjects(s, a, "", 0, -1, false)
	if err != nil {
		return err
	}

	for _, project := range allProjects {
		projectIDs = append(projectIDs, project.ID)
	}

	buckets := make(map[int64]*Bucket)
	err = s.
		Where(builder.In("id", builder.Select("bucket_id").
			From("task_buckets").
			Where(builder.In("task_id", taskIDs)))).
		And(builder.In("project_view_id", builder.Select("id").
			From("project_views").
			Where(builder.In("project_id", projectIDs)))).
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

// Checks if adding a new task would exceed the bucket limit
func checkBucketLimit(s *xorm.Session, a web.Auth, t *Task, bucket *Bucket) (taskCount int64, err error) {
	view, err := GetProjectViewByID(s, bucket.ProjectViewID)
	if err != nil {
		return 0, err
	}

	if view.ProjectID < 0 || (view.Filter != nil && view.Filter.Filter != "") {
		tc := &TaskCollection{
			ProjectID:     view.ProjectID,
			ProjectViewID: bucket.ProjectViewID,
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

	if bucket.Limit > 0 && taskCount >= bucket.Limit {
		return 0, ErrBucketLimitExceeded{TaskID: t.ID, BucketID: bucket.ID, Limit: bucket.Limit}
	}

	return
}

func calculateDefaultPosition(entityID int64, position float64) float64 {
	if position == 0 {
		return float64(entityID) * math.Pow(2, 16)
	}

	return position
}

func calculateNextTaskIndex(s *xorm.Session, projectID int64) (nextIndex int64, err error) {
	latestTask := &Task{}
	_, err = s.
		Where("project_id = ?", projectID).
		OrderBy("`index` desc").
		Get(latestTask)
	if err != nil {
		return 0, err
	}

	return latestTask.Index + 1, nil
}

func setNewTaskIndex(s *xorm.Session, t *Task) (err error) {
	// Check if an index was provided, otherwise calculate a new one
	if t.Index == 0 {
		t.Index, err = calculateNextTaskIndex(s, t.ProjectID)
		return
	}

	// Check if the provided index is already taken
	exists, err := s.Where("project_id = ? AND `index` = ?", t.ProjectID, t.Index).Exist(&Task{})
	if err != nil {
		return err
	}
	if exists {
		// If the index is taken, calculate a new one
		t.Index, err = calculateNextTaskIndex(s, t.ProjectID)
		if err != nil {
			return err
		}
	}

	return
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

	t.ID = 0

	// Check if we have at least a title
	if t.Title == "" {
		return ErrTaskCannotBeEmpty{}
	}

	// Check if the project exists
	p, err := GetProjectSimpleByID(s, t.ProjectID)
	if err != nil {
		return err
	}

	createdBy, err := GetUserOrLinkShareUser(s, a)
	if err != nil {
		return err
	}
	t.CreatedByID = createdBy.ID

	// Generate a uuid if we don't already have one
	if t.UID == "" {
		t.UID = uuid.NewString()
	}

	err = setNewTaskIndex(s, t)
	if err != nil {
		return err
	}

	t.HexColor = utils.NormalizeHex(t.HexColor)

	_, err = s.Insert(t)
	if err != nil {
		return err
	}

	var providedBucket *Bucket
	if t.BucketID != 0 {
		providedBucket, err = getBucketByID(s, t.BucketID)
		if err != nil {
			return
		}

		_, err = checkBucketLimit(s, a, t, providedBucket)
		if err != nil {
			return
		}
	}

	positions, taskBuckets, err := setTaskInBucketInViews(s, t, a, setBucket, providedBucket)
	if err != nil {
		return err
	}

	if len(positions) > 0 {
		_, err = s.Insert(&positions)
		if err != nil {
			return
		}
	}

	if len(taskBuckets) > 0 {
		_, err = s.Insert(&taskBuckets)
		if err != nil {
			return
		}
	}

	t.CreatedBy = createdBy

	// Update the assignees
	if updateAssignees {
		if err := t.updateTaskAssignees(s, t.Assignees, a); err != nil {
			return err
		}
	}

	// Update the reminders
	if err := t.updateReminders(s, t); err != nil {
		return err
	}

	t.setIdentifier(p)

	if t.IsFavorite {
		if err := addToFavorites(s, t.ID, createdBy, FavoriteKindTask); err != nil {
			return err
		}
	}

	err = events.Dispatch(&TaskCreatedEvent{
		Task: t,
		Doer: createdBy,
	})
	if err != nil {
		return err
	}

	err = updateProjectLastUpdated(s, &Project{ID: t.ProjectID})
	return
}

func setTaskInBucketInViews(s *xorm.Session, t *Task, a web.Auth, setBucket bool, providedBucket *Bucket) ([]*TaskPosition, []*TaskBucket, error) {
	views, err := getViewsForProject(s, t.ProjectID)
	if err != nil {
		return nil, nil, err
	}

	positions := []*TaskPosition{}
	taskBuckets := []*TaskBucket{}

	var moveToDone bool

	for _, view := range views {
		if setBucket && !moveToDone &&
			view.ViewKind == ProjectViewKindKanban &&
			view.BucketConfigurationMode == BucketConfigurationModeManual {

			bucketID := view.DoneBucketID
			if !t.Done || view.DoneBucketID == 0 {
				if providedBucket != nil && view.ID == providedBucket.ProjectViewID {
					bucketID = providedBucket.ID
				} else {
					bucketID, err = getDefaultBucketID(s, view)
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

			taskBuckets = append(taskBuckets, &TaskBucket{
				BucketID:      bucketID,
				TaskID:        t.ID,
				ProjectViewID: view.ID,
			})
		}

		newPosition, err := calculateNewPositionForTask(s, a, t, view)
		if err != nil {
			return nil, nil, err
		}

		positions = append(positions, newPosition)
	}

	if moveToDone {
		taskBuckets = []*TaskBucket{}
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
//
//nolint:gocyclo
func (t *Task) Update(s *xorm.Session, a web.Auth) (err error) {

	// Check if the task exists and get the old values
	ot, err := GetTaskByIDSimple(s, t.ID)
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

	// Update the assignees
	if err := ot.updateTaskAssignees(s, t.Assignees, a); err != nil {
		return err
	}

	// All columns to update in a separate variable to be able to add to them
	colsToUpdate := []string{
		"title",
		"description",
		"done",
		"due_date",
		"repeat_after",
		"priority",
		"start_date",
		"end_date",
		"hex_color",
		"done_at",
		"percent_done",
		"project_id",
		"bucket_id",
		"repeat_mode",
		"cover_image_attachment_id",
	}

	// If the task is being moved between projects, make sure to move the bucket + index as well
	if t.ProjectID != 0 && ot.ProjectID != t.ProjectID {
		t.Index, err = calculateNextTaskIndex(s, t.ProjectID)
		if err != nil {
			return err
		}
		t.BucketID = 0
		colsToUpdate = append(colsToUpdate, "index")
	}

	views := []*ProjectView{}
	if (!t.isRepeating() && t.Done != ot.Done) || t.ProjectID != ot.ProjectID {
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
			err = tb.Update(s, a)
			if err != nil {
				return err
			}

			tp, err := calculateNewPositionForTask(s, a, t, view)
			if err != nil {
				return err
			}

			err = tp.Update(s, a)
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

	// When a repeating task is marked as done, we update all deadlines and reminders and set it as undone
	updateDone(&ot, t)

	// Update the reminders
	if err := ot.updateReminders(s, t); err != nil {
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
	if err := mergo.Merge(&ot, t, mergo.WithOverride); err != nil {
		return err
	}

	t.HexColor = utils.NormalizeHex(t.HexColor)

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
	// Repeat after
	if t.RepeatAfter == 0 {
		ot.RepeatAfter = 0
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
	// Repeat from current date
	if t.RepeatMode == TaskRepeatModeDefault {
		ot.RepeatMode = TaskRepeatModeDefault
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
		Update(ot)
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

	doer, _ := user.GetFromAuth(a)
	err = events.Dispatch(&TaskUpdatedEvent{
		Task: t,
		Doer: doer,
	})
	if err != nil {
		return err
	}

	return updateProjectLastUpdated(s, &Project{ID: t.ProjectID})
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
		err = tb.Update(s, a)
		if err != nil {
			return err
		}

		tp := TaskPosition{
			TaskID:        t.ID,
			ProjectViewID: view.ID,
			Position:      calculateDefaultPosition(t.Index, t.Position),
		}
		err = tp.Update(s, a)
		if err != nil {
			return err
		}
	}
	return nil
}

func addOneMonthToDate(d time.Time) time.Time {
	return time.Date(d.Year(), d.Month()+1, d.Day(), d.Hour(), d.Minute(), d.Second(), d.Nanosecond(), config.GetTimeZone())
}

func addRepeatIntervalToTime(now, t time.Time, duration time.Duration) time.Time {
	for {
		t = t.Add(duration)
		if t.After(now) {
			break
		}
	}

	return t
}

func setTaskDatesDefault(oldTask, newTask *Task) {
	if oldTask.RepeatAfter == 0 {
		return
	}

	// Current time in an extra variable to base all calculations on the same time
	now := time.Now()

	repeatDuration := time.Duration(oldTask.RepeatAfter) * time.Second

	// assuming we'll merge the new task over the old task
	if !oldTask.DueDate.IsZero() {
		newTask.DueDate = addRepeatIntervalToTime(now, oldTask.DueDate, repeatDuration)
	}

	newTask.Reminders = oldTask.Reminders
	// When repeating from the current date, all reminders should keep their difference to each other.
	// To make this easier, we sort them first because we can then rely on the fact the first is the smallest
	if len(oldTask.Reminders) > 0 {
		for in, r := range oldTask.Reminders {
			newTask.Reminders[in].Reminder = addRepeatIntervalToTime(now, r.Reminder, repeatDuration)
		}
	}

	// If a task has a start and end date, the end date should keep the difference to the start date when setting them as new
	if !oldTask.StartDate.IsZero() {
		newTask.StartDate = addRepeatIntervalToTime(now, oldTask.StartDate, repeatDuration)
	}

	if !oldTask.EndDate.IsZero() {
		newTask.EndDate = addRepeatIntervalToTime(now, oldTask.EndDate, repeatDuration)
	}

	newTask.Done = false
}

func setTaskDatesMonthRepeat(oldTask, newTask *Task) {
	if !oldTask.DueDate.IsZero() {
		newTask.DueDate = addOneMonthToDate(oldTask.DueDate)
	}

	newTask.Reminders = oldTask.Reminders
	if len(oldTask.Reminders) > 0 {
		for in, r := range oldTask.Reminders {
			newTask.Reminders[in].Reminder = addOneMonthToDate(r.Reminder)
		}
	}

	if !oldTask.StartDate.IsZero() && !oldTask.EndDate.IsZero() {
		diff := oldTask.EndDate.Sub(oldTask.StartDate)
		newTask.StartDate = addOneMonthToDate(oldTask.StartDate)
		newTask.EndDate = newTask.StartDate.Add(diff)
	} else {
		if !oldTask.StartDate.IsZero() {
			newTask.StartDate = addOneMonthToDate(oldTask.StartDate)
		}

		if !oldTask.EndDate.IsZero() {
			newTask.EndDate = addOneMonthToDate(oldTask.EndDate)
		}
	}

	newTask.Done = false
}

func setTaskDatesFromCurrentDateRepeat(oldTask, newTask *Task) {
	if oldTask.RepeatAfter == 0 {
		return
	}

	// Current time in an extra variable to base all calculations on the same time
	now := time.Now()

	repeatDuration := time.Duration(oldTask.RepeatAfter) * time.Second

	// assuming we'll merge the new task over the old task
	if !oldTask.DueDate.IsZero() {
		newTask.DueDate = now.Add(repeatDuration)
	}

	newTask.Reminders = oldTask.Reminders
	// When repeating from the current date, all reminders should keep their difference to each other.
	// To make this easier, we sort them first because we can then rely on the fact the first is the smallest
	if len(oldTask.Reminders) > 0 {
		sort.Slice(oldTask.Reminders, func(i, j int) bool {
			return oldTask.Reminders[i].Reminder.Unix() < oldTask.Reminders[j].Reminder.Unix()
		})
		first := oldTask.Reminders[0].Reminder
		for in, r := range oldTask.Reminders {
			diff := r.Reminder.Sub(first)
			newTask.Reminders[in].Reminder = now.Add(repeatDuration + diff)
		}
	}

	// We want to preserve intervals among the due, start and end dates.
	// The due date is used as a reference point for all new dates, so the
	// behaviour depends on whether the due date is set at all.
	if oldTask.DueDate.IsZero() {
		// If a task has no due date, but does have a start and end date, the
		// end date should keep the difference to the start date when setting
		// them as new
		if !oldTask.StartDate.IsZero() && !oldTask.EndDate.IsZero() {
			diff := oldTask.EndDate.Sub(oldTask.StartDate)
			newTask.StartDate = now.Add(repeatDuration)
			newTask.EndDate = now.Add(repeatDuration + diff)
		} else {
			if !oldTask.StartDate.IsZero() {
				newTask.StartDate = now.Add(repeatDuration)
			}

			if !oldTask.EndDate.IsZero() {
				newTask.EndDate = now.Add(repeatDuration)
			}
		}
	} else {
		// If the old task has a start and due date, we set the new start date
		// to preserve the interval between them.
		if !oldTask.StartDate.IsZero() {
			diff := oldTask.DueDate.Sub(oldTask.StartDate)
			newTask.StartDate = newTask.DueDate.Add(-diff)
		}

		// If the old task has an end and due date, we set the new end date
		// to preserve the interval between them.
		if !oldTask.EndDate.IsZero() {
			diff := oldTask.DueDate.Sub(oldTask.EndDate)
			newTask.EndDate = newTask.DueDate.Add(-diff)
		}
	}

	newTask.Done = false
}

// This helper function updates the reminders, doneAt, start and end dates of the *old* task
// and saves the new values in the newTask object.
// We make a few assumptions here:
//  1. Everything in oldTask is the truth - we figure out if we update anything at all if oldTask.RepeatAfter has a value > 0
//  2. Because of 1., this functions should not be used to update values other than Done in the same go
func updateDone(oldTask *Task, newTask *Task) {
	if !oldTask.Done && newTask.Done {
		switch oldTask.RepeatMode {
		case TaskRepeatModeMonth:
			setTaskDatesMonthRepeat(oldTask, newTask)
		case TaskRepeatModeFromCurrentDate:
			setTaskDatesFromCurrentDateRepeat(oldTask, newTask)
		case TaskRepeatModeDefault:
			setTaskDatesDefault(oldTask, newTask)
		}

		newTask.DoneAt = time.Now()
	}

	// When unmarking a task as done, reset the timestamp
	if oldTask.Done && !newTask.Done {
		newTask.DoneAt = time.Time{}
	}
}

// Set the absolute trigger dates for Reminders with relative period
func updateRelativeReminderDates(task *Task) (err error) {
	for _, reminder := range task.Reminders {
		relativeDuration := time.Duration(reminder.RelativePeriod) * time.Second
		if reminder.RelativeTo != "" {
			reminder.Reminder = time.Time{}
		}
		switch reminder.RelativeTo {
		case ReminderRelationDueDate:
			if !task.DueDate.IsZero() {
				reminder.Reminder = task.DueDate.Add(relativeDuration)
			}
		case ReminderRelationStartDate:
			if !task.StartDate.IsZero() {
				reminder.Reminder = task.StartDate.Add(relativeDuration)
			}
		case ReminderRelationEndDate:
			if !task.EndDate.IsZero() {
				reminder.Reminder = task.EndDate.Add(relativeDuration)
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

// Removes all old reminders and adds the new ones. This is a lot easier and less buggy than
// trying to figure out which reminders changed and then only re-add those needed. And since it does
// not make a performance difference we'll just do that.
// The parameter is a slice which holds the new reminders.
func (t *Task) updateReminders(s *xorm.Session, task *Task) (err error) {

	_, err = s.
		Where("task_id = ?", t.ID).
		Delete(&TaskReminder{})
	if err != nil {
		return
	}

	err = updateRelativeReminderDates(task)
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

	// Delete assignees
	if _, err = s.Where("task_id = ?", t.ID).Delete(&TaskAssginee{}); err != nil {
		return err
	}

	// Delete Favorites
	err = removeFromFavorite(s, t.ID, a, FavoriteKindTask)
	if err != nil {
		return
	}

	// Delete label associations
	_, err = s.Where("task_id = ?", t.ID).Delete(&LabelTask{})
	if err != nil {
		return
	}

	// Delete task attachments
	attachments, err := getTaskAttachmentsByTaskIDs(s, []int64{t.ID})
	if err != nil {
		return err
	}
	for _, attachment := range attachments {
		// Using the attachment delete method here because that takes care of removing all files properly
		err = attachment.Delete(s, a)
		if err != nil && !IsErrTaskAttachmentDoesNotExist(err) {
			return err
		}
	}

	// Delete all comments
	_, err = s.Where("task_id = ?", t.ID).Delete(&TaskComment{})
	if err != nil {
		return
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

	// Delete all positions
	_, err = s.Where("task_id = ?", t.ID).Delete(&TaskPosition{})
	if err != nil {
		return
	}

	// Delete all bucket relations
	_, err = s.Where("task_id = ?", t.ID).Delete(&TaskBucket{})
	if err != nil {
		return
	}

	// Actually delete the task
	_, err = s.ID(t.ID).Delete(Task{})
	if err != nil {
		return err
	}

	doer, _ := user.GetFromAuth(a)
	err = events.Dispatch(&TaskDeletedEvent{
		Task: fullTask,
		Doer: doer,
	})
	if err != nil {
		return
	}

	err = updateProjectLastUpdated(s, &Project{ID: t.ProjectID})
	return
}

// ReadOne gets one task by its ID
// @Summary Get one task
// @Description Returns one task by its ID
// @tags task
// @Accept json
// @Produce json
// @Param id path int true "The task ID"
// @Param expand query []string false "If set to `subtasks`, Vikunja will fetch only tasks which do not have subtasks and then in a second step, will fetch all of these subtasks. This may result in more tasks than the pagination limit being returned, but all subtasks will be present in the response. If set to `buckets`, the buckets of each task will be present in the response. If set to `reactions`, the reactions of each task will be present in the response. If set to `comments`, the first 50 comments of each task will be present in the response. You can set this multiple times with different values."
// @Security JWTKeyAuth
// @Success 200 {object} models.Task "The task"
// @Failure 404 {object} models.Message "Task not found"
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{id} [get]
func (t *Task) ReadOne(s *xorm.Session, a web.Auth) (err error) {

	expand := t.Expand
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

	doer, _ := user.GetFromAuth(auth)
	err = events.Dispatch(&TaskUpdatedEvent{
		Task: &t,
		Doer: doer,
	})
	return err
}
