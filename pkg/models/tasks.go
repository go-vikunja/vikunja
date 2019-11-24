//  Vikunja is a todo-list application to facilitate your life.
//  Copyright 2018 Vikunja and contributors. All rights reserved.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU General Public License as published by
//  the Free Software Foundation, either version 3 of the License, or
//  (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU General Public License for more details.
//
//  You should have received a copy of the GNU General Public License
//  along with this program.  If not, see <https://www.gnu.org/licenses/>.

package models

import (
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/metrics"
	"code.vikunja.io/api/pkg/utils"
	"code.vikunja.io/web"
	"github.com/imdario/mergo"
	"sort"
	"time"
)

// Task represents an task in a todolist
type Task struct {
	// The unique, numeric id of this task.
	ID int64 `xorm:"int(11) autoincr not null unique pk" json:"id" param:"listtask"`
	// The task text. This is what you'll see in the list.
	Text string `xorm:"varchar(250) not null" json:"text" valid:"runelength(3|250)" minLength:"3" maxLength:"250"`
	// The task description.
	Description string `xorm:"longtext null" json:"description"`
	// Whether a task is done or not.
	Done bool `xorm:"INDEX null" json:"done"`
	// The unix timestamp when a task was marked as done.
	DoneAtUnix int64 `xorm:"INDEX null" json:"doneAt"`
	// A unix timestamp when the task is due.
	DueDateUnix int64 `xorm:"int(11) INDEX null" json:"dueDate"`
	// An array of unix timestamps when the user wants to be reminded of the task.
	RemindersUnix []int64 `xorm:"-" json:"reminderDates"`
	CreatedByID   int64   `xorm:"int(11) not null" json:"-"` // ID of the user who put that task on the list
	// The list this task belongs to.
	ListID int64 `xorm:"int(11) INDEX not null" json:"listID" param:"list"`
	// An amount in seconds this task repeats itself. If this is set, when marking the task as done, it will mark itself as "undone" and then increase all remindes and the due date by its amount.
	RepeatAfter int64 `xorm:"int(11) INDEX null" json:"repeatAfter"`
	// The task priority. Can be anything you want, it is possible to sort by this later.
	Priority int64 `xorm:"int(11) null" json:"priority"`
	// When this task starts.
	StartDateUnix int64 `xorm:"int(11) INDEX null" json:"startDate" query:"-"`
	// When this task ends.
	EndDateUnix int64 `xorm:"int(11) INDEX null" json:"endDate" query:"-"`
	// An array of users who are assigned to this task
	Assignees []*User `xorm:"-" json:"assignees"`
	// An array of labels which are associated with this task.
	Labels []*Label `xorm:"-" json:"labels"`
	// The task color in hex
	HexColor string `xorm:"varchar(6) null" json:"hexColor" valid:"runelength(0|6)" maxLength:"6"`
	// Determines how far a task is left from being done
	PercentDone float64 `xorm:"DOUBLE null" json:"percentDone"`

	// The UID is currently not used for anything other than caldav, which is why we don't expose it over json
	UID string `xorm:"varchar(250) null" json:"-"`

	Sorting           string `xorm:"-" json:"-" query:"sort"` // Parameter to sort by
	StartDateSortUnix int64  `xorm:"-" json:"-" query:"startdate"`
	EndDateSortUnix   int64  `xorm:"-" json:"-" query:"enddate"`

	// All related tasks, grouped by their relation kind
	RelatedTasks RelatedTaskMap `xorm:"-" json:"related_tasks"`

	// All attachments this task has
	Attachments []*TaskAttachment `xorm:"-" json:"attachments"`

	// A unix timestamp when this task was created. You cannot change this value.
	Created int64 `xorm:"created not null" json:"created"`
	// A unix timestamp when this task was last updated. You cannot change this value.
	Updated int64 `xorm:"updated not null" json:"updated"`

	// The user who initially created the task.
	CreatedBy *User `xorm:"-" json:"createdBy" valid:"-"`

	web.CRUDable `xorm:"-" json:"-"`
	web.Rights   `xorm:"-" json:"-"`
}

// TableName returns the table name for listtasks
func (Task) TableName() string {
	return "tasks"
}

// TaskReminder holds a reminder on a task
type TaskReminder struct {
	ID           int64 `xorm:"int(11) autoincr not null unique pk"`
	TaskID       int64 `xorm:"int(11) not null INDEX"`
	ReminderUnix int64 `xorm:"int(11) not null INDEX"`
	Created      int64 `xorm:"created not null"`
}

// TableName returns a pretty table name
func (TaskReminder) TableName() string {
	return "task_reminders"
}

// SortBy declares constants to sort
type SortBy int

// These are possible sort options
const (
	SortTasksByUnsorted   SortBy = -1
	SortTasksByDueDateAsc        = iota
	SortTasksByDueDateDesc
	SortTasksByPriorityAsc
	SortTasksByPriorityDesc
)

// ReadAll gets all tasks for a user
// @Summary Get tasks
// @Description Returns all tasks on any list the user has access to.
// @tags task
// @Accept json
// @Produce json
// @Param page query int false "The page number. Used for pagination. If not provided, the first page of results is returned."
// @Param per_page query int false "The maximum number of items per page. Note this parameter is limited by the configured maximum of items per page."
// @Param s query string false "Search tasks by task text."
// @Param sort query string false "The sorting parameter. Possible values to sort by are priority, prioritydesc, priorityasc, duedate, duedatedesc, duedateasc."
// @Param startdate query int false "The start date parameter to filter by. Expects a unix timestamp. If no end date, but a start date is specified, the end date is set to the current time."
// @Param enddate query int false "The end date parameter to filter by. Expects a unix timestamp. If no start date, but an end date is specified, the start date is set to the current time."
// @Security JWTKeyAuth
// @Success 200 {array} models.Task "The tasks"
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/all [get]
func (t *Task) ReadAll(a web.Auth, search string, page int, perPage int) (result interface{}, resultCount int, totalItems int64, err error) {
	var sortby SortBy
	switch t.Sorting {
	case "priority":
		sortby = SortTasksByPriorityDesc
	case "prioritydesc":
		sortby = SortTasksByPriorityDesc
	case "priorityasc":
		sortby = SortTasksByPriorityAsc
	case "duedate":
		sortby = SortTasksByDueDateDesc
	case "duedatedesc":
		sortby = SortTasksByDueDateDesc
	case "duedateasc":
		sortby = SortTasksByDueDateAsc
	default:
		sortby = SortTasksByUnsorted
	}

	taskopts := &taskOptions{
		search:    search,
		sortby:    sortby,
		startDate: time.Unix(t.StartDateSortUnix, 0),
		endDate:   time.Unix(t.EndDateSortUnix, 0),
		page:      page,
		perPage:   perPage,
	}

	shareAuth, is := a.(*LinkSharing)
	if is {
		list := &List{ID: shareAuth.ListID}
		err := list.GetSimpleByID()
		if err != nil {
			return nil, 0, 0, err
		}
		return getTasksForLists([]*List{list}, taskopts)
	}

	// Get all lists for the user
	lists, _, _, err := getRawListsForUser("", &User{ID: a.GetID()}, -1, 0)
	if err != nil {
		return nil, 0, 0, err
	}

	return getTasksForLists(lists, taskopts)
}

type taskOptions struct {
	search    string
	sortby    SortBy
	startDate time.Time
	endDate   time.Time
	page      int
	perPage   int
}

func getRawTasksForLists(lists []*List, opts *taskOptions) (taskMap map[int64]*Task, resultCount int, totalItems int64, err error) {

	// Get all list IDs and get the tasks
	var listIDs []int64
	for _, l := range lists {
		listIDs = append(listIDs, l.ID)
	}

	var orderby string
	switch opts.sortby {
	case SortTasksByPriorityDesc:
		orderby = "priority desc"
	case SortTasksByPriorityAsc:
		orderby = "priority asc"
	case SortTasksByDueDateDesc:
		orderby = "due_date_unix desc"
	case SortTasksByDueDateAsc:
		orderby = "due_date_unix asc"
	}

	taskMap = make(map[int64]*Task)

	// Then return all tasks for that lists
	if opts.startDate.Unix() != 0 || opts.endDate.Unix() != 0 {

		startDateUnix := time.Now().Unix()
		if opts.startDate.Unix() != 0 {
			startDateUnix = opts.startDate.Unix()
		}

		endDateUnix := time.Now().Unix()
		if opts.endDate.Unix() != 0 {
			endDateUnix = opts.endDate.Unix()
		}

		err := x.In("list_id", listIDs).
			Where("text LIKE ?", "%"+opts.search+"%").
			And("((due_date_unix BETWEEN ? AND ?) OR "+
				"(start_date_unix BETWEEN ? and ?) OR "+
				"(end_date_unix BETWEEN ? and ?))", startDateUnix, endDateUnix, startDateUnix, endDateUnix, startDateUnix, endDateUnix).
			OrderBy(orderby).
			Limit(getLimitFromPageIndex(opts.page, opts.perPage)).
			Find(&taskMap)
		if err != nil {
			return nil, 0, 0, err
		}

		totalItems, err = x.In("list_id", listIDs).
			Where("text LIKE ?", "%"+opts.search+"%").
			And("((due_date_unix BETWEEN ? AND ?) OR "+
				"(start_date_unix BETWEEN ? and ?) OR "+
				"(end_date_unix BETWEEN ? and ?))", startDateUnix, endDateUnix, startDateUnix, endDateUnix, startDateUnix, endDateUnix).
			Count(&Task{})
		if err != nil {
			return nil, 0, 0, err
		}
	} else {
		err := x.In("list_id", listIDs).
			Where("text LIKE ?", "%"+opts.search+"%").
			OrderBy(orderby).
			Limit(getLimitFromPageIndex(opts.page, opts.perPage)).
			Find(&taskMap)
		if err != nil {
			return nil, 0, 0, err
		}
		totalItems, err = x.In("list_id", listIDs).
			Where("text LIKE ?", "%"+opts.search+"%").
			Count(&Task{})
		if err != nil {
			return nil, 0, 0, err
		}
	}
	return taskMap, len(taskMap), totalItems, nil
}

func getTasksForLists(lists []*List, opts *taskOptions) (tasks []*Task, resultCount int, totalItems int64, err error) {

	taskMap, resultCount, totalItems, err := getRawTasksForLists(lists, opts)
	if err != nil {
		return nil, 0, 0, err
	}

	tasks, err = addMoreInfoToTasks(taskMap)
	if err != nil {
		return nil, 0, 0, err
	}
	// Because the list is sorted by id which we don't want (since we're dealing with maps)
	// we have to manually sort the tasks again here.
	sortTasks(tasks, opts.sortby)

	return tasks, resultCount, totalItems, err
}

func sortTasks(tasks []*Task, by SortBy) {
	switch by {
	case SortTasksByPriorityDesc:
		sort.Slice(tasks, func(i, j int) bool {
			return tasks[i].Priority > tasks[j].Priority
		})
	case SortTasksByPriorityAsc:
		sort.Slice(tasks, func(i, j int) bool {
			return tasks[i].Priority < tasks[j].Priority
		})
	case SortTasksByDueDateDesc:
		sort.Slice(tasks, func(i, j int) bool {
			return tasks[i].DueDateUnix > tasks[j].DueDateUnix
		})
	case SortTasksByDueDateAsc:
		sort.Slice(tasks, func(i, j int) bool {
			return tasks[i].DueDateUnix < tasks[j].DueDateUnix
		})
	}
}

// GetTasksByListID gets all todotasks for a list
func GetTasksByListID(listID int64) (tasks []*Task, err error) {
	// make a map so we can put in a lot of other stuff more easily
	taskMap := make(map[int64]*Task, len(tasks))
	err = x.Where("list_id = ?", listID).Find(&taskMap)
	if err != nil {
		return
	}

	tasks, err = addMoreInfoToTasks(taskMap)
	return
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

	taskMap := make(map[int64]*Task, len(bt.Tasks))
	err = x.In("id", bt.IDs).Find(&taskMap)
	if err != nil {
		return
	}

	bt.Tasks, err = addMoreInfoToTasks(taskMap)
	return
}

// GetTasksByUIDs gets all tasks from a bunch of uids
func GetTasksByUIDs(uids []string) (tasks []*Task, err error) {
	taskMap := make(map[int64]*Task)
	err = x.In("uid", uids).Find(&taskMap)
	if err != nil {
		return
	}

	tasks, err = addMoreInfoToTasks(taskMap)
	return
}

func getRemindersForTasks(taskIDs []int64) (reminders []*TaskReminder, err error) {
	reminders = []*TaskReminder{}
	err = x.Table("task_reminders").In("task_id", taskIDs).Find(&reminders)
	return
}

// This function takes a map with pointers and returns a slice with pointers to tasks
// It adds more stuff like assignees/labels/etc to a bunch of tasks
func addMoreInfoToTasks(taskMap map[int64]*Task) (tasks []*Task, err error) {

	// No need to iterate over users and stuff if the list doesn't has tasks
	if len(taskMap) == 0 {
		return
	}

	// Get all users & task ids and put them into the array
	var userIDs []int64
	var taskIDs []int64
	for _, i := range taskMap {
		taskIDs = append(taskIDs, i.ID)
		userIDs = append(userIDs, i.CreatedByID)
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
	attachments := []*TaskAttachment{}
	err = x.
		In("task_id", taskIDs).
		Find(&attachments)
	if err != nil {
		return nil, err
	}

	fileIDs := []int64{}
	for _, a := range attachments {
		userIDs = append(userIDs, a.CreatedByID)
		fileIDs = append(fileIDs, a.FileID)
	}

	// Get all files
	fs := make(map[int64]*files.File)
	err = x.In("id", fileIDs).Find(&fs)
	if err != nil {
		return
	}

	// Get all users of a task
	// aka the ones who created a task
	users := make(map[int64]*User)
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
		a.CreatedBy = users[a.CreatedByID]
		a.File = fs[a.FileID]
		taskMap[a.TaskID].Attachments = append(taskMap[a.TaskID].Attachments, a)
	}

	// Get all reminders and put them in a map to have it easier later
	reminders, err := getRemindersForTasks(taskIDs)
	if err != nil {
		return
	}

	taskRemindersUnix := make(map[int64][]int64)
	for _, r := range reminders {
		taskRemindersUnix[r.TaskID] = append(taskRemindersUnix[r.TaskID], r.ReminderUnix)
	}

	// Add all user objects to the appropriate tasks
	for _, task := range taskMap {

		// Make created by user objects
		task.CreatedBy = users[task.CreatedByID]

		// Add the reminders
		task.RemindersUnix = taskRemindersUnix[task.ID]

		// Prepare the subtasks
		task.RelatedTasks = make(RelatedTaskMap)
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

	// make a complete slice from the map
	tasks = []*Task{}
	for _, t := range taskMap {
		tasks = append(tasks, t)
	}

	// Sort the output. In Go, contents on a map are put on that map in no particular order.
	// Because of this, tasks are not sorted anymore in the output, this leads to confiusion.
	// To avoid all this, we need to sort the slice afterwards
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].ID < tasks[j].ID
	})

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
// @Failure 400 {object} code.vikunja.io/web.HTTPError "Invalid task object provided."
// @Failure 403 {object} code.vikunja.io/web.HTTPError "The user does not have access to the list"
// @Failure 500 {object} models.Message "Internal error"
// @Router /lists/{id} [put]
func (t *Task) Create(a web.Auth) (err error) {

	t.ID = 0

	// Check if we have at least a text
	if t.Text == "" {
		return ErrTaskCannotBeEmpty{}
	}

	// Check if the list exists
	l := &List{ID: t.ListID}
	if err = l.GetSimpleByID(); err != nil {
		return
	}

	u, err := GetUserByID(a.GetID())
	if err != nil {
		return err
	}

	// Generate a uuid if we don't already have one
	if t.UID == "" {
		t.UID = utils.MakeRandomString(40)
	}

	t.CreatedByID = u.ID
	t.CreatedBy = u
	if _, err = x.Insert(t); err != nil {
		return err
	}

	// Update the assignees
	if err := t.updateTaskAssignees(t.Assignees); err != nil {
		return err
	}

	// Update the reminders
	if err := t.updateReminders(t.RemindersUnix); err != nil {
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
// @Failure 400 {object} code.vikunja.io/web.HTTPError "Invalid task object provided."
// @Failure 403 {object} code.vikunja.io/web.HTTPError "The user does not have access to the task (aka its list)"
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{id} [post]
func (t *Task) Update() (err error) {
	// Check if the task exists
	ot, err := GetTaskByIDSimple(t.ID)
	if err != nil {
		return
	}

	// When a repeating task is marked as done, we update all deadlines and reminders and set it as undone
	updateDone(&ot, t)

	// Update the assignees
	if err := ot.updateTaskAssignees(t.Assignees); err != nil {
		return err
	}

	// Update the reminders
	if err := ot.updateReminders(t.RemindersUnix); err != nil {
		return err
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
	// Which is why we merge the actual task struct with the one we got from the
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
	if t.DueDateUnix == 0 {
		ot.DueDateUnix = 0
	}
	// Repeat after
	if t.RepeatAfter == 0 {
		ot.RepeatAfter = 0
	}
	// Start date
	if t.StartDateUnix == 0 {
		ot.StartDateUnix = 0
	}
	// End date
	if t.EndDateUnix == 0 {
		ot.EndDateUnix = 0
	}
	// Color
	if t.HexColor == "" {
		ot.HexColor = ""
	}
	// Percent DOnw
	if t.PercentDone == 0 {
		ot.PercentDone = 0
	}

	_, err = x.ID(t.ID).
		Cols("text",
			"description",
			"done",
			"due_date_unix",
			"repeat_after",
			"priority",
			"start_date_unix",
			"end_date_unix",
			"hex_color",
			"done_at_unix",
			"percent_done").
		Update(ot)
	*t = ot
	if err != nil {
		return err
	}

	err = updateListLastUpdated(&List{ID: t.ListID})
	return
}

// This helper function updates the reminders and doneAtUnix of the *old* task (since that's the one we're inserting
// with updated values into the db)
func updateDone(oldTask *Task, newTask *Task) {
	if !oldTask.Done && newTask.Done && oldTask.RepeatAfter > 0 {
		oldTask.DueDateUnix = oldTask.DueDateUnix + oldTask.RepeatAfter // assuming we'll save the old task (merged)

		for in, r := range oldTask.RemindersUnix {
			oldTask.RemindersUnix[in] = r + oldTask.RepeatAfter
		}

		newTask.Done = false
	}

	// Update the "done at" timestamp
	if !oldTask.Done && newTask.Done {
		oldTask.DoneAtUnix = time.Now().Unix()
	}
	// When unmarking a task as done, reset the timestamp
	if oldTask.Done && !newTask.Done {
		oldTask.DoneAtUnix = 0
	}
}

// Creates or deletes all necessary remindes without unneded db operations.
// The parameter is a slice with unix dates which holds the new reminders.
func (t *Task) updateReminders(reminders []int64) (err error) {

	// Load the current reminders
	taskReminders, err := getRemindersForTasks([]int64{t.ID})
	if err != nil {
		return err
	}

	t.RemindersUnix = make([]int64, 0, len(taskReminders))
	for _, reminder := range taskReminders {
		t.RemindersUnix = append(t.RemindersUnix, reminder.ReminderUnix)
	}

	// If we're removing everything, delete all reminders right away
	if len(reminders) == 0 && len(t.RemindersUnix) > 0 {
		_, err = x.Where("task_id = ?", t.ID).
			Delete(TaskReminder{})
		t.RemindersUnix = nil
		return err
	}

	// If we didn't change anything (from 0 to zero) don't do anything.
	if len(reminders) == 0 && len(t.RemindersUnix) == 0 {
		return nil
	}

	// Make a hashmap of the new reminders for easier comparison
	newReminders := make(map[int64]*TaskReminder, len(reminders))
	for _, newReminder := range reminders {
		newReminders[newReminder] = &TaskReminder{ReminderUnix: newReminder}
	}

	// Get old reminders to delete
	var found bool
	var remindersToDelete []int64
	oldReminders := make(map[int64]*TaskReminder, len(t.RemindersUnix))
	for _, oldReminder := range t.RemindersUnix {
		found = false
		// If a new reminder is already in the list with old reminders
		if newReminders[oldReminder] != nil {
			found = true
		}

		// Put all reminders which are only on the old list to the trash
		if !found {
			remindersToDelete = append(remindersToDelete, oldReminder)
		}

		oldReminders[oldReminder] = &TaskReminder{ReminderUnix: oldReminder}
	}

	// Delete all reminders not passed
	if len(remindersToDelete) > 0 {
		_, err = x.In("reminder_unix", remindersToDelete).
			And("task_id = ?", t.ID).
			Delete(TaskReminder{})
		if err != nil {
			return err
		}
	}

	// Loop through our users and add them
	for _, r := range reminders {
		// Check if the reminder already exists and only inserts it if not
		if oldReminders[r] != nil {
			// continue outer loop
			continue
		}

		// Add the new reminder
		_, err = x.Insert(TaskReminder{TaskID: t.ID, ReminderUnix: r})
		if err != nil {
			return err
		}
	}

	t.RemindersUnix = reminders
	if len(reminders) == 0 {
		t.RemindersUnix = nil
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
// @Failure 400 {object} code.vikunja.io/web.HTTPError "Invalid task ID provided."
// @Failure 403 {object} code.vikunja.io/web.HTTPError "The user does not have access to the list"
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
// @Router /tasks/all [get]
func (t *Task) ReadOne() (err error) {

	taskMap := make(map[int64]*Task, 1)
	taskMap[t.ID] = &Task{}
	*taskMap[t.ID], err = GetTaskByIDSimple(t.ID)
	if err != nil {
		return
	}

	tasks, err := addMoreInfoToTasks(taskMap)
	if err != nil {
		return
	}

	if len(tasks) == 0 {
		return ErrTaskDoesNotExist{t.ID}
	}

	*t = *tasks[0]

	return
}
