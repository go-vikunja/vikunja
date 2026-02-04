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

package files

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/metrics"
	"code.vikunja.io/api/pkg/modules/keyvalue"

	"code.vikunja.io/api/pkg/web"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/c2h5oh/datasize"
	"github.com/spf13/afero"
	"xorm.io/xorm"
)

// File holds all information about a file
type File struct {
	ID   int64  `xorm:"bigint autoincr not null unique pk" json:"id"`
	Name string `xorm:"text not null" json:"name"`
	Mime string `xorm:"text null" json:"mime"`
	Size uint64 `xorm:"bigint not null" json:"size"`

	Created     time.Time `xorm:"created" json:"created"`
	CreatedByID int64     `xorm:"bigint not null" json:"-"`

	File afero.File `xorm:"-" json:"-"`
	// This ReadCloser is only used for migration purposes. Use with care!
	// There is currentlc no better way of doing this.
	FileContent []byte `xorm:"-" json:"-"`
}

// TableName is the table name for the files table
func (*File) TableName() string {
	return "files"
}

func (f *File) getAbsoluteFilePath() string {
	return filepath.Join(
		config.FilesBasePath.GetString(),
		strconv.FormatInt(f.ID, 10),
	)
}

// LoadFileByID returns a file by its ID
func (f *File) LoadFileByID() (err error) {
	f.File, err = afs.Open(f.getAbsoluteFilePath())
	return
}

// LoadFileMetaByID loads everything about a file without loading the actual file
func (f *File) LoadFileMetaByID() (err error) {
	exists, err := x.Where("id = ?", f.ID).Get(f)
	if !exists {
		return ErrFileDoesNotExist{FileID: f.ID}
	}
	return
}

// Create creates a new file from an FileHeader
func Create(f io.ReadSeeker, realname string, realsize uint64, a web.Auth) (file *File, err error) {
	return CreateWithMime(f, realname, realsize, a, "")
}

// CreateWithMime creates a new file from an FileHeader and sets its mime type
func CreateWithMime(f io.ReadSeeker, realname string, realsize uint64, a web.Auth, mime string) (file *File, err error) {
	s := db.NewSession()
	defer s.Close()

	file, err = CreateWithMimeAndSession(s, f, realname, realsize, a, mime, true)
	if err != nil {
		_ = s.Rollback()
		return
	}
	return
}

func CreateWithMimeAndSession(s *xorm.Session, f io.ReadSeeker, realname string, realsize uint64, a web.Auth, mime string, checkFileSizeLimit bool) (file *File, err error) {
	if realsize > config.GetMaxFileSizeInMBytes()*uint64(datasize.MB) && checkFileSizeLimit {
		return nil, ErrFileIsTooLarge{Size: realsize}
	}

	// We first insert the file into the db to get it's ID
	file = &File{
		Name:        realname,
		Size:        realsize,
		CreatedByID: a.GetID(),
		Mime:        mime,
	}

	_, err = s.Insert(file)
	if err != nil {
		return
	}

	// Save the file to storage with its new ID as path
	err = file.Save(f)
	return
}

// Delete removes a file from the DB and the file system
func (f *File) Delete(s *xorm.Session) (err error) {
	deleted, err := s.Where("id = ?", f.ID).Delete(&File{})
	if err != nil {
		_ = s.Rollback()
		return err
	}
	if deleted == 0 {
		_ = s.Rollback()
		return ErrFileDoesNotExist{FileID: f.ID}
	}

	err = afs.Remove(f.getAbsoluteFilePath())
	if err != nil {
		var perr *os.PathError
		if errors.As(err, &perr) {
			// Don't fail when removing the file failed
			log.Errorf("Error deleting file %d: %s", f.ID, err)
			return s.Commit()
		}

		_ = s.Rollback()
		return err
	}

	return keyvalue.DecrBy(metrics.FilesCountKey, 1)
}

// writeToStorage writes content to the given path, handling both local and S3 backends.
// The reader is always seeked to position 0 before writing to ensure consistent behavior.
func writeToStorage(path string, content io.ReadSeeker, size uint64) error {
	// Seek to start to ensure we write the complete content regardless of
	// the reader's current position
	if _, err := content.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("failed to seek to start of content: %w", err)
	}

	if s3Client == nil {
		return afs.WriteReader(path, content)
	}

	contentLength, err := contentLengthFromReadSeeker(content, size)
	if err != nil {
		return fmt.Errorf("failed to determine S3 upload content length: %w", err)
	}

	_, err = content.Seek(0, io.SeekStart)
	if err != nil {
		return fmt.Errorf("failed to seek S3 upload body to start: %w", err)
	}

	_, err = s3Client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket:        aws.String(s3Bucket),
		Key:           aws.String(path),
		Body:          content,
		ContentLength: aws.Int64(contentLength),
	})
	if err != nil {
		return fmt.Errorf("failed to upload file to S3: %w", err)
	}
	return nil
}

// Save saves a file to storage
func (f *File) Save(fcontent io.ReadSeeker) error {
	err := writeToStorage(f.getAbsoluteFilePath(), fcontent, f.Size)
	if err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}
	return keyvalue.IncrBy(metrics.FilesCountKey, 1)
}

func contentLengthFromReadSeeker(seeker io.ReadSeeker, expectedSize uint64) (int64, error) {
	currentOffset, err := seeker.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0, err
	}

	endOffset, err := seeker.Seek(0, io.SeekEnd)
	if err != nil {
		return 0, err
	}

	_, err = seeker.Seek(currentOffset, io.SeekStart)
	if err != nil {
		return 0, err
	}

	if expectedSize > 0 && expectedSize <= uint64(math.MaxInt64) && endOffset != int64(expectedSize) {
		log.Warningf("File size mismatch for S3 upload: expected %d bytes but reader reports %d bytes", expectedSize, endOffset)
	}

	return endOffset, nil
}
