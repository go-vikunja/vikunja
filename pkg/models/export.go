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

package models

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strconv"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/cron"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/notifications"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/utils"
	"code.vikunja.io/api/pkg/version"

	"xorm.io/xorm"
)

func ExportUserData(s *xorm.Session, u *user.User) (err error) {
	exportDir := config.FilesBasePath.GetString() + "/user-export-tmp/"
	err = os.MkdirAll(exportDir, 0700)
	if err != nil {
		return err
	}

	tmpFilename := exportDir + strconv.FormatInt(u.ID, 10) + "_" + time.Now().Format("2006-01-02_15-03-05") + ".zip"

	// Open zip
	dumpFile, err := os.Create(tmpFilename)
	if err != nil {
		return fmt.Errorf("error opening dump file: %w", err)
	}
	defer dumpFile.Close()

	dumpWriter := zip.NewWriter(dumpFile)
	defer dumpWriter.Close()

	// Get the data
	taskIDs, err := exportProjectsAndTasks(s, u, dumpWriter)
	if err != nil {
		return err
	}
	// Task attachment files
	err = exportTaskAttachments(s, dumpWriter, taskIDs)
	if err != nil {
		return err
	}
	// Saved filters
	err = exportSavedFilters(s, u, dumpWriter)
	if err != nil {
		return err
	}
	// Background files
	err = exportProjectBackgrounds(s, u, dumpWriter)
	if err != nil {
		return err
	}
	// Vikunja Version
	err = utils.WriteBytesToZip("VERSION", []byte(version.Version), dumpWriter)
	if err != nil {
		return err
	}

	// If we reuse the same file again, saving it as a file in Vikunja will save it as a file with 0 bytes in size.
	// Closing and reopening does work.
	dumpWriter.Close()
	dumpFile.Close()

	exported, err := os.Open(tmpFilename)
	if err != nil {
		return err
	}

	stat, err := exported.Stat()
	if err != nil {
		return err
	}

	exportFile, err := files.CreateWithMimeAndSession(s, exported, tmpFilename, uint64(stat.Size()), u, "application/zip", false)
	if err != nil {
		return err
	}

	// Save the file id with the user
	u.ExportFileID = exportFile.ID
	_, err = s.
		Where("id = ?", u.ID).
		Cols("export_file_id").
		Update(u)
	if err != nil {
		return
	}

	// Remove the old file
	err = os.Remove(exported.Name())
	if err != nil {
		return err
	}

	// Send a notification
	return notifications.Notify(u, &DataExportReadyNotification{
		User: u,
	})
}

func exportProjectsAndTasks(s *xorm.Session, u *user.User, wr *zip.Writer) (taskIDs []int64, err error) {

	// Get all projects
	rawProjects, _, _, err := getRawProjectsForUser(
		s,
		&projectOptions{
			search:      "",
			user:        u,
			page:        0,
			perPage:     -1,
			getArchived: true,
		})
	if err != nil {
		return taskIDs, err
	}

	if len(rawProjects) == 0 {
		return
	}

	projects := []*ProjectWithTasksAndBuckets{}
	projectsMap := make(map[int64]*ProjectWithTasksAndBuckets, len(rawProjects))
	projectIDs := []int64{}
	for _, p := range rawProjects {
		if p.BackgroundFileID > 0 {
			p.BackgroundInformation = &files.File{
				ID: p.BackgroundFileID,
			}
		}
		pp := &ProjectWithTasksAndBuckets{
			Project: *p,
		}
		projects = append(projects, pp)
		projectsMap[p.ID] = pp
		projectIDs = append(projectIDs, p.ID)
	}

	views := map[int64]*ProjectView{}
	err = s.In("project_id", projectIDs).Find(&views)
	if err != nil {
		return
	}

	viewIDs := []int64{}
	for _, v := range views {
		if projectsMap[v.ProjectID].Views == nil {
			projectsMap[v.ProjectID].Views = []*ProjectView{}
		}
		projectsMap[v.ProjectID].Views = append(projectsMap[v.ProjectID].Views, v)
		viewIDs = append(viewIDs, v.ID)
	}

	tasks, _, _, err := getTasksForProjects(s, rawProjects, u, &taskSearchOptions{
		page:    0,
		perPage: -1,
	}, nil)
	if err != nil {
		return taskIDs, err
	}

	taskMap := make(map[int64]*TaskWithComments, len(tasks))
	for _, t := range tasks {
		taskMap[t.ID] = &TaskWithComments{
			Task: *t,
		}
		if _, exists := projectsMap[t.ProjectID]; !exists {
			log.Debugf("[User Data Export] Project %d does not exist for task %d, omitting", t.ProjectID, t.ID)
			continue
		}
		projectsMap[t.ProjectID].Tasks = append(projectsMap[t.ProjectID].Tasks, taskMap[t.ID])
		taskIDs = append(taskIDs, t.ID)
	}

	comments := []*TaskComment{}
	err = s.
		Join("LEFT", "tasks", "tasks.id = task_comments.task_id").
		In("tasks.project_id", projectIDs).
		Find(&comments)
	if err != nil {
		return
	}

	for _, c := range comments {
		if _, exists := taskMap[c.TaskID]; !exists {
			log.Debugf("[User Data Export] Task %d does not exist for comment %d, omitting", c.TaskID, c.ID)
			continue
		}
		taskMap[c.TaskID].Comments = append(taskMap[c.TaskID].Comments, c)
	}

	buckets := []*Bucket{}
	err = s.In("project_view_id", viewIDs).Find(&buckets)
	if err != nil {
		return
	}

	bucketIDs := []int64{}
	for _, b := range buckets {
		view, exists := views[b.ProjectViewID]
		if !exists {
			log.Debugf("[User Data Export] Project view %d does not exist for bucket %d, omitting", b.ProjectViewID, b.ID)
			continue
		}
		_, exists = projectsMap[view.ProjectID]
		if !exists {
			log.Debugf("[User Data Export] Project %d does not exist for bucket %d, omitting", view.ProjectID, b.ID)
			continue
		}
		projectsMap[view.ProjectID].Buckets = append(projectsMap[view.ProjectID].Buckets, b)
		bucketIDs = append(bucketIDs, b.ID)
	}

	taskBuckets := []*TaskBucket{}
	err = s.In("bucket_id", bucketIDs).Find(&taskBuckets)
	if err != nil {
		return
	}

	for _, tb := range taskBuckets {
		view, exists := views[tb.ProjectViewID]
		if !exists {
			log.Debugf("[User Data Export] Project view %d does not exist, omitting", tb.ProjectViewID)
			continue
		}
		_, exists = projectsMap[view.ProjectID]
		if !exists {
			log.Debugf("[User Data Export] Project %d does not exist, omitting", view.ProjectID)
			continue
		}

		if projectsMap[view.ProjectID].TaskBuckets == nil {
			projectsMap[view.ProjectID].TaskBuckets = []*TaskBucket{}
		}

		projectsMap[view.ProjectID].TaskBuckets = append(projectsMap[view.ProjectID].TaskBuckets, tb)
	}

	taskPositions := []*TaskPosition{}
	err = s.In("project_view_id", viewIDs).Find(&taskPositions)
	if err != nil {
		return
	}

	for _, p := range taskPositions {
		view, exists := views[p.ProjectViewID]
		if !exists {
			log.Debugf("[User Data Export] Project view %d does not exist, omitting", p.ProjectViewID)
			continue
		}
		_, exists = projectsMap[view.ProjectID]
		if !exists {
			log.Debugf("[User Data Export] Project %d does not exist, omitting", view.ProjectID)
			continue
		}

		if projectsMap[view.ProjectID].Positions == nil {
			projectsMap[view.ProjectID].Positions = []*TaskPosition{}
		}

		projectsMap[view.ProjectID].Positions = append(projectsMap[view.ProjectID].Positions, p)
	}

	data, err := json.Marshal(projects)
	if err != nil {
		return taskIDs, err
	}

	return taskIDs, utils.WriteBytesToZip("data.json", data, wr)
}

func exportTaskAttachments(s *xorm.Session, wr *zip.Writer, taskIDs []int64) (err error) {
	tas, err := getTaskAttachmentsByTaskIDs(s, taskIDs)
	if err != nil {
		return err
	}

	attachmentFiles := make(map[int64]io.ReadCloser)
	for _, ta := range tas {
		err = ta.File.LoadFileByID()
		if err != nil {
			var pathError *fs.PathError
			if errors.As(err, &pathError) {
				continue
			}
			return err
		}
		attachmentFiles[ta.FileID] = ta.File.File
	}

	return utils.WriteFilesToZip(attachmentFiles, wr)
}

func exportSavedFilters(s *xorm.Session, u *user.User, wr *zip.Writer) (err error) {
	filters, err := getSavedFiltersForUser(s, u, "")
	if err != nil {
		return err
	}

	data, err := json.Marshal(filters)
	if err != nil {
		return err
	}

	return utils.WriteBytesToZip("filters.json", data, wr)
}

func exportProjectBackgrounds(s *xorm.Session, u *user.User, wr *zip.Writer) (err error) {
	projects, _, _, err := getRawProjectsForUser(
		s,
		&projectOptions{
			user: u,
			page: -1,
		},
	)
	if err != nil {
		return err
	}

	backgroundFiles := make(map[int64]io.ReadCloser)
	for _, l := range projects {
		if l.BackgroundFileID == 0 {
			continue
		}

		bgFile := &files.File{
			ID: l.BackgroundFileID,
		}
		err = bgFile.LoadFileByID()
		if err != nil {
			var pathError *fs.PathError
			if errors.As(err, &pathError) {
				continue
			}
			return err
		}

		backgroundFiles[l.BackgroundFileID] = bgFile.File
	}

	return utils.WriteFilesToZip(backgroundFiles, wr)
}

func RegisterOldExportCleanupCron() {
	const logPrefix = "[User Export Cleanup Cron] "

	err := cron.Schedule("0 * * * *", func() {
		s := db.NewSession()
		defer s.Close()

		users := []*user.User{}
		err := s.Where("export_file_id IS NOT NULL AND export_file_id != ?", 0).Find(&users)
		if err != nil {
			log.Errorf(logPrefix+"Could not get users with export files: %s", err)
			return
		}

		fileIDs := []int64{}
		for _, u := range users {
			fileIDs = append(fileIDs, u.ExportFileID)
		}

		fs := []*files.File{}
		err = s.Where("created < ?", time.Now().Add(-time.Hour*24*7)).In("id", fileIDs).Find(&fs)
		if err != nil {
			log.Errorf(logPrefix+"Could not get users with export files: %s", err)
			return
		}

		if len(fs) == 0 {
			return
		}

		log.Debugf(logPrefix+"Removing %d old user data exports...", len(fs))

		for _, f := range fs {
			err = f.Delete(s)
			if err != nil {
				log.Errorf(logPrefix+"Could not remove user export file %d: %s", f.ID, err)
				return
			}
		}

		_, err = s.In("export_file_id", fileIDs).Cols("export_file_id").Update(&user.User{})
		if err != nil {
			log.Errorf(logPrefix+"Could not update user export file state: %s", err)
			return
		}

		log.Debugf(logPrefix+"Removed %d old user data exports...", len(fs))

	})
	if err != nil {
		log.Fatalf("Could not register old export cleanup cron: %s", err)
	}
}
