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
	"io"
	"os"
	"path"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	"github.com/spf13/afero"
)

// s3Fs is a minimal afero.Fs implementation backed by S3.
// It only supports Open (read), Remove, and Stat — the operations
// Vikunja actually uses. All other methods return ErrS3NotSupported.
type s3Fs struct {
	client *s3.Client
	bucket string
}

var ErrS3NotSupported = errors.New("operation not supported on S3 filesystem")

func newS3Fs(bucket string, client *s3.Client) afero.Fs {
	return &s3Fs{bucket: bucket, client: client}
}

func (*s3Fs) Name() string { return "s3" }

// Open opens a file for reading via S3 GetObject.
// The actual GetObject call is deferred until the first Read.
func (f *s3Fs) Open(name string) (afero.File, error) {
	// Verify the object exists
	_, err := f.client.HeadObject(context.Background(), &s3.HeadObjectInput{
		Bucket: aws.String(f.bucket),
		Key:    aws.String(name),
	})
	if err != nil {
		return nil, s3ToPathError("open", name, err)
	}

	return &s3File{
		fs:   f,
		name: name,
	}, nil
}

// Stat returns file info via S3 HeadObject.
func (f *s3Fs) Stat(name string) (os.FileInfo, error) {
	head, err := f.client.HeadObject(context.Background(), &s3.HeadObjectInput{
		Bucket: aws.String(f.bucket),
		Key:    aws.String(name),
	})
	if err != nil {
		return nil, s3ToPathError("stat", name, err)
	}

	var size int64
	if head.ContentLength != nil {
		size = *head.ContentLength
	}
	var modTime time.Time
	if head.LastModified != nil {
		modTime = *head.LastModified
	}

	return &s3FileInfo{
		name:    path.Base(name),
		size:    size,
		modTime: modTime,
	}, nil
}

// Remove deletes a file via S3 DeleteObject.
func (f *s3Fs) Remove(name string) error {
	// Check existence first so we return a proper error for missing files
	if _, err := f.Stat(name); err != nil {
		return err
	}

	_, err := f.client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(f.bucket),
		Key:    aws.String(name),
	})
	return err
}

// Unsupported operations

func (*s3Fs) Create(string) (afero.File, error)                     { return nil, ErrS3NotSupported }
func (*s3Fs) Mkdir(string, os.FileMode) error                       { return ErrS3NotSupported }
func (*s3Fs) MkdirAll(string, os.FileMode) error                    { return ErrS3NotSupported }
func (*s3Fs) OpenFile(string, int, os.FileMode) (afero.File, error) { return nil, ErrS3NotSupported }
func (*s3Fs) RemoveAll(string) error                                { return ErrS3NotSupported }
func (*s3Fs) Rename(string, string) error                           { return ErrS3NotSupported }
func (*s3Fs) Chmod(string, os.FileMode) error                       { return ErrS3NotSupported }
func (*s3Fs) Chown(string, int, int) error                          { return ErrS3NotSupported }
func (*s3Fs) Chtimes(string, time.Time, time.Time) error            { return ErrS3NotSupported }

// s3ToPathError converts S3 SDK errors into os-compatible path errors.
func s3ToPathError(op, name string, err error) error {
	var respErr *smithyhttp.ResponseError
	if errors.As(err, &respErr) && respErr.HTTPStatusCode() == 404 {
		return &os.PathError{Op: op, Path: name, Err: os.ErrNotExist}
	}
	return &os.PathError{Op: op, Path: name, Err: err}
}

// s3File is a minimal afero.File for reading from S3.
// It lazily opens a GetObject stream on the first Read call.
type s3File struct {
	fs     *s3Fs
	name   string
	body   io.ReadCloser
	closed bool
}

func (f *s3File) Name() string { return f.name }

func (f *s3File) Read(p []byte) (int, error) {
	if f.closed {
		return 0, afero.ErrFileClosed
	}

	// Lazily open stream on first Read
	if f.body == nil {
		if err := f.openStream(); err != nil {
			return 0, err
		}
	}

	return f.body.Read(p)
}

func (f *s3File) Close() error {
	f.closed = true
	f.closeStream()
	return nil
}

func (f *s3File) Stat() (os.FileInfo, error) {
	return f.fs.Stat(f.name)
}

func (f *s3File) openStream() error {
	out, err := f.fs.client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(f.fs.bucket),
		Key:    aws.String(f.name),
	})
	if err != nil {
		return err
	}
	f.body = out.Body
	return nil
}

func (f *s3File) closeStream() {
	if f.body != nil {
		f.body.Close()
		f.body = nil
	}
}

// Unsupported file operations

func (*s3File) ReadAt([]byte, int64) (int, error)  { return 0, ErrS3NotSupported }
func (*s3File) Seek(int64, int) (int64, error)     { return 0, ErrS3NotSupported }
func (*s3File) Write([]byte) (int, error)          { return 0, ErrS3NotSupported }
func (*s3File) WriteAt([]byte, int64) (int, error) { return 0, ErrS3NotSupported }
func (*s3File) WriteString(string) (int, error)    { return 0, ErrS3NotSupported }
func (*s3File) Truncate(int64) error               { return ErrS3NotSupported }
func (*s3File) Sync() error                        { return nil }
func (*s3File) Readdir(int) ([]os.FileInfo, error) { return nil, ErrS3NotSupported }
func (*s3File) Readdirnames(int) ([]string, error) { return nil, ErrS3NotSupported }

// s3FileInfo implements os.FileInfo for S3 objects.
type s3FileInfo struct {
	name    string
	size    int64
	modTime time.Time
}

func (fi *s3FileInfo) Name() string       { return fi.name }
func (fi *s3FileInfo) Size() int64        { return fi.size }
func (fi *s3FileInfo) Mode() os.FileMode  { return 0664 }
func (fi *s3FileInfo) ModTime() time.Time { return fi.modTime }
func (fi *s3FileInfo) IsDir() bool        { return false }
func (fi *s3FileInfo) Sys() interface{}   { return nil }
