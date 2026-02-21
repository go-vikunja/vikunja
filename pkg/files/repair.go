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
	"fmt"

	"code.vikunja.io/api/pkg/log"

	"github.com/gabriel-vasile/mimetype"
	"github.com/schollz/progressbar/v3"
	"xorm.io/xorm"
)

// RepairMimeTypesResult holds the summary of a MIME type repair run.
type RepairMimeTypesResult struct {
	Total   int
	Updated int
	Errors  []string
}

// RepairFileMimeTypes finds all files with no MIME type set, detects it from
// the stored file content, and updates the database.
func RepairFileMimeTypes(s *xorm.Session) (*RepairMimeTypesResult, error) {
	var files []*File
	err := s.Where("mime = '' OR mime IS NULL").Find(&files)
	if err != nil {
		return nil, fmt.Errorf("failed to query files with missing mime type: %w", err)
	}

	result := &RepairMimeTypesResult{
		Total: len(files),
	}

	if len(files) == 0 {
		return result, nil
	}

	bar := progressbar.Default(int64(len(files)), "Detecting MIME types")

	for _, f := range files {
		file, err := afs.Open(f.getAbsoluteFilePath())
		if err != nil {
			msg := fmt.Sprintf("file %d: failed to open: %s", f.ID, err)
			log.Errorf("file %d: failed to open: %s", f.ID, err)
			result.Errors = append(result.Errors, msg)
			_ = bar.Add(1)
			continue
		}

		mime, err := mimetype.DetectReader(file)
		_ = file.Close()
		if err != nil {
			msg := fmt.Sprintf("file %d: failed to detect mime type: %s", f.ID, err)
			log.Errorf("file %d: failed to detect mime type: %s", f.ID, err)
			result.Errors = append(result.Errors, msg)
			_ = bar.Add(1)
			continue
		}

		f.Mime = mime.String()
		_, err = s.ID(f.ID).Cols("mime").Update(f)
		if err != nil {
			msg := fmt.Sprintf("file %d: failed to update mime type: %s", f.ID, err)
			log.Errorf("file %d: failed to update mime type: %s", f.ID, err)
			result.Errors = append(result.Errors, msg)
			_ = bar.Add(1)
			continue
		}

		result.Updated++
		_ = bar.Add(1)
	}

	_ = bar.Finish()

	return result, nil
}
