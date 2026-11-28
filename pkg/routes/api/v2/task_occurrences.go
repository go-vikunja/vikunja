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
	"errors"
	"net/http"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/web"

	"github.com/danielgtaylor/huma/v2"
	"xorm.io/xorm"
)

// RegisterTaskOccurrencesRoutes wires the occurrence query and the unsaved-rule preview onto the
// Huma API. Both are read-only; the preview is a POST only because it takes a body.
func RegisterTaskOccurrencesRoutes(api huma.API) {
	tags := []string{"tasks"}

	Register(api, huma.Operation{
		OperationID: "tasks-occurrences",
		Summary:     "List a task's next occurrences",
		Description: "Returns the next due dates this task would take if it were completed once per date, " +
			"computed by the same code completion uses (same time zone order: the caller, then the project " +
			"owner, then the server; same daylight-saving and safety limits). Read-only: the task is not " +
			"changed. Requires read access to the task. When fewer than count dates exist, reason says why. " +
			"A rule that ends by UNTIL within the list reports no_occurrence; until_passed means UNTIL was " +
			"already behind the request time. A task that is already done lists no dates and reports " +
			"task_done, because completing it again changes nothing; that includes a task that stayed done " +
			"because its UNTIL had passed. Without read access the answer is 403 with the same body as a task " +
			"read; a task that does not exist is 404; a count outside 1 to 100 is 422. " +
			occurrencesForecastSentence,
		Method: http.MethodGet,
		Path:   "/tasks/{task}/occurrences",
		Tags:   tags,
	}, tasksOccurrences)

	Register(api, huma.Operation{
		OperationID: "tasks-occurrences-preview",
		Summary:     "Preview the occurrences of an unsaved rule",
		Description: "Computes the due dates a task with this rule and due date would take, without saving " +
			"anything. Send exactly one of repeat or repeats_rrule, and exactly one of due_date or " +
			"initial_due_time. With initial_due_time (HH:MM), the due date is the one creating the task with " +
			"that initial_due_time would select: the first occurrence strictly after now, searched from today " +
			"at that time in the effective zone; due_date_source is then generated, otherwise explicit. When " +
			"no date can be selected, due_date is absent, occurrences is empty and reason says why. " +
			"occurrences are strictly after due_date; render [due_date, ...occurrences] in timezone, as " +
			"returned. A rule that saving a task would refuse is refused here with the same error: 400 with " +
			"code 2002 and the same detail. The time zone is timezone when sent; otherwise, with task_id or " +
			"project_id, the zone a change to a task in that project uses (the caller, then the project " +
			"owner, then the server); without either, the caller's time zone is used, then the server's. An " +
			"unknown timezone is 400 with code 2003. task_id and project_id are zone context only and need read " +
			"access, refused as a task or project read is. Sending both or neither of repeat and " +
			"repeats_rrule, or a count outside 1 to 100, is 422; so is sending both or neither of due_date and " +
			"initial_due_time, both task_id and project_id, a negative project_id (a saved filter has no " +
			"zone), or an initial_due_time that is not HH:MM from 00:00 to 23:59. " + occurrencesForecastSentence +
			" evaluated_at is the request time the answer was computed at, for audit only.",
		Method: http.MethodPost,
		Path:   "/tasks/occurrences/preview",
		// A POST that computes rather than creates.
		DefaultStatus: http.StatusOK,
		Tags:          tags,
	}, tasksOccurrencesPreview)
}

func init() { AddRouteRegistrar(RegisterTaskOccurrencesRoutes) }

// occurrencesForecastSentence is the on-time-completion assumption both descriptions state.
const occurrencesForecastSentence = "For a task that repeats from the current date, the list assumes each " +
	"completion happens on time, so the first simulated completion is the later of now and the due date; " +
	"it is not a promise about actual completion."

// occurrenceAnswer is the response of tasks-occurrences (schema OccurrenceAnswer).
type occurrenceAnswer struct {
	Occurrences []time.Time `json:"occurrences" required:"true" doc:"Due dates in ascending order, at most count."`
	Reason      string      `json:"reason,omitempty" doc:"Present only when fewer than count dates come back. One of count_exhausted, until_passed, no_rule, no_due_date, task_done, no_occurrence, invalid_rule, interval_beyond_horizon, anchor_beyond_horizon, outside_bounds, policy_rejected, rule_too_long, evaluation_failed, outside_storable_range, no_zone."`
}

func newOccurrenceAnswer(answer *models.OccurrenceAnswer) *singleBody[occurrenceAnswer] {
	occurrences := answer.Occurrences
	if occurrences == nil {
		// The model never returns nil; an empty list must still serialise as [].
		occurrences = []time.Time{}
	}
	return &singleBody[occurrenceAnswer]{Body: &occurrenceAnswer{Occurrences: occurrences, Reason: answer.Reason}}
}

// previewAnswer is the response of tasks-occurrences-preview.
type previewAnswer struct {
	Occurrences   []time.Time `json:"occurrences" required:"true" doc:"Due dates in ascending order, strictly after due_date, at most count."`
	Reason        string      `json:"reason,omitempty" doc:"Present only when fewer than count dates come back: either no due date could be selected (due_date is then absent), or the list after due_date ended early (count_exhausted, for example, beside a valid due_date). One of count_exhausted, until_passed, no_rule, no_due_date, task_done, no_occurrence, invalid_rule, interval_beyond_horizon, anchor_beyond_horizon, outside_bounds, policy_rejected, rule_too_long, evaluation_failed, outside_storable_range, no_zone."`
	DueDate       *time.Time  `json:"due_date,omitempty" doc:"The anchor: the due_date sent (explicit) or the one initial_due_time selected (generated). Absent when there is none."`
	DueDateSource string      `json:"due_date_source,omitempty" enum:"explicit,generated" doc:"explicit when due_date was sent; generated when initial_due_time selected it. Absent without a due_date."`
	TimeZone      string      `json:"timezone" required:"true" doc:"The IANA name of the time zone the preview was evaluated in."`
	EvaluatedAt   time.Time   `json:"evaluated_at" required:"true" doc:"The request time the answer was computed at, to the second. For audit only."`
}

func newPreviewAnswer(answer *models.PreviewAnswer) *singleBody[previewAnswer] {
	occurrences := answer.Occurrences
	if occurrences == nil {
		occurrences = []time.Time{}
	}
	return &singleBody[previewAnswer]{Body: &previewAnswer{
		Occurrences:   occurrences,
		Reason:        answer.Reason,
		DueDate:       answer.DueDate,
		DueDateSource: answer.DueDateSource,
		TimeZone:      answer.TimeZone,
		EvaluatedAt:   answer.EvaluatedAt,
	}}
}

// authorizeTaskRead is the required task read of tasks-occurrences, and of the preview's task_id:
// exactly the check tasks-read runs (CanRead, then ReadOne, which adds the reminders completion
// steps), whatever the id. A denial is the same 403 and a missing task, id 0 or below included, the
// same 404 as tasks-read. On success the task is never nil.
func authorizeTaskRead(s *xorm.Session, a web.Auth, taskID int64) (*models.Task, error) {
	task := &models.Task{ID: taskID}
	canRead, _, err := task.CanRead(s, a)
	if err != nil {
		return nil, translateDomainError(err)
	}
	if !canRead {
		return nil, errReadForbidden(a)
	}
	if err := task.ReadOne(s, a); err != nil {
		return nil, translateDomainError(err)
	}
	return task, nil
}

// authorizeZoneContext is the read check on the preview's optional zone context, so a zone is never
// resolved for a task or project the caller cannot read. A taskID other than 0 is read
// as tasks-occurrences reads its task (authorizeTaskRead); a projectID other than 0 runs the
// project's CanRead, a denial being the same 403. It returns the project whose zone applies and,
// with taskID, the loaded task; with neither, zeros and no task. Only the preview calls it: an
// absent task_id is 0 there, while tasks-occurrences always has a task and calls authorizeTaskRead.
// The caller refuses sending both.
func authorizeZoneContext(s *xorm.Session, a web.Auth, taskID, projectID int64) (zoneProjectID int64, task *models.Task, err error) {
	switch {
	case taskID != 0:
		task, err = authorizeTaskRead(s, a, taskID)
		if err != nil {
			return 0, nil, err
		}
		return task.ProjectID, task, nil
	case projectID != 0:
		canRead, _, err := (&models.Project{ID: projectID}).CanRead(s, a)
		if err != nil {
			return 0, nil, translateDomainError(err)
		}
		if !canRead {
			return 0, nil, errReadForbidden(a)
		}
		return projectID, nil, nil
	}
	return 0, nil, nil
}

// The count bounds here and on previewInput are struct-tag literals, which cannot name a constant;
// TestTaskOccurrences_CountTagsMatchModel keeps them equal to models.DefaultOccurrencesCount and
// models.MaxOccurrencesCount.
type taskOccurrencesInput struct {
	TaskID int64 `path:"task" doc:"The numeric id of the task."`
	Count  int   `query:"count" minimum:"1" maximum:"100" default:"10" doc:"How many occurrences to return, 1 to 100. Defaults to 10."`
}

// tasksOccurrences owns the session and the read check: there is no handler.Do* for a query on a
// task. authorizeTaskRead loads the task as tasks-read does, for every id, so a missing task
// (id 0 included) is the same 404 and a denial the same 403.
func tasksOccurrences(ctx context.Context, in *taskOccurrencesInput) (*singleBody[occurrenceAnswer], error) {
	// The operation's one clock sample.
	now := models.EngineNow()
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	s := db.NewReadSession()
	defer s.Close()

	task, err := authorizeTaskRead(s, a, in.TaskID)
	if err != nil {
		_ = s.Rollback()
		return nil, err
	}

	answer, err := models.TaskOccurrences(s, a, task, in.Count, now)
	// Read-only: nothing to commit.
	_ = s.Rollback()
	if err != nil {
		return nil, translateDomainError(err)
	}
	return newOccurrenceAnswer(answer), nil
}

// previewInput is the preview request body (schema PreviewInput).
type previewInput struct {
	Repeat                 *models.TaskRepeat `json:"repeat,omitempty" doc:"Structured recurrence, as on a task. Send exactly one of repeat and repeats_rrule."`
	RepeatsRRule           *string            `json:"repeats_rrule,omitempty" doc:"RFC 5545 rule text, as on a task (see the task's repeats_rrule). Send exactly one of repeat and repeats_rrule. An empty value is refused here (400, code 2002), unlike on a task write, where it clears the rule."`
	DueDate                *time.Time         `json:"due_date,omitempty" doc:"The anchor: the due date the rule starts from. Send exactly one of due_date and initial_due_time."`
	InitialDueTime         string             `json:"initial_due_time,omitempty" pattern:"^([01][0-9]|2[0-3]):[0-5][0-9]$" doc:"HH:MM, 00:00 to 23:59: preview the due date creating the task with this initial_due_time would select, the first occurrence strictly after now from today at this time in the effective zone. Send exactly one of due_date and initial_due_time."`
	RepeatsFromCurrentDate bool               `json:"repeats_from_current_date,omitempty" default:"false" doc:"Same meaning as on a task: when true, the next occurrence is counted from the moment the task is completed instead of from its previous due date."`
	TimeZone               string             `json:"timezone,omitempty" doc:"IANA time zone name, such as Europe/Berlin. Without it, the caller's time zone is used, then the server's. An unknown name is refused with code 2003."`
	Count                  int                `json:"count,omitempty" minimum:"1" maximum:"100" default:"10" doc:"How many occurrences to return, 1 to 100. Defaults to 10."`
	TaskID                 int64              `json:"task_id,omitempty" doc:"Zone context only: evaluate in the zone of this task's project. Needs read access to the task. Not with project_id."`
	ProjectID              int64              `json:"project_id,omitempty" doc:"Zone context only: evaluate in the zone of this project, as a task created in it would be. Needs read access to the project; a saved filter (negative id) is refused. Not with task_id."`
}

// tasksOccurrencesPreview refuses shape conflicts with 422 before anything else, then validates
// the rule with the validator a task write uses and nothing else, so a refusal carries the same code
// and message as saving it would. Any signed-in caller may preview; with task_id or
// project_id the caller must be able to read it (authorizeZoneContext), and its project's zone is
// used.
func tasksOccurrencesPreview(ctx context.Context, in *struct {
	Body previewInput
}) (*singleBody[previewAnswer], error) {
	// The operation's one clock sample; evaluated_at echoes it.
	now := models.EngineNow()
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	body := &in.Body
	if err := previewShapeError(body); err != nil {
		return nil, translateDomainError(err)
	}

	var rule string
	if body.Repeat != nil {
		rule, err = models.ValidateRepeat(body.Repeat)
	} else {
		rule, err = models.ValidateRuleText(*body.RepeatsRRule)
	}
	if err != nil {
		return nil, translateDomainError(err)
	}

	s := db.NewReadSession()
	defer s.Close()

	zoneProjectID, _, err := authorizeZoneContext(s, a, body.TaskID, body.ProjectID)
	if err != nil {
		_ = s.Rollback()
		return nil, err
	}

	answer, err := models.PreviewOccurrences(s, a, models.PreviewInput{
		Rule:           rule,
		DueDate:        body.DueDate,
		InitialDueTime: body.InitialDueTime,
		FromCompletion: body.RepeatsFromCurrentDate,
		TimeZone:       body.TimeZone,
		ProjectID:      zoneProjectID,
		Count:          body.Count,
	}, now)
	// Read-only: nothing to commit.
	_ = s.Rollback()
	if err != nil {
		return nil, translateDomainError(err)
	}
	return newPreviewAnswer(answer), nil
}

// previewShapeError is the preview's first refusal, in a fixed order, or nil: the rule
// forms, the anchor forms (both is the initial-due conflict, InitialDueConflict), the zone contexts,
// then a saved filter as zone context. Each is 422.
func previewShapeError(body *previewInput) error {
	if (body.Repeat == nil) == (body.RepeatsRRule == nil) {
		const msg = "send exactly one of repeat and repeats_rrule"
		return models.InvalidFieldErrorWithMessage(
			[]string{"repeat: " + msg, "repeats_rrule: " + msg}, "Send exactly one of repeat and repeats_rrule.")
	}
	if body.DueDate == nil && body.InitialDueTime == "" {
		const msg = "send exactly one of due_date and initial_due_time"
		return models.InvalidFieldErrorWithMessage(
			[]string{"due_date: " + msg, "initial_due_time: " + msg}, "Send exactly one of due_date and initial_due_time.")
	}
	// A rule is present and a preview is never done; only the due date can conflict here.
	var conflict models.ErrInitialDueConflict
	if err := models.InitialDueConflict(body.DueDate != nil, body.InitialDueTime, true, false); errors.As(err, &conflict) {
		return conflict.PreviewError()
	}
	if body.TaskID != 0 && body.ProjectID != 0 {
		const msg = "send at most one of task_id and project_id"
		return models.InvalidFieldErrorWithMessage(
			[]string{"task_id: " + msg, "project_id: " + msg}, "Send at most one of task_id and project_id.")
	}
	if body.ProjectID < 0 {
		return models.InvalidFieldErrorWithMessage(
			[]string{"project_id: a saved filter has no time zone"}, "project_id cannot be a saved filter.")
	}
	return nil
}
