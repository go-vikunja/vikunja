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
	"io"
	"os"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFileStorageIntegration tests end-to-end file storage and retrieval
// with S3/MinIO storage backend. This test specifically validates S3 functionality
// and will fail if S3 is not properly configured.
func TestFileStorageIntegration(t *testing.T) {
	if os.Getenv("VIKUNJA_TESTS_S3") != "1" {
		t.Skip("Skipping integration tests - set VIKUNJA_TESTS_S3=1 to run")
	}

	// Initialize config to load environment variables
	config.InitConfig()

	// Ensure S3 is configured for this test
	storageType := config.FilesType.GetString()
	if storageType != "s3" {
		t.Skip("Skipping S3 integration tests - VIKUNJA_FILES_TYPE must be set to 's3'")
	}

	// Validate S3 configuration is present
	if config.FilesS3Endpoint.GetString() == "" {
		t.Fatal("S3 integration test requires VIKUNJA_FILES_S3_ENDPOINT to be set")
	}

	// Save original config values
	originalType := config.FilesType.GetString()
	originalBasePath := config.FilesBasePath.GetString()
	originalEndpoint := config.FilesS3Endpoint.GetString()
	originalBucket := config.FilesS3Bucket.GetString()
	originalRegion := config.FilesS3Region.GetString()
	originalAccessKey := config.FilesS3AccessKey.GetString()
	originalSecretKey := config.FilesS3SecretKey.GetString()
	originalUsePathStyle := config.FilesS3UsePathStyle.GetBool()

	// Restore config after test
	defer func() {
		config.FilesType.Set(originalType)
		config.FilesBasePath.Set(originalBasePath)
		config.FilesS3Endpoint.Set(originalEndpoint)
		config.FilesS3Bucket.Set(originalBucket)
		config.FilesS3Region.Set(originalRegion)
		config.FilesS3AccessKey.Set(originalAccessKey)
		config.FilesS3SecretKey.Set(originalSecretKey)
		config.FilesS3UsePathStyle.Set(originalUsePathStyle)
		// Reinitialize with original settings
		_ = InitFileHandler()
	}()

	t.Run("Initialize file handler with s3", func(t *testing.T) {
		err := InitFileHandler()
		require.NoError(t, err, "Failed to initialize file handler with type: s3")
		assert.NotNil(t, afs, "File system should be initialized")
	})

	t.Run("Create and retrieve file with s3", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		// Test data
		testContent := []byte("This is a test file for storage integration testing with s3")
		testFileName := "integration-test-file.txt"
		testAuth := &testauth{id: 1}

		// Create file
		fileReader := bytes.NewReader(testContent)
		createdFile, err := Create(fileReader, testFileName, uint64(len(testContent)), testAuth)
		require.NoError(t, err, "Failed to create file")
		require.NotNil(t, createdFile, "Created file should not be nil")
		assert.Positive(t, createdFile.ID, "File ID should be assigned")
		assert.Equal(t, testFileName, createdFile.Name, "File name should match")
		assert.Equal(t, uint64(len(testContent)), createdFile.Size, "File size should match")
		assert.Equal(t, int64(1), createdFile.CreatedByID, "Creator ID should match")

		// Load file metadata from database
		loadedFile := &File{ID: createdFile.ID}
		err = loadedFile.LoadFileMetaByID()
		require.NoError(t, err, "Failed to load file metadata")
		assert.Equal(t, testFileName, loadedFile.Name, "Loaded file name should match")
		assert.Equal(t, uint64(len(testContent)), loadedFile.Size, "Loaded file size should match")

		// Load and verify file content
		err = loadedFile.LoadFileByID()
		require.NoError(t, err, "Failed to load file content")
		require.NotNil(t, loadedFile.File, "File handle should not be nil")

		retrievedContent, err := io.ReadAll(loadedFile.File)
		require.NoError(t, err, "Failed to read file content")
		assert.Equal(t, testContent, retrievedContent, "Retrieved content should match original")

		_ = loadedFile.File.Close()

		// Verify file exists in storage
		fileInfo, err := FileStat(loadedFile)
		require.NoError(t, err, "File should exist in storage")
		assert.NotNil(t, fileInfo, "File info should not be nil")

		// Delete file
		s := db.NewSession()
		defer s.Close()
		err = loadedFile.Delete(s)
		require.NoError(t, err, "Failed to delete file")

		// Verify file is deleted from storage
		_, err = FileStat(loadedFile)
		require.Error(t, err, "File should not exist after deletion")
		assert.True(t, os.IsNotExist(err), "Error should indicate file does not exist")
	})

	t.Run("Create multiple files with s3", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		testAuth := &testauth{id: 1}
		fileIDs := make([]int64, 0, 3)

		// Create multiple files
		for i := 1; i <= 3; i++ {
			content := []byte("Test file content number " + string(rune('0'+i)))
			fileName := "test-file-" + string(rune('0'+i)) + ".txt"

			file, err := Create(bytes.NewReader(content), fileName, uint64(len(content)), testAuth)
			require.NoError(t, err, "Failed to create file %d", i)
			fileIDs = append(fileIDs, file.ID)
		}

		// Verify all files exist and can be retrieved
		for i, fileID := range fileIDs {
			file := &File{ID: fileID}
			err := file.LoadFileByID()
			require.NoError(t, err, "Failed to load file %d", i+1)

			content, err := io.ReadAll(file.File)
			require.NoError(t, err, "Failed to read file %d", i+1)
			expectedContent := "Test file content number " + string(rune('0'+i+1))
			assert.Equal(t, []byte(expectedContent), content, "Content should match for file %d", i+1)

			_ = file.File.Close()
		}

		// Clean up: delete all files
		s := db.NewSession()
		defer s.Close()
		for _, fileID := range fileIDs {
			file := &File{ID: fileID}
			err := file.Delete(s)
			require.NoError(t, err, "Failed to delete file")
		}
	})

	t.Run("Handle large file with s3", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		testAuth := &testauth{id: 1}
		// Create a 1MB file
		largeContent := bytes.Repeat([]byte("X"), 1024*1024)
		fileName := "large-test-file.bin"

		file, err := Create(bytes.NewReader(largeContent), fileName, uint64(len(largeContent)), testAuth)
		require.NoError(t, err, "Failed to create large file")
		assert.Equal(t, uint64(len(largeContent)), file.Size, "File size should match")

		// Retrieve and verify
		loadedFile := &File{ID: file.ID}
		err = loadedFile.LoadFileByID()
		require.NoError(t, err, "Failed to load large file")

		retrievedContent, err := io.ReadAll(loadedFile.File)
		require.NoError(t, err, "Failed to read large file")
		assert.Len(t, retrievedContent, len(largeContent), "Retrieved file size should match")
		assert.Equal(t, largeContent, retrievedContent, "Large file content should match")

		_ = loadedFile.File.Close()

		// Clean up
		s := db.NewSession()
		defer s.Close()
		err = loadedFile.Delete(s)
		require.NoError(t, err, "Failed to delete large file")
	})

	t.Run("File not found with s3", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		// Try to load a file that doesn't exist
		nonExistentFile := &File{ID: 999999}
		err := nonExistentFile.LoadFileByID()
		require.Error(t, err, "Loading non-existent file should error")
		assert.True(t, os.IsNotExist(err), "Error should indicate file does not exist")

		// Try to load metadata for non-existent file
		err = nonExistentFile.LoadFileMetaByID()
		require.Error(t, err, "Loading metadata for non-existent file should error")
		assert.True(t, IsErrFileDoesNotExist(err), "Error should be ErrFileDoesNotExist")
	})
}

// TestInitFileHandler_S3Configuration tests S3 configuration validation
func TestInitFileHandler_S3Configuration(t *testing.T) {
	// Save original config values
	originalType := config.FilesType.GetString()
	originalEndpoint := config.FilesS3Endpoint.GetString()
	originalBucket := config.FilesS3Bucket.GetString()
	originalRegion := config.FilesS3Region.GetString()
	originalAccessKey := config.FilesS3AccessKey.GetString()
	originalSecretKey := config.FilesS3SecretKey.GetString()

	// Restore config after test
	defer func() {
		config.FilesType.Set(originalType)
		config.FilesS3Endpoint.Set(originalEndpoint)
		config.FilesS3Bucket.Set(originalBucket)
		config.FilesS3Region.Set(originalRegion)
		config.FilesS3AccessKey.Set(originalAccessKey)
		config.FilesS3SecretKey.Set(originalSecretKey)
		_ = InitFileHandler()
	}()

	t.Run("valid S3 configuration", func(t *testing.T) {
		config.FilesType.Set("s3")
		config.FilesS3Endpoint.Set("https://s3.amazonaws.com")
		config.FilesS3Bucket.Set("test-bucket")
		config.FilesS3Region.Set("us-east-1")
		config.FilesS3AccessKey.Set("test-access-key")
		config.FilesS3SecretKey.Set("test-secret-key")

		// This should not return an error with valid configuration
		err := InitFileHandler()
		assert.NoError(t, err)
	})

	t.Run("missing S3 endpoint", func(t *testing.T) {
		config.FilesType.Set("s3")
		config.FilesS3Endpoint.Set("")
		config.FilesS3Bucket.Set("test-bucket")
		config.FilesS3AccessKey.Set("test-access-key")
		config.FilesS3SecretKey.Set("test-secret-key")

		// This should return an error for missing endpoint
		err := InitFileHandler()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "endpoint")
	})

	t.Run("missing S3 bucket", func(t *testing.T) {
		config.FilesType.Set("s3")
		config.FilesS3Endpoint.Set("https://s3.amazonaws.com")
		config.FilesS3Bucket.Set("")
		config.FilesS3AccessKey.Set("test-access-key")
		config.FilesS3SecretKey.Set("test-secret-key")

		// This should return an error for missing bucket
		err := InitFileHandler()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "bucket")
	})

	t.Run("missing S3 access key", func(t *testing.T) {
		config.FilesType.Set("s3")
		config.FilesS3Endpoint.Set("https://s3.amazonaws.com")
		config.FilesS3Bucket.Set("test-bucket")
		config.FilesS3AccessKey.Set("")
		config.FilesS3SecretKey.Set("test-secret-key")

		// This should return an error for missing access key
		err := InitFileHandler()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "access key")
	})

	t.Run("missing S3 secret key", func(t *testing.T) {
		config.FilesType.Set("s3")
		config.FilesS3Endpoint.Set("https://s3.amazonaws.com")
		config.FilesS3Bucket.Set("test-bucket")
		config.FilesS3AccessKey.Set("test-access-key")
		config.FilesS3SecretKey.Set("")

		// This should return an error for missing secret key
		err := InitFileHandler()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "secret key")
	})
}

func TestInitFileHandler_LocalFilesystem(t *testing.T) {
	// Save original config values
	originalType := config.FilesType.GetString()

	// Restore config after test
	defer func() {
		config.FilesType.Set(originalType)
	}()

	// Test with local filesystem
	config.FilesType.Set("local")

	// This should not return an error
	err := InitFileHandler()
	require.NoError(t, err)

	// Verify that afs is initialized
	assert.NotNil(t, afs)
}
