// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2020 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package models

import (
	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/metrics"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/utils"
	"code.vikunja.io/web"
	"github.com/imdario/mergo"
	"math"
	"sort"
	"strconv"
	"time"
	"xorm.io/builder"
	"xorm.io/xorm/schemas"
)

// Task represents an task in a todolist
type Task struct {
	// The unique, numeric id of this task.
	ID int64 `xorm:"int(11) autoincr not null unique pk" json:"id" param:"listtask"`
	// The task text. This is what you'll see in the list.
	Title string `xorm:"varchar(250) not null" json:"title" valid:"runelength(1|250)" minLength:"3" maxLength:"250"`
	// The task description.
	Description string `xorm:"longtext null" json:"description"`
	// Whether a task is done or not.
	Done bool `xorm:"INDEX null" json:"done"`
	// The time when a task was marked as done.
	DoneAt time.Time `xorm:"INDEX null 'done_at'" json:"done_at"`
	// The time when the task is due.
	DueDate time.Time `xorm:"DATETIME INDEX null 'due_date'" json:"due_date"`
	// An array of datetimes when the user wants to be reminded of the task.
	Reminders   []time.Time `xorm:"-" json:"reminder_dates"`
	CreatedByID int64       `xorm:"int(11) not null" json:"-"` // ID of the user who put that task on the list
	// The list this task belongs to.
	ListID int64 `xorm:"int(11) INDEX not null" json:"list_id" param:"list"`
	// An amount in seconds this task repeats itself. If this is set, when marking the task as done, it will mark itself as "undone" and then increase all remindes and the due date by its amount.
	RepeatAfter int64 `xorm:"int(11) INDEX null" json:"repeat_after"`
	// If specified, a repeating task will repeat from the current date rather than the last set date.
	RepeatFromCurrentDate bool `xorm:"null" json:"repeat_from_current_date"`
	// The task priority. Can be anything you want, it is possible to sort by this later.
	Priority int64 `xorm:"int(11) null" json:"priority"`
	// When this task starts.
	StartDate time.Time `xorm:"DATETIME INDEX null 'start_date'" json:"start_date" query:"-"`
	// When this task ends.
	EndDate time.Time `xorm:"DATETIME INDEX null 'end_date'" json:"end_date" query:"-"`
	// An array of users who are assigned to this task
	Assignees []*user.User `xorm:"-" json:"assignees"`
	// An array of labels which are associated with this task.
	Labels []*Label `xorm:"-" json:"labels"`
	// The task color in hex
	HexColor string `xorm:"varchar(6) null" json:"hex_color" valid:"runelength(0|6)" maxLength:"6"`
	// Determines how far a task is left from being done
	PercentDone float64 `xorm:"DOUBLE null" json:"percent_done"`

	// The task identifier, based on the list identifier and the task's index
	Identifier string `xorm:"-" json:"identifier"`
	// The task index, calculated per list
	Index int64 `xorm:"int(11) not null default 0" json:"index"`

	// The UID is currently not used for anything other than caldav, which is why we don't expose it over json
	UID string `xorm:"varchar(250) null" json:"-"`

	// All related tasks, grouped by their relation kind
	RelatedTasks RelatedTaskMap `xorm:"-" json:"related_tasks"`

	// All attachments this task has
	Attachments []*TaskAttachment `xorm:"-" json:"attachments"`

	// A timestamp when this task was created. You cannot change this value.
	Created time.Time `xorm:"created not null" json:"created"`
	// A timestamp when this task was last updated. You cannot change this value.
	Updated time.Time `xorm:"updated not null" json:"updated"`

	// BucketID is the ID of the kanban bucket this task belongs to.
	BucketID int64 `xorm:"int(11) null" json:"bucket_id"`

	// The position of the task - any task list can be sorted as usual by this parameter.
	// When accessing tasks via kanban buckets, this is primarily used to sort them based on a range
	// We're using a float64 here to make it possible to put any task within any two other tasks (by changing the number).
	// You would calculate the new position between two tasks with something like task3.position = (task2.position - task1.position) / 2.
	// A 64-Bit float leaves plenty of room to initially give tasks a position with 2^16 difference to the previous task
	// which also leaves a lot of room for rearranging and sorting later.
	Position float64 `xorm:"double null" json:"position"`

	// The user who initially created the task.
	CreatedBy *user.User `xorm:"-" json:"created_by" valid:"-"`

	web.CRUDable `xorm:"-" json:"-"`
	web.Rights   `xorm:"-" json:"-"`
}

// TableName returns the table name for listtasks
func (Task) TableName() string {
	return "tasks"
}

// TaskReminder holds a reminder on a task
type TaskReminder struct {
	ID       int64     `xorm:"int(11) autoincr not null unique pk"`
	TaskID   int64     `xorm:"int(11) not null INDEX"`
	Reminder time.Time `xorm:"DATETIME not null INDEX 'reminder'"`
	Created  time.Time `xorm:"created not null"`
}

// TableName returns a pretty table name
func (TaskReminder) TableName() string {
	return "task_reminders"
}

type taskFilterConcatinator string

const (
	filterConcatAnd = "and"
	filterConcatOr  = "or"
)

type taskOptions struct {
	search             string
	page               int
	perPage            int
	sortby             []*sortParam
	filters            []*taskFilter
	filterConcat       taskFilterConcatinator
	filterIncludeNulls bool
}

// ReadAll is a dummy function to still have that endpoint documented
// @Summary Get tasks
// @Description Returns all tasks on any list the user has access to.
// @tags task
// @Accept json
// @Produce json
// @Param page query int false "The page number. Used for pagination. If not provided, the first page of results is returned."
// @Param per_page query int false "The maximum number of items per page. Note this parameter is limited by the configured maximum of items per page."
// @Param s query string false "Search tasks by task text."
// @Param sort_by query string false "The sorting parameter. You can pass this multiple times to get the tasks ordered by multiple different parametes, along with `order_by`. Possible values to sort by are `id`, `text`, `description`, `done`, `done_at`, `due_date`, `created_by_id`, `list_id`, `repeat_after`, `priority`, `start_date`, `end_date`, `hex_color`, `percent_done`, `uid`, `created`, `updated`. Default is `id`."
// @Param order_by query string false "The ordering parameter. Possible values to order by are `asc` or `desc`. Default is `asc`."
// @Param filter_by query string false "The name of the field to filter by. Accepts an array for multiple filters which will be chanied together, all supplied filter must match."
// @Param filter_value query string false "The value to filter for."
// @Param filter_comparator query string false "The comparator to use for a filter. Available values are `equals`, `greater`, `greater_equals`, `less` and `less_equals`. Defaults to `equals`"
// @Param filter_concat query string false "The concatinator to use for filters. Available values are `and` or `or`. Defaults to `or`."
// @Security JWTKeyAuth
// @Success 200 {array} models.Task "The tasks"
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/all [get]
func (t *Task) ReadAll(a web.Auth, search string, page int, perPage int) (result interface{}, resultCount int, totalItems int64, err error) {
	return nil, 0, 0, nil
}

func getRawTasksForLists(lists []*List, opts *taskOptions) (tasks []*Task, resultCount int, totalItems int64, err error) {

	// If the user does not have any lists, don't try to get any tasks
	if len(lists) == 0 {
		return nil, 0, 0, nil
	}

	// Set the default concatinator of filter variables to or if none was provided
	if opts.filterConcat == "" {
		opts.filterConcat = filterConcatOr
	}

	// Get all list IDs and get the tasks
	var listIDs []int64
	for _, l := range lists {
		listIDs = append(listIDs, l.ID)
	}

	// Add the id parameter as the last parameter to sorty by default, but only if it is not already passed as the last parameter.
	if len(opts.sortby) == 0 ||
		len(opts.sortby) > 0 && opts.sortby[len(opts.sortby)-1].sortBy != taskPropertyID {
		opts.sortby = append(opts.sortby, &sortParam{
			sortBy:  taskPropertyID,
			orderBy: orderAscending,
		})
	}

	// Since xorm does not use placeholders for order by, it is possible to expose this with sql injection if we're directly
	// passing user input to the db.
	// As a workaround to prevent this, we check for valid column names here prior to passing it to the db.
	var orderby string
	for i, param := range opts.sortby {
		// Validate the params
		if err := param.validate(); err != nil {
			return nil, 0, 0, err
		}
		orderby += param.sortBy + " " + param.orderBy.String()

		// Postgres sorts by default entries with null values after ones with values.
		// To make that consistent with the sort order we have and other dbms, we're adding a separate clause here.
		if x.Dialect().URI().DBType == schemas.POSTGRES {
			if param.orderBy == orderAscending {
				orderby += " NULLS FIRST"
			}
			if param.orderBy == orderDescending {
				orderby += " NULLS LAST"
			}
		}

		if (i + 1) < len(opts.sortby) {
			orderby += ", "
		}
	}

	var filters = make([]builder.Cond, 0, len(opts.filters))
	// To still find tasks with nil values, we exclude 0s when comparing with >/< values.
	for _, f := range opts.filters {
		switch f.comparator {
		case taskFilterComparatorEquals:
			filters = append(filters, &builder.Eq{f.field: f.value})
		case taskFilterComparatorNotEquals:
			filters = append(filters, &builder.Neq{f.field: f.value})
		case taskFilterComparatorGreater:
			if opts.filterIncludeNulls {
				filters = append(filters, builder.Or(&builder.Gt{f.field: f.value}, &builder.IsNull{f.field}))
			} else {
				filters = append(filters, &builder.Gt{f.field: f.value})
			}
		case taskFilterComparatorGreateEquals:
			if opts.filterIncludeNulls {
				filters = append(filters, builder.Or(&builder.Gte{f.field: f.value}, &builder.IsNull{f.field}))
			} else {
				filters = append(filters, &builder.Gte{f.field: f.value})
			}
		case taskFilterComparatorLess:
			if opts.filterIncludeNulls {
				filters = append(filters, builder.Or(&builder.Lt{f.field: f.value}, &builder.IsNull{f.field}))
			} else {
				filters = append(filters, &builder.Lt{f.field: f.value})
			}
		case taskFilterComparatorLessEquals:
			if opts.filterIncludeNulls {
				filters = append(filters, builder.Or(&builder.Lte{f.field: f.value}, &builder.IsNull{f.field}))
			} else {
				filters = append(filters, &builder.Lte{f.field: f.value})
			}
		}
	}

	// Then return all tasks for that lists
	query := x.NewSession().
		OrderBy(orderby)
	queryCount := x.NewSession()

	if len(opts.search) > 0 {
		// Postgres' is case sensitive by default.
		// To work around this, we're using ILIKE as opposed to normal LIKE statements.
		// ILIKE is preferred over LOWER(text) LIKE for performance reasons.
		// See https://stackoverflow.com/q/7005302/10924593
		// Seems okay to use that now, we may need to find a better solution overall in the future.
		if config.DatabaseType.GetString() == "postgres" {
			query = query.Where("title ILIKE ?", "%"+opts.search+"%")
			queryCount = queryCount.Where("title ILIKE ?", "%"+opts.search+"%")
		} else {
			query = query.Where("title LIKE ?", "%"+opts.search+"%")
			queryCount = queryCount.Where("title LIKE ?", "%"+opts.search+"%")
		}
	}

	if len(listIDs) > 0 {
		query = query.In("list_id", listIDs)
		queryCount = queryCount.In("list_id", listIDs)
	}

	if len(filters) > 0 {
		if opts.filterConcat == filterConcatOr {
			query = query.Where(builder.Or(filters...))
			queryCount = queryCount.Where(builder.Or(filters...))
		}
		if opts.filterConcat == filterConcatAnd {
			query = query.Where(builder.And(filters...))
			queryCount = queryCount.Where(builder.And(filters...))
		}
	}

	limit, start := getLimitFromPageIndex(opts.page, opts.perPage)

	if limit > 0 {
		query = query.Limit(limit, start)
	}

	tasks = []*Task{}
	err = query.Find(&tasks)
	if err != nil {
		return nil, 0, 0, err
	}

	totalItems, err = queryCount.
		Count(&Task{})
	if err != nil {
		return nil, 0, 0, err
	}

	return tasks, len(tasks), totalItems, nil
}

func getTasksForLists(lists []*List, opts *taskOptions) (tasks []*Task, resultCount int, totalItems int64, err error) {

	tasks, resultCount, totalItems, err = getRawTasksForLists(lists, opts)
	if err != nil {
		return nil, 0, 0, err
	}

	taskMap := make(map[int64]*Task, len(tasks))
	for _, t := range tasks {
		taskMap[t.ID] = t
	}

	err = addMoreInfoToTasks(taskMap)
	if err != nil {
		return nil, 0, 0, err
	}

	return tasks, resultCount, totalItems, err
}

// GetTaskByIDSimple returns a raw task without extra data by the task ID
func GetTaskByIDSimple(taskID int64) (task Task, err error) {
	if taskID < 1 {
		return Task{}, ErrTaskDoesNotExist{taskID}
	}

	return GetTaskSimple(&Task{ID: taskID})
}

// GetTaskSimple returns a raw task without extra data
func GetTaskSimple(t *Task) (task Task, err error) {
	task = *t
	exists, err := x.Get(&task)
	if err != nil {
		return Task{}, err
	}

	if !exists {
		return Task{}, ErrTaskDoesNotExist{t.ID}
	}
	return
}

// GetTasksByIDs returns all tasks for a list of ids
func (bt *BulkTask) GetTasksByIDs() (err error) {
	for _, id := range bt.IDs {
		if id < 1 {
			return ErrTaskDoesNotExist{id}
		}
	}

	err = x.In("id", bt.IDs).Find(&bt.Tasks)
	if err != nil {
		return
	}

	return
}

// GetTasksByUIDs gets all tasks from a bunch of uids
func GetTasksByUIDs(uids []string) (tasks []*Task, err error) {
	tasks = []*Task{}
	err = x.In("uid", uids).Find(&tasks)
	if err != nil {
		return
	}

	taskMap := make(map[int64]*Task, len(tasks))
	for _, t := range tasks {
		taskMap[t.ID] = t
	}

	err = addMoreInfoToTasks(taskMap)
	return
}

func getRemindersForTasks(taskIDs []int64) (reminders []*TaskReminder, err error) {
	reminders = []*TaskReminder{}
	err = x.In("task_id", taskIDs).Find(&reminders)
	return
}

// This function takes a map with pointers and returns a slice with pointers to tasks
// It adds more stuff like assignees/labels/etc to a bunch of tasks
func addMoreInfoToTasks(taskMap map[int64]*Task) (err error) {

	// No need to iterate over users and stuff if the list doesn't has tasks
	if len(taskMap) == 0 {
		return
	}

	// Get all users & task ids and put them into the array
	var userIDs []int64
	var taskIDs []int64
	var listIDs []int64
	for _, i := range taskMap {
		taskIDs = append(taskIDs, i.ID)
		userIDs = append(userIDs, i.CreatedByID)
		listIDs = append(listIDs, i.ListID)
	}

	// Get all assignees
	taskAssignees, err := getRawTaskAssigneesForTasks(taskIDs)
	if err != nil {
		return
	}
	// Put the assignees in the task map
	for _, a := range taskAssignees {
		if a != nil {
			a.Email = "" // Obfuscate the email
			taskMap[a.TaskID].Assignees = append(taskMap[a.TaskID].Assignees, &a.User)
		}
	}

	// Get all labels for all the tasks
	labels, _, _, err := getLabelsByTaskIDs(&LabelByTaskIDsOptions{
		TaskIDs: taskIDs,
		Page:    -1,
	})
	if err != nil {
		return
	}
	for _, l := range labels {
		if l != nil {
			taskMap[l.TaskID].Labels = append(taskMap[l.TaskID].Labels, &l.Label)
		}
	}

	// Get task attachments
	attachments, err := getTaskAttachmentsByTaskIDs(taskIDs)

	// Get all users of a task
	// aka the ones who created a task
	users := make(map[int64]*user.User)
	err = x.In("id", userIDs).Find(&users)
	if err != nil {
		return
	}

	// Obfuscate all user emails
	for _, u := range users {
		u.Email = ""
	}

	// Put the users and files in task attachments
	for _, a := range attachments {
		taskMap[a.TaskID].Attachments = append(taskMap[a.TaskID].Attachments, a)
	}

	// Get all reminders and put them in a map to have it easier later
	reminders, err := getRemindersForTasks(taskIDs)
	if err != nil {
		return
	}

	taskReminders := make(map[int64][]time.Time)
	for _, r := range reminders {
		taskReminders[r.TaskID] = append(taskReminders[r.TaskID], r.Reminder)
	}

	// Get all identifiers
	lists := make(map[int64]*List, len(listIDs))
	err = x.In("id", listIDs).Find(&lists)
	if err != nil {
		return
	}

	// Add all user objects to the appropriate tasks
	for _, task := range taskMap {

		// Make created by user objects
		task.CreatedBy = users[task.CreatedByID]

		// Add the reminders
		task.Reminders = taskReminders[task.ID]

		// Prepare the subtasks
		task.RelatedTasks = make(RelatedTaskMap)

		// Build the task identifier from the list identifier and task index
		task.Identifier = lists[task.ListID].Identifier + "-" + strconv.FormatInt(task.Index, 10)
	}

	// Get all related tasks
	relatedTasks := []*TaskRelation{}
	err = x.In("task_id", taskIDs).Find(&relatedTasks)
	if err != nil {
		return
	}

	// Collect all related task IDs, so we can get all related task headers in one go
	var relatedTaskIDs []int64
	for _, rt := range relatedTasks {
		relatedTaskIDs = append(relatedTaskIDs, rt.OtherTaskID)
	}
	fullRelatedTasks := make(map[int64]*Task)
	err = x.In("id", relatedTaskIDs).Find(&fullRelatedTasks)
	if err != nil {
		return
	}

	// NOTE: while it certainly be possible to run this function on	fullRelatedTasks again, we don't do this for performance reasons.

	// Go through all task relations and put them into the task objects
	for _, rt := range relatedTasks {
		taskMap[rt.TaskID].RelatedTasks[rt.RelationKind] = append(taskMap[rt.TaskID].RelatedTasks[rt.RelationKind], fullRelatedTasks[rt.OtherTaskID])
	}

	return
}

func checkBucketAndTaskBelongToSameList(fullTask *Task, bucketID int64) (err error) {
	if bucketID != 0 {
		b, err := getBucketByID(bucketID)
		if err != nil {
			return err
		}
		if fullTask.ListID != b.ListID {
			return ErrBucketDoesNotBelongToList{
				ListID:   fullTask.ListID,
				BucketID: fullTask.BucketID,
			}
		}
	}
	return
}

// Create is the implementation to create a list task
// @Summary Create a task
// @Description Inserts a task into a list.
// @tags task
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "List ID"
// @Param task body models.Task true "The task object"
// @Success 200 {object} models.Task "The created task object."
// @Failure 400 {object} web.HTTPError "Invalid task object provided."
// @Failure 403 {object} web.HTTPError "The user does not have access to the list"
// @Failure 500 {object} models.Message "Internal error"
// @Router /lists/{id} [put]
func (t *Task) Create(a web.Auth) (err error) {
	return createTask(t, a, true)
}

func createTask(t *Task, a web.Auth, updateAssignees bool) (err error) {

	t.ID = 0

	// Check if we have at least a text
	if t.Title == "" {
		return ErrTaskCannotBeEmpty{}
	}

	// Check if the list exists
	l := &List{ID: t.ListID}
	if err = l.GetSimpleByID(); err != nil {
		return
	}

	if _, is := a.(*LinkSharing); is {
		// A negative user id indicates user share links
		t.CreatedByID = a.GetID() * -1
	} else {
		u, err := user.GetUserByID(a.GetID())
		if err != nil {
			return err
		}
		t.CreatedByID = u.ID
		t.CreatedBy = u
	}

	// Generate a uuid if we don't already have one
	if t.UID == "" {
		t.UID = utils.MakeRandomString(40)
	}

	// If there is a bucket set, make sure they belong to the same list as the task
	err = checkBucketAndTaskBelongToSameList(t, t.BucketID)
	if err != nil {
		return
	}

	// Get the default bucket and move the task there
	if t.BucketID == 0 {
		defaultBucket, err := getDefaultBucket(t.ListID)
		if err != nil {
			return err
		}
		t.BucketID = defaultBucket.ID
	}

	// Get the index for this task
	latestTask := &Task{}
	_, err = x.Where("list_id = ?", t.ListID).OrderBy("id desc").Get(latestTask)
	if err != nil {
		return err
	}

	t.Index = latestTask.Index + 1
	// If no position was supplied, set a default one
	if t.Position == 0 {
		t.Position = float64(latestTask.ID+1) * math.Pow(2, 16)
	}
	if _, err = x.Insert(t); err != nil {
		return err
	}

	// Update the assignees
	if updateAssignees {
		if err := t.updateTaskAssignees(t.Assignees); err != nil {
			return err
		}
	}

	// Update the reminders
	if err := t.updateReminders(t.Reminders); err != nil {
		return err
	}

	metrics.UpdateCount(1, metrics.TaskCountKey)

	err = updateListLastUpdated(&List{ID: t.ListID})
	return
}

// Update updates a list task
// @Summary Update a task
// @Description Updates a task. This includes marking it as done. Assignees you pass will be updated, see their individual endpoints for more details on how this is done. To update labels, see the description of the endpoint.
// @tags task
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "Task ID"
// @Param task body models.Task true "The task object"
// @Success 200 {object} models.Task "The updated task object."
// @Failure 400 {object} web.HTTPError "Invalid task object provided."
// @Failure 403 {object} web.HTTPError "The user does not have access to the task (aka its list)"
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{id} [post]
func (t *Task) Update() (err error) {
	// Check if the task exists and get the old values
	ot, err := GetTaskByIDSimple(t.ID)
	if err != nil {
		return
	}

	// Get the reminders
	reminders, err := getRemindersForTasks([]int64{t.ID})
	if err != nil {
		return
	}

	ot.Reminders = make([]time.Time, len(reminders))
	for i, r := range reminders {
		ot.Reminders[i] = r.Reminder
	}

	// When a repeating task is marked as done, we update all deadlines and reminders and set it as undone
	updateDone(&ot, t)

	// Update the assignees
	if err := ot.updateTaskAssignees(t.Assignees); err != nil {
		return err
	}

	// Update the reminders
	if err := ot.updateReminders(t.Reminders); err != nil {
		return err
	}

	// If there is a bucket set, make sure they belong to the same list as the task
	err = checkBucketAndTaskBelongToSameList(&ot, t.BucketID)
	if err != nil {
		return
	}

	// If the task is being moved between lists, make sure to move the bucket as well
	if t.ListID != 0 && ot.ListID != t.ListID {
		b, err := getDefaultBucket(t.ListID)
		if err != nil {
			return err
		}
		t.BucketID = b.ID
	}

	// Update the labels
	//
	// Maybe FIXME:
	// I've disabled this for now, because it requires significant changes in the way we do updates (using the
	// Update() function. We need a user object in updateTaskLabels to check if the user has the right to see
	// the label it is currently adding. To do this, we'll need to update the webhandler to let it pass the current
	// user object (like it's already the case with the create method). However when we change it, that'll break
	// a lot of existing code which we'll then need to refactor.
	// This is why.
	//
	//if err := ot.updateTaskLabels(t.Labels); err != nil {
	//	return err
	//}
	// set the labels to ot.Labels because our updateTaskLabels function puts the full label objects in it pretty nicely
	// We also set this here to prevent it being overwritten later on.
	//t.Labels = ot.Labels

	// For whatever reason, xorm dont detect if done is updated, so we need to update this every time by hand
	// Which is why we merge the actual task struct with the one we got from the db
	// The user struct overrides values in the actual one.
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
	// Position
	if t.Position == 0 {
		ot.Position = 0
	}
	// Repeat from current date
	if !t.RepeatFromCurrentDate {
		ot.RepeatFromCurrentDate = false
	}

	_, err = x.ID(t.ID).
		Cols("title",
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
			"list_id",
			"bucket_id",
			"position",
			"repeat_from_current_date",
		).
		Update(ot)
	*t = ot
	if err != nil {
		return err
	}

	err = updateListLastUpdated(&List{ID: t.ListID})
	return
}

// This helper function updates the reminders, doneAt, start and end dates of the *old* task
// and saves the new values in the newTask object.
// We make a few assumtions here:
//   1. Everything in oldTask is the truth - we figure out if we update anything at all if oldTask.RepeatAfter has a value > 0
//   2. Because of 1., this functions should not be used to update values other than Done in the same go
func updateDone(oldTask *Task, newTask *Task) {
	if !oldTask.Done && newTask.Done && oldTask.RepeatAfter > 0 {

		repeatDuration := time.Duration(oldTask.RepeatAfter) * time.Second

		// Current time in an extra variable to base all calculations on the same time
		now := time.Now()

		// assuming we'll merge the new task over the old task
		if !oldTask.DueDate.IsZero() {
			if oldTask.RepeatFromCurrentDate {
				newTask.DueDate = now.Add(repeatDuration)
			} else {
				// Always add one instance of the repeating interval to catch cases where a due date is already in the future
				// but not the repeating interval
				newTask.DueDate = oldTask.DueDate.Add(repeatDuration)
				// Add the repeating interval until the new due date is in the future
				for !newTask.DueDate.After(now) {
					newTask.DueDate = newTask.DueDate.Add(repeatDuration)
				}
			}
		}

		newTask.Reminders = oldTask.Reminders
		// When repeating from the current date, all reminders should keep their difference to each other.
		// To make this easier, we sort them first because we can then rely on the fact the first is the smallest
		if len(oldTask.Reminders) > 0 {
			if oldTask.RepeatFromCurrentDate {
				sort.Slice(oldTask.Reminders, func(i, j int) bool {
					return oldTask.Reminders[i].Unix() < oldTask.Reminders[j].Unix()
				})
				first := oldTask.Reminders[0]
				for in, r := range oldTask.Reminders {
					diff := r.Sub(first)
					newTask.Reminders[in] = now.Add(repeatDuration + diff)
				}
			} else {
				for in, r := range oldTask.Reminders {
					newTask.Reminders[in] = r.Add(repeatDuration)
					for !newTask.Reminders[in].After(now) {
						newTask.Reminders[in] = newTask.Reminders[in].Add(repeatDuration)
					}
				}
			}
		}

		// If a task has a start and end date, the end date should keep the difference to the start date when setting them as new
		if oldTask.RepeatFromCurrentDate && !oldTask.StartDate.IsZero() && !oldTask.EndDate.IsZero() {
			diff := oldTask.EndDate.Sub(oldTask.StartDate)
			newTask.StartDate = now.Add(repeatDuration)
			newTask.EndDate = now.Add(repeatDuration + diff)
		} else {
			if !oldTask.StartDate.IsZero() {
				if oldTask.RepeatFromCurrentDate {
					newTask.StartDate = now.Add(repeatDuration)
				} else {
					newTask.StartDate = oldTask.StartDate.Add(repeatDuration)
					for !newTask.StartDate.After(now) {
						newTask.StartDate = newTask.StartDate.Add(repeatDuration)
					}
				}
			}

			if !oldTask.EndDate.IsZero() {
				if oldTask.RepeatFromCurrentDate {
					newTask.EndDate = now.Add(repeatDuration)
				} else {
					newTask.EndDate = oldTask.EndDate.Add(repeatDuration)
					for !newTask.EndDate.After(now) {
						newTask.EndDate = newTask.EndDate.Add(repeatDuration)
					}
				}
			}
		}

		newTask.Done = false
	}

	// Update the "done at" timestamp
	if !oldTask.Done && newTask.Done {
		newTask.DoneAt = time.Now()
	}
	// When unmarking a task as done, reset the timestamp
	if oldTask.Done && !newTask.Done {
		newTask.DoneAt = time.Time{}
	}
}

// Creates or deletes all necessary reminders without unneded db operations.
// The parameter is a slice with unix dates which holds the new reminders.
func (t *Task) updateReminders(reminders []time.Time) (err error) {

	// Load the current reminders
	taskReminders, err := getRemindersForTasks([]int64{t.ID})
	if err != nil {
		return err
	}

	t.Reminders = make([]time.Time, 0, len(taskReminders))
	for _, reminder := range taskReminders {
		t.Reminders = append(t.Reminders, reminder.Reminder)
	}

	// If we're removing everything, delete all reminders right away
	if len(reminders) == 0 && len(t.Reminders) > 0 {
		_, err = x.Where("task_id = ?", t.ID).
			Delete(TaskReminder{})
		t.Reminders = nil
		return err
	}

	// If we didn't change anything (from 0 to zero) don't do anything.
	if len(reminders) == 0 && len(t.Reminders) == 0 {
		return nil
	}

	// Make a hashmap of the new reminders for easier comparison
	newReminders := make(map[time.Time]*TaskReminder, len(reminders))
	for _, newReminder := range reminders {
		newReminders[newReminder] = &TaskReminder{Reminder: newReminder}
	}

	// Get old reminders to delete
	var found bool
	var remindersToDelete []time.Time
	oldReminders := make(map[time.Time]*TaskReminder, len(t.Reminders))
	for _, oldReminder := range t.Reminders {
		found = false
		// If a new reminder is already in the list with old reminders
		if newReminders[oldReminder] != nil {
			found = true
		}

		// Put all reminders which are only on the old list to the trash
		if !found {
			remindersToDelete = append(remindersToDelete, oldReminder)
		}

		oldReminders[oldReminder] = &TaskReminder{Reminder: oldReminder}
	}

	// Delete all reminders not passed
	if len(remindersToDelete) > 0 {
		_, err = x.In("reminder", remindersToDelete).
			And("task_id = ?", t.ID).
			Delete(TaskReminder{})
		if err != nil {
			return err
		}
	}

	// Loop through our reminders and add them
	for _, r := range reminders {
		// Check if the reminder already exists and only inserts it if not
		if oldReminders[r] != nil {
			// continue outer loop
			continue
		}

		// Add the new reminder
		_, err = x.Insert(TaskReminder{TaskID: t.ID, Reminder: r})
		if err != nil {
			return err
		}
	}

	t.Reminders = reminders
	if len(reminders) == 0 {
		t.Reminders = nil
	}

	err = updateListLastUpdated(&List{ID: t.ListID})
	return
}

// Delete implements the delete method for listTask
// @Summary Delete a task
// @Description Deletes a task from a list. This does not mean "mark it done".
// @tags task
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "Task ID"
// @Success 200 {object} models.Message "The created task object."
// @Failure 400 {object} web.HTTPError "Invalid task ID provided."
// @Failure 403 {object} web.HTTPError "The user does not have access to the list"
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{id} [delete]
func (t *Task) Delete() (err error) {

	if _, err = x.ID(t.ID).Delete(Task{}); err != nil {
		return err
	}

	// Delete assignees
	if _, err = x.Where("task_id = ?", t.ID).Delete(TaskAssginee{}); err != nil {
		return err
	}

	metrics.UpdateCount(-1, metrics.TaskCountKey)

	err = updateListLastUpdated(&List{ID: t.ListID})
	return
}

// ReadOne gets one task by its ID
// @Summary Get one task
// @Description Returns one task by its ID
// @tags task
// @Accept json
// @Produce json
// @Param ID path int true "The task ID"
// @Security JWTKeyAuth
// @Success 200 {object} models.Task "The task"
// @Failure 404 {object} models.Message "Task not found"
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{ID} [get]
func (t *Task) ReadOne() (err error) {

	taskMap := make(map[int64]*Task, 1)
	taskMap[t.ID] = &Task{}
	*taskMap[t.ID], err = GetTaskByIDSimple(t.ID)
	if err != nil {
		return
	}

	err = addMoreInfoToTasks(taskMap)
	if err != nil {
		return
	}

	if len(taskMap) == 0 {
		return ErrTaskDoesNotExist{t.ID}
	}

	*t = *taskMap[t.ID]

	return
}
