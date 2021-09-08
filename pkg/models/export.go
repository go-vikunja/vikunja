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

package models

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
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
		return fmt.Errorf("error opening dump file: %s", err)
	}
	defer dumpFile.Close()

	dumpWriter := zip.NewWriter(dumpFile)
	defer dumpWriter.Close()

	// Get the data
	err = exportListsAndTasks(s, u, dumpWriter)
	if err != nil {
		return err
	}
	// Task attachment files
	err = exportTaskAttachments(s, u, dumpWriter)
	if err != nil {
		return err
	}
	// Saved filters
	err = exportSavedFilters(s, u, dumpWriter)
	if err != nil {
		return err
	}
	// Background files
	err = exportListBackgrounds(s, u, dumpWriter)
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

	exportFile, err := files.CreateWithMimeAndSession(s, exported, tmpFilename, uint64(stat.Size()), u, "application/zip")
	if err != nil {
		return err
	}

	// Save the file id with the user
	u.ExportFileID = exportFile.ID
	_, err = s.Cols("export_file_id").Update(u)
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

func exportListsAndTasks(s *xorm.Session, u *user.User, wr *zip.Writer) (err error) {

	namspaces, _, _, err := (&Namespace{IsArchived: true}).ReadAll(s, u, "", -1, 0)
	if err != nil {
		return err
	}

	namespaceIDs := []int64{}
	namespaces := []*NamespaceWithListsAndTasks{}
	listMap := make(map[int64]*ListWithTasksAndBuckets)
	listIDs := []int64{}
	for _, n := range namspaces.([]*NamespaceWithLists) {
		if n.ID < 1 {
			// Don't include filters
			continue
		}

		nn := &NamespaceWithListsAndTasks{
			Namespace: n.Namespace,
			Lists:     []*ListWithTasksAndBuckets{},
		}

		for _, l := range n.Lists {
			ll := &ListWithTasksAndBuckets{
				List:             *l,
				BackgroundFileID: l.BackgroundFileID,
				Tasks:            []*TaskWithComments{},
			}
			nn.Lists = append(nn.Lists, ll)
			listMap[l.ID] = ll
			listIDs = append(listIDs, l.ID)
		}

		namespaceIDs = append(namespaceIDs, n.ID)
		namespaces = append(namespaces, nn)
	}

	if len(namespaceIDs) == 0 {
		return nil
	}

	// Get all lists
	lists, err := getListsForNamespaces(s, namespaceIDs, true)
	if err != nil {
		return err
	}

	tasks, _, _, err := getTasksForLists(s, lists, u, &taskOptions{
		page:    0,
		perPage: -1,
	})
	if err != nil {
		return err
	}

	taskMap := make(map[int64]*TaskWithComments, len(tasks))
	for _, t := range tasks {
		taskMap[t.ID] = &TaskWithComments{
			Task: *t,
		}
		if _, exists := listMap[t.ListID]; !exists {
			log.Debugf("[User Data Export] List %d does not exist for task %d, omitting", t.ListID, t.ID)
			continue
		}
		listMap[t.ListID].Tasks = append(listMap[t.ListID].Tasks, taskMap[t.ID])
	}

	comments := []*TaskComment{}
	err = s.
		Join("LEFT", "tasks", "tasks.id = task_comments.task_id").
		In("tasks.list_id", listIDs).
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
	err = s.In("list_id", listIDs).Find(&buckets)
	if err != nil {
		return
	}

	for _, b := range buckets {
		if _, exists := listMap[b.ListID]; !exists {
			log.Debugf("[User Data Export] List %d does not exist for bucket %d, omitting", b.ListID, b.ID)
			continue
		}
		listMap[b.ListID].Buckets = append(listMap[b.ListID].Buckets, b)
	}

	data, err := json.Marshal(namespaces)
	if err != nil {
		return err
	}

	return utils.WriteBytesToZip("data.json", data, wr)
}

func exportTaskAttachments(s *xorm.Session, u *user.User, wr *zip.Writer) (err error) {
	lists, _, _, err := getRawListsForUser(
		s,
		&listOptions{
			user: u,
			page: -1,
		},
	)
	if err != nil {
		return err
	}

	tasks, _, _, err := getRawTasksForLists(s, lists, u, &taskOptions{page: -1})
	if err != nil {
		return err
	}

	taskIDs := []int64{}
	for _, t := range tasks {
		taskIDs = append(taskIDs, t.ID)
	}

	tas, err := getTaskAttachmentsByTaskIDs(s, taskIDs)
	if err != nil {
		return err
	}

	fs := make(map[int64]io.ReadCloser)
	for _, ta := range tas {
		if err := ta.File.LoadFileByID(); err != nil {
			return err
		}
		fs[ta.FileID] = ta.File.File
	}

	return utils.WriteFilesToZip(fs, wr)
}

func exportSavedFilters(s *xorm.Session, u *user.User, wr *zip.Writer) (err error) {
	filters, err := getSavedFiltersForUser(s, u)
	if err != nil {
		return err
	}

	data, err := json.Marshal(filters)
	if err != nil {
		return err
	}

	return utils.WriteBytesToZip("filters.json", data, wr)
}

func exportListBackgrounds(s *xorm.Session, u *user.User, wr *zip.Writer) (err error) {
	lists, _, _, err := getRawListsForUser(
		s,
		&listOptions{
			user: u,
			page: -1,
		},
	)
	if err != nil {
		return err
	}

	fs := make(map[int64]io.ReadCloser)
	for _, l := range lists {
		if l.BackgroundFileID == 0 {
			continue
		}

		bgFile := &files.File{
			ID: l.BackgroundFileID,
		}
		err = bgFile.LoadFileByID()
		if err != nil {
			return
		}

		fs[l.BackgroundFileID] = bgFile.File
	}

	return utils.WriteFilesToZip(fs, wr)
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
			err = f.Delete()
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
		log.Fatalf("Could not old export cleanup cron: %s", err)
	}
}
