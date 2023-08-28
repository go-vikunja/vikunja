// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2023 Vikunja and contributors. All rights reserved.
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
	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"
	"fmt"

	"github.com/typesense/typesense-go/typesense"
	"github.com/typesense/typesense-go/typesense/api"
	"github.com/typesense/typesense-go/typesense/api/pointer"
)

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
			},
			{
				Name: "title",
				Type: "string",
			},
			{
				Name: "description",
				Type: "string",
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
			},
			{
				Name: "percent_done",
				Type: "float",
			},
			{
				Name: "identifier",
				Type: "string",
			},
			{
				Name: "index",
				Type: "int64",
			},
			{
				Name: "uid",
				Type: "string",
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
	_, _ = typesenseClient.Collection("tasks").Delete()

	_, err := typesenseClient.Collections().Create(taskSchema)
	return err
}

func ReindexAllTasks() (err error) {
	tasks := make(map[int64]*Task)

	s := db.NewSession()
	defer s.Close()

	err = s.Find(tasks)
	if err != nil {
		return err
	}

	err = addMoreInfoToTasks(s, tasks, &user.User{ID: 1})
	if err != nil {
		return err
	}

	for _, task := range tasks {
		searchTask := convertTaskToTypesenseTask(task)

		comment := &TaskComment{TaskID: task.ID}
		searchTask.Comments, _, _, err = comment.ReadAll(s, task.CreatedBy, "", -1, -1)
		if err != nil {
			return err
		}

		_, err = typesenseClient.Collection("tasks").
			Documents().
			Create(searchTask)
		if err != nil {
			return err
		}
	}

	return nil
}

type typesenseTask struct {
	ID                     string      `json:"id"`
	Title                  string      `json:"title"`
	Description            string      `json:"description"`
	Done                   bool        `json:"done"`
	DoneAt                 int64       `json:"done_at"`
	DueDate                int64       `json:"due_date"`
	ProjectID              int64       `json:"project_id"`
	RepeatAfter            int64       `json:"repeat_after"`
	RepeatMode             int         `json:"repeat_mode"`
	Priority               int64       `json:"priority"`
	StartDate              int64       `json:"start_date"`
	EndDate                int64       `json:"end_date"`
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
		DoneAt:                 task.DoneAt.UTC().Unix(),
		DueDate:                task.DueDate.UTC().Unix(),
		ProjectID:              task.ProjectID,
		RepeatAfter:            task.RepeatAfter,
		RepeatMode:             int(task.RepeatMode),
		Priority:               task.Priority,
		StartDate:              task.StartDate.UTC().Unix(),
		EndDate:                task.EndDate.UTC().Unix(),
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
		tt.DoneAt = 0
	}
	if task.DueDate.IsZero() {
		tt.DueDate = 0
	}
	if task.StartDate.IsZero() {
		tt.StartDate = 0
	}
	if task.EndDate.IsZero() {
		tt.EndDate = 0
	}

	return tt
}
