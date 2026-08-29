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
	"fmt"
	"net/http"

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/web/handler"

	"github.com/danielgtaylor/huma/v2"
)

const taskListFilterDoc = "Filtering, sorting and search apply to every variant. See https://vikunja.io/docs/filters for the filter language."

type taskListBody struct {
	Body Paginated[*models.Task]
}

// bucketsWithTasksBody is the buckets-with-tasks response. It is not paginated:
// the view's bucket configuration bounds how many tasks each bucket carries, so
// page/per_page don't apply and total is simply the number of buckets.
type bucketsWithTasksBody struct {
	Body struct {
		Items []*models.Bucket `json:"items"`
		Total int64            `json:"total" doc:"The number of buckets returned."`
	}
}

// TaskListQueryParams is the shared filter/sort/search/expand query block for
// every task-list variant. It must stay EXPORTED: Huma promotes an anonymous
// embed's params only when the embed field is itself exported, and an embed
// field is exported iff its type name is (a lowercase type name silently drops
// all of its params from binding and the spec).
//
// The three input structs below embed it but keep their path params inline:
// Huma lists every path:"" field regardless of the route template, so a shared
// project/view field would leak onto a narrower route as a phantom path param.
// taskListViewInput is shared by both view-scoped endpoints.
type TaskListQueryParams struct {
	ListParams
	Filter             string   `query:"filter" doc:"Filter query to match tasks by. See https://vikunja.io/docs/filters."`
	FilterTimezone     string   `query:"filter_timezone" doc:"Timezone used to resolve relative date filters like \"now\"."`
	FilterIncludeNulls bool     `query:"filter_include_nulls" doc:"If true, also include tasks whose filtered field is null."`
	SortBy             []string `query:"sort_by,explode" doc:"Fields to sort by (e.g. done, priority). Repeatable; pair positionally with order_by. The special value relevance sorts by search relevance (most relevant first, requires s; ignored when the database cannot score the query)."`
	OrderBy            []string `query:"order_by,explode" doc:"Sort order per sort_by field, asc or desc. Repeatable; defaults to asc."`
	Expand             []string `query:"expand,explode" enum:"subtasks,buckets,reactions,comments,comment_count,time_entries_count,is_unread" doc:"Embed extra, more expensive data per task. Repeatable."`
	Format             string   `query:"format" enum:"html,markdown" doc:"How rich-text fields are exchanged. See the API description."`
}

type taskListAllInput struct {
	TaskListQueryParams
}

type taskListProjectInput struct {
	ProjectID int64 `path:"project" doc:"The numeric id of the project."`
	TaskListQueryParams
}

type taskListViewInput struct {
	ProjectID int64 `path:"project" doc:"The numeric id of the project."`
	ViewID    int64 `path:"view" doc:"The numeric id of the project view."`
	TaskListQueryParams
}

// taskListFilters is the bound query carried into the shared collection builder.
// The three input structs convert into it so the collection logic lives once.
type taskListFilters struct {
	Q                  string
	Filter             string
	FilterTimezone     string
	FilterIncludeNulls bool
	SortBy             []string
	OrderBy            []string
	Expand             []string
}

func (in taskListAllInput) filters() taskListFilters {
	return taskListFilters{in.Q, in.Filter, in.FilterTimezone, in.FilterIncludeNulls, in.SortBy, in.OrderBy, in.Expand}
}

func (in taskListProjectInput) filters() taskListFilters {
	return taskListFilters{in.Q, in.Filter, in.FilterTimezone, in.FilterIncludeNulls, in.SortBy, in.OrderBy, in.Expand}
}

func (in taskListViewInput) filters() taskListFilters {
	return taskListFilters{in.Q, in.Filter, in.FilterTimezone, in.FilterIncludeNulls, in.SortBy, in.OrderBy, in.Expand}
}

// collection turns the bound query into a TaskCollection. The search term
// arrives as `q` but reaches the model through DoReadAll's search argument, not
// the collection's Search field. forceFlat keeps a kanban view path returning
// flat tasks; the buckets endpoint leaves it false for the polymorphic shape.
func (f taskListFilters) collection(projectID, viewID int64, forceFlat bool) (*models.TaskCollection, error) {
	expand, err := parseTaskExpand(f.Expand)
	if err != nil {
		return nil, translateDomainError(err)
	}
	tc := &models.TaskCollection{
		ProjectID:          projectID,
		ProjectViewID:      viewID,
		Filter:             f.Filter,
		FilterTimezone:     f.FilterTimezone,
		FilterIncludeNulls: f.FilterIncludeNulls,
		SortBy:             f.SortBy,
		OrderBy:            f.OrderBy,
		Expand:             expand,
	}
	if forceFlat {
		tc.SetForceFlatTasks()
	}
	return tc, nil
}

func RegisterTaskCollectionRoutes(api huma.API) {
	tags := []string{"tasks"}

	Register(api, huma.Operation{
		OperationID: "tasks-list",
		Summary:     "List tasks across all projects",
		Description: "Returns the tasks the authenticated user can see across every project they have access to, paginated and flat. " + taskListFilterDoc,
		Method:      http.MethodGet,
		Path:        "/tasks",
		Tags:        tags,
	}, tasksListAll)

	Register(api, huma.Operation{
		OperationID: "project-tasks-list",
		Summary:     "List tasks in a project",
		Description: "Returns the tasks in a project, paginated and flat. Requires read access to the project. " + taskListFilterDoc,
		Method:      http.MethodGet,
		Path:        "/projects/{project}/tasks",
		Tags:        tags,
	}, projectTasksList)

	Register(api, huma.Operation{
		OperationID: "project-view-tasks-list",
		Summary:     "List tasks in a project view",
		Description: "Returns the tasks in a project view, paginated and flat. The view's own filter, sort and search are applied on top of the query. Always returns flat tasks, even for a kanban view — use the buckets endpoint to get tasks grouped by bucket. " + taskListFilterDoc,
		Method:      http.MethodGet,
		Path:        "/projects/{project}/views/{view}/tasks",
		Tags:        tags,
	}, projectViewTasksList)

	Register(api, huma.Operation{
		OperationID: "project-view-buckets-tasks-list",
		Summary:     "List a kanban view's buckets with their tasks",
		Description: "Returns the buckets of a project's kanban view, each populated with the tasks in it. Requires read access to the project. Not paginated: the number and size of buckets follow the view's bucket configuration, so page/per_page do not apply. " + taskListFilterDoc,
		Method:      http.MethodGet,
		Path:        "/projects/{project}/views/{view}/buckets/tasks",
		Tags:        tags,
	}, projectViewBucketsTasksList)
}

func init() { AddRouteRegistrar(RegisterTaskCollectionRoutes) }

func tasksListAll(ctx context.Context, in *taskListAllInput) (*taskListBody, error) {
	return readFlatTasks(ctx, in.filters(), in.Page, in.PerPage, 0, 0)
}

func projectTasksList(ctx context.Context, in *taskListProjectInput) (*taskListBody, error) {
	return readFlatTasks(ctx, in.filters(), in.Page, in.PerPage, in.ProjectID, 0)
}

func projectViewTasksList(ctx context.Context, in *taskListViewInput) (*taskListBody, error) {
	return readFlatTasks(ctx, in.filters(), in.Page, in.PerPage, in.ProjectID, in.ViewID)
}

// readFlatTasks runs DoReadAll for a flat-task endpoint and unwraps the result.
// The model authorizes (project/view CanRead) inside ReadAll, so there's no
// Can* call here.
func readFlatTasks(ctx context.Context, f taskListFilters, page, perPage int, projectID, viewID int64) (*taskListBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	tc, err := f.collection(projectID, viewID, true)
	if err != nil {
		return nil, err
	}
	result, _, total, err := handler.DoReadAll(ctx, tc, a, f.Q, page, perPage)
	if err != nil {
		return nil, translateDomainError(err)
	}
	tasks, ok := result.([]*models.Task)
	if !ok {
		return nil, fmt.Errorf("taskCollection.ReadAll returned unexpected type %T (expected []*models.Task)", result)
	}
	convertTasksToMarkdown(ctx, tasks...)
	return &taskListBody{Body: NewPaginated(tasks, total, page, perPage)}, nil
}

func projectViewBucketsTasksList(ctx context.Context, in *taskListViewInput) (*bucketsWithTasksBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	f := in.filters()
	tc, err := f.collection(in.ProjectID, in.ViewID, false)
	if err != nil {
		return nil, err
	}
	result, _, total, err := handler.DoReadAll(ctx, tc, a, f.Q, in.Page, in.PerPage)
	if err != nil {
		return nil, translateDomainError(err)
	}
	buckets, ok := result.([]*models.Bucket)
	if !ok {
		// ReadAll only yields []*Bucket from the kanban branch; a flat []*Task
		// here means the view has no bucket configuration, so there are no
		// buckets to return. That's a client error, not a 500.
		if _, isTasks := result.([]*models.Task); isTasks {
			return nil, huma.Error400BadRequest("this view has no buckets; use the tasks endpoint for non-kanban views")
		}
		return nil, fmt.Errorf("taskCollection.ReadAll returned unexpected type %T (expected []*models.Bucket)", result)
	}
	var bucketTasks []*models.Task
	for _, bucket := range buckets {
		bucketTasks = append(bucketTasks, bucket.Tasks...)
	}
	convertTasksToMarkdown(ctx, bucketTasks...)
	out := &bucketsWithTasksBody{}
	out.Body.Items = buckets
	out.Body.Total = total
	return out, nil
}
