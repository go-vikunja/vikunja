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
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"xorm.io/builder"
	"xorm.io/xorm"
)

// TaskComment represents a task comment
type TaskComment struct {
	ID       int64      `xorm:"autoincr pk unique not null" json:"id" param:"commentid"`
	Comment  string     `xorm:"text not null" json:"comment" valid:"dbtext,required"`
	AuthorID int64      `xorm:"not null" json:"-"`
	Author   *user.User `xorm:"-" json:"author"`
	TaskID   int64      `xorm:"not null" json:"-" param:"task"`

	Reactions ReactionMap `xorm:"-" json:"reactions"`

	Created time.Time `xorm:"created" json:"created"`
	Updated time.Time `xorm:"updated" json:"updated"`

	web.CRUDable    `xorm:"-" json:"-"`
	web.Permissions `xorm:"-" json:"-"`
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

	tc.ID = 0
	tc.Created = time.Time{}
	tc.Updated = time.Time{}

	return tc.CreateWithTimestamps(s, a)
}

func (tc *TaskComment) CreateWithTimestamps(s *xorm.Session, a web.Auth) (err error) {
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

	if !tc.Created.IsZero() && !tc.Updated.IsZero() {
		_, err = s.NoAutoTime().Insert(tc)
		if err != nil {
			return
		}
	} else {
		_, err = s.Insert(tc)
		if err != nil {
			return
		}
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
func (tc *TaskComment) Delete(s *xorm.Session, _ web.Auth) error {
	deleted, err := s.
		ID(tc.ID).
		NoAutoCondition().
		Delete(tc)
	if deleted == 0 {
		return ErrTaskCommentDoesNotExist{ID: tc.ID}
	}

	if err != nil {
		return err
	}

	task, err := GetTaskByIDSimple(s, tc.TaskID)
	if err != nil {
		return err
	}

	return events.Dispatch(&TaskCommentDeletedEvent{
		Task:    &task,
		Comment: tc,
		Doer:    tc.Author,
	})
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
func (tc *TaskComment) Update(s *xorm.Session, _ web.Auth) error {
	updated, err := s.
		ID(tc.ID).
		Cols("comment").
		Update(tc)
	if updated == 0 {
		return ErrTaskCommentDoesNotExist{ID: tc.ID}
	}

	if err != nil {
		return err
	}

	task, err := GetTaskSimple(s, &Task{ID: tc.TaskID})
	if err != nil {
		return err
	}

	return events.Dispatch(&TaskCommentUpdatedEvent{
		Task:    &task,
		Comment: tc,
		Doer:    tc.Author,
	})
}

func getTaskCommentSimple(s *xorm.Session, tc *TaskComment) error {
	exists, err := s.
		Where("id = ?", tc.ID).
		NoAutoCondition().
		Get(tc)
	if err != nil {
		return err
	}
	if !exists {
		return ErrTaskCommentDoesNotExist{
			ID:     tc.ID,
			TaskID: tc.TaskID,
		}
	}

	return nil
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
func (tc *TaskComment) ReadOne(s *xorm.Session, _ web.Auth) (err error) {
	err = getTaskCommentSimple(s, tc)
	if err != nil {
		return err
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

	return getAllCommentsForTasksWithoutPermissionCheck(s, []int64{tc.TaskID}, search, page, perPage)
}

func addCommentsToTasks(s *xorm.Session, taskIDs []int64, taskMap map[int64]*Task) (err error) {
	comments, _, _, err := getAllCommentsForTasksWithoutPermissionCheck(s, taskIDs, "", 0, 50)
	if err != nil {
		return err
	}

	for _, comment := range comments {
		if task, exists := taskMap[comment.TaskID]; exists {
			if task.Comments == nil {
				task.Comments = []*TaskComment{}
			}
			task.Comments = append(task.Comments, comment)
		}
	}

	return nil
}

func getAllCommentsForTasksWithoutPermissionCheck(s *xorm.Session, taskIDs []int64, search string, page int, perPage int) (result []*TaskComment, resultCount int, numberOfTotalItems int64, err error) {
	// Because we can't extend the type in general, we need to do this here.
	// Not a good solution, but saves performance.
	type TaskCommentWithAuthor struct {
		TaskComment
		AuthorFromDB *user.User `xorm:"extends" json:"-"`
	}

	limit, start := getLimitFromPageIndex(page, perPage)
	comments := []*TaskComment{}
	where := []builder.Cond{
		builder.In("task_id", taskIDs),
	}

	if search != "" {
		where = append(where, db.ILIKE("comment", search))
	}
	query := s.
		Where(builder.And(where...)).
		Join("LEFT", "users", "users.id = task_comments.author_id").
		OrderBy("task_comments.created asc")
	if limit > 0 {
		query = query.Limit(limit, start)
	}
	err = query.Find(&comments)
	if err != nil {
		return
	}

	var authorIDs []int64
	var commentIDs []int64
	for _, comment := range comments {
		authorIDs = append(authorIDs, comment.AuthorID)
		commentIDs = append(commentIDs, comment.ID)
	}

	authors, err := getUsersOrLinkSharesFromIDs(s, authorIDs)
	if err != nil {
		return
	}

	reactions, err := getReactionsForEntityIDs(s, ReactionKindComment, commentIDs)
	if err != nil {
		return
	}

	for _, comment := range comments {
		comment.Author = authors[comment.AuthorID]
		r, has := reactions[comment.ID]
		if has {
			comment.Reactions = r
		}
	}

	var totalItemsQuery = s.In("task_id", taskIDs)
	if search != "" {
		totalItemsQuery = totalItemsQuery.And("comment like ?", "%"+search+"%")
	}
	numberOfTotalItems, err = totalItemsQuery.Count(&TaskCommentWithAuthor{})
	return comments, len(comments), numberOfTotalItems, err
}
