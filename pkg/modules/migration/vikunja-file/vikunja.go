// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2021 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package vikunjafile

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/migration"
	"code.vikunja.io/api/pkg/user"
)

const logPrefix = "[Vikunja File Import] "

type FileMigrator struct {
}

// Name is used to get the name of the vikunja-file migration - we're using the docs here to annotate the status route.
// @Summary Get migration status
// @Description Returns if the current user already did the migation or not. This is useful to show a confirmation message in the frontend if the user is trying to do the same migration again.
// @tags migration
// @Produce json
// @Security JWTKeyAuth
// @Success 200 {object} migration.Status "The migration status"
// @Failure 500 {object} models.Message "Internal server error"
// @Router /migration/vikunja-file/status [get]
func (v *FileMigrator) Name() string {
	return "vikunja-file"
}

// Migrate takes a vikunja file export, parses it and imports everything in it into Vikunja.
// @Summary Import all lists, tasks etc. from a Vikunja data export
// @Description Imports all projects, tasks, notes, reminders, subtasks and files from a Vikunjda data export into Vikunja.
// @tags migration
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param import formData string true "The Vikunja export zip file."
// @Success 200 {object} models.Message "A message telling you everything was migrated successfully."
// @Failure 500 {object} models.Message "Internal server error"
// @Router /migration/vikunja-file/migrate [post]
func (v *FileMigrator) Migrate(user *user.User, file io.ReaderAt, size int64) error {
	r, err := zip.NewReader(file, size)
	if err != nil {
		return fmt.Errorf("could not open import file: %s", err)
	}

	log.Debugf(logPrefix+"Importing a zip file containing %d files", len(r.File))

	var dataFile *zip.File
	var filterFile *zip.File
	storedFiles := make(map[int64]*zip.File)
	for _, f := range r.File {
		if strings.HasPrefix(f.Name, "files/") {
			fname := strings.ReplaceAll(f.Name, "files/", "")
			id, err := strconv.ParseInt(fname, 10, 64)
			if err != nil {
				return fmt.Errorf("could not convert file id: %s", err)
			}
			storedFiles[id] = f
			log.Debugf(logPrefix + "Found a blob file")
			continue
		}
		if f.Name == "data.json" {
			dataFile = f
			log.Debugf(logPrefix + "Found a data file")
			continue
		}
		if f.Name == "filters.json" {
			filterFile = f
			log.Debugf(logPrefix + "Found a filter file")
		}
	}

	if dataFile == nil {
		return fmt.Errorf("no data file provided")
	}

	log.Debugf(logPrefix + "")

	//////
	// Import the bulk of Vikunja data
	df, err := dataFile.Open()
	if err != nil {
		return fmt.Errorf("could not open data file: %s", err)
	}
	defer df.Close()

	var bufData bytes.Buffer
	if _, err := bufData.ReadFrom(df); err != nil {
		return fmt.Errorf("could not read data file: %s", err)
	}

	namespaces := []*models.NamespaceWithListsAndTasks{}
	if err := json.Unmarshal(bufData.Bytes(), &namespaces); err != nil {
		return fmt.Errorf("could not read data: %s", err)
	}

	for _, n := range namespaces {
		for _, l := range n.Lists {
			if b, exists := storedFiles[l.BackgroundFileID]; exists {
				bf, err := b.Open()
				if err != nil {
					return fmt.Errorf("could not open list background file %d for reading: %s", l.BackgroundFileID, err)
				}
				var buf bytes.Buffer
				if _, err := buf.ReadFrom(bf); err != nil {
					return fmt.Errorf("could not read list background file %d: %s", l.BackgroundFileID, err)
				}

				l.BackgroundInformation = &buf
			}

			for _, t := range l.Tasks {
				for _, label := range t.Labels {
					label.ID = 0
				}
				for _, comment := range t.Comments {
					comment.ID = 0
				}
				for _, attachment := range t.Attachments {
					af, err := storedFiles[attachment.File.ID].Open()
					if err != nil {
						return fmt.Errorf("could not open attachment %d for reading: %s", attachment.ID, err)
					}
					var buf bytes.Buffer
					if _, err := buf.ReadFrom(af); err != nil {
						return fmt.Errorf("could not read attachment %d: %s", attachment.ID, err)
					}

					attachment.ID = 0
					attachment.File.ID = 0
					attachment.File.FileContent = buf.Bytes()
				}
			}
		}
	}

	err = migration.InsertFromStructure(namespaces, user)
	if err != nil {
		return fmt.Errorf("could not insert data: %s", err)
	}

	if filterFile == nil {
		log.Debugf(logPrefix + "No filter file found")
		return nil
	}

	///////
	// Import filters
	ff, err := filterFile.Open()
	if err != nil {
		return fmt.Errorf("could not open filters file: %s", err)
	}
	defer ff.Close()

	var bufFilter bytes.Buffer
	if _, err := bufFilter.ReadFrom(ff); err != nil {
		return fmt.Errorf("could not read filters file: %s", err)
	}

	filters := []*models.SavedFilter{}
	if err := json.Unmarshal(bufFilter.Bytes(), &filters); err != nil {
		return fmt.Errorf("could not read filter data: %s", err)
	}

	log.Debugf(logPrefix+"Importing %d saved filters", len(filters))

	s := db.NewSession()
	defer s.Close()

	for _, f := range filters {
		f.ID = 0
		err = f.Create(s, user)
		if err != nil {
			_ = s.Rollback()
			return err
		}
	}

	return s.Commit()
}
