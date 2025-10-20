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
	"testing"

	"code.vikunja.io/api/pkg/config"

	"github.com/stretchr/testify/assert"
)

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
	}()

	// Test with S3 configuration
	config.FilesType.Set("s3")
	config.FilesS3Endpoint.Set("https://s3.amazonaws.com")
	config.FilesS3Bucket.Set("test-bucket")
	config.FilesS3Region.Set("us-east-1")
	config.FilesS3AccessKey.Set("test-access-key")
	config.FilesS3SecretKey.Set("test-secret-key")

	// This should not panic with valid configuration
	assert.NotPanics(t, func() {
		InitFileHandler()
	})

	// Test with missing S3 configuration
	config.FilesS3Endpoint.Set("")
	config.FilesS3Bucket.Set("")
	config.FilesS3AccessKey.Set("")
	config.FilesS3SecretKey.Set("")

	// This should panic with missing configuration
	assert.Panics(t, func() {
		InitFileHandler()
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

	// This should not panic
	assert.NotPanics(t, func() {
		InitFileHandler()
	})

	// Verify that afs is initialized
	assert.NotNil(t, afs)
}
