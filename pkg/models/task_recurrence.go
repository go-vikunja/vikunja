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
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"slices"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/web"

	"xorm.io/xorm"
)

// TaskRecurrence completes an occurrence or restores it using the token returned by completion.
type TaskRecurrence struct {
	TaskID    int64  `json:"-"`
	UndoToken string `json:"undo_token,omitempty" doc:"Omit to complete the occurrence. To undo, send the token returned by completion."`
	Task      *Task  `json:"task,omitempty" readOnly:"true" doc:"The task after completion or undo."`

	web.Permissions `json:"-"`
	web.CRUDable    `json:"-"`
}

func (r *TaskRecurrence) CanCreate(s *xorm.Session, a web.Auth) (bool, error) {
	return (&Task{ID: r.TaskID}).CanUpdate(s, a)
}

type taskOccurrence struct {
	DoneAt, DueDate, StartDate, EndDate time.Time
	Description                         string
	Reminders                           []*TaskReminder
	Buckets                             []*TaskBucket
	Positions                           []*TaskPosition
}

type taskOccurrenceUndo struct {
	TaskID, ProjectID, ActorID int64
	LinkShare                  bool
	Expires                    time.Time
	Before                     taskOccurrence
	After                      [sha256.Size]byte
}

type errTaskOccurrenceConflict struct{}

func (errTaskOccurrenceConflict) Error() string {
	return "The task occurrence has changed or the undo has expired."
}

func (e errTaskOccurrenceConflict) HTTPError() web.HTTPError {
	return web.HTTPError{HTTPCode: http.StatusConflict, Message: e.Error()}
}

func snapshotTaskOccurrence(s *xorm.Session, task Task, viewIDs []int64) (state taskOccurrence, err error) {
	state = taskOccurrence{
		DoneAt: task.DoneAt, DueDate: task.DueDate, StartDate: task.StartDate,
		EndDate: task.EndDate, Description: task.Description,
	}
	state.Reminders, err = getRemindersForTasks(s, []int64{task.ID})
	if err != nil || len(viewIDs) == 0 {
		return
	}
	err = s.Where("task_id = ?", task.ID).In("project_view_id", viewIDs).
		OrderBy("project_view_id").Find(&state.Buckets)
	if err != nil {
		return
	}
	err = s.Where("task_id = ?", task.ID).In("project_view_id", viewIDs).
		OrderBy("project_view_id").Find(&state.Positions)
	return
}

func taskOccurrenceHash(task Task, state taskOccurrence, viewIDs []int64) ([sha256.Size]byte, error) {
	data, err := json.Marshal(struct {
		Updated time.Time
		State   taskOccurrence
		Views   []int64
	}{task.Updated, state, viewIDs})
	return sha256.Sum256(data), err
}

func signTaskOccurrence(data string) []byte {
	mac := hmac.New(sha256.New, []byte(config.ServiceSecret.GetString()))
	_, _ = mac.Write([]byte("vikunja-task-occurrence-undo:" + data))
	return mac.Sum(nil)
}

func (r *TaskRecurrence) Create(s *xorm.Session, a web.Auth) error {
	tasks, views, err := lockViewsAndTasksForWrite(s, a, []int64{r.TaskID})
	if err != nil {
		return err
	}
	task := tasks[r.TaskID]
	if task.Done || !task.isRepeating() {
		return errTaskOccurrenceConflict{}
	}

	var viewIDs []int64
	for _, view := range views {
		if view.ProjectID == task.ProjectID && view.ViewKind == ProjectViewKindKanban && view.BucketConfigurationMode == BucketConfigurationModeManual {
			viewIDs = append(viewIDs, view.ID)
		}
	}
	slices.Sort(viewIDs)

	state, err := snapshotTaskOccurrence(s, task, viewIDs)
	if err != nil {
		return err
	}
	_, linkShare := a.(*LinkSharing)
	if r.UndoToken == "" {
		updated := Task{ID: task.ID, Done: true}
		if err := updated.updateSingleTaskInViews(s, a, []string{"done"}, task, views); err != nil {
			return err
		}
		stored, err := GetTaskSimple(s, &Task{ID: task.ID})
		if err != nil {
			return err
		}
		stateAfter, err := snapshotTaskOccurrence(s, stored, viewIDs)
		if err != nil {
			return err
		}
		hash, err := taskOccurrenceHash(stored, stateAfter, viewIDs)
		if err != nil {
			return err
		}
		payload, err := json.Marshal(taskOccurrenceUndo{
			TaskID: task.ID, ProjectID: task.ProjectID, ActorID: a.GetID(), LinkShare: linkShare,
			Expires: time.Now().Add(5 * time.Minute), Before: state, After: hash,
		})
		if err != nil {
			return err
		}
		data := base64.RawURLEncoding.EncodeToString(payload)
		r.UndoToken = data + "." + base64.RawURLEncoding.EncodeToString(signTaskOccurrence(data))
		r.Task = &updated
		return nil
	}

	data, signature, ok := strings.Cut(r.UndoToken, ".")
	mac, err := base64.RawURLEncoding.DecodeString(signature)
	if !ok || err != nil || !hmac.Equal(mac, signTaskOccurrence(data)) {
		return errTaskOccurrenceConflict{}
	}
	payload, err := base64.RawURLEncoding.DecodeString(data)
	if err != nil {
		return errTaskOccurrenceConflict{}
	}
	var undo taskOccurrenceUndo
	if err := json.Unmarshal(payload, &undo); err != nil || time.Now().After(undo.Expires) ||
		undo.TaskID != task.ID || undo.ProjectID != task.ProjectID || undo.ActorID != a.GetID() || undo.LinkShare != linkShare {
		return errTaskOccurrenceConflict{}
	}

	hash, err := taskOccurrenceHash(task, state, viewIDs)
	if err != nil {
		return err
	}
	if hash != undo.After {
		return errTaskOccurrenceConflict{}
	}
	if err := restoreTaskOccurrence(s, a, &task, undo.Before, viewIDs); err != nil {
		return err
	}
	r.Task = &task
	r.UndoToken = ""
	return nil
}

func restoreTaskOccurrence(s *xorm.Session, a web.Auth, task *Task, before taskOccurrence, viewIDs []int64) error {
	// Restore associations directly. Moving into a bucket through the normal
	// move action can trigger another completion and is subject to its WIP limit.
	for _, association := range before.Buckets {
		bucket, err := getBucketByID(lockingSession(s), association.BucketID)
		if IsErrBucketDoesNotExist(err) {
			return errTaskOccurrenceConflict{}
		}
		if err != nil {
			return err
		}
		if bucket.ProjectViewID != association.ProjectViewID {
			return errTaskOccurrenceConflict{}
		}
	}
	if len(viewIDs) > 0 {
		if _, err := s.Where("task_id = ?", task.ID).In("project_view_id", viewIDs).Delete(&TaskBucket{}); err != nil {
			return err
		}
		if _, err := s.Where("task_id = ?", task.ID).In("project_view_id", viewIDs).Delete(&TaskPosition{}); err != nil {
			return err
		}
	}
	if len(before.Buckets) > 0 {
		if _, err := s.Insert(&before.Buckets); err != nil {
			return err
		}
	}
	for _, position := range before.Positions {
		conflicts, err := findPositionConflicts(s, position.ProjectViewID, position.Position)
		if err != nil {
			return err
		}
		if len(conflicts) > 0 {
			return errTaskOccurrenceConflict{}
		}
		if err := upsertTaskPosition(s, position); err != nil {
			return err
		}
	}
	task.DoneAt, task.DueDate, task.StartDate, task.EndDate = before.DoneAt, before.DueDate, before.StartDate, before.EndDate
	task.Description = before.Description
	if _, err := s.Where("task_id = ?", task.ID).Delete(&TaskReminder{}); err != nil {
		return err
	}
	task.Reminders = before.Reminders
	for _, reminder := range task.Reminders {
		reminder.TaskID = task.ID
	}
	if len(task.Reminders) > 0 {
		if _, err := s.Insert(&task.Reminders); err != nil {
			return err
		}
	}
	if _, err := s.ID(task.ID).Cols("done_at", "due_date", "start_date", "end_date", "description").Update(task); err != nil {
		return err
	}
	stored, err := GetTaskSimple(s, &Task{ID: task.ID})
	if err != nil {
		return err
	}
	task.Updated = stored.Updated
	events.DispatchOnCommit(s, &TaskUpdatedEvent{Task: task, Doer: doerFromAuth(s, a)})
	return updateProjectLastUpdated(s, &Project{ID: task.ProjectID})
}
