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
	"sort"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/models"

	"github.com/danielgtaylor/huma/v2"
)

type rescheduleRequest struct {
	StartDate *time.Time `json:"start_date,omitempty" nullable:"true" doc:"The new start. Omit or send null for none. Ignored for a milestone, which sits at its end date."`
	EndDate   *time.Time `json:"end_date,omitempty" nullable:"true" doc:"The new end. Omit or send null for none."`
}

type rescheduledTask struct {
	ID        int64      `json:"id" doc:"The task id."`
	ProjectID int64      `json:"project_id" doc:"The project the task is in."`
	StartDate *time.Time `json:"start_date,omitempty" doc:"The new start."`
	EndDate   *time.Time `json:"end_date,omitempty" doc:"The new end."`
}

type rescheduleResponse struct {
	Task    rescheduledTask   `json:"task" doc:"The task that was moved."`
	Changed []rescheduledTask `json:"changed" doc:"The tasks that depend on it and were pushed later, with their new dates."`
	Skipped []int64           `json:"skipped" doc:"Ids of dependent tasks that had to move but that you may not edit. They were left as they are."`
}

type rescheduleBody struct{ Body rescheduleResponse }

type baselineEntry struct {
	TaskID    int64      `json:"task_id" doc:"The task id."`
	StartDate *time.Time `json:"start_date,omitempty" doc:"The planned start when the baseline was saved."`
	EndDate   *time.Time `json:"end_date,omitempty" doc:"The planned end when the baseline was saved."`
}

type baselineResponse struct {
	SavedAt *time.Time      `json:"saved_at,omitempty" doc:"When the baseline was saved. Missing if the project has none."`
	Items   []baselineEntry `json:"items" doc:"The baseline, one entry per task that had dates."`
}

type baselineBody struct{ Body baselineResponse }

type baselineSavedBody struct {
	Body struct {
		Tasks int `json:"tasks" doc:"The number of tasks in the new baseline."`
	}
}

func init() { AddRouteRegistrar(RegisterTaskScheduleRoutes) }

// RegisterTaskScheduleRoutes registers rescheduling and baselines, the scheduling features of the Gantt chart.
func RegisterTaskScheduleRoutes(api huma.API) {
	Register(api, huma.Operation{
		OperationID:   "tasks-reschedule",
		Summary:       "Move a task and the tasks that depend on it",
		Description:   "Sets the start and end of a task and pushes tasks that depend on it (finish-to-start: A precedes B means B cannot start before A ends). A dependent task that would start before the new end moves to start at it and keeps its duration, and so on down the chain. Tasks are never pulled earlier, tasks without dates and done tasks stay, and dependent tasks you may not edit are left alone and listed in skipped. A milestone sits at its end date. Needs write access to the task.",
		Method:        http.MethodPost,
		Path:          "/tasks/{task}/reschedule",
		DefaultStatus: http.StatusOK,
		Tags:          []string{"tasks"},
	}, tasksReschedule)

	tags := []string{"projects"}
	Register(api, huma.Operation{
		OperationID: "project-baseline-read",
		Summary:     "Get the baseline of a project",
		Description: "Returns the planned start and end of every task as they were when the baseline was saved, to compare with the current plan. Empty if there is none. Needs read access.",
		Method:      http.MethodGet,
		Path:        "/projects/{project}/baseline",
		Tags:        tags,
	}, projectBaselineRead)
	Register(api, huma.Operation{
		OperationID:   "project-baseline-save",
		Summary:       "Save the current plan as the baseline of a project",
		Description:   "Records the current start and end of every dated task of the project and replaces the previous baseline. A project has one baseline. Needs write access.",
		Method:        http.MethodPut,
		Path:          "/projects/{project}/baseline",
		DefaultStatus: http.StatusOK,
		Tags:          tags,
	}, projectBaselineSave)
	Register(api, huma.Operation{
		OperationID: "project-baseline-delete",
		Summary:     "Remove the baseline of a project",
		Description: "Deletes the saved baseline. Needs write access.",
		Method:      http.MethodDelete,
		Path:        "/projects/{project}/baseline",
		Tags:        tags,
	}, projectBaselineDelete)
}

func optTime(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

func toRescheduled(t *models.Task) rescheduledTask {
	return rescheduledTask{ID: t.ID, ProjectID: t.ProjectID, StartDate: optTime(t.StartDate), EndDate: optTime(t.EndDate)}
}

func tasksReschedule(ctx context.Context, in *struct {
	Task int64 `path:"task" doc:"The numeric ID of the task."`
	Body rescheduleRequest
}) (*rescheduleBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var start, end time.Time
	if in.Body.StartDate != nil {
		start = *in.Body.StartDate
	}
	if in.Body.EndDate != nil {
		end = *in.Body.EndDate
	}

	s := db.NewSession()
	defer s.Close()

	res, err := models.RescheduleTask(s, a, in.Task, start, end)
	if err != nil {
		_ = s.Rollback()
		events.CleanupPending(s)
		return nil, translateDomainError(err)
	}
	err = s.Commit()
	if err != nil {
		events.CleanupPending(s)
		return nil, translateDomainError(err)
	}
	events.DispatchPending(ctx, s)

	out := &rescheduleBody{}
	out.Body.Task = toRescheduled(res.Task)
	out.Body.Changed = make([]rescheduledTask, 0, len(res.Changed))
	for _, t := range res.Changed {
		out.Body.Changed = append(out.Body.Changed, toRescheduled(t))
	}
	sort.Slice(out.Body.Changed, func(i, j int) bool { return out.Body.Changed[i].ID < out.Body.Changed[j].ID })
	out.Body.Skipped = res.Skipped
	if out.Body.Skipped == nil {
		out.Body.Skipped = []int64{}
	}
	sort.Slice(out.Body.Skipped, func(i, j int) bool { return out.Body.Skipped[i] < out.Body.Skipped[j] })
	return out, nil
}

func projectBaselineRead(ctx context.Context, in *struct {
	Project int64 `path:"project" doc:"The numeric ID of the project."`
}) (*baselineBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	s := db.NewReadSession()
	defer s.Close()

	rows, savedAt, err := models.GetProjectBaseline(s, a, in.Project)
	if err != nil {
		return nil, translateDomainError(err)
	}
	out := &baselineBody{}
	out.Body.SavedAt = optTime(savedAt)
	out.Body.Items = make([]baselineEntry, 0, len(rows))
	for _, r := range rows {
		out.Body.Items = append(out.Body.Items, baselineEntry{TaskID: r.TaskID, StartDate: optTime(r.StartDate), EndDate: optTime(r.EndDate)})
	}
	return out, nil
}

func projectBaselineSave(ctx context.Context, in *struct {
	Project int64 `path:"project" doc:"The numeric ID of the project."`
}) (*baselineSavedBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	s := db.NewSession()
	defer s.Close()

	n, err := models.SaveProjectBaseline(s, a, in.Project)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	err = s.Commit()
	if err != nil {
		return nil, translateDomainError(err)
	}
	out := &baselineSavedBody{}
	out.Body.Tasks = n
	return out, nil
}

func projectBaselineDelete(ctx context.Context, in *struct {
	Project int64 `path:"project" doc:"The numeric ID of the project."`
}) (*emptyBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	s := db.NewSession()
	defer s.Close()

	err = models.ClearProjectBaseline(s, a, in.Project)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	err = s.Commit()
	if err != nil {
		return nil, translateDomainError(err)
	}
	return &emptyBody{}, nil
}
