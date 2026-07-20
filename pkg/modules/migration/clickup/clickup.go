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

package clickup

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/migration"
	"code.vikunja.io/api/pkg/user"
)

const apiBase = "https://api.clickup.com/api/v2"

// Migration represents the ClickUp migration.
//
// Unlike the OAuth-based migrators, ClickUp is authenticated with a personal
// API token the user pastes in directly - there is no redirect flow, so
// AuthURL returns an empty string. The frontend treats an empty AuthURL as
// the signal to render a plain text input instead of a "connect" button.
type Migration struct {
	Code string `json:"code"`
}

// Name is used to get the name of the ClickUp migration - we're using the docs here to annotate the status route.
// @Summary Get migration status
// @Description Returns if the current user already did the migation or not. This is useful to show a confirmation message in the frontend if the user is trying to do the same migration again.
// @tags migration
// @Produce json
// @Security JWTKeyAuth
// @Success 200 {object} migration.Status "The migration status"
// @Failure 500 {object} models.Message "Internal server error"
// @Router /migration/clickup/status [get]
func (m *Migration) Name() string {
	return "clickup"
}

// AuthURL is empty for ClickUp - it uses a pasted personal token, not an OAuth redirect.
// @Summary Get the auth url from ClickUp
// @Description Empty for ClickUp: it authenticates with a personal API token pasted directly into the frontend, not an OAuth redirect.
// @tags migration
// @Produce json
// @Security JWTKeyAuth
// @Success 200 {object} handler.AuthURL "The (empty) auth url."
// @Failure 500 {object} models.Message "Internal server error"
// @Router /migration/clickup/auth [get]
func (m *Migration) AuthURL() string {
	return ""
}

type clickupTeam struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type teamsResponse struct {
	Teams []*clickupTeam `json:"teams"`
}

type clickupSpace struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Archived bool   `json:"archived"`
}

type spacesResponse struct {
	Spaces []*clickupSpace `json:"spaces"`
}

type clickupFolder struct {
	ID       string         `json:"id"`
	Name     string         `json:"name"`
	Archived bool           `json:"archived"`
	Lists    []*clickupList `json:"lists"`
}

type foldersResponse struct {
	Folders []*clickupFolder `json:"folders"`
}

type clickupList struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type listsResponse struct {
	Lists []*clickupList `json:"lists"`
}

type clickupStatus struct {
	Status string `json:"status"`
	Type   string `json:"type"`
}

type clickupPriority struct {
	Priority string `json:"priority"`
}

type clickupTag struct {
	Name string `json:"name"`
}

type clickupAttachment struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Extension string `json:"extension"`
	URL       string `json:"url"`
}

type clickupTask struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Status      clickupStatus       `json:"status"`
	DateClosed  string              `json:"date_closed"`
	DueDate     string              `json:"due_date"`
	StartDate   string              `json:"start_date"`
	Priority    *clickupPriority    `json:"priority"`
	Tags        []clickupTag        `json:"tags"`
	Parent      string              `json:"parent"`
	Attachments []clickupAttachment `json:"attachments"`
}

type tasksResponse struct {
	Tasks    []*clickupTask `json:"tasks"`
	LastPage bool           `json:"last_page"`
}

var priorityMap = map[string]int64{
	"low":    1,
	"normal": 2,
	"high":   3,
	"urgent": 4,
}

// authHeader returns the headers clickup expects - a bare token, no "Bearer" prefix.
func (m *Migration) authHeader() map[string]string {
	return map[string]string{"Authorization": m.Code}
}

func (m *Migration) getTeams() ([]*clickupTeam, error) {
	resp, err := migration.DoGetWithHeaders(apiBase+"/team", m.authHeader())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	r := &teamsResponse{}
	if err := json.NewDecoder(resp.Body).Decode(r); err != nil {
		return nil, err
	}
	return r.Teams, nil
}

func (m *Migration) getSpaces(teamID string) ([]*clickupSpace, error) {
	resp, err := migration.DoGetWithHeaders(apiBase+"/team/"+teamID+"/space?archived=false", m.authHeader())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	r := &spacesResponse{}
	if err := json.NewDecoder(resp.Body).Decode(r); err != nil {
		return nil, err
	}
	return r.Spaces, nil
}

func (m *Migration) getFolders(spaceID string) ([]*clickupFolder, error) {
	resp, err := migration.DoGetWithHeaders(apiBase+"/space/"+spaceID+"/folder?archived=false", m.authHeader())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	r := &foldersResponse{}
	if err := json.NewDecoder(resp.Body).Decode(r); err != nil {
		return nil, err
	}
	return r.Folders, nil
}

func (m *Migration) getFolderlessLists(spaceID string) ([]*clickupList, error) {
	resp, err := migration.DoGetWithHeaders(apiBase+"/space/"+spaceID+"/list?archived=false", m.authHeader())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	r := &listsResponse{}
	if err := json.NewDecoder(resp.Body).Decode(r); err != nil {
		return nil, err
	}
	return r.Lists, nil
}

func (m *Migration) getTasks(listID string) ([]*clickupTask, error) {
	var all []*clickupTask
	page := 0
	for {
		url := fmt.Sprintf("%s/list/%s/task?archived=false&include_closed=true&subtasks=true&page=%d", apiBase, listID, page)
		resp, err := migration.DoGetWithHeaders(url, m.authHeader())
		if err != nil {
			return nil, err
		}

		r := &tasksResponse{}
		decodeErr := json.NewDecoder(resp.Body).Decode(r)
		resp.Body.Close()
		if decodeErr != nil {
			return nil, decodeErr
		}

		all = append(all, r.Tasks...)
		if r.LastPage || len(r.Tasks) == 0 {
			break
		}
		page++
	}
	return all, nil
}

func parseClickupTimestamp(ms string) (t time.Time, ok bool) {
	if ms == "" {
		return time.Time{}, false
	}
	millis, err := strconv.ParseInt(ms, 10, 64)
	if err != nil {
		return time.Time{}, false
	}
	return time.UnixMilli(millis), true
}

// convertTaskToVikunja converts a single ClickUp task, downloading its attachments.
// A failed attachment download is logged and skipped rather than failing the
// whole task. Subtask relations are resolved by the caller once every task
// has been converted, since a task's parent may be fetched after the task
// itself.
func (m *Migration) convertTaskToVikunja(ct *clickupTask, bucketID int64) *models.TaskWithComments {
	task := &models.TaskWithComments{
		Task: models.Task{
			Title:       ct.Name,
			Description: ct.Description,
			BucketID:    bucketID,
		},
	}

	if ct.Status.Type == "closed" || ct.Status.Type == "done" {
		task.Done = true
		if closed, ok := parseClickupTimestamp(ct.DateClosed); ok {
			task.DoneAt = closed
		}
	}

	if ct.Priority != nil {
		task.Priority = priorityMap[ct.Priority.Priority]
	}

	if due, ok := parseClickupTimestamp(ct.DueDate); ok {
		task.DueDate = due.In(config.GetTimeZone())
	}
	if start, ok := parseClickupTimestamp(ct.StartDate); ok {
		task.StartDate = start.In(config.GetTimeZone())
	}

	for _, tag := range ct.Tags {
		task.Labels = append(task.Labels, &models.Label{Title: tag.Name})
	}

	for _, att := range ct.Attachments {
		if att.URL == "" {
			continue
		}
		buf, err := migration.DownloadFile(att.URL)
		if err != nil {
			log.Debugf("[ClickUp Migration] Could not download attachment %s for task %s: %s", att.ID, ct.ID, err)
			continue
		}
		name := att.Title
		if name == "" {
			name = att.ID + "." + att.Extension
		}
		task.Attachments = append(task.Attachments, &models.TaskAttachment{
			File: &files.File{
				Name:        name,
				Size:        uint64(buf.Len()),
				FileContent: buf.Bytes(),
			},
		})
	}

	return task
}

// Migrate gets all tasks from ClickUp for a user and puts them into vikunja
// @Summary Migrate all tasks, lists, spaces etc. from ClickUp
// @Description Migrates all spaces, folders, lists, tasks, tags, subtasks and attachments from ClickUp into Vikunja.
// @tags migration
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param migrationCode body clickup.Migration true "A personal ClickUp API token (Settings > Apps in ClickUp)."
// @Success 200 {object} models.Message "A message telling you everything was migrated successfully."
// @Failure 500 {object} models.Message "Internal server error"
// @Router /migration/clickup/migrate [post]
func (m *Migration) Migrate(u *user.User) (err error) {
	log.Debugf("[ClickUp Migration] Starting migration for user %d", u.ID)

	var pseudoParentID int64 = 1
	hierarchy := []*models.ProjectWithTasksAndBuckets{{
		Project: models.Project{
			ID:    pseudoParentID,
			Title: "Migrated from ClickUp",
		},
	}}

	var nextID = pseudoParentID + 1
	// tasksByClickupID and pendingParents let a final pass resolve subtask
	// relations once every task in every list has been converted, regardless
	// of which list/folder/space a task's parent happens to live in.
	tasksByClickupID := make(map[string]*models.Task)
	var pendingParents []subtaskLink

	teams, err := m.getTeams()
	if err != nil {
		return err
	}

	for _, team := range teams {
		spaces, spaceErr := m.getSpaces(team.ID)
		if spaceErr != nil {
			return spaceErr
		}

		for _, space := range spaces {
			spaceProjectID := nextID
			nextID++
			spaceProject := &models.ProjectWithTasksAndBuckets{
				Project: models.Project{
					ID:              spaceProjectID,
					ParentProjectID: &pseudoParentID,
					Title:           space.Name,
					IsArchived:      space.Archived,
				},
			}
			hierarchy = append(hierarchy, spaceProject)

			folders, folderErr := m.getFolders(space.ID)
			if folderErr != nil {
				return folderErr
			}

			for _, folder := range folders {
				folderProjectID := nextID
				nextID++
				folderProject := &models.ProjectWithTasksAndBuckets{
					Project: models.Project{
						ID:              folderProjectID,
						ParentProjectID: &spaceProjectID,
						Title:           folder.Name,
						IsArchived:      folder.Archived,
					},
				}
				hierarchy = append(hierarchy, folderProject)

				folderPending, listErr := m.migrateLists(folder.Lists, folderProject, tasksByClickupID, &nextID)
				if listErr != nil {
					return listErr
				}
				pendingParents = append(pendingParents, folderPending...)
			}

			folderlessLists, listErr := m.getFolderlessLists(space.ID)
			if listErr != nil {
				return listErr
			}
			spacePending, migrateErr := m.migrateLists(folderlessLists, spaceProject, tasksByClickupID, &nextID)
			if migrateErr != nil {
				return migrateErr
			}
			pendingParents = append(pendingParents, spacePending...)
		}
	}

	resolveSubtaskRelations(pendingParents, tasksByClickupID)

	log.Debugf("[ClickUp Migration] Done converting data for user %d", u.ID)
	log.Debugf("[ClickUp Migration] Start inserting data for user %d", u.ID)

	if err = migration.InsertFromStructure(hierarchy, u); err != nil {
		return err
	}

	log.Debugf("[ClickUp Migration] ClickUp migration done for user %d", u.ID)
	return nil
}

// subtaskLink records a converted task's ClickUp parent id so relations can
// be resolved in one final pass once every list in every folder/space has
// been fetched - a subtask's parent may live in a list this function hasn't
// reached yet.
type subtaskLink struct {
	task     *models.TaskWithComments
	parentID string
}

// migrateLists fetches every task in each of the given lists (each becomes a
// bucket in project), converts them, and records ClickUp-id -> Task so
// subtask relations can be resolved once every list has been processed.
func (m *Migration) migrateLists(lists []*clickupList, project *models.ProjectWithTasksAndBuckets, tasksByClickupID map[string]*models.Task, nextID *int64) (pending []subtaskLink, err error) {
	for _, list := range lists {
		bucketID := *nextID
		*nextID++
		project.Buckets = append(project.Buckets, &models.Bucket{ID: bucketID, Title: list.Name})

		tasks, err := m.getTasks(list.ID)
		if err != nil {
			return nil, err
		}

		for _, ct := range tasks {
			task := m.convertTaskToVikunja(ct, bucketID)
			project.Tasks = append(project.Tasks, task)
			tasksByClickupID[ct.ID] = &task.Task
			if ct.Parent != "" {
				pending = append(pending, subtaskLink{task: task, parentID: ct.Parent})
			}
		}
	}

	return pending, nil
}

// resolveSubtaskRelations links each pending subtask to its parent now that
// every list has been converted. A parent outside the token's accessible
// spaces/lists simply leaves that task as a normal top-level task.
func resolveSubtaskRelations(pending []subtaskLink, tasksByClickupID map[string]*models.Task) {
	for _, p := range pending {
		parentTask, ok := tasksByClickupID[p.parentID]
		if !ok {
			continue
		}
		if parentTask.RelatedTasks == nil {
			parentTask.RelatedTasks = make(models.RelatedTaskMap)
		}
		parentTask.RelatedTasks[models.RelationKindSubtask] = append(parentTask.RelatedTasks[models.RelationKindSubtask], &p.task.Task)
	}
}
