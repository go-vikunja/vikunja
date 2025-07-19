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

package todoist

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/migration"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/utils"
)

const paginationLimit = 200

// Migration is the todoist migration struct
type Migration struct {
	Code string `json:"code"`
}

type apiTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

type label struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Color      string `json:"color"`
	ItemOrder  int64  `json:"item_order"`
	IsFavorite bool   `json:"is_favorite"`
}

type project struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Color      string `json:"color"`
	ParentID   string `json:"parent_id"`
	ChildOrder int64  `json:"child_order"`
	Collapsed  bool   `json:"collapsed"`
	Shared     bool   `json:"shared"`
	IsDeleted  bool   `json:"is_deleted"`
	IsArchived bool   `json:"is_archived"`
	IsFavorite bool   `json:"is_favorite"`
}

type dueDate struct {
	Date        string      `json:"date"`
	Timezone    interface{} `json:"timezone"`
	String      string      `json:"string"`
	Lang        string      `json:"lang"`
	IsRecurring bool        `json:"is_recurring"`
}

type item struct {
	ID             string      `json:"id"`
	UserID         string      `json:"user_id"`
	ProjectID      string      `json:"project_id"`
	Content        string      `json:"content"`
	Priority       int64       `json:"priority"`
	Due            *dueDate    `json:"due"`
	ParentID       string      `json:"parent_id"`
	ChildOrder     int64       `json:"child_order"`
	SectionID      string      `json:"section_id"`
	Children       interface{} `json:"children"`
	Labels         []string    `json:"labels"`
	AddedByUID     string      `json:"added_by_uid"`
	AssignedByUID  string      `json:"assigned_by_uid"`
	ResponsibleUID string      `json:"responsible_uid"`
	Checked        bool        `json:"checked"`
	DateAdded      time.Time   `json:"added_at"`
	HasMoreNotes   bool        `json:"has_more_notes"`
	DateCompleted  time.Time   `json:"completed_at"`
}

type itemWrapper struct {
	Item *item `json:"item"`
}

type doneItem struct {
	CompletedDate time.Time `json:"completed_at"`
	Content       string    `json:"content"`
	ID            string    `json:"id"`
	ProjectID     string    `json:"project_id"`
	TaskID        string    `json:"task_id"`
}

type doneItemSync struct {
	Items    []*doneItem         `json:"items"`
	Projects map[string]*project `json:"projects"`
}

type fileAttachment struct {
	FileType    string `json:"file_type"`
	FileName    string `json:"file_name"`
	FileSize    int64  `json:"file_size"`
	FileURL     string `json:"file_url"`
	UploadState string `json:"upload_state"`
}

type note struct {
	ID             string          `json:"id"`
	ProjectID      string          `json:"project_id"`
	ItemID         string          `json:"item_id"`
	Content        string          `json:"content"`
	FileAttachment *fileAttachment `json:"file_attachment"`
	Posted         time.Time       `json:"posted_at"`
}

type projectNote struct {
	Content        string          `json:"content"`
	FileAttachment *fileAttachment `json:"file_attachment"`
	ID             string          `json:"id"`
	Posted         time.Time       `json:"posted"`
	ProjectID      string          `json:"project_id"`
}

type reminder struct {
	ID       string   `json:"id"`
	ItemID   string   `json:"item_id"`
	Type     string   `json:"type"`
	Due      *dueDate `json:"due"`
	MmOffset int64    `json:"mm_offset"`
}

type section struct {
	ID           string    `json:"id"`
	DateAdded    time.Time `json:"added_at"`
	IsDeleted    bool      `json:"is_deleted"`
	Name         string    `json:"name"`
	ProjectID    string    `json:"project_id"`
	SectionOrder int64     `json:"section_order"`
}

type sync struct {
	Projects     []*project     `json:"projects"`
	Items        []*item        `json:"items"`
	Labels       []*label       `json:"labels"`
	Notes        []*note        `json:"notes"`
	ProjectNotes []*projectNote `json:"project_notes"`
	Reminders    []*reminder    `json:"reminders"`
	Sections     []*section     `json:"sections"`
}

var todoistColors = map[string]string{}

func init() {
	todoistColors = make(map[string]string, 19)
	// The todoists colors are static, taken from https://developer.todoist.com/guides/#colors
	todoistColors = map[string]string{
		"berry_red":   "b8256f",
		"red":         "db4035",
		"orange":      "ff9933",
		"yellow":      "fad000",
		"olive_green": "afb83b",
		"lime_green":  "7ecc49",
		"green":       "299438",
		"mint_green":  "6accbc",
		"teal":        "158fad",
		"sky_blue":    "14aaf5",
		"light_blue":  "96c3eb",
		"blue":        "4073ff",
		"grape":       "884dff",
		"violet":      "af38eb",
		"lavender":    "eb96eb",
		"magenta":     "e05194",
		"salmon":      "ff8d85",
		"charcoal":    "808080",
		"grey":        "b8b8b8",
		"taupe":       "ccac93",
	}
}

// Name is used to get the name of the todoist migration - we're using the docs here to annotate the status route.
// @Summary Get migration status
// @Description Returns if the current user already did the migation or not. This is useful to show a confirmation message in the frontend if the user is trying to do the same migration again.
// @tags migration
// @Produce json
// @Security JWTKeyAuth
// @Success 200 {object} migration.Status "The migration status"
// @Failure 500 {object} models.Message "Internal server error"
// @Router /migration/todoist/status [get]
func (m *Migration) Name() string {
	return "todoist"
}

// AuthURL returns the url users need to authenticate against
// @Summary Get the auth url from todoist
// @Description Returns the auth url where the user needs to get its auth code. This code can then be used to migrate everything from todoist to Vikunja.
// @tags migration
// @Produce json
// @Security JWTKeyAuth
// @Success 200 {object} handler.AuthURL "The auth url."
// @Failure 500 {object} models.Message "Internal server error"
// @Router /migration/todoist/auth [get]
func (m *Migration) AuthURL() string {
	state, err := utils.CryptoRandomString(32)
	if err != nil {
		state = "todoist-migration"
	}
	return "https://todoist.com/oauth/authorize" +
		"?client_id=" + config.MigrationTodoistClientID.GetString() +
		"&scope=data:read" +
		"&state=" + state
}

func parseDate(dateString string) (date time.Time, err error) {
	if len(dateString) == 10 {
		// We're probably dealing with a date in the form of 2021-11-23 without a time
		date, err = time.Parse("2006-01-02", dateString)
		if err == nil {
			// round the day to eod
			return date.Add(time.Hour*23 + time.Minute*59), nil
		}
	}

	date, err = time.Parse("2006-01-02T15:04:05Z", dateString)
	if err != nil {
		date, err = time.Parse("2006-01-02T15:04:05", dateString)
	}
	if err != nil {
		date, err = time.Parse("2006-01-02", dateString)
	}

	return date, err
}

func convertTodoistToVikunja(sync *sync, doneItems map[string]*doneItem) (fullVikunjaHierachie []*models.ProjectWithTasksAndBuckets, err error) {

	var pseudoParentID int64 = 1

	parent := &models.ProjectWithTasksAndBuckets{
		Project: models.Project{
			ID:    pseudoParentID,
			Title: "Migrated from todoist",
		},
	}
	fullVikunjaHierachie = append(fullVikunjaHierachie, parent)

	// A map for all vikunja lists with the project id they're coming from as key
	lists := make(map[string]*models.ProjectWithTasksAndBuckets, len(sync.Projects))

	// A map for all vikunja tasks with the todoist task id as key to find them easily and add more data
	tasks := make(map[string]*models.TaskWithComments, len(sync.Items))

	// A map for all vikunja labels with the todoist id as key to find them easier
	labels := make(map[string]*models.Label, len(sync.Labels))

	sections := make(map[string]int64)

	for index, p := range sync.Projects {
		project := &models.ProjectWithTasksAndBuckets{
			Project: models.Project{
				ID:              int64(index+1) + pseudoParentID,
				ParentProjectID: pseudoParentID,
				Title:           p.Name,
				HexColor:        todoistColors[p.Color],
				IsArchived:      p.IsArchived,
			},
		}

		lists[p.ID] = project

		fullVikunjaHierachie = append(fullVikunjaHierachie, project)
	}

	sort.Slice(sync.Sections, func(i, j int) bool {
		return sync.Sections[i].SectionOrder < sync.Sections[j].SectionOrder
	})

	var fabricatedSectionID int64 = 1
	for _, section := range sync.Sections {
		if section.IsDeleted || section.ProjectID == "" {
			continue
		}

		lists[section.ProjectID].Buckets = append(lists[section.ProjectID].Buckets, &models.Bucket{
			ID:      fabricatedSectionID,
			Title:   section.Name,
			Created: section.DateAdded,
		})
		sections[section.ID] = fabricatedSectionID
		fabricatedSectionID++
	}

	for _, label := range sync.Labels {
		labels[label.Name] = &models.Label{
			Title:    label.Name,
			HexColor: todoistColors[label.Color],
		}
	}

	for _, i := range sync.Items {

		if i == nil {
			// This should never happen
			continue
		}

		task := &models.TaskWithComments{
			Task: models.Task{
				Title:    i.Content,
				Created:  i.DateAdded.In(config.GetTimeZone()),
				Done:     i.Checked,
				BucketID: sections[i.SectionID],
			},
		}

		// Only try to parse the task done at date if the task is actually done
		// Sometimes weired things happen if we try to parse nil dates.
		if task.Done {
			task.DoneAt = i.DateCompleted.In(config.GetTimeZone())
		}

		done, has := doneItems[i.ID]
		if has {
			task.Done = true
			task.DoneAt = done.CompletedDate.In(config.GetTimeZone())
		}

		// Todoist priorities only range from 1 (lowest) and max 4 (highest), so we need to make slight adjustments
		if i.Priority > 1 {
			task.Priority = i.Priority
		}

		// Put the due date together
		if i.Due != nil {
			dueDate, err := parseDate(i.Due.Date)
			if err != nil {
				return nil, err
			}
			task.DueDate = dueDate.In(config.GetTimeZone())
		}

		// Put all labels together from earlier
		for _, lName := range i.Labels {
			task.Labels = append(task.Labels, labels[lName])
		}

		tasks[i.ID] = task

		if _, exists := lists[i.ProjectID]; !exists {
			log.Debugf("[Todoist Migration] Tried to put item %s in project %s but the project does not exist", i.ID, i.ProjectID)
			continue
		}

		lists[i.ProjectID].Tasks = append(lists[i.ProjectID].Tasks, task)
	}

	// If the parenId of a task is not 0, create a task relation
	// We're looping again here to make sure we have seen all tasks before and have them in our map
	for _, i := range sync.Items {
		if i.ParentID == "" {
			continue
		}

		if _, exists := tasks[i.ParentID]; !exists {
			log.Debugf("[Todoist Migration] Could not find task %s in tasks map while trying to resolve subtasks for task %s", i.ParentID, i.ID)
			continue
		}

		// Prevent all those nil errors
		if tasks[i.ParentID].RelatedTasks == nil {
			tasks[i.ParentID].RelatedTasks = make(models.RelatedTaskMap)
		}

		if _, exists := tasks[i.ID]; !exists {
			log.Debugf("[Todoist Migration] Could not find task %s in tasks map while trying to add it as subtask", i.ID)
			continue
		}

		tasks[i.ParentID].RelatedTasks[models.RelationKindSubtask] = append(tasks[i.ParentID].RelatedTasks[models.RelationKindSubtask], &tasks[i.ID].Task)

		// Remove the task from the top level structure, otherwise it is added twice
	outer:
		for _, list := range lists {
			for in, t := range list.Tasks {
				if t == tasks[i.ID] {
					list.Tasks = append(list.Tasks[:in], list.Tasks[in+1:]...)
					break outer
				}
			}
		}
		delete(tasks, i.ID)
	}

	// Task Notes -> Task Descriptions
	// FIXME: Should be comments
	for _, n := range sync.Notes {
		if _, exists := tasks[n.ItemID]; !exists {
			log.Debugf("[Todoist Migration] Could not find task %s for note %s", n.ItemID, n.ID)
			continue
		}

		if tasks[n.ItemID].Description != "" {
			tasks[n.ItemID].Description += "\n"
		}
		tasks[n.ItemID].Description += n.Content

		if n.FileAttachment == nil {
			continue
		}

		// Only add the attachment if there's something to download
		if len(n.FileAttachment.FileURL) > 0 {
			// Download the attachment and put it in the file
			buf, err := migration.DownloadFile(n.FileAttachment.FileURL)
			if err != nil {
				return nil, err
			}

			tasks[n.ItemID].Attachments = append(tasks[n.ItemID].Attachments, &models.TaskAttachment{
				File: &files.File{
					Name:    n.FileAttachment.FileName,
					Mime:    n.FileAttachment.FileType,
					Size:    uint64(n.FileAttachment.FileSize),
					Created: n.Posted,
					// We directly pass the file contents here to have a way to link the attachment to the file later.
					// Because we don't have an ID for our task at this point of the migration, we cannot just throw all
					// attachments in a slice and do the work of downloading and properly storing them later.
					FileContent: buf.Bytes(),
				},
				Created: n.Posted,
			})
		}
	}

	// Project Notes -> Project Descriptions
	for _, pn := range sync.ProjectNotes {
		if lists[pn.ProjectID].Description != "" {
			lists[pn.ProjectID].Description += "\n"
		}

		lists[pn.ProjectID].Description += pn.Content
	}

	// Reminders -> vikunja reminders
	for _, r := range sync.Reminders {
		if r.Due == nil {
			continue
		}

		if _, exists := tasks[r.ItemID]; !exists {
			log.Debugf("Could not find task %s for reminder %s while trying to resolve reminders", r.ItemID, r.ID)
			continue
		}

		date, err := parseDate(r.Due.Date)
		if err != nil {
			return nil, err
		}

		tasks[r.ItemID].Reminders = append(tasks[r.ItemID].Reminders, &models.TaskReminder{
			Reminder: date.In(config.GetTimeZone()),
		},
		)
	}

	return
}

func getAccessTokenFromAuthToken(authToken string) (accessToken string, err error) {

	form := url.Values{
		"client_id":     []string{config.MigrationTodoistClientID.GetString()},
		"client_secret": []string{config.MigrationTodoistClientSecret.GetString()},
		"code":          []string{authToken},
		"redirect_uri":  []string{config.MigrationTodoistRedirectURL.GetString()},
	}
	resp, err := migration.DoPost("https://todoist.com/oauth/access_token", form)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode > 399 {
		buf := &bytes.Buffer{}
		_, _ = buf.ReadFrom(resp.Body)
		return "", fmt.Errorf("got http status %d while trying to get token, error was %s", resp.StatusCode, buf.String())
	}

	token := &apiTokenResponse{}
	err = json.NewDecoder(resp.Body).Decode(token)
	return token.AccessToken, err
}

// Migrate gets all tasks from todoist for a user and puts them into vikunja
// @Summary Migrate all lists, tasks etc. from todoist
// @Description Migrates all projects, tasks, notes, reminders, subtasks and files from todoist to vikunja.
// @tags migration
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param migrationCode body todoist.Migration true "The auth code previously obtained from the auth url. See the docs for /migration/todoist/auth."
// @Success 200 {object} models.Message "A message telling you everything was migrated successfully."
// @Failure 500 {object} models.Message "Internal server error"
// @Router /migration/todoist/migrate [post]
func (m *Migration) Migrate(u *user.User) (err error) {

	log.Debugf("[Todoist Migration] Starting migration for user %d", u.ID)

	// 0. Get an api token from the obtained auth token
	token, err := getAccessTokenFromAuthToken(m.Code)
	if err != nil {
		return
	}

	if token == "" {
		log.Debugf("[Todoist Migration] Could not get token")
		return
	}

	log.Debugf("[Todoist Migration] Got user token for user %d", u.ID)
	log.Debugf("[Todoist Migration] Getting todoist data for user %d", u.ID)

	// Get everything with the sync api
	form := url.Values{
		"sync_token":     []string{"*"},
		"resource_types": []string{"[\"all\"]"},
	}
	bearerHeader := map[string]string{
		"Authorization": "Bearer " + token,
	}

	resp, err := migration.DoPostWithHeaders("https://api.todoist.com/sync/v9/sync", form, bearerHeader)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	syncResponse := &sync{}
	err = json.NewDecoder(resp.Body).Decode(syncResponse)
	if err != nil {
		return
	}

	log.Debugf("[Todoist Migration] Getting done items for user %d", u.ID)

	// Get all done tasks and projects
	offset := 0
	doneItems := make(map[string]*doneItem)

	for {
		resp, err = migration.DoPostWithHeaders("https://api.todoist.com/sync/v9/completed/get_all?limit="+strconv.Itoa(paginationLimit)+"&offset="+strconv.Itoa(offset*paginationLimit), form, bearerHeader)
		if err != nil {
			return
		}
		defer resp.Body.Close()

		completedSyncResponse := &doneItemSync{}
		err = json.NewDecoder(resp.Body).Decode(completedSyncResponse)
		if err != nil {
			return
		}

		sort.Slice(completedSyncResponse.Items, func(i, j int) bool {
			return completedSyncResponse.Items[i].CompletedDate.After(completedSyncResponse.Items[j].CompletedDate)
		})

		for _, i := range completedSyncResponse.Items {

			// Don't try to fetch task details from deleted projects (that will fail anyway)
			_, hasProject := completedSyncResponse.Projects[i.ProjectID]
			if hasProject && completedSyncResponse.Projects[i.ProjectID].IsDeleted {
				log.Debugf("[Todoist Migration] Not fetching task details from task %s because its project (%s) is deleted already.", i.TaskID, i.ProjectID)
				continue
			}

			if _, has := doneItems[i.TaskID]; has {
				// Only set the newest completion date
				continue
			}
			doneItems[i.TaskID] = i

			// need to get done item data
			resp, err = migration.DoPostWithHeaders("https://api.todoist.com/sync/v9/items/get", url.Values{
				"item_id": []string{i.TaskID},
			}, bearerHeader)
			if err != nil {
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusNotFound {
				// Done items of deleted projects may show up here but since the project is already deleted
				// we can't show them individually and the api returns a 404.
				buf := bytes.Buffer{}
				_, _ = buf.ReadFrom(resp.Body)
				log.Debugf("[Todoist Migration] Could not retrieve task details for task %s: %s", i.TaskID, buf.String())
				continue
			}

			doneI := &itemWrapper{}
			err = json.NewDecoder(resp.Body).Decode(doneI)
			if err != nil {
				return
			}
			log.Debugf("[Todoist Migration] Retrieved full task data for done task %s", i.TaskID)
			syncResponse.Items = append(syncResponse.Items, doneI.Item)
		}

		if len(completedSyncResponse.Items) < paginationLimit {
			break
		}
		offset++
		log.Debugf("[Todoist Migration] User %d has more than 200 done tasks or projects, looping to get more; iteration %d", u.ID, offset)
	}

	log.Debugf("[Todoist Migration] Got %d done items for user %d", len(doneItems), u.ID)
	log.Debugf("[Todoist Migration] Getting archived projects for user %d", u.ID)

	// Get all archived projects
	resp, err = migration.DoPostWithHeaders("https://api.todoist.com/sync/v9/projects/get_archived", form, bearerHeader)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	archivedProjects := []*project{}
	err = json.NewDecoder(resp.Body).Decode(&archivedProjects)
	if err != nil {
		return
	}
	syncResponse.Projects = append(syncResponse.Projects, archivedProjects...)

	log.Debugf("[Todoist Migration] Got %d archived projects for user %d", len(archivedProjects), u.ID)
	log.Debugf("[Todoist Migration] Getting data for archived projects for user %d", u.ID)

	// Project data is not included in the regular sync for archived projects, so we need to get all of those by hand
	for _, p := range archivedProjects {
		resp, err = migration.DoPostWithHeaders("https://api.todoist.com/sync/v9/projects/get_data?project_id="+p.ID, form, bearerHeader)
		if err != nil {
			return
		}

		archivedProjectData := &sync{}
		err = json.NewDecoder(resp.Body).Decode(&archivedProjectData)
		if err != nil {
			return
		}
		resp.Body.Close()

		syncResponse.Items = append(syncResponse.Items, archivedProjectData.Items...)
		syncResponse.Labels = append(syncResponse.Labels, archivedProjectData.Labels...)
		syncResponse.Notes = append(syncResponse.Notes, archivedProjectData.Notes...)
		syncResponse.ProjectNotes = append(syncResponse.ProjectNotes, archivedProjectData.ProjectNotes...)
		syncResponse.Reminders = append(syncResponse.Reminders, archivedProjectData.Reminders...)
		syncResponse.Sections = append(syncResponse.Sections, archivedProjectData.Sections...)
	}

	log.Debugf("[Todoist Migration] Got all todoist user data for user %d", u.ID)
	log.Debugf("[Todoist Migration] Start converting data for user %d", u.ID)

	fullVikunjaHierachie, err := convertTodoistToVikunja(syncResponse, doneItems)
	if err != nil {
		return
	}

	log.Debugf("[Todoist Migration] Done converting data for user %d", u.ID)
	log.Debugf("[Todoist Migration] Start inserting data for user %d", u.ID)

	err = migration.InsertFromStructure(fullVikunjaHierachie, u)
	if err != nil {
		return
	}

	log.Debugf("[Todoist Migration] Done inserting data for user %d", u.ID)
	log.Debugf("[Todoist Migration] Todoist migration done for user %d", u.ID)

	return nil
}
