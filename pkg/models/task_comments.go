// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2021 Vikunja and contributors. All rights reserved.
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
	"time"

	"code.vikunja.io/api/pkg/events"

	"xorm.io/xorm"

	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/web"
)

// TaskComment represents a task comment
type TaskComment struct {
	ID       int64      `xorm:"autoincr pk unique not null" json:"id" param:"commentid"`
	Comment  string     `xorm:"text not null" json:"comment"`
	AuthorID int64      `xorm:"not null" json:"-"`
	Author   *user.User `xorm:"-" json:"author"`
	TaskID   int64      `xorm:"not null" json:"-" param:"task"`

	Created time.Time `xorm:"created" json:"created"`
	Updated time.Time `xorm:"updated" json:"updated"`

	web.CRUDable `xorm:"-" json:"-"`
	web.Rights   `xorm:"-" json:"-"`
}

// TableName holds the table name for the task comments table
func (tc *TaskComment) TableName() string {
	return "task_comments"
}

// Create creates a new task comment
// @Summary Create a new task comment
// @Description Create a new task comment. The user doing this need to have at least write access to the task this comment should belong to.
// @tags task
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param relation body models.TaskComment true "The task comment object"
// @Param taskID path int true "Task ID"
// @Success 201 {object} models.TaskComment "The created task comment object."
// @Failure 400 {object} web.HTTPError "Invalid task comment object provided."
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{taskID}/comments [put]
func (tc *TaskComment) Create(s *xorm.Session, a web.Auth) (err error) {
	// Check if the task exists
	task, err := GetTaskSimple(s, &Task{ID: tc.TaskID})
	if err != nil {
		return err
	}

	tc.Author, err = GetUserOrLinkShareUser(s, a)
	if err != nil {
		return err
	}
	tc.AuthorID = tc.Author.ID

	_, err = s.Insert(tc)
	if err != nil {
		return
	}

	return events.Dispatch(&TaskCommentCreatedEvent{
		Task:    &task,
		Comment: tc,
		Doer:    tc.Author,
	})
}

// Delete removes a task comment
// @Summary Remove a task comment
// @Description Remove a task comment. The user doing this need to have at least write access to the task this comment belongs to.
// @tags task
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param taskID path int true "Task ID"
// @Param commentID path int true "Comment ID"
// @Success 200 {object} models.Message "The task comment was successfully deleted."
// @Failure 400 {object} web.HTTPError "Invalid task comment object provided."
// @Failure 404 {object} web.HTTPError "The task comment was not found."
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{taskID}/comments/{commentID} [delete]
func (tc *TaskComment) Delete(s *xorm.Session, a web.Auth) error {
	deleted, err := s.
		ID(tc.ID).
		NoAutoCondition().
		Delete(tc)
	if deleted == 0 {
		return ErrTaskCommentDoesNotExist{ID: tc.ID}
	}
	return err
}

// Update updates a task text by its ID
// @Summary Update an existing task comment
// @Description Update an existing task comment. The user doing this need to have at least write access to the task this comment belongs to.
// @tags task
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param taskID path int true "Task ID"
// @Param commentID path int true "Comment ID"
// @Success 200 {object} models.TaskComment "The updated task comment object."
// @Failure 400 {object} web.HTTPError "Invalid task comment object provided."
// @Failure 404 {object} web.HTTPError "The task comment was not found."
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{taskID}/comments/{commentID} [post]
func (tc *TaskComment) Update(s *xorm.Session, a web.Auth) error {
	updated, err := s.
		ID(tc.ID).
		Cols("comment").
		Update(tc)
	if updated == 0 {
		return ErrTaskCommentDoesNotExist{ID: tc.ID}
	}
	return err
}

// ReadOne handles getting a single comment
// @Summary Remove a task comment
// @Description Remove a task comment. The user doing this need to have at least read access to the task this comment belongs to.
// @tags task
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param taskID path int true "Task ID"
// @Param commentID path int true "Comment ID"
// @Success 200 {object} models.TaskComment "The task comment object."
// @Failure 400 {object} web.HTTPError "Invalid task comment object provided."
// @Failure 404 {object} web.HTTPError "The task comment was not found."
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{taskID}/comments/{commentID} [get]
func (tc *TaskComment) ReadOne(s *xorm.Session, a web.Auth) (err error) {
	exists, err := s.Get(tc)
	if err != nil {
		return
	}
	if !exists {
		return ErrTaskCommentDoesNotExist{
			ID:     tc.ID,
			TaskID: tc.TaskID,
		}
	}

	// Get the author
	author := &user.User{}
	_, err = s.
		Where("id = ?", tc.AuthorID).
		Get(author)
	tc.Author = author
	return
}

// ReadAll returns all comments for a task
// @Summary Get all task comments
// @Description Get all task comments. The user doing this need to have at least read access to the task.
// @tags task
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param taskID path int true "Task ID"
// @Success 200 {array} models.TaskComment "The array with all task comments"
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{taskID}/comments [get]
func (tc *TaskComment) ReadAll(s *xorm.Session, auth web.Auth, search string, page int, perPage int) (result interface{}, resultCount int, numberOfTotalItems int64, err error) {

	// Check if the user has access to the task
	canRead, _, err := tc.CanRead(s, auth)
	if err != nil {
		return nil, 0, 0, err
	}
	if !canRead {
		return nil, 0, 0, ErrGenericForbidden{}
	}

	// Because we can't extend the type in general, we need to do this here.
	// Not a good solution, but saves performance.
	type TaskCommentWithAuthor struct {
		TaskComment
		AuthorFromDB *user.User `xorm:"extends" json:"-"`
	}

	limit, start := getLimitFromPageIndex(page, perPage)

	comments := []*TaskComment{}
	query := s.
		Where("task_id = ? AND comment like ?", tc.TaskID, "%"+search+"%").
		Join("LEFT", "users", "users.id = task_comments.author_id")
	if limit > 0 {
		query = query.Limit(limit, start)
	}
	err = query.Find(&comments)
	if err != nil {
		return
	}

	var authorIDs []int64
	for _, comment := range comments {
		authorIDs = append(authorIDs, comment.AuthorID)
	}

	authors, err := getUsersOrLinkSharesFromIDs(s, authorIDs)
	if err != nil {
		return
	}

	for _, comment := range comments {
		comment.Author = authors[comment.AuthorID]
	}

	numberOfTotalItems, err = s.
		Where("task_id = ? AND comment like ?", tc.TaskID, "%"+search+"%").
		Count(&TaskCommentWithAuthor{})
	return comments, len(comments), numberOfTotalItems, err
}
