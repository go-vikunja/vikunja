// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package models

import (
	"context"
	"fmt"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/cron"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/user"

	"github.com/typesense/typesense-go/typesense"
	"github.com/typesense/typesense-go/typesense/api"
	"github.com/typesense/typesense-go/typesense/api/pointer"
	"xorm.io/xorm"
)

type TypesenseSync struct {
	Collection     string    `xorm:"not null"`
	SyncStartedAt  time.Time `xorm:"not null"`
	SyncFinishedAt time.Time `xorm:"null"`
}

var typesenseClient *typesense.Client

func InitTypesense() {
	if !config.TypesenseEnabled.GetBool() {
		return
	}

	typesenseClient = typesense.NewClient(
		typesense.WithServer(config.TypesenseURL.GetString()),
		typesense.WithAPIKey(config.TypesenseAPIKey.GetString()))
}

func CreateTypesenseCollections() error {
	taskSchema := &api.CollectionSchema{
		Name:               "tasks",
		EnableNestedFields: pointer.True(),
		Fields: []api.Field{
			{
				Name: "id",
				Type: "string",
				Sort: pointer.True(),
			},
			{
				Name: "title",
				Type: "string",
				Sort: pointer.True(),
			},
			{
				Name: "description",
				Type: "string",
				Sort: pointer.True(),
			},
			{
				Name: "done",
				Type: "bool",
			},
			{
				Name:     "done_at",
				Type:     "int64", // unix timestamp
				Optional: pointer.True(),
			},
			{
				Name:     "due_date",
				Type:     "int64", // unix timestamp
				Optional: pointer.True(),
			},
			{
				Name: "project_id",
				Type: "int64",
			},
			{
				Name: "repeat_after",
				Type: "int64",
			},
			{
				Name: "repeat_mode",
				Type: "int32",
			},
			{
				Name: "priority",
				Type: "int64",
			},
			{
				Name:     "start_date",
				Type:     "int64", // unix timestamp
				Optional: pointer.True(),
			},
			{
				Name:     "end_date",
				Type:     "int64", // unix timestamp
				Optional: pointer.True(),
			},
			{
				Name: "hex_color",
				Type: "string",
				Sort: pointer.True(),
			},
			{
				Name: "percent_done",
				Type: "float",
			},
			{
				Name: "identifier",
				Type: "string",
				Sort: pointer.True(),
			},
			{
				Name: "index",
				Type: "int64",
			},
			{
				Name: "uid",
				Type: "string",
				Sort: pointer.True(),
			},
			{
				Name: "cover_image_attachment_id",
				Type: "int64",
			},
			{
				Name: "created",
				Type: "int64", // unix timestamp
			},
			{
				Name: "updated",
				Type: "int64", // unix timestamp
			},
			{
				Name: "bucket_id",
				Type: "int64",
			},
			{
				Name: "position",
				Type: "float",
			},
			{
				Name: "kanban_position",
				Type: "float",
			},
			{
				Name: "created_by_id",
				Type: "int64",
			},
			{
				Name:     "reminders",
				Type:     "object[]", // TODO
				Optional: pointer.True(),
			},
			{
				Name:     "assignees",
				Type:     "object[]", // TODO
				Optional: pointer.True(),
			},
			{
				Name:     "labels",
				Type:     "object[]", // TODO
				Optional: pointer.True(),
			},
			{
				Name:     "related_tasks",
				Type:     "object[]", // TODO
				Optional: pointer.True(),
			},
			{
				Name:     "attachments",
				Type:     "object[]", // TODO
				Optional: pointer.True(),
			},
			{
				Name:     "comments",
				Type:     "object[]", // TODO
				Optional: pointer.True(),
			},
		},
	}

	// delete any collection which might exist
	_, _ = typesenseClient.Collection("tasks").Delete(context.Background())

	_, err := typesenseClient.Collections().Create(context.Background(), taskSchema)
	return err
}

func ReindexAllTasks() (err error) {
	tasks := make(map[int64]*Task)

	s := db.NewSession()
	defer s.Close()

	_, err = s.Where("collection = ?", "tasks").Delete(&TypesenseSync{})
	if err != nil {
		return fmt.Errorf("could not delete old sync status: %s", err.Error())
	}

	currentSync := &TypesenseSync{
		Collection:    "tasks",
		SyncStartedAt: time.Now(),
	}
	_, err = s.Insert(currentSync)
	if err != nil {
		return fmt.Errorf("could not update last sync: %s", err.Error())
	}

	err = s.Find(tasks)
	if err != nil {
		return fmt.Errorf("could not get all tasks: %s", err.Error())
	}

	err = indexDummyTask()
	if err != nil {
		return fmt.Errorf("could not index dummy task: %w", err)
	}

	err = reindexTasks(s, tasks)
	if err != nil {
		return fmt.Errorf("could not reindex all tasks: %s", err.Error())
	}

	currentSync.SyncFinishedAt = time.Now()
	_, err = s.Where("collection = ?", "tasks").
		Cols("sync_finished_at").
		Update(currentSync)
	if err != nil {
		return fmt.Errorf("could update last sync state: %s", err.Error())
	}

	return
}

func getTypesenseTaskForTask(s *xorm.Session, task *Task, projectsCache map[int64]*Project) (ttask *typesenseTask, err error) {
	ttask = convertTaskToTypesenseTask(task)

	var p *Project
	if projectsCache == nil {
		p, err = GetProjectSimpleByID(s, task.ProjectID)
		if err != nil {
			return nil, fmt.Errorf("could not fetch project %d: %s", task.ProjectID, err.Error())
		}
	} else {
		var has bool
		p, has = projectsCache[task.ProjectID]
		if !has {
			p, err = GetProjectSimpleByID(s, task.ProjectID)
			if err != nil {
				return nil, fmt.Errorf("could not fetch project %d: %s", task.ProjectID, err.Error())
			}
			projectsCache[task.ProjectID] = p
		}
	}

	comment := &TaskComment{TaskID: task.ID}
	ttask.Comments, _, _, err = comment.ReadAll(s, &user.User{ID: p.OwnerID}, "", -1, -1)
	if err != nil {
		return nil, fmt.Errorf("could not fetch comments for task %d: %s", task.ID, err.Error())
	}

	return
}

func reindexTasks(s *xorm.Session, tasks map[int64]*Task) (err error) {

	if len(tasks) == 0 {
		log.Infof("No tasks to index")
		return
	}

	err = addMoreInfoToTasks(s, tasks, &user.User{ID: 1})
	if err != nil {
		return fmt.Errorf("could not fetch more task info: %s", err.Error())
	}

	projects := make(map[int64]*Project)

	typesenseTasks := []interface{}{}
	for _, task := range tasks {

		ttask, err := getTypesenseTaskForTask(s, task, projects)
		if err != nil {
			return err
		}

		typesenseTasks = append(typesenseTasks, ttask)
	}

	_, err = typesenseClient.Collection("tasks").
		Documents().
		Import(context.Background(), typesenseTasks, &api.ImportDocumentsParams{
			Action:    pointer.String("upsert"),
			BatchSize: pointer.Int(100),
		})
	if err != nil {
		log.Errorf("Could not upsert task into Typesense", err)
		return err
	}

	return nil
}

func indexDummyTask() (err error) {
	// The initial sync should contain one dummy task with all related fields populated so that typesense
	// creates the indexes properly. A little hacky, but gets the job done.
	dummyTask := &typesenseTask{
		ID:      "-100",
		Title:   "Dummytask",
		Created: time.Now().Unix(),
		Updated: time.Now().Unix(),
		Reminders: []*TaskReminder{
			{
				ID:             -10,
				TaskID:         -100,
				Reminder:       time.Now(),
				RelativePeriod: 10,
				RelativeTo:     ReminderRelationDueDate,
				Created:        time.Now(),
			},
		},
		Assignees: []*user.User{
			{
				ID:       -100,
				Username: "dummy",
				Name:     "dummy",
				Email:    "dummy@vikunja",
				Created:  time.Now(),
				Updated:  time.Now(),
			},
		},
		Labels: []*Label{
			{
				ID:          -110,
				Title:       "dummylabel",
				Description: "Lorem Ipsum Dummy",
				HexColor:    "000000",
				Created:     time.Now(),
				Updated:     time.Now(),
			},
		},
		Attachments: []*TaskAttachment{
			{
				ID:      -120,
				TaskID:  -100,
				Created: time.Now(),
			},
		},
		Comments: []*TaskComment{
			{
				ID:      -220,
				Comment: "Lorem Ipsum Dummy",
				Created: time.Now(),
				Updated: time.Now(),
				Author: &user.User{
					ID:       -100,
					Username: "dummy",
					Name:     "dummy",
					Email:    "dummy@vikunja",
					Created:  time.Now(),
					Updated:  time.Now(),
				},
			},
		},
	}

	_, err = typesenseClient.Collection("tasks").
		Documents().
		Create(context.Background(), dummyTask)
	if err != nil {
		return
	}

	_, err = typesenseClient.Collection("tasks").
		Document(dummyTask.ID).
		Delete(context.Background())
	return
}

type typesenseTask struct {
	ID                     string      `json:"id"`
	Title                  string      `json:"title"`
	Description            string      `json:"description"`
	Done                   bool        `json:"done"`
	DoneAt                 *int64      `json:"done_at"`
	DueDate                *int64      `json:"due_date"`
	ProjectID              int64       `json:"project_id"`
	RepeatAfter            int64       `json:"repeat_after"`
	RepeatMode             int         `json:"repeat_mode"`
	Priority               int64       `json:"priority"`
	StartDate              *int64      `json:"start_date"`
	EndDate                *int64      `json:"end_date"`
	HexColor               string      `json:"hex_color"`
	PercentDone            float64     `json:"percent_done"`
	Identifier             string      `json:"identifier"`
	Index                  int64       `json:"index"`
	UID                    string      `json:"uid"`
	CoverImageAttachmentID int64       `json:"cover_image_attachment_id"`
	Created                int64       `json:"created"`
	Updated                int64       `json:"updated"`
	BucketID               int64       `json:"bucket_id"`
	Position               float64     `json:"position"`
	KanbanPosition         float64     `json:"kanban_position"`
	CreatedByID            int64       `json:"created_by_id"`
	Reminders              interface{} `json:"reminders"`
	Assignees              interface{} `json:"assignees"`
	Labels                 interface{} `json:"labels"`
	//RelatedTasks           interface{} `json:"related_tasks"` // TODO
	Attachments interface{} `json:"attachments"`
	Comments    interface{} `json:"comments"`
}

func convertTaskToTypesenseTask(task *Task) *typesenseTask {
	tt := &typesenseTask{
		ID:                     fmt.Sprintf("%d", task.ID),
		Title:                  task.Title,
		Description:            task.Description,
		Done:                   task.Done,
		DoneAt:                 pointer.Int64(task.DoneAt.UTC().Unix()),
		DueDate:                pointer.Int64(task.DueDate.UTC().Unix()),
		ProjectID:              task.ProjectID,
		RepeatAfter:            task.RepeatAfter,
		RepeatMode:             int(task.RepeatMode),
		Priority:               task.Priority,
		StartDate:              pointer.Int64(task.StartDate.UTC().Unix()),
		EndDate:                pointer.Int64(task.EndDate.UTC().Unix()),
		HexColor:               task.HexColor,
		PercentDone:            task.PercentDone,
		Identifier:             task.Identifier,
		Index:                  task.Index,
		UID:                    task.UID,
		CoverImageAttachmentID: task.CoverImageAttachmentID,
		Created:                task.Created.UTC().Unix(),
		Updated:                task.Updated.UTC().Unix(),
		BucketID:               task.BucketID,
		Position:               task.Position,
		KanbanPosition:         task.KanbanPosition,
		CreatedByID:            task.CreatedByID,
		Reminders:              task.Reminders,
		Assignees:              task.Assignees,
		Labels:                 task.Labels,
		//RelatedTasks:           task.RelatedTasks,
		Attachments: task.Attachments,
	}

	if task.DoneAt.IsZero() {
		tt.DoneAt = nil
	}
	if task.DueDate.IsZero() {
		tt.DueDate = nil
	}
	if task.StartDate.IsZero() {
		tt.StartDate = nil
	}
	if task.EndDate.IsZero() {
		tt.EndDate = nil
	}

	return tt
}

func SyncUpdatedTasksIntoTypesense() (err error) {
	tasks := make(map[int64]*Task)

	s := db.NewSession()
	_ = s.Begin()
	defer s.Close()

	lastSync := &TypesenseSync{}
	has, err := s.Where("collection = ?", "tasks").
		Get(lastSync)
	if err != nil {
		_ = s.Rollback()
		return err
	}

	if !has {
		log.Errorf("[Typesense Sync] No typesense sync stats yet, please run a full index via the CLI first")
		_ = s.Rollback()
		return
	}

	currentSync := &TypesenseSync{SyncStartedAt: time.Now()}
	_, err = s.Where("collection = ?", "tasks").
		Cols("sync_started_at", "sync_finished_at").
		Update(currentSync)
	if err != nil {
		_ = s.Rollback()
		return
	}

	err = s.
		Where("updated >= ?", lastSync.SyncStartedAt).
		And("updated != created"). // new tasks are already indexed via the event handler
		Find(tasks)
	if err != nil {
		_ = s.Rollback()
		return
	}

	if len(tasks) > 0 {
		log.Debugf("[Typesense Sync] Updating %d tasks", len(tasks))

		err = reindexTasks(s, tasks)
		if err != nil {
			_ = s.Rollback()
			return
		}
	}

	if len(tasks) == 0 {
		log.Debugf("[Typesense Sync] No tasks changed since the last sync, not syncing")
	}

	currentSync.SyncFinishedAt = time.Now()
	_, err = s.Where("collection = ?", "tasks").
		Cols("sync_finished_at").
		Update(currentSync)
	if err != nil {
		_ = s.Rollback()
		return err
	}

	return s.Commit()
}

func RegisterPeriodicTypesenseResyncCron() {
	if !config.TypesenseEnabled.GetBool() {
		log.Debugf("[Typesense Sync] Typesense is disabled, not setting up sync cron")
		return
	}

	err := cron.Schedule("* * * * *", func() {
		err := SyncUpdatedTasksIntoTypesense()
		if err != nil {
			log.Fatalf("[Typesense Sync] Could not sync updated tasks into typesense: %s", err)
		}
	})
	if err != nil {
		log.Fatalf("[Typesense Sync] Could not register typesense resync cron: %s", err)
	}
}
