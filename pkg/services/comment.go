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

package services

import (
	"time"

	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"
	"xorm.io/builder"
	"xorm.io/xorm"
)

// CommentService represents a service for managing task comments.
type CommentService struct {
	DB *xorm.Engine
}

// NewCommentService creates a new CommentService.
func NewCommentService(db *xorm.Engine) *CommentService {
	return &CommentService{
		DB: db,
	}
}

// CommentPermissions represents permission checking for comments.
// This implements the "Move Logic, Don't Expose It" principle by moving permission logic from models to services.
type CommentPermissions struct {
	s       *xorm.Session
	comment *models.TaskComment
	user    *user.User
	cs      *CommentService
}

// Can returns a new CommentPermissions struct.
func (cs *CommentService) Can(s *xorm.Session, comment *models.TaskComment, u *user.User) *CommentPermissions {
	return &CommentPermissions{s: s, comment: comment, user: u, cs: cs}
}

// Read checks if the user can read the comment.
// This implements the "Move Logic, Don't Expose It" principle by moving permission logic from models to services.
func (cp *CommentPermissions) Read() (bool, int, error) {
	if cp.user == nil {
		return false, 0, nil
	}

	// Check if user has read access to the task
	task := &models.Task{ID: cp.comment.TaskID}
	canRead, maxPermission, err := task.CanRead(cp.s, cp.user)
	if err != nil {
		return false, 0, err
	}

	return canRead, maxPermission, nil
}

// Create checks if the user can create a comment.
func (cp *CommentPermissions) Create() (bool, error) {
	if cp.user == nil {
		return false, nil
	}

	// User needs write access to the task to create comments
	task := &models.Task{ID: cp.comment.TaskID}
	return task.CanWrite(cp.s, cp.user)
}

// Update checks if the user can update the comment.
func (cp *CommentPermissions) Update() (bool, error) {
	if cp.user == nil {
		return false, nil
	}

	return cp.canUserModifyTaskComment()
}

// Delete checks if the user can delete the comment.
func (cp *CommentPermissions) Delete() (bool, error) {
	if cp.user == nil {
		return false, nil
	}

	return cp.canUserModifyTaskComment()
}

// canUserModifyTaskComment checks if a user can modify (update/delete) a task comment.
// This logic is moved from models.TaskComment.canUserModifyTaskComment.
func (cp *CommentPermissions) canUserModifyTaskComment() (bool, error) {
	// First check if user has write access to the task
	task := &models.Task{ID: cp.comment.TaskID}
	canWriteTask, err := task.CanWrite(cp.s, cp.user)
	if err != nil {
		return false, err
	}
	if !canWriteTask {
		return false, nil
	}

	// Then check if the user is the author of the comment
	savedComment := &models.TaskComment{
		ID:     cp.comment.ID,
		TaskID: cp.comment.TaskID,
	}
	err = cp.getTaskCommentSimple(savedComment)
	if err != nil {
		return false, err
	}

	return cp.user.GetID() == savedComment.AuthorID, nil
}

// getTaskCommentSimple retrieves a comment by ID. Logic moved from models.
func (cp *CommentPermissions) getTaskCommentSimple(tc *models.TaskComment) error {
	exists, err := cp.s.
		Where("id = ?", tc.ID).
		NoAutoCondition().
		Get(tc)
	if err != nil {
		return err
	}
	if !exists {
		return models.ErrTaskCommentDoesNotExist{
			ID:     tc.ID,
			TaskID: tc.TaskID,
		}
	}

	return nil
}

// Create creates a new task comment.
func (cs *CommentService) Create(s *xorm.Session, comment *models.TaskComment, u *user.User) (*models.TaskComment, error) {
	// Check permissions
	can, err := cs.Can(s, comment, u).Create()
	if err != nil {
		return nil, err
	}
	if !can {
		return nil, models.ErrGenericForbidden{}
	}

	// Verify the task exists
	task, err := models.GetTaskSimple(s, &models.Task{ID: comment.TaskID})
	if err != nil {
		return nil, err
	}

	// Set up the comment
	comment.ID = 0
	comment.Created = time.Time{}
	comment.Updated = time.Time{}
	comment.Author, err = models.GetUserOrLinkShareUser(s, u)
	if err != nil {
		return nil, err
	}
	comment.AuthorID = comment.Author.ID

	// Insert the comment
	_, err = s.Insert(comment)
	if err != nil {
		return nil, err
	}

	// Dispatch event
	err = events.Dispatch(&models.TaskCommentCreatedEvent{
		Task:    &task,
		Comment: comment,
		Doer:    comment.Author,
	})
	if err != nil {
		return nil, err
	}

	return comment, nil
}

// GetByID retrieves a single comment by ID.
func (cs *CommentService) GetByID(s *xorm.Session, commentID int64, u *user.User) (*models.TaskComment, error) {
	comment := &models.TaskComment{ID: commentID}

	// Get the comment first to get the task ID
	err := cs.getTaskCommentSimple(s, comment)
	if err != nil {
		return nil, err
	}

	// Check permissions
	can, _, err := cs.Can(s, comment, u).Read()
	if err != nil {
		return nil, err
	}
	if !can {
		return nil, models.ErrGenericForbidden{}
	}

	// Get full comment with author
	exists, err := s.
		Where("task_comments.id = ?", commentID).
		Join("LEFT", "users", "users.id = task_comments.author_id").
		Get(comment)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, models.ErrTaskCommentDoesNotExist{ID: commentID}
	}

	// Get author details
	if comment.AuthorID > 0 {
		comment.Author, err = user.GetUserByID(s, comment.AuthorID)
		if err != nil {
			return nil, err
		}
	} else {
		// Link share user
		comment.Author = &user.User{
			ID:       comment.AuthorID,
			Username: "link-share",
			Name:     "Link Share",
		}
	}

	return comment, nil
}

// GetAllForTask retrieves all comments for a specific task.
func (cs *CommentService) GetAllForTask(s *xorm.Session, taskID int64, u *user.User, search string, page, perPage int) (interface{}, int, int64, error) {
	// Create a comment with task ID for permission checking
	testComment := &models.TaskComment{TaskID: taskID}

	// Check permissions
	can, _, err := cs.Can(s, testComment, u).Read()
	if err != nil {
		return nil, 0, 0, err
	}
	if !can {
		return nil, 0, 0, models.ErrGenericForbidden{}
	}

	return cs.getAllCommentsForTasksWithoutPermissionCheck(s, []int64{taskID}, search, page, perPage)
}

// Update updates an existing task comment.
func (cs *CommentService) Update(s *xorm.Session, comment *models.TaskComment, u *user.User) (*models.TaskComment, error) {
	// Get existing comment to verify it exists and get task ID
	existing := &models.TaskComment{ID: comment.ID}
	err := cs.getTaskCommentSimple(s, existing)
	if err != nil {
		return nil, err
	}
	comment.TaskID = existing.TaskID

	// Check permissions
	can, err := cs.Can(s, comment, u).Update()
	if err != nil {
		return nil, err
	}
	if !can {
		return nil, models.ErrGenericForbidden{}
	}

	// Update the comment
	updated, err := s.
		ID(comment.ID).
		Cols("comment").
		Update(comment)
	if updated == 0 {
		return nil, models.ErrTaskCommentDoesNotExist{ID: comment.ID}
	}
	if err != nil {
		return nil, err
	}

	// Get the task for event
	task, err := models.GetTaskSimple(s, &models.Task{ID: comment.TaskID})
	if err != nil {
		return nil, err
	}

	// Get author for event
	comment.Author, err = user.GetUserByID(s, u.ID)
	if err != nil {
		return nil, err
	}

	// Dispatch event
	err = events.Dispatch(&models.TaskCommentUpdatedEvent{
		Task:    &task,
		Comment: comment,
		Doer:    comment.Author,
	})
	if err != nil {
		return nil, err
	}

	return comment, nil
}

// Delete removes a task comment.
func (cs *CommentService) Delete(s *xorm.Session, commentID int64, u *user.User) error {
	// Get the comment to verify it exists and get task ID
	comment := &models.TaskComment{ID: commentID}
	err := cs.getTaskCommentSimple(s, comment)
	if err != nil {
		return err
	}

	// Check permissions
	can, err := cs.Can(s, comment, u).Delete()
	if err != nil {
		return err
	}
	if !can {
		return models.ErrGenericForbidden{}
	}

	// Delete the comment
	deleted, err := s.
		ID(commentID).
		NoAutoCondition().
		Delete(&models.TaskComment{})
	if deleted == 0 {
		return models.ErrTaskCommentDoesNotExist{ID: commentID}
	}
	if err != nil {
		return err
	}

	// Get the task for event
	task, err := models.GetTaskByIDSimple(s, comment.TaskID)
	if err != nil {
		return err
	}

	// Get author for event
	comment.Author, err = user.GetUserByID(s, u.ID)
	if err != nil {
		return err
	}

	// Dispatch event
	return events.Dispatch(&models.TaskCommentDeletedEvent{
		Task:    &task,
		Comment: comment,
		Doer:    comment.Author,
	})
}

// getAllCommentsForTasksWithoutPermissionCheck retrieves comments for tasks without permission checks.
// This is used internally after permissions have been verified.
func (cs *CommentService) getAllCommentsForTasksWithoutPermissionCheck(s *xorm.Session, taskIDs []int64, search string, page int, perPage int) (result []*models.TaskComment, resultCount int, numberOfTotalItems int64, err error) {
	// Helper struct for joins
	type TaskCommentWithAuthor struct {
		models.TaskComment
		AuthorFromDB *user.User `xorm:"extends" json:"-"`
	}

	limit, start := cs.getLimitFromPageIndex(page, perPage)
	comments := []*models.TaskComment{}
	where := []builder.Cond{
		builder.In("task_id", taskIDs),
	}

	if search != "" {
		where = append(where, builder.Like{"comment", "%" + search + "%"})
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
		return comments, 0, 0, err
	}

	// Set up authors
	for _, comment := range comments {
		if comment.AuthorID > 0 {
			comment.Author, _ = user.GetUserByID(s, comment.AuthorID)
		} else {
			// Link share user
			comment.Author = &user.User{
				ID:       comment.AuthorID,
				Username: "link-share",
				Name:     "Link Share",
			}
		}
	}

	// Get total count
	numberOfTotalItems, err = s.
		Where(builder.And(where...)).
		Count(&models.TaskComment{})
	return comments, len(comments), numberOfTotalItems, err
}

// getTaskCommentSimple retrieves a comment by ID. Logic moved from models.
func (cs *CommentService) getTaskCommentSimple(s *xorm.Session, tc *models.TaskComment) error {
	exists, err := s.
		Where("id = ?", tc.ID).
		NoAutoCondition().
		Get(tc)
	if err != nil {
		return err
	}
	if !exists {
		return models.ErrTaskCommentDoesNotExist{
			ID:     tc.ID,
			TaskID: tc.TaskID,
		}
	}

	return nil
}

// getLimitFromPageIndex calculates limit and start offset for pagination.
func (cs *CommentService) getLimitFromPageIndex(page int, perPage int) (limit int, start int) {
	if page == 0 {
		page = 1
	}
	if perPage == 0 {
		perPage = 20
	}
	start = (page - 1) * perPage
	return perPage, start
}

// InitCommentService sets up dependency injection for comment-related model functions.
// This function must be called during test initialization to ensure models can call services.
func InitCommentService() {
	// Wire model functions to service implementations
	models.TaskCommentCreateFunc = func(s *xorm.Session, comment *models.TaskComment, u *user.User) error {
		_, err := NewCommentService(s.Engine()).Create(s, comment, u)
		return err
	}

	models.TaskCommentUpdateFunc = func(s *xorm.Session, comment *models.TaskComment, u *user.User) error {
		_, err := NewCommentService(s.Engine()).Update(s, comment, u)
		return err
	}

	models.TaskCommentDeleteFunc = func(s *xorm.Session, commentID int64, u *user.User) error {
		return NewCommentService(s.Engine()).Delete(s, commentID, u)
	}

	models.TaskCommentReadAllFunc = func(s *xorm.Session, taskID int64, auth web.Auth, search string, page int, perPage int) (interface{}, int, int64, error) {
		// Convert web.Auth to *user.User
		var u *user.User
		if auth != nil {
			var err error
			u, err = models.GetUserOrLinkShareUser(s, auth)
			if err != nil {
				return nil, 0, 0, err
			}
		}
		return NewCommentService(s.Engine()).GetAllForTask(s, taskID, u, search, page, perPage)
	}
}

// AddCommentsToTasks adds comments to the provided task map for the given task IDs
func (cs *CommentService) AddCommentsToTasks(s *xorm.Session, taskIDs []int64, taskMap map[int64]*models.Task) error {
	if len(taskIDs) == 0 {
		return nil
	}

	var comments []*models.TaskComment
	err := s.In("task_id", taskIDs).Find(&comments)
	if err != nil {
		return err
	}

	// Group comments by task ID
	commentMap := make(map[int64][]*models.TaskComment)
	for _, comment := range comments {
		commentMap[comment.TaskID] = append(commentMap[comment.TaskID], comment)
	}

	// Add comments to each task
	for taskID, task := range taskMap {
		if comments, exists := commentMap[taskID]; exists {
			task.Comments = comments
		}
	}

	return nil
}
