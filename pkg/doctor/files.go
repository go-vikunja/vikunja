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
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/files"
)

// s3ProbeTimeout bounds every S3 round trip the check makes. An unreachable but
// well-formed endpoint would otherwise block until the OS TCP timeout.
const s3ProbeTimeout = 12 * time.Second

// CheckFiles returns file storage checks.
func CheckFiles() CheckGroup {
	fileType := config.FilesType.GetString()

	ctx := context.Background()
	if fileType == "s3" {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, s3ProbeTimeout)
		defer cancel()
	}

	// Not InitFileHandler: a diagnostic must not create the storage it reports on.
	if err := files.InitStorageBackend(ctx); err != nil {
		return CheckGroup{
			Name: fmt.Sprintf("Files (%s)", fileType),
			Results: []CheckResult{
				{
					Name:   "Initialization",
					Passed: false,
					Error:  storageError(ctx, err),
				},
			},
		}
	}

	var results []CheckResult

	switch fileType {
	case "local":
		results = checkLocalStorage(ctx)
	case "s3":
		results = checkS3Storage(ctx)
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

func checkLocalStorage(ctx context.Context) []CheckResult {
	basePath := config.FilesBasePath.GetString()

	results := []CheckResult{
		{
			Name:   "Path",
			Passed: true,
			Value:  basePath,
		},
	}

	info, err := os.Stat(basePath)
	if err != nil {
		results = append(results, CheckResult{
			Name:   "Directory exists",
			Passed: false,
			Error:  err.Error(),
		})
		return results
	}

	if !info.IsDir() {
		results = append(results, CheckResult{
			Name:   "Directory exists",
			Passed: false,
			Error:  fmt.Sprintf("%s exists but is not a directory", basePath),
		})
		return results
	}

	results = append(results, CheckResult{
		Name:   "Directory exists",
		Passed: true,
		Value:  "yes",
	})

	results = append(results, CheckResult{
		Name:   "Directory permissions",
		Passed: true,
		Value:  fmt.Sprintf("%04o", info.Mode().Perm()),
	})

	results = append(results, checkDirectoryOwnership(info)...)

	if err := files.ValidateFileStorage(ctx); err != nil {
		results = append(results, CheckResult{
			Name:   "Writable",
			Passed: false,
			Error:  storageError(ctx, err),
		})
	} else {
		results = append(results, CheckResult{
			Name:   "Writable",
			Passed: true,
			Value:  "yes",
		})
	}

	results = append(results, checkDiskSpace(basePath))
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

func checkS3Storage(ctx context.Context) []CheckResult {
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

	if err := files.ValidateFileStorage(ctx); err != nil {
		results = append(results, CheckResult{
			Name:   "Writable",
			Passed: false,
			Error:  storageError(ctx, err),
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

// storageError replaces the raw backend error with an actionable one when our own
// probe deadline fired.
func storageError(ctx context.Context, err error) string {
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return fmt.Sprintf("S3 endpoint %s did not respond within %s", config.FilesS3Endpoint.GetString(), s3ProbeTimeout)
	}
	return err.Error()
}
