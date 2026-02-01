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
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/modules/keyvalue"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	aferos3 "github.com/fclairamb/afero-s3"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

// This file handles storing and retrieving a file for different backends
var fs afero.Fs
var afs *afero.Afero

// S3 client and bucket for direct uploads with Content-Length
type s3PutObjectClient interface {
	PutObject(ctx context.Context, input *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
}

var s3Client s3PutObjectClient
var s3Bucket string

func setDefaultLocalConfig() {
	if !strings.HasPrefix(config.FilesBasePath.GetString(), "/") {
		config.FilesBasePath.Set(filepath.Join(
			config.ServiceRootpath.GetString(),
			config.FilesBasePath.GetString(),
		))
	}
}

// initS3FileHandler initializes the S3 file backend
func initS3FileHandler() error {
	// Get S3 configuration
	endpoint := config.FilesS3Endpoint.GetString()
	bucket := config.FilesS3Bucket.GetString()
	region := config.FilesS3Region.GetString()
	accessKey := config.FilesS3AccessKey.GetString()
	secretKey := config.FilesS3SecretKey.GetString()

	if endpoint == "" {
		return errors.New("S3 endpoint is not configured. Please set files.s3.endpoint")
	}
	if bucket == "" {
		return errors.New("S3 bucket is not configured. Please set files.s3.bucket")
	}
	if accessKey == "" {
		return errors.New("S3 access key is not configured. Please set files.s3.accesskey")
	}
	if secretKey == "" {
		return errors.New("S3 secret key is not configured. Please set files.s3.secretkey")
	}

	// Create AWS SDK v2 config
	cfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client with custom endpoint and path style options
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = config.FilesS3UsePathStyle.GetBool()
	})

	// Initialize S3 filesystem using afero-s3
	fs = aferos3.NewFsFromClient(bucket, client)
	afs = &afero.Afero{Fs: fs}

	// Store S3 client and bucket for direct uploads with Content-Length
	s3Client = client
	s3Bucket = bucket

	return nil
}

// initLocalFileHandler initializes the local filesystem backend
func initLocalFileHandler() {
	fs = afero.NewOsFs()
	afs = &afero.Afero{Fs: fs}
	s3Client = nil
	setDefaultLocalConfig()
}

// InitFileHandler creates a new file handler for the file backend we want to use
func InitFileHandler() error {
	fileType := config.FilesType.GetString()

	switch fileType {
	case "s3":
		if err := initS3FileHandler(); err != nil {
			return err
		}
	case "local":
		initLocalFileHandler()
	default:
		return fmt.Errorf("invalid file storage type '%s': must be 'local' or 's3'", fileType)
	}

	if err := ValidateFileStorage(); err != nil {
		return fmt.Errorf("storage validation failed: %w", err)
	}

	return nil
}

// InitTestFileHandler initializes a new memory file system for testing
func InitTestFileHandler() {
	fs = afero.NewMemMapFs()
	afs = &afero.Afero{Fs: fs}
	setDefaultLocalConfig()
}

func initFixtures(t *testing.T) {
	// DB fixtures
	db.LoadAndAssertFixtures(t)
	// File fixtures
	InitTestFileFixtures(t)
	err := config.SetMaxFileSizeMBytesFromString("20MB")
	require.NoError(t, err)
}

// InitTestFileFixtures initializes file fixtures
func InitTestFileFixtures(t *testing.T) {
	testfile := &File{ID: 1}
	err := afero.WriteFile(afs, testfile.getAbsoluteFilePath(), []byte("testfile1"), 0644)
	require.NoError(t, err)
}

// InitTests handles the actual bootstrapping of the test env
func InitTests() {
	var err error
	x, err = db.CreateTestEngine()
	if err != nil {
		log.Fatal(err)
	}

	err = x.Sync2(GetTables()...)
	if err != nil {
		log.Fatal(err)
	}

	err = db.InitTestFixtures("files")
	if err != nil {
		log.Fatal(err)
	}

	InitTestFileHandler()

	keyvalue.InitStorage()
}

// FileStat stats a file. This is an exported function to be able to test this from outide of the package
func FileStat(file *File) (os.FileInfo, error) {
	return afs.Stat(file.getAbsoluteFilePath())
}

// ValidateFileStorage checks that the configured file storage is writable
// by creating and removing a temporary file.
func ValidateFileStorage() error {
	basePath := config.FilesBasePath.GetString()

	diag := storageDiagnosticInfo(basePath)
	if diag != "" {
		diag = "\n" + diag
	}

	// For local filesystem, ensure the base directory exists
	if config.FilesType.GetString() == "local" {
		// Check if directory exists
		info, err := afs.Stat(basePath)
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				// Error other than "file doesn't exist"
				return fmt.Errorf("failed to access file storage directory at %s: %w%s", basePath, err, diag)
			}

			// Directory doesn't exist, try to create it
			err = afs.MkdirAll(basePath, 0755)
			if err != nil {
				return fmt.Errorf("failed to create file storage directory at %s: %w%s", basePath, err, diag)
			}
		} else if !info.IsDir() {
			// Path exists but is not a directory
			return fmt.Errorf("file storage path exists but is not a directory: %s", basePath)
		}
	}

	filename := fmt.Sprintf(".vikunja-check-%d", time.Now().UnixNano())
	path := filepath.Join(basePath, filename)

	err := writeToStorage(path, bytes.NewReader([]byte{}), 0)
	if err != nil {
		return fmt.Errorf("failed to create test file at %s: %w%s", path, err, diag)
	}

	err = afs.Remove(path)
	if err != nil {
		return fmt.Errorf("failed to remove test file at %s: %w", path, err)
	}

	return nil
}
