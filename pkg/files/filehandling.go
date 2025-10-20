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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/modules/keyvalue"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	s3 "github.com/fclairamb/afero-s3"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

// This file handles storing and retrieving a file for different backends
var fs afero.Fs
var afs *afero.Afero

func setDefaultConfig() {
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

	// Create AWS session for afero-s3
	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String(region),
		Credentials:      credentials.NewStaticCredentials(accessKey, secretKey, ""),
		Endpoint:         aws.String(endpoint),
		S3ForcePathStyle: aws.Bool(config.FilesS3UsePathStyle.GetBool()),
	})
	if err != nil {
		return fmt.Errorf("failed to create AWS session: %w", err)
	}

	// Initialize S3 filesystem using afero-s3
	fs = s3.NewFs(bucket, sess)
	afs = &afero.Afero{Fs: fs}
	
	return nil
}

// initLocalFileHandler initializes the local filesystem backend
func initLocalFileHandler() error {
	fs = afero.NewOsFs()
	afs = &afero.Afero{Fs: fs}
	return nil
}

// InitFileHandler creates a new file handler for the file backend we want to use
func InitFileHandler() error {
	fileType := config.FilesType.GetString()

	var err error
	switch fileType {
	case "s3":
		err = initS3FileHandler()
	case "local":
		err = initLocalFileHandler()
	default:
		err = initLocalFileHandler()
	}

	if err != nil {
		return err
	}

	setDefaultConfig()
	return nil
}

// InitTestFileHandler initializes a new memory file system for testing
func InitTestFileHandler() {
	fs = afero.NewMemMapFs()
	afs = &afero.Afero{Fs: fs}
	setDefaultConfig()
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
