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
	"os"
	"path"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	smithyhttp "github.com/aws/smithy-go/transport/http"

	"code.vikunja.io/api/pkg/log"
)

// s3Storage implements FileStorage backed by S3.
// All paths are prefixed with basePath to form S3 object keys.
type s3Storage struct {
	client   *s3.Client
	bucket   string
	basePath string
}

func newS3Storage(bucket, basePath string, client *s3.Client) *s3Storage {
	return &s3Storage{bucket: bucket, basePath: basePath, client: client}
}

func (s *s3Storage) key(name string) string {
	return path.Join(s.basePath, name)
}

func (s *s3Storage) Open(name string) (io.ReadCloser, error) {
	key := s.key(name)
	out, err := s.client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, s3ToPathError("open", name, err)
	}
	return out.Body, nil
}

func (s *s3Storage) Write(name string, content io.ReadSeeker, size uint64) error {
	contentLength, err := contentLengthFromReadSeeker(content, size)
	if err != nil {
		return fmt.Errorf("failed to determine S3 upload content length: %w", err)
	}

	if _, err = content.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("failed to seek to start before S3 upload: %w", err)
	}

	_, err = s.client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(s.key(name)),
		Body:          content,
		ContentLength: aws.Int64(contentLength),
	})
	if err != nil {
		return fmt.Errorf("failed to upload file to S3: %w", err)
	}
	return nil
}

func (s *s3Storage) Stat(name string) (os.FileInfo, error) {
	key := s.key(name)
	head, err := s.client.HeadObject(context.Background(), &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
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

func (s *s3Storage) Remove(name string) error {
	// Check existence first for proper error on missing files
	if _, err := s.Stat(name); err != nil {
		return err
	}

	_, err := s.client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.key(name)),
	})
	return err
}

func (*s3Storage) MkdirAll(string, os.FileMode) error {
	return nil // S3 has no directories
}

// s3ToPathError converts S3 SDK errors into os-compatible path errors.
func s3ToPathError(op, name string, err error) error {
	var respErr *smithyhttp.ResponseError
	if errors.As(err, &respErr) && respErr.HTTPStatusCode() == 404 {
		return &os.PathError{Op: op, Path: name, Err: os.ErrNotExist}
	}
	return &os.PathError{Op: op, Path: name, Err: err}
}

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

// contentLengthFromReadSeeker determines the content length by seeking to the end.
func contentLengthFromReadSeeker(seeker io.ReadSeeker, expectedSize uint64) (int64, error) {
	endOffset, err := seeker.Seek(0, io.SeekEnd)
	if err != nil {
		return 0, err
	}

	if expectedSize > 0 && expectedSize <= uint64(maxInt64) && endOffset != int64(expectedSize) {
		log.Warningf("File size mismatch for S3 upload: expected %d bytes but reader reports %d bytes", expectedSize, endOffset)
	}

	return endOffset, nil
}

const maxInt64 = 1<<63 - 1
