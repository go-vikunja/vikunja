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
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"io/fs"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/modules/imageutils"
	"code.vikunja.io/api/pkg/modules/keyvalue"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"github.com/disintegration/imaging"
	"github.com/gabriel-vasile/mimetype"
	"golang.org/x/sync/semaphore"
	"golang.org/x/sync/singleflight"
	"xorm.io/xorm"
)

// TaskAttachment is the definition of a task attachment
type TaskAttachment struct {
	ID     int64 `xorm:"bigint autoincr not null unique pk" json:"id" param:"attachment" readOnly:"true" doc:"The unique, numeric id of this attachment."`
	TaskID int64 `xorm:"bigint not null" json:"task_id" param:"task" readOnly:"true" doc:"The id of the task this attachment belongs to. Taken from the URL, not the body."`
	FileID int64 `xorm:"bigint not null" json:"-"`

	CreatedByID int64      `xorm:"bigint not null" json:"-"`
	CreatedBy   *user.User `xorm:"-" json:"created_by" readOnly:"true" doc:"The user who uploaded this attachment."`

	File *files.File `xorm:"-" json:"file" readOnly:"true" doc:"Metadata of the uploaded file (name, mime type, size). The bytes are fetched from the download endpoint, not this field."`

	Created time.Time `xorm:"created" json:"created" readOnly:"true" doc:"A timestamp when this attachment was uploaded. You cannot change this value."`

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

// AttachmentToUpload is a transport-neutral file to attach, so the upload logic
// can be shared by the multipart v1 handler and the Huma v2 handler.
type AttachmentToUpload struct {
	Reader   io.ReadSeeker
	Filename string
	Size     uint64
}

// UploadTaskAttachments checks create access to the task, then stores each file,
// collecting per-file failures rather than aborting. The caller owns the session
// and the commit. A returned err means the request as a whole failed (e.g.
// forbidden); per-file failures come back in failures instead.
func UploadTaskAttachments(s *xorm.Session, a web.Auth, taskID int64, uploads []*AttachmentToUpload) (success []*TaskAttachment, failures []error, err error) {
	ta := &TaskAttachment{TaskID: taskID}
	can, err := ta.CanCreate(s, a)
	if err != nil {
		return nil, nil, err
	}
	if !can {
		return nil, nil, ErrGenericForbidden{}
	}

	for _, upload := range uploads {
		attachment := &TaskAttachment{TaskID: taskID}
		if err := attachment.NewAttachment(s, upload.Reader, upload.Filename, upload.Size, a); err != nil {
			failures = append(failures, err)
			continue
		}
		success = append(success, attachment)
	}
	return success, failures, nil
}

// TaskAttachmentDownload is either a *TaskAttachmentPreview or a *TaskAttachmentOriginal. The caller must close it.
//
//sumtype:decl
type TaskAttachmentDownload interface {
	io.Closer
	taskAttachmentDownload()
}

// TaskAttachmentPreview is an encoded preview image.
type TaskAttachmentPreview struct {
	io.ReadCloser
	Mime string
	Size int64 // -1 when the storage did not report it
}

func (*TaskAttachmentPreview) taskAttachmentDownload() {}

// TaskAttachmentOriginal is the uploaded file, opened for reading.
type TaskAttachmentOriginal struct {
	File *files.File
}

func (original *TaskAttachmentOriginal) Close() error {
	return original.File.File.Close()
}

func (*TaskAttachmentOriginal) taskAttachmentDownload() {}

// GetTaskAttachmentForDownload returns a preview instead of the file when previewSize
// is set and the attachment is an image. Returns ErrGenericForbidden on denied access.
// Owns its own session and commits it before touching storage so no pool connection is held.
func GetTaskAttachmentForDownload(ctx context.Context, a web.Auth, taskID, attachmentID int64, previewSize PreviewSize) (TaskAttachmentDownload, error) {
	s := db.NewSession()
	defer s.Close()

	ta := &TaskAttachment{ID: attachmentID, TaskID: taskID}
	can, _, err := ta.CanRead(s, a)
	if err != nil {
		_ = s.Rollback()
		return nil, err
	}
	if !can {
		_ = s.Rollback()
		return nil, ErrGenericForbidden{}
	}

	if err := ta.ReadOne(s, a); err != nil {
		_ = s.Rollback()
		return nil, err
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return nil, err
	}

	return ta.openPreviewOrFile(ctx, previewSize)
}

func (ta *TaskAttachment) openPreviewOrFile(ctx context.Context, previewSize PreviewSize) (TaskAttachmentDownload, error) {
	if previewSize != PreviewSizeUnknown && strings.HasPrefix(ta.File.Mime, "image") {
		preview, err := ta.GetPreview(ctx, previewSize)
		if err == nil {
			return preview, nil
		}
		if !IsErrTaskAttachmentNotPreviewable(err) {
			return nil, err
		}
	}

	if err := ta.File.LoadFileByID(); err != nil {
		return nil, err
	}
	return &TaskAttachmentOriginal{File: ta.File}, nil
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
	err = ta.File.LoadFileMetaByID(s)
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
		return attachments, 0, 0, nil
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

// RemoveLegacyTaskAttachmentPreviews drops the previews older versions put into keyvalue without a TTL,
// which Redis would keep forever. Not a migration: instances still running the old version keep writing them.
func RemoveLegacyTaskAttachmentPreviews() {
	if err := keyvalue.DelPrefix("task_attachment_preview_"); err != nil {
		log.Errorf("Could not remove the attachment previews left in keyvalue: %v", err)
	}
}

// previewStore keeps generated previews, in the file storage outside of tests.
type previewStore interface {
	// open returns an error wrapping fs.ErrNotExist when the preview is not stored.
	open(fileID int64, size PreviewSize) (preview io.ReadCloser, length int64, err error)
	save(fileID int64, size PreviewSize, preview []byte) error
}

type filePreviewStore struct{}

func (filePreviewStore) open(fileID int64, size PreviewSize) (io.ReadCloser, int64, error) {
	return (&files.File{ID: fileID}).OpenPreview(size)
}

func (filePreviewStore) save(fileID int64, size PreviewSize, preview []byte) error {
	return (&files.File{ID: fileID}).SavePreview(size, preview)
}

const (
	// Below imageutils.MaxPixels the largest allowed image could never acquire.
	previewPixelBudget    = 2 * imageutils.MaxPixels
	previewQueueTimeout   = 30 * time.Second
	maxPreviewSourceBytes = 50 << 20
	maxPreviewHeaderBytes = 1 << 20
)

var attachmentPreviews = newPreviewGenerator(previewPixelBudget, previewQueueTimeout)

type previewGenerator struct {
	slots        *semaphore.Weighted
	queueTimeout time.Duration
	flights      singleflight.Group
	store        previewStore
}

func newPreviewGenerator(pixelBudget int64, queueTimeout time.Duration) *previewGenerator {
	return &previewGenerator{
		slots:        semaphore.NewWeighted(pixelBudget),
		queueTimeout: queueTimeout,
		store:        filePreviewStore{},
	}
}

// GetPreview returns ErrTaskAttachmentNotPreviewable when the original should be served instead.
func (ta *TaskAttachment) GetPreview(ctx context.Context, previewSize PreviewSize) (*TaskAttachmentPreview, error) {
	if ta.File.Size > maxPreviewSourceBytes {
		return nil, ErrTaskAttachmentNotPreviewable{Err: fmt.Errorf("file is %d bytes, previews are limited to %d", ta.File.Size, maxPreviewSourceBytes)}
	}
	return attachmentPreviews.get(ctx, ta.File.ID, fileOpener(ta.File.ID), previewSize)
}

func newTaskAttachmentPreview(preview []byte) *TaskAttachmentPreview {
	return &TaskAttachmentPreview{
		ReadCloser: io.NopCloser(bytes.NewReader(preview)),
		Mime:       mimetype.Detect(preview).String(),
		Size:       int64(len(preview)),
	}
}

func (g *previewGenerator) get(ctx context.Context, fileID int64, open sourceOpener, previewSize PreviewSize) (*TaskAttachmentPreview, error) {
	preview, err := g.openStored(fileID, previewSize)
	if err == nil {
		return preview, nil
	}
	if !errors.Is(err, fs.ErrNotExist) {
		log.Errorf("Could not use the stored %s preview of file %d: %v", previewSize, fileID, err)
	}

	flightKey := strconv.FormatInt(fileID, 10) + "_" + string(previewSize)
	flight := g.flights.DoChan(flightKey, func() (any, error) {
		// Detached: the flight is shared, so the caller who started it leaving must not fail the others.
		return g.generateAndStore(context.WithoutCancel(ctx), fileID, open, previewSize)
	})

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("waiting for attachment preview: %w", context.Cause(ctx))
	case result := <-flight:
		if result.Err != nil {
			return nil, result.Err
		}
		// The flight is shared, so it hands out bytes and every caller gets its own reader.
		preview, _ := result.Val.([]byte)
		return newTaskAttachmentPreview(preview), nil
	}
}

// Enough for every image format's magic number.
const previewHeaderLength = 512

// Any error means the preview is generated instead of failing the request.
func (g *previewGenerator) openStored(fileID int64, previewSize PreviewSize) (*TaskAttachmentPreview, error) {
	content, length, err := g.store.open(fileID, previewSize)
	if err != nil {
		return nil, err
	}

	buffered := bufio.NewReaderSize(content, previewHeaderLength)
	header, err := buffered.Peek(previewHeaderLength)
	if err != nil && !errors.Is(err, io.EOF) {
		_ = content.Close()
		return nil, fmt.Errorf("reading its header: %w", err)
	}

	mime := mimetype.Detect(header).String()
	if mime != "image/jpeg" && mime != "image/png" {
		_ = content.Close()
		return nil, fmt.Errorf("not a jpeg or png but %s", mime)
	}

	return &TaskAttachmentPreview{
		ReadCloser: struct {
			io.Reader
			io.Closer
		}{
			Reader: buffered,
			Closer: content,
		},
		Mime: mime,
		Size: length,
	}, nil
}

// The caller's own lookup already logged why the stored preview is unusable.
func (g *previewGenerator) readStored(fileID int64, previewSize PreviewSize) (preview []byte, stored bool) {
	content, err := g.openStored(fileID, previewSize)
	if err != nil {
		return nil, false
	}
	defer content.Close()

	preview, err = io.ReadAll(content)
	if err != nil {
		log.Errorf("Could not read the stored %s preview of file %d: %v", previewSize, fileID, err)
		return nil, false
	}
	return preview, true
}

func (g *previewGenerator) generateAndStore(ctx context.Context, fileID int64, open sourceOpener, previewSize PreviewSize) (preview []byte, err error) {
	// singleflight re-panics a DoChan callback on its own goroutine, which would crash the whole API.
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("Panic while generating the %s preview of file %d: %v\n%s", previewSize, fileID, r, debug.Stack())
			err = fmt.Errorf("panic while generating the %s preview of file %d: %v", previewSize, fileID, r)
		}
	}()

	// A flight that just finished may have stored it after this caller's lookup.
	if preview, stored := g.readStored(fileID, previewSize); stored {
		return preview, nil
	}

	preview, err = g.generate(ctx, open, previewSize)
	if err != nil {
		return nil, err
	}

	if err := g.store.save(fileID, previewSize, preview); err != nil {
		log.Errorf("Could not store the %s preview of file %d: %v", previewSize, fileID, err)
		return preview, nil
	}
	log.Infof("Created and stored the %s preview of file %d", previewSize, fileID)
	return preview, nil
}

type sourceOpener func() (io.ReadCloser, error)

func fileOpener(fileID int64) sourceOpener {
	return func() (io.ReadCloser, error) {
		file := &files.File{ID: fileID}
		if err := file.LoadFileByID(); err != nil {
			return nil, err
		}
		return file.File, nil
	}
}

// Opens the source twice so callers queued for capacity hold no storage handle or S3 body while waiting.
func (g *previewGenerator) generate(ctx context.Context, open sourceOpener, previewSize PreviewSize) ([]byte, error) {
	pixels, err := sourcePixels(open)
	if err != nil {
		return nil, err
	}

	if err := g.acquire(ctx, pixels); err != nil {
		return nil, err
	}
	defer g.slots.Release(pixels)

	src, err := open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	preview, err := renderPreview(src, previewSize)
	if err != nil {
		return nil, ErrTaskAttachmentNotPreviewable{Err: err}
	}
	return preview, nil
}

func sourcePixels(open sourceOpener) (int64, error) {
	header, err := open()
	if err != nil {
		return 0, err
	}
	defer header.Close()

	// Rejects decompression bombs (small file, huge pixel dimensions) before the full decode.
	// DecodeConfig buffers the whole file for formats whose metadata may sit at the end, like TIFF.
	cfg, err := imageutils.ValidateReader(io.LimitReader(header, maxPreviewHeaderBytes))
	if err != nil {
		return 0, ErrTaskAttachmentNotPreviewable{Err: err}
	}
	return int64(cfg.Width) * int64(cfg.Height), nil
}

func (g *previewGenerator) acquire(ctx context.Context, pixels int64) error {
	queueCtx, cancel := context.WithTimeout(ctx, g.queueTimeout)
	defer cancel()

	if err := g.slots.Acquire(queueCtx, pixels); err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("waiting to generate attachment preview: %w", context.Cause(ctx))
		}
		return ErrTaskAttachmentPreviewBusy{}
	}
	return nil
}

func renderPreview(src io.Reader, previewSize PreviewSize) ([]byte, error) {
	img, _, err := image.Decode(src)
	if err != nil {
		return nil, err
	}

	return encodePreview(resizeImage(img, previewSize.GetSize()))
}

const previewJPEGQuality = 85

// encodePreview alone decides the format of a preview; its mime type is read back from the bytes.
func encodePreview(img *image.NRGBA) ([]byte, error) {
	buf := &bytes.Buffer{}

	// JPEG is much smaller for photos but has no alpha channel.
	var err error
	if img.Opaque() {
		err = jpeg.Encode(buf, img, &jpeg.Options{Quality: previewJPEGQuality})
	} else {
		err = png.Encode(buf, img)
	}
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

type PreviewSize = files.PreviewSize

const (
	PreviewSizeUnknown = files.PreviewSizeUnknown
	PreviewSmall       = files.PreviewSmall
	PreviewMedium      = files.PreviewMedium
	PreviewLarge       = files.PreviewLarge
	PreviewExtraLarge  = files.PreviewExtraLarge
)

func GetPreviewSizeFromString(size string) PreviewSize {
	return files.GetPreviewSizeFromString(size)
}

func resizeImage(img image.Image, width int) *image.NRGBA {
	resizedImg := imaging.Resize(img, min(width, img.Bounds().Dx()), 0, imaging.Lanczos)
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
