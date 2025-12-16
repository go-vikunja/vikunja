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
	"bytes"
	"errors"
	"io"
	"os"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"github.com/aws/aws-sdk-go/service/s3" //nolint:staticcheck // afero-s3 still requires aws-sdk-go v1
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeS3PutObjectClient struct {
	lastInput *s3.PutObjectInput
	err       error
}

func (f *fakeS3PutObjectClient) PutObject(input *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
	f.lastInput = input
	if f.err != nil {
		return nil, f.err
	}
	return &s3.PutObjectOutput{}, nil
}

type readerOnly struct {
	r io.Reader
}

func (r *readerOnly) Read(p []byte) (int, error) {
	return r.r.Read(p)
}

func TestFileSave_S3_UsesSeekableReaderWithoutTempFile(t *testing.T) {
	originalClient := s3Client
	originalBucket := s3Bucket
	originalTempDir := config.FilesS3TempDir.GetString()
	t.Cleanup(func() {
		s3Client = originalClient
		s3Bucket = originalBucket
		config.FilesS3TempDir.Set(originalTempDir)
	})

	tempDir := t.TempDir()
	config.FilesS3TempDir.Set(tempDir)

	client := &fakeS3PutObjectClient{}
	s3Client = client
	s3Bucket = "test-bucket"

	content := []byte("seekable-content")
	file := &File{ID: 123, Size: uint64(len(content))}

	err := file.Save(bytes.NewReader(content))
	require.NoError(t, err)

	require.NotNil(t, client.lastInput)
	assert.Equal(t, "test-bucket", *client.lastInput.Bucket)
	assert.Equal(t, file.getAbsoluteFilePath(), *client.lastInput.Key)
	require.NotNil(t, client.lastInput.ContentLength)
	assert.Equal(t, int64(len(content)), *client.lastInput.ContentLength)
	assert.IsType(t, &bytes.Reader{}, client.lastInput.Body)

	entries, err := os.ReadDir(tempDir)
	require.NoError(t, err)
	assert.Len(t, entries, 0)
}

func TestFileSave_S3_BuffersNonSeekableReaderAndCleansUpTempFile(t *testing.T) {
	originalClient := s3Client
	originalBucket := s3Bucket
	originalTempDir := config.FilesS3TempDir.GetString()
	t.Cleanup(func() {
		s3Client = originalClient
		s3Bucket = originalBucket
		config.FilesS3TempDir.Set(originalTempDir)
	})

	tempDir := t.TempDir()
	config.FilesS3TempDir.Set(tempDir)

	client := &fakeS3PutObjectClient{}
	s3Client = client
	s3Bucket = "test-bucket"

	content := []byte("non-seekable-content")
	file := &File{ID: 456, Size: 0}

	err := file.Save(&readerOnly{r: bytes.NewReader(content)})
	require.NoError(t, err)

	require.NotNil(t, client.lastInput)
	require.NotNil(t, client.lastInput.ContentLength)
	assert.Equal(t, int64(len(content)), *client.lastInput.ContentLength)
	assert.IsType(t, &os.File{}, client.lastInput.Body)

	entries, err := os.ReadDir(tempDir)
	require.NoError(t, err)
	assert.Len(t, entries, 0)
}

func TestFileSave_S3_CleansUpTempFileOnPutObjectError(t *testing.T) {
	originalClient := s3Client
	originalBucket := s3Bucket
	originalTempDir := config.FilesS3TempDir.GetString()
	t.Cleanup(func() {
		s3Client = originalClient
		s3Bucket = originalBucket
		config.FilesS3TempDir.Set(originalTempDir)
	})

	tempDir := t.TempDir()
	config.FilesS3TempDir.Set(tempDir)

	client := &fakeS3PutObjectClient{err: errors.New("boom")}
	s3Client = client
	s3Bucket = "test-bucket"

	content := []byte("non-seekable-content")
	file := &File{ID: 789, Size: 0}

	err := file.Save(&readerOnly{r: bytes.NewReader(content)})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to upload file to S3")

	entries, readErr := os.ReadDir(tempDir)
	require.NoError(t, readErr)
	assert.Len(t, entries, 0)
}

func TestFileSave_S3_UsesBufferedSizeWhenExpectedSizeMismatch(t *testing.T) {
	originalClient := s3Client
	originalBucket := s3Bucket
	originalTempDir := config.FilesS3TempDir.GetString()
	t.Cleanup(func() {
		s3Client = originalClient
		s3Bucket = originalBucket
		config.FilesS3TempDir.Set(originalTempDir)
	})

	tempDir := t.TempDir()
	config.FilesS3TempDir.Set(tempDir)

	client := &fakeS3PutObjectClient{}
	s3Client = client
	s3Bucket = "test-bucket"

	content := []byte("mismatch-content")
	file := &File{ID: 999, Size: uint64(len(content) + 10)}

	err := file.Save(&readerOnly{r: bytes.NewReader(content)})
	require.NoError(t, err)

	require.NotNil(t, client.lastInput)
	require.NotNil(t, client.lastInput.ContentLength)
	assert.Equal(t, int64(len(content)), *client.lastInput.ContentLength)

	entries, readErr := os.ReadDir(tempDir)
	require.NoError(t, readErr)
	assert.Len(t, entries, 0)
}
