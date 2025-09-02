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
	"io"
	"time"

	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"xorm.io/xorm"
)

// AttachmentService represents a service for managing task attachments.
type AttachmentService struct {
	DB *xorm.Engine
}

// NewAttachmentService creates a new AttachmentService.
func NewAttachmentService(db *xorm.Engine) *AttachmentService {
	return &AttachmentService{
		DB: db,
	}
}

// AttachmentPermissions represents permission checking for attachments.
// This implements the "Move Logic, Don't Expose It" principle by moving permission logic from models to services.
type AttachmentPermissions struct {
	s          *xorm.Session
	attachment *models.TaskAttachment
	user       *user.User
	as         *AttachmentService
}

// Can returns a new AttachmentPermissions struct.
func (as *AttachmentService) Can(s *xorm.Session, attachment *models.TaskAttachment, u *user.User) *AttachmentPermissions {
	return &AttachmentPermissions{s: s, attachment: attachment, user: u, as: as}
}

// Read checks if the user can read the attachment.
// This implements the "Move Logic, Don't Expose It" principle by moving permission logic from models to services.
func (ap *AttachmentPermissions) Read() (bool, int, error) {
	if ap.user == nil {
		return false, 0, nil
	}

	// Check if user has read access to the task
	task := &models.Task{ID: ap.attachment.TaskID}
	return task.CanRead(ap.s, ap.user)
}

// Create checks if the user can create an attachment.
func (ap *AttachmentPermissions) Create() (bool, error) {
	if ap.user == nil {
		return false, nil
	}

	// User needs write access to the task to create attachments
	task, err := models.GetTaskByIDSimple(ap.s, ap.attachment.TaskID)
	if err != nil {
		return false, err
	}
	return task.CanCreate(ap.s, ap.user)
}

// Delete checks if the user can delete the attachment.
func (ap *AttachmentPermissions) Delete() (bool, error) {
	if ap.user == nil {
		return false, nil
	}

	// User needs write access to the task to delete attachments
	task := &models.Task{ID: ap.attachment.TaskID}
	return task.CanWrite(ap.s, ap.user)
}

// Create creates a new task attachment with file upload.
func (as *AttachmentService) Create(s *xorm.Session, attachment *models.TaskAttachment, f io.ReadCloser, filename string, size uint64, u *user.User) (*models.TaskAttachment, error) {
	// Check permissions
	can, err := as.Can(s, attachment, u).Create()
	if err != nil {
		return nil, err
	}
	if !can {
		return nil, models.ErrGenericForbidden{}
	}

	// Store the file
	file, err := files.Create(f, filename, size, u)
	if err != nil {
		if files.IsErrFileIsTooLarge(err) {
			return nil, models.ErrTaskAttachmentIsTooLarge{Size: size}
		}
		return nil, err
	}
	attachment.File = file
	attachment.FileID = file.ID

	// Set up the attachment
	attachment.ID = 0
	attachment.Created = time.Time{}
	attachment.CreatedBy, err = models.GetUserOrLinkShareUser(s, u)
	if err != nil {
		// Remove the uploaded file if adding it to the db fails
		if err2 := file.Delete(s); err2 != nil {
			return nil, err2
		}
		return nil, err
	}
	attachment.CreatedByID = attachment.CreatedBy.ID

	// Insert the attachment
	_, err = s.Insert(attachment)
	if err != nil {
		// Remove the uploaded file if adding it to the db fails
		if err2 := file.Delete(s); err2 != nil {
			return nil, err2
		}
		return nil, err
	}

	// Get the task for event dispatch
	task, err := models.GetTaskByIDSimple(s, attachment.TaskID)
	if err != nil {
		return nil, err
	}

	// Dispatch event
	err = events.Dispatch(&models.TaskAttachmentCreatedEvent{
		Task:       &task,
		Attachment: attachment,
		Doer:       attachment.CreatedBy,
	})
	if err != nil {
		return nil, err
	}

	return attachment, nil
}

// CreateWithoutPermissionCheck creates a new task attachment with file upload, bypassing permission checks.
// This method is intended for internal service operations like project duplication where permissions
// are already verified at a higher level.
func (as *AttachmentService) CreateWithoutPermissionCheck(s *xorm.Session, attachment *models.TaskAttachment, f io.ReadCloser, filename string, size uint64, u *user.User) (*models.TaskAttachment, error) {
	// Store the file
	file, err := files.Create(f, filename, size, u)
	if err != nil {
		if files.IsErrFileIsTooLarge(err) {
			return nil, models.ErrTaskAttachmentIsTooLarge{Size: size}
		}
		return nil, err
	}
	attachment.File = file
	attachment.FileID = file.ID

	// Set up the attachment
	attachment.ID = 0
	attachment.Created = time.Time{}
	attachment.CreatedBy, err = models.GetUserOrLinkShareUser(s, u)
	if err != nil {
		// Remove the uploaded file if adding it to the db fails
		if err2 := file.Delete(s); err2 != nil {
			return nil, err2
		}
		return nil, err
	}
	attachment.CreatedByID = attachment.CreatedBy.ID

	// Insert the attachment
	_, err = s.Insert(attachment)
	if err != nil {
		// Remove the uploaded file if adding it to the db fails
		if err2 := file.Delete(s); err2 != nil {
			return nil, err2
		}
		return nil, err
	}

	// Get the task for event dispatch
	task, err := models.GetTaskByIDSimple(s, attachment.TaskID)
	if err != nil {
		return nil, err
	}

	// Dispatch event
	err = events.Dispatch(&models.TaskAttachmentCreatedEvent{
		Task:       &task,
		Attachment: attachment,
		Doer:       attachment.CreatedBy,
	})
	if err != nil {
		return nil, err
	}

	return attachment, nil
}

// GetByID retrieves a single attachment by ID.
func (as *AttachmentService) GetByID(s *xorm.Session, attachmentID int64, taskID int64, u *user.User) (*models.TaskAttachment, error) {
	attachment := &models.TaskAttachment{
		ID:     attachmentID,
		TaskID: taskID,
	}

	// Check permissions
	can, _, err := as.Can(s, attachment, u).Read()
	if err != nil {
		return nil, err
	}
	if !can {
		return nil, models.ErrGenericForbidden{}
	}

	// Get the attachment
	err = as.getTaskAttachmentSimple(s, attachment)
	if err != nil {
		return nil, err
	}

	// Load file metadata
	attachment.File = &files.File{ID: attachment.FileID}
	err = attachment.File.LoadFileMetaByID()
	if err != nil {
		return nil, err
	}

	// Load user information
	users, err := models.GetUsersOrLinkSharesFromIDs(s, []int64{attachment.CreatedByID})
	if err != nil {
		return nil, err
	}
	if createdBy, has := users[attachment.CreatedByID]; has {
		attachment.CreatedBy = createdBy
	}

	return attachment, nil
}

// GetAllForTask retrieves all attachments for a task with pagination.
func (as *AttachmentService) GetAllForTask(s *xorm.Session, taskID int64, u *user.User, page int, perPage int) ([]*models.TaskAttachment, int, int64, error) {
	attachment := &models.TaskAttachment{TaskID: taskID}

	// Check permissions - user needs read access to the task
	can, _, err := as.Can(s, attachment, u).Read()
	if err != nil {
		return nil, 0, 0, err
	}
	if !can {
		return nil, 0, 0, models.ErrGenericForbidden{}
	}

	attachments := []*models.TaskAttachment{}

	limit, start := as.getLimitFromPageIndex(page, perPage)

	query := s.Where("task_id = ?", taskID)
	if limit > 0 {
		query = query.Limit(limit, start)
	}
	err = query.Find(&attachments)
	if err != nil {
		return nil, 0, 0, err
	}

	if len(attachments) == 0 {
		return attachments, 0, 0, nil
	}

	// Get file and user information
	fileIDs := make([]int64, 0, len(attachments))
	userIDs := make([]int64, 0, len(attachments))
	for _, r := range attachments {
		fileIDs = append(fileIDs, r.FileID)
		userIDs = append(userIDs, r.CreatedByID)
	}

	fs := make(map[int64]*files.File)
	err = s.In("id", fileIDs).Find(&fs)
	if err != nil {
		return nil, 0, 0, err
	}

	users, err := models.GetUsersOrLinkSharesFromIDs(s, userIDs)
	if err != nil {
		return nil, 0, 0, err
	}

	for _, r := range attachments {
		if createdBy, has := users[r.CreatedByID]; has {
			r.CreatedBy = createdBy
		}

		// If the actual file does not exist, don't try to load it as that would fail with nil panic
		if _, exists := fs[r.FileID]; !exists {
			continue
		}
		r.File = fs[r.FileID]
	}

	numberOfTotalItems, err := s.
		Where("task_id = ?", taskID).
		Count(&models.TaskAttachment{})

	return attachments, len(attachments), numberOfTotalItems, err
}

// Delete removes an attachment and its underlying file.
func (as *AttachmentService) Delete(s *xorm.Session, attachmentID int64, taskID int64, u *user.User) error {
	attachment := &models.TaskAttachment{
		ID:     attachmentID,
		TaskID: taskID,
	}

	// Check permissions
	can, err := as.Can(s, attachment, u).Delete()
	if err != nil {
		return err
	}
	if !can {
		return models.ErrGenericForbidden{}
	}

	// Load the attachment first
	err = as.getTaskAttachmentSimple(s, attachment)
	if err != nil {
		return err
	}

	// Load file information
	attachment.File = &files.File{ID: attachment.FileID}
	err = attachment.File.LoadFileMetaByID()
	if err != nil && !files.IsErrFileDoesNotExist(err) {
		return err
	}

	// Delete from database first
	_, err = s.
		Where("task_id = ? AND id = ?", taskID, attachmentID).
		Delete(&models.TaskAttachment{})
	if err != nil {
		return err
	}

	// Delete the underlying file
	if attachment.File != nil {
		err = attachment.File.Delete(s)
		// If the file does not exist, we don't want to error out
		if err != nil && !files.IsErrFileDoesNotExist(err) {
			return err
		}
	}

	// Get task and user for event dispatch
	task, err := models.GetTaskByIDSimple(s, taskID)
	if err != nil {
		return err
	}

	doer, _ := user.GetFromAuth(u)

	return events.Dispatch(&models.TaskAttachmentDeletedEvent{
		Task:       &task,
		Attachment: attachment,
		Doer:       doer,
	})
}

// GetPreview returns a preview image for an attachment if it's an image.
func (as *AttachmentService) GetPreview(s *xorm.Session, attachmentID int64, taskID int64, previewSize models.PreviewSize, u *user.User) ([]byte, error) {
	attachment := &models.TaskAttachment{
		ID:     attachmentID,
		TaskID: taskID,
	}

	// Check permissions
	can, _, err := as.Can(s, attachment, u).Read()
	if err != nil {
		return nil, err
	}
	if !can {
		return nil, models.ErrGenericForbidden{}
	}

	// Get the attachment
	err = as.getTaskAttachmentSimple(s, attachment)
	if err != nil {
		return nil, err
	}

	// Load the file
	attachment.File = &files.File{ID: attachment.FileID}
	err = attachment.File.LoadFileByID()
	if err != nil {
		return nil, err
	}

	// Generate preview
	return attachment.GetPreview(previewSize), nil
}

// getTaskAttachmentSimple retrieves an attachment by ID. Logic moved from models.
func (as *AttachmentService) getTaskAttachmentSimple(s *xorm.Session, ta *models.TaskAttachment) error {
	exists, err := s.
		Where("id = ? AND task_id = ?", ta.ID, ta.TaskID).
		NoAutoCondition().
		Get(ta)
	if err != nil {
		return err
	}
	if !exists {
		return models.ErrTaskAttachmentDoesNotExist{
			TaskID:       ta.TaskID,
			AttachmentID: ta.ID,
		}
	}

	return nil
}

// getLimitFromPageIndex calculates limit and start offset for pagination.
func (as *AttachmentService) getLimitFromPageIndex(page int, perPage int) (limit int, start int) {
	if page == 0 {
		page = 1
	}
	if perPage == 0 {
		perPage = 50
	}
	start = (page - 1) * perPage
	return perPage, start
}

// init sets up dependency injection for attachment-related model functions.
func init() {
	// Wire model functions to service implementations
	models.AttachmentCreateFunc = func(s *xorm.Session, attachment *models.TaskAttachment, f io.ReadCloser, filename string, size uint64, u *user.User) (*models.TaskAttachment, error) {
		return NewAttachmentService(s.Engine()).Create(s, attachment, f, filename, size, u)
	}

	models.AttachmentDeleteFunc = func(s *xorm.Session, attachmentID int64, taskID int64, u *user.User) error {
		return NewAttachmentService(s.Engine()).Delete(s, attachmentID, taskID, u)
	}
}

// InitAttachmentService sets up dependency injection for attachment-related model functions.
// This function must be called during test initialization to ensure models can call services.
func InitAttachmentService() {
	models.AttachmentCreateFunc = func(s *xorm.Session, attachment *models.TaskAttachment, f io.ReadCloser, filename string, size uint64, u *user.User) (*models.TaskAttachment, error) {
		return NewAttachmentService(s.Engine()).Create(s, attachment, f, filename, size, u)
	}

	models.AttachmentDeleteFunc = func(s *xorm.Session, attachmentID int64, taskID int64, u *user.User) error {
		return NewAttachmentService(s.Engine()).Delete(s, attachmentID, taskID, u)
	}
}
