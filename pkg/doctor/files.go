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

package doctor

import (
	"fmt"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/files"

	"golang.org/x/sys/unix"
)

// CheckFiles returns file storage checks.
func CheckFiles() CheckGroup {
	fileType := config.FilesType.GetString()

	var results []CheckResult

	switch fileType {
	case "local":
		results = checkLocalStorage()
	case "s3":
		results = checkS3Storage()
	default:
		results = []CheckResult{
			{
				Name:   "Type",
				Passed: false,
				Error:  fmt.Sprintf("unknown storage type: %s", fileType),
			},
		}
	}

	return CheckGroup{
		Name:    fmt.Sprintf("Files (%s)", fileType),
		Results: results,
	}
}

func checkLocalStorage() []CheckResult {
	basePath := config.FilesBasePath.GetString()

	results := []CheckResult{
		{
			Name:   "Path",
			Passed: true,
			Value:  basePath,
		},
	}

	// Check writable using the existing ValidateFileStorage function
	if err := files.ValidateFileStorage(); err != nil {
		results = append(results, CheckResult{
			Name:   "Writable",
			Passed: false,
			Error:  err.Error(),
		})
	} else {
		results = append(results, CheckResult{
			Name:   "Writable",
			Passed: true,
			Value:  "yes",
		})
	}

	// Check disk space
	results = append(results, checkDiskSpace(basePath))

	return results
}

func checkDiskSpace(path string) CheckResult {
	var stat unix.Statfs_t
	if err := unix.Statfs(path, &stat); err != nil {
		return CheckResult{
			Name:   "Disk space",
			Passed: false,
			Error:  err.Error(),
		}
	}

	// Available space in bytes
	availableBytes := stat.Bavail * uint64(stat.Bsize)
	availableGB := float64(availableBytes) / (1024 * 1024 * 1024)

	return CheckResult{
		Name:   "Disk space",
		Passed: true,
		Value:  fmt.Sprintf("%.1f GB available", availableGB),
	}
}

func checkS3Storage() []CheckResult {
	endpoint := config.FilesS3Endpoint.GetString()
	bucket := config.FilesS3Bucket.GetString()

	results := []CheckResult{
		{
			Name:   "Endpoint",
			Passed: true,
			Value:  endpoint,
		},
		{
			Name:   "Bucket",
			Passed: true,
			Value:  bucket,
		},
	}

	// Check writable using the existing ValidateFileStorage function
	if err := files.ValidateFileStorage(); err != nil {
		results = append(results, CheckResult{
			Name:   "Writable",
			Passed: false,
			Error:  err.Error(),
		})
	} else {
		results = append(results, CheckResult{
			Name:   "Writable",
			Passed: true,
			Value:  "yes",
		})
	}

	return results
}
