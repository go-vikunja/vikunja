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
	"bytes"
	"fmt"
	"image"
	"image/png"
	"io"
	"strconv"
	"time"

	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/modules/keyvalue"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"github.com/disintegration/imaging"
	"xorm.io/xorm"
)

// TaskAttachment is the definition of a task attachment
type TaskAttachment struct {
	ID     int64 `xorm:"bigint autoincr not null unique pk" json:"id" param:"attachment"`
	TaskID int64 `xorm:"bigint not null" json:"task_id" param:"task"`
	FileID int64 `xorm:"bigint not null" json:"-"`

	CreatedByID int64      `xorm:"bigint not null" json:"-"`
	CreatedBy   *user.User `xorm:"-" json:"created_by"`

	File *files.File `xorm:"-" json:"file"`

	Created time.Time `xorm:"created" json:"created"`

	web.CRUDable    `xorm:"-" json:"-"`
	web.Permissions `xorm:"-" json:"-"`
}

// TableName returns the table name for task attachments
func (*TaskAttachment) TableName() string {
	return "task_attachments"
}

// NewAttachment creates a new task attachment
func (ta *TaskAttachment) NewAttachment(s *xorm.Session, f io.ReadSeeker, realname string, realsize uint64, a web.Auth) error {

	// Store the file using the existing session to avoid nested transactions
	file, err := files.CreateWithSession(s, f, realname, realsize, a)
	if err != nil {
		if files.IsErrFileIsTooLarge(err) {
			return ErrTaskAttachmentIsTooLarge{Size: realsize}
		}
		return err
	}
	ta.File = file

	// Add an entry to the db
	ta.FileID = file.ID

	ta.CreatedBy, err = GetUserOrLinkShareUser(s, a)
	if err != nil {
		// remove the  uploaded file if adding it to the db fails
		if err2 := file.Delete(s); err2 != nil {
			return err2
		}
		return err
	}
	ta.CreatedByID = ta.CreatedBy.ID

	_, err = s.Insert(ta)
	if err != nil {
		// remove the  uploaded file if adding it to the db fails
		if err2 := file.Delete(s); err2 != nil {
			return err2
		}
		return err
	}

	task, err := GetTaskByIDSimple(s, ta.TaskID)
	if err != nil {
		return err
	}

	events.DispatchOnCommit(s, &TaskAttachmentCreatedEvent{
		Task:       &task,
		Attachment: ta,
		Doer:       ta.CreatedBy,
	})
	return nil
}

// ReadOne returns a task attachment
func (ta *TaskAttachment) ReadOne(s *xorm.Session, _ web.Auth) (err error) {
	query := s.Where("id = ?", ta.ID).NoAutoCondition()

	// When TaskID is provided (e.g. from URL parameters), verify the attachment
	// belongs to that task to prevent IDOR attacks.
	if ta.TaskID != 0 {
		query = query.And("task_id = ?", ta.TaskID)
	}

	exists, err := query.Get(ta)
	if err != nil {
		return
	}
	if !exists {
		return ErrTaskAttachmentDoesNotExist{
			TaskID:       ta.TaskID,
			AttachmentID: ta.ID,
		}
	}

	// Get the file
	ta.File = &files.File{ID: ta.FileID}
	err = ta.File.LoadFileMetaByID()
	if err != nil {
		return
	}

	// Swallow missing/disabled/locked so the user-delete cascade can complete.
	ta.CreatedBy, err = user.GetUserByID(s, ta.CreatedByID)
	if err != nil && !user.IsErrUserDoesNotExist(err) && !user.IsErrUserStatusError(err) {
		return err
	}

	return nil
}

// ReadAll returns a project with all attachments
// @Summary Get  all attachments for one task.
// @Description Get all task attachments for one task.
// @tags task
// @Accept json
// @Produce json
// @Param id path int true "Task ID"
// @Param page query int false "The page number. Used for pagination. If not provided, the first page of results is returned."
// @Param per_page query int false "The maximum number of items per page. Note this parameter is limited by the configured maximum of items per page."
// @Security JWTKeyAuth
// @Success 200 {array} models.TaskAttachment "All attachments for this task"
// @Failure 403 {object} models.Message "No access to this task."
// @Failure 404 {object} models.Message "The task does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{id}/attachments [get]
func (ta *TaskAttachment) ReadAll(s *xorm.Session, a web.Auth, _ string, page int, perPage int) (result interface{}, resultCount int, numberOfTotalItems int64, err error) {
	task := Task{ID: ta.TaskID}
	canRead, _, err := task.CanRead(s, a)
	if err != nil {
		return nil, 0, 0, err
	}
	if !canRead {
		return nil, 0, 0, ErrGenericForbidden{}
	}

	attachments := []*TaskAttachment{}

	limit, start := getLimitFromPageIndex(page, perPage)

	query := s.
		Where("task_id = ?", ta.TaskID)
	if limit > 0 {
		query = query.Limit(limit, start)
	}
	err = query.Find(&attachments)
	if err != nil {
		return nil, 0, 0, err
	}

	if len(attachments) == 0 {
		return
	}

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

	users, err := getUsersOrLinkSharesFromIDs(s, userIDs)
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

	numberOfTotalItems, err = s.
		Where("task_id = ?", ta.TaskID).
		Count(&TaskAttachment{})
	return attachments, len(attachments), numberOfTotalItems, err
}

func cacheKeyForTaskAttachmentPreview(id int64, size PreviewSize) string {
	return "task_attachment_preview_" + strconv.FormatInt(id, 10) + "_size_" + string(size)
}

func (ta *TaskAttachment) GetPreview(previewSize PreviewSize) []byte {
	cacheKey := cacheKeyForTaskAttachmentPreview(ta.ID, previewSize)

	result, err := keyvalue.RememberValue(cacheKey, func() ([]byte, error) {
		// Read all bytes up front so we can inspect dimensions without seeking.
		// The file is an io.ReadCloser (no Seek), so we buffer it once.
		data, err := io.ReadAll(ta.File.File)
		if err != nil {
			return nil, err
		}

		// Check image dimensions before full decode to prevent DoS
		// from decompression bombs (small file, huge pixel dimensions)
		const maxPixels = 50_000_000 // 50 megapixels
		cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		if cfg.Width*cfg.Height > maxPixels {
			return nil, fmt.Errorf("image dimensions %dx%d exceed maximum of %d pixels", cfg.Width, cfg.Height, maxPixels)
		}

		img, _, err := image.Decode(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}

		// Scale down the image to a minimum size
		resizedImg := resizeImage(img, previewSize.GetSize())

		// Get the raw bytes of the resized image
		buf := &bytes.Buffer{}
		if err := png.Encode(buf, resizedImg); err != nil {
			return nil, err
		}
		previewImage, err := io.ReadAll(buf)
		if err != nil {
			return nil, err
		}

		log.Infof("Attachment image preview for task attachment %v of size %v created and cached", ta.ID, previewSize)
		return previewImage, nil
	})
	if err != nil {
		return nil
	}

	return result
}

type PreviewSize string

const (
	PreviewSizeUnknown PreviewSize = "unknown"
	PreviewSmall       PreviewSize = "sm"
	PreviewMedium      PreviewSize = "md"
	PreviewLarge       PreviewSize = "lg"
	PreviewExtraLarge  PreviewSize = "xl"
)

func (previewSize PreviewSize) GetSize() int {
	switch previewSize {
	case PreviewSmall:
		return 100
	case PreviewMedium:
		return 200
	case PreviewLarge:
		return 400
	case PreviewExtraLarge:
		return 800
	case PreviewSizeUnknown:
		return 0
	default:
		return 200
	}
}

func GetPreviewSizeFromString(size string) PreviewSize {
	switch size {
	case "sm":
		return PreviewSmall
	case "md":
		return PreviewMedium
	case "lg":
		return PreviewLarge
	case "xl":
		return PreviewExtraLarge
	}

	return PreviewSizeUnknown
}

func resizeImage(img image.Image, width int) *image.NRGBA {
	resizedImg := imaging.Resize(img, width, 0, imaging.Lanczos)
	log.Debugf(
		"Resized attachment image from %vx%v to %vx%v for a preview",
		img.Bounds().Size().X,
		img.Bounds().Size().Y,
		resizedImg.Bounds().Size().X,
		resizedImg.Bounds().Size().Y,
	)

	return resizedImg
}

// Delete removes an attachment
// @Summary Delete an attachment
// @Description Delete an attachment.
// @tags task
// @Accept json
// @Produce json
// @Param id path int true "Task ID"
// @Param attachmentID path int true "Attachment ID"
// @Security JWTKeyAuth
// @Success 200 {object} models.Message "The attachment was deleted successfully."
// @Failure 403 {object} models.Message "No access to this task."
// @Failure 404 {object} models.Message "The task does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{id}/attachments/{attachmentID} [delete]
func (ta *TaskAttachment) Delete(s *xorm.Session, a web.Auth) error {
	// Load the attachment
	err := ta.ReadOne(s, a)
	if err != nil && !files.IsErrFileDoesNotExist(err) {
		return err
	}

	// Delete it
	_, err = s.
		Where("task_id = ? AND id = ?", ta.TaskID, ta.ID).
		Delete(ta)
	if err != nil {
		return err
	}

	// Delete the underlying file
	err = ta.File.Delete(s)
	// If the file does not exist, we don't want to error out
	if err != nil && files.IsErrFileDoesNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	doer, _ := user.GetFromAuth(a)
	task, err := GetTaskByIDSimple(s, ta.TaskID)
	if err != nil {
		return err
	}

	events.DispatchOnCommit(s, &TaskAttachmentDeletedEvent{
		Task:       &task,
		Attachment: ta,
		Doer:       doer,
	})
	return nil
}

func getTaskAttachmentsByTaskIDs(s *xorm.Session, taskIDs []int64) (attachments []*TaskAttachment, err error) {
	attachments = []*TaskAttachment{}
	err = s.
		In("task_id", taskIDs).
		Find(&attachments)
	if err != nil {
		return
	}

	if len(attachments) == 0 {
		return
	}

	fileIDs := []int64{}
	userIDs := []int64{}
	for _, a := range attachments {
		userIDs = append(userIDs, a.CreatedByID)
		fileIDs = append(fileIDs, a.FileID)
	}

	// Get all files
	fs := make(map[int64]*files.File)
	err = s.In("id", fileIDs).Find(&fs)
	if err != nil {
		return
	}

	users, err := getUsersOrLinkSharesFromIDs(s, userIDs)
	if err != nil {
		return nil, err
	}

	// Obfuscate all user emails
	for _, u := range users {
		u.Email = ""
	}

	for _, a := range attachments {
		if createdBy, has := users[a.CreatedByID]; has {
			a.CreatedBy = createdBy
		}
		a.File = fs[a.FileID]
	}

	return
}
