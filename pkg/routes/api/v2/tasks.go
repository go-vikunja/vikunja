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

package apiv2

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/humabridge"
	"code.vikunja.io/api/pkg/web"
	"code.vikunja.io/api/pkg/web/handler"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/conditional"
	"xorm.io/xorm"
)

// expandDoc lists the accepted expand values; shared between the by-id and
// by-index operations so the docs stay in sync.
const expandDoc = "Embed extra, more expensive data in each task. Repeatable. One of: subtasks, buckets, reactions, comments, comment_count, time_entries_count, is_unread. Expanding can return more tasks than the page limit (subtasks) and inflate the response."

// parseTaskExpand turns the raw `expand` query values into validated
// TaskCollectionExpandable entries. Kept package-level for the TaskCollection
// list endpoint, which accepts the same option. An invalid value returns the
// model's own validation error, which translateDomainError maps to 422.
func parseTaskExpand(raw []string) ([]models.TaskCollectionExpandable, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	expand := make([]models.TaskCollectionExpandable, 0, len(raw))
	for _, e := range raw {
		v := models.TaskCollectionExpandable(e)
		if err := v.Validate(); err != nil {
			return nil, err
		}
		expand = append(expand, v)
	}
	return expand, nil
}

// RegisterTaskRoutes wires Task CRUD onto the Huma API. The list lives on
// TaskCollection, not here.
func RegisterTaskRoutes(api huma.API) {
	tags := []string{"tasks"}

	Register(api, huma.Operation{
		OperationID: "tasks-read",
		Summary:     "Get a task",
		Description: "Returns a single task by its numeric id. Sends an ETag; pass it as If-None-Match on a later read to get a 304 Not Modified. " + expandDoc,
		Method:      "GET",
		Path:        "/tasks/{task}",
		Tags:        tags,
	}, tasksRead)

	Register(api, huma.Operation{
		OperationID: "tasks-read-by-index",
		Summary:     "Get a task by its project index",
		Description: "Returns a single task addressed by its per-project index. Historical addresses redirect. The {project} segment accepts either a numeric project id or a textual project identifier (e.g. \"PROJ\"); a value made solely of digits is always treated as an id. " + expandDoc,
		Method:      "GET",
		Path:        "/projects/{project}/tasks/by-index/{index}",
		Tags:        tags,
		Responses: map[string]*huma.Response{
			"default": defaultErrorResponse(api),
			"307": {
				Description: "Historical address; Location holds the task's current one.",
				Headers: map[string]*huma.Param{
					"Location": {
						Description: "Current v2 by-index address, query string preserved.",
						Schema:      &huma.Schema{Type: huma.TypeString, Format: "uri-reference"},
					},
					"Cache-Control": {
						Description: "no-store",
						Schema:      &huma.Schema{Type: huma.TypeString},
					},
				},
			},
		},
	}, tasksReadByIndex)

	Register(api, huma.Operation{
		OperationID: "tasks-create",
		Summary:     "Create a task",
		Description: "Creates a task in the project from the URL. The authenticated user needs write access to that project and becomes the task's creator.",
		Method:      "POST",
		Path:        "/projects/{project}/tasks",
		Tags:        tags,
	}, tasksCreate)

	Register(api, huma.Operation{
		OperationID: "tasks-update",
		Summary:     "Update a task",
		Description: "Replaces all of a task's fields; requires write access. Setting project_id to a different project moves the task and also requires write access to the target project. Use PATCH for a partial update.",
		Method:      "PUT",
		Path:        "/tasks/{task}",
		Tags:        tags,
	}, tasksUpdate)

	Register(api, huma.Operation{
		OperationID: "tasks-delete",
		Summary:     "Delete a task",
		Description: "Moves a task to the trash. Requires write access to its project. Trashed tasks can be restored for 30 days before they are deleted permanently.",
		Method:      "DELETE",
		Path:        "/tasks/{task}",
		Tags:        tags,
	}, tasksDelete)
}

func init() { AddRouteRegistrar(RegisterTaskRoutes) }

// repeats_rrule is declared on each v2 single-task body (read, create, write) rather than in a
// shared embedded struct: Huma builds its $schema wrapper with reflect.StructOf, which panics on
// an embedded single-pointer struct (TestRegisterAllCreatesAllSchemaLinks). It is never on
// models.Task, so v1's wire shape and the list and bulk endpoints are unchanged.
// TestTaskRuleTextSchema_OpenAPI keeps the three doc tags identical.

// ruleTextOf is the one fill helper for repeats_rrule on every v2 single-task response: the
// stored rule, or nil (absent) when the task does not repeat.
func ruleTextOf(task *models.Task) *string {
	if task.Repeats == "" {
		return nil
	}
	text := task.Repeats
	return &text
}

// applyRuleText is the one inbound helper: it hands a v2 body's repeats_rrule to the model as
// Task.RepeatsText (nil when absent, "" to clear), which resolveV2Repeat maps onto the write
// path. It never writes Task.Repeats, the trusted raw channel CalDAV, import and duplication
// use: copying the text there would skip the rule-text validation and precedence.
func applyRuleText(task *models.Task, text *string) {
	task.RepeatsText = text
}

type taskReadOneBody struct {
	models.Task
	MaxPermission  models.Permission `json:"max_permission" readOnly:"true" doc:"The maximum permission the requesting user has on this task (0=read, 1=read/write, 2=admin)."`
	RepeatsRRule   *string           `json:"repeats_rrule,omitempty" nullable:"true" doc:"The task's recurrence as RFC 5545 RRULE text (without DTSTART), e.g. FREQ=WEEKLY;BYDAY=MO. Any RFC 5545 rule part is accepted, plus BYEASTER as a documented extension. UNTIL must be a UTC date-time (ending in Z). On read, the stored rule exactly; absent when the task does not repeat. On write, whichever of repeat and repeats_rrule differs from the stored rule is the change; sending both changed and disagreeing is refused. An empty string clears the recurrence; null is treated as absent. Unchanged text (as a PATCH echoes it) is ignored and repeat decides, exactly as without this field: on PUT, omitting repeat clears the rule. A higher COUNT on otherwise unchanged text counts as unchanged, as for repeat, so COUNT cannot be raised through this field alone. Only single-task reads, create and update carry this field; list and bulk endpoints do not."`
	InitialDueTime string            `json:"initial_due_time,omitempty" pattern:"^([01][0-9]|2[0-3]):[0-5][0-9]$" writeOnly:"true" doc:"Time of day, HH:MM in the task's effective zone, to seed the first occurrence of the task's rule when the task has no due date. The server picks the first instance strictly after now and it is stored as due_date. Never stored or returned. Needs a repeat rule after the write (repeat or repeats_rrule, on PATCH including the stored one); refused with 400 when the task has a due date (on PATCH send due_date null to replace one), has no rule, or is done. When no instance exists the task is saved without a due date and the response says why in reason."`
}

// taskCreateBody is the create input: the task plus its rule as text.
type taskCreateBody struct {
	models.Task
	RepeatsRRule   *string `json:"repeats_rrule,omitempty" nullable:"true" doc:"The task's recurrence as RFC 5545 RRULE text (without DTSTART), e.g. FREQ=WEEKLY;BYDAY=MO. Any RFC 5545 rule part is accepted, plus BYEASTER as a documented extension. UNTIL must be a UTC date-time (ending in Z). On read, the stored rule exactly; absent when the task does not repeat. On write, whichever of repeat and repeats_rrule differs from the stored rule is the change; sending both changed and disagreeing is refused. An empty string clears the recurrence; null is treated as absent. Unchanged text (as a PATCH echoes it) is ignored and repeat decides, exactly as without this field: on PUT, omitting repeat clears the rule. A higher COUNT on otherwise unchanged text counts as unchanged, as for repeat, so COUNT cannot be raised through this field alone. Only single-task reads, create and update carry this field; list and bulk endpoints do not."`
	InitialDueTime string  `json:"initial_due_time,omitempty" pattern:"^([01][0-9]|2[0-3]):[0-5][0-9]$" writeOnly:"true" doc:"Time of day, HH:MM in the task's effective zone, to seed the first occurrence of the task's rule when the task has no due date. The server picks the first instance strictly after now and it is stored as due_date. Never stored or returned. Needs a repeat rule after the write (repeat or repeats_rrule, on PATCH including the stored one); refused with 400 when the task has a due date (on PATCH send due_date null to replace one), has no rule, or is done. When no instance exists the task is saved without a due date and the response says why in reason."`
}

// taskWriteBody is the create and update output. It adds repeats_rrule, due_date_source and
// reason to the task, not max_permission.
type taskWriteBody struct {
	models.Task
	RepeatsRRule  *string `json:"repeats_rrule,omitempty" nullable:"true" doc:"The task's recurrence as RFC 5545 RRULE text (without DTSTART), e.g. FREQ=WEEKLY;BYDAY=MO. Any RFC 5545 rule part is accepted, plus BYEASTER as a documented extension. UNTIL must be a UTC date-time (ending in Z). On read, the stored rule exactly; absent when the task does not repeat. On write, whichever of repeat and repeats_rrule differs from the stored rule is the change; sending both changed and disagreeing is refused. An empty string clears the recurrence; null is treated as absent. Unchanged text (as a PATCH echoes it) is ignored and repeat decides, exactly as without this field: on PUT, omitting repeat clears the rule. A higher COUNT on otherwise unchanged text counts as unchanged, as for repeat, so COUNT cannot be raised through this field alone. Only single-task reads, create and update carry this field; list and bulk endpoints do not."`
	DueDateSource string  `json:"due_date_source,omitempty" enum:"explicit,generated" readOnly:"true" doc:"Where the stored due_date came from: generated when this write chose it from initial_due_time, explicit for any other due date (sent by the client or already stored). Absent when the task has no due date."`
	Reason        string  `json:"reason,omitempty" readOnly:"true" doc:"Why initial_due_time chose no due date, in the next-occurrence vocabulary, for example until_passed (the rule's UNTIL is before now) or no_occurrence (no instance between now and UNTIL or in the search window). Present only then; the task is saved without a due date."`
}

// newTaskWriteBody fills a write response. sel is the initial_due_time selection, zero when the
// body had none: a date it chose is generated, any other stored due date is explicit.
func newTaskWriteBody(task *models.Task, sel initialDueResult) *taskWriteBody {
	body := &taskWriteBody{Task: *task, RepeatsRRule: ruleTextOf(task), Reason: sel.reason}
	switch {
	case sel.source != "":
		body.DueDateSource = sel.source
	case !task.DueDate.IsZero():
		body.DueDateSource = models.DueDateSourceExplicit
	}
	return body
}

// initialDueTask is the CObject a v2 create or update runs when the body carries
// initial_due_time: the task, with Create and Update selecting the first due date
// (models.SelectInitialDue) in the handler's session before delegating, so the selection and the
// write share one transaction and the write sees the chosen date as if the client had
// sent it. Every other method, the Can* checks included, is the embedded task's.
type initialDueTask struct {
	*models.Task
	hhmm string
	now  time.Time
	initialDueResult
}

// initialDueResult is what the selection reports to the write response: the source when it chose
// a date, the reason when it chose none.
type initialDueResult struct {
	source, reason string
}

func (t *initialDueTask) Create(s *xorm.Session, a web.Auth) (err error) {
	t.source, t.reason, err = models.SelectInitialDue(s, a, t.Task, t.hhmm, t.now, models.InitialDueCreate)
	if err != nil {
		return err
	}
	return t.Task.Create(s, a)
}

func (t *initialDueTask) Update(s *xorm.Session, a web.Auth) (err error) {
	t.source, t.reason, err = models.SelectInitialDue(s, a, t.Task, t.hhmm, t.now, models.InitialDueUpdate)
	if err != nil {
		return err
	}
	return t.Task.Update(s, a)
}

// taskWrite is the CObject of a v2 single-task write: the task itself, exactly as without the
// field, unless the body carries initial_due_time. now is the operation's one clock reading.
func taskWrite(task *models.Task, hhmm string) (handler.CObject, *initialDueResult) {
	if hhmm == "" {
		return task, &initialDueResult{}
	}
	sel := &initialDueTask{Task: task, hhmm: hhmm, now: models.EngineNow()}
	return sel, &sel.initialDueResult
}

func tasksRead(ctx context.Context, in *struct {
	ID     int64    `path:"task" doc:"The numeric id of the task."`
	Expand []string `query:"expand,explode" enum:"subtasks,buckets,reactions,comments,comment_count,time_entries_count,is_unread" doc:"Embed extra data per task. Repeatable."`
	Format string   `query:"format" enum:"html,markdown" doc:"How rich-text fields are exchanged. See the API description."`
	conditional.Params
}) (*singleReadBody[taskReadOneBody], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	expand, err := parseTaskExpand(in.Expand)
	if err != nil {
		return nil, translateDomainError(err)
	}
	task := &models.Task{ID: in.ID, Expand: expand}
	maxPermission, err := handler.DoReadOne(ctx, task, a)
	if err != nil {
		return nil, translateDomainError(err)
	}
	body := &taskReadOneBody{Task: *task, MaxPermission: models.Permission(maxPermission), RepeatsRRule: ruleTextOf(task)}
	convertTasksToMarkdown(ctx, &body.Task)
	return conditionalReadResponse(&in.Params, body, task.Updated, maxPermission)
}

type taskReadByIndexResponse struct {
	Status       int
	ETag         string  `header:"ETag"`
	Location     *string `header:"Location" hidden:"true"`
	CacheControl *string `header:"Cache-Control" hidden:"true"`
	Body         *taskReadOneBody
}

func tasksReadByIndex(ctx context.Context, in *struct {
	Project string   `path:"project" doc:"A numeric project id or a textual project identifier (e.g. \"PROJ\")."`
	Index   int64    `path:"index" doc:"The per-project task index."`
	Expand  []string `query:"expand,explode" enum:"subtasks,buckets,reactions,comments,comment_count,time_entries_count,is_unread" doc:"Embed extra data per task. Repeatable."`
	Format  string   `query:"format" enum:"html,markdown" doc:"How rich-text fields are exchanged. See the API description."`
	conditional.Params
}) (*taskReadByIndexResponse, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	expand, err := parseTaskExpand(in.Expand)
	if err != nil {
		return nil, translateDomainError(err)
	}
	projectID, err := resolveProjectIdentifier(in.Project)
	if err != nil {
		return nil, err
	}

	task := &models.Task{ProjectID: projectID, Index: in.Index, Expand: expand}
	maxPermission, err := handler.DoReadOne(ctx, task, a)
	if err == nil {
		body := &taskReadOneBody{Task: *task, MaxPermission: models.Permission(maxPermission), RepeatsRRule: ruleTextOf(task)}
		convertTasksToMarkdown(ctx, &body.Task)
		response, err := conditionalReadResponse(&in.Params, body, task.Updated, maxPermission)
		if err != nil {
			return nil, err
		}
		return &taskReadByIndexResponse{
			Status: http.StatusOK,
			ETag:   response.ETag,
			Body:   response.Body,
		}, nil
	}
	if !models.IsErrTaskDoesNotExist(err) {
		return nil, translateDomainError(err)
	}

	s := db.NewSession()
	defer s.Close()

	// Otherwise 307/403 vs 404 lets callers probe retired indexes of projects they cannot read.
	canReadProject, _, err := (&models.Project{ID: projectID}).CanRead(s, a)
	if err != nil {
		return nil, translateDomainError(err)
	}
	if !canReadProject {
		return nil, errReadForbidden(a)
	}

	taskID, err := models.GetTaskIDByIndexAlias(s, projectID, in.Index)
	if err != nil {
		return nil, translateDomainError(err)
	}

	task = &models.Task{ID: taskID}
	canRead, _, err := task.CanRead(s, a)
	if err != nil {
		return nil, translateDomainError(err)
	}
	if !canRead {
		return nil, errReadForbidden(a)
	}

	location := &url.URL{
		Path: GroupPrefix + "/projects/" + strconv.FormatInt(task.ProjectID, 10) +
			"/tasks/by-index/" + strconv.FormatInt(task.Index, 10),
	}
	if ec := humabridge.EchoContextFrom(ctx); ec != nil {
		location.RawQuery = (*ec).Request().URL.RawQuery
	}
	locationValue := location.String()
	cacheControl := "no-store"
	return &taskReadByIndexResponse{
		Status:       http.StatusTemporaryRedirect,
		Location:     &locationValue,
		CacheControl: &cacheControl,
	}, nil
}

func tasksCreate(ctx context.Context, in *struct {
	Project int64  `path:"project" doc:"The numeric id of the project to create the task in."`
	Format  string `query:"format" enum:"html,markdown" doc:"How rich-text fields are exchanged. See the API description."`
	Body    taskCreateBody
}) (*singleBody[taskWriteBody], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	task := &in.Body.Task
	obj, sel := taskWrite(task, in.Body.InitialDueTime)
	task.ProjectID = in.Project // URL wins over body
	applyRuleText(task, in.Body.RepeatsRRule)
	if err := convertToHTML(ctx, &task.Description); err != nil {
		return nil, translateDomainError(err)
	}
	if err := handler.DoCreate(ctx, obj, a); err != nil {
		return nil, translateDomainError(err)
	}
	if err := addPositionsToTasks(task); err != nil {
		return nil, translateDomainError(err)
	}
	convertTasksToMarkdown(ctx, task)
	return &singleBody[taskWriteBody]{Body: newTaskWriteBody(task, *sel)}, nil
}

func addPositionsToTasks(tasks ...*models.Task) error {
	s := db.NewSession()
	defer s.Close()
	return models.AddPositionsToTasks(s, tasks)
}

// Body matches the read shape so AutoPatch's GET→PUT echo of max_permission validates.
func tasksUpdate(ctx context.Context, in *struct {
	ID     int64  `path:"task"`
	Format string `query:"format" enum:"html,markdown" doc:"How rich-text fields are exchanged. See the API description."`
	Body   taskReadOneBody
}) (*singleBody[taskWriteBody], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	task := &in.Body.Task
	obj, sel := taskWrite(task, in.Body.InitialDueTime)
	task.ID = in.ID // URL wins over body
	applyRuleText(task, in.Body.RepeatsRRule)
	if err := convertToHTML(ctx, &task.Description); err != nil {
		return nil, translateDomainError(err)
	}
	if err := handler.DoUpdate(ctx, obj, a); err != nil {
		return nil, translateDomainError(err)
	}
	convertTasksToMarkdown(ctx, task)
	return &singleBody[taskWriteBody]{Body: newTaskWriteBody(task, *sel)}, nil
}

func tasksDelete(ctx context.Context, in *struct {
	ID int64 `path:"task"`
}) (*emptyBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	if err := handler.DoDelete(ctx, &models.Task{ID: in.ID}, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &emptyBody{}, nil
}

// resolveProjectIdentifier turns the {project} path segment into a numeric
// project id. A pure-digit value is always an id (mirroring v1's
// ResolveProjectIdentifier middleware); anything else is looked up as a
// case-insensitive identifier and 404s if unknown.
func resolveProjectIdentifier(raw string) (int64, error) {
	if id, err := strconv.ParseInt(raw, 10, 64); err == nil {
		return id, nil
	}
	s := db.NewSession()
	defer s.Close()
	project, err := models.GetProjectSimpleByIdentifier(s, raw)
	if err != nil {
		return 0, translateDomainError(err)
	}
	return project.ID, nil
}
