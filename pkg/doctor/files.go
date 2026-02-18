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
	"io/fs"
	"os"
	"path/filepath"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/files"
)

// CheckFiles returns file storage checks.
func CheckFiles() CheckGroup {
	fileType := config.FilesType.GetString()

	// Initialize file handler
	if err := files.InitFileHandler(); err != nil {
		return CheckGroup{
			Name: fmt.Sprintf("Files (%s)", fileType),
			Results: []CheckResult{
				{
					Name:   "Initialization",
					Passed: false,
					Error:  err.Error(),
				},
			},
		}
	}

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

	// Check if the directory exists
	info, err := os.Stat(basePath)
	if err != nil {
		results = append(results, CheckResult{
			Name:   "Directory exists",
			Passed: false,
			Error:  err.Error(),
		})
		// If the directory doesn't exist, skip the remaining checks
		return results
	}

	results = append(results, CheckResult{
		Name:   "Directory exists",
		Passed: true,
		Value:  "yes",
	})

	// Directory permissions (octal mode)
	results = append(results, CheckResult{
		Name:   "Directory permissions",
		Passed: true,
		Value:  fmt.Sprintf("%04o", info.Mode().Perm()),
	})

	// Directory ownership (platform-specific)
	results = append(results, checkDirectoryOwnership(info)...)

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

	// Check disk space (platform-specific)
	results = append(results, checkDiskSpace(basePath))

	// Count files and total size in the directory
	results = append(results, checkFileStats(basePath))

	return results
}

func checkFileStats(basePath string) CheckResult {
	var totalFiles int
	var totalSize int64

	err := filepath.WalkDir(basePath, func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			totalFiles++
			info, err := d.Info()
			if err != nil {
				return err
			}
			totalSize += info.Size()
		}
		return nil
	})

	if err != nil {
		return CheckResult{
			Name:   "Stored files",
			Passed: false,
			Error:  fmt.Sprintf("error scanning directory: %s", err.Error()),
		}
	}

	return CheckResult{
		Name:   "Stored files",
		Passed: true,
		Value:  fmt.Sprintf("%d files, %s total", totalFiles, formatBytes(totalSize)),
	}
}

func formatBytes(b int64) string {
	const (
		kb = 1024
		mb = kb * 1024
		gb = mb * 1024
	)

	switch {
	case b >= gb:
		return fmt.Sprintf("%.1f GB", float64(b)/float64(gb))
	case b >= mb:
		return fmt.Sprintf("%.1f MB", float64(b)/float64(mb))
	case b >= kb:
		return fmt.Sprintf("%.1f KB", float64(b)/float64(kb))
	default:
		return fmt.Sprintf("%d B", b)
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
