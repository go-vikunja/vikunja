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

package todoist

import (
	"bytes"
	"encoding/json"
	"fmt"
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

// Migration is the todoist migration struct
type Migration struct {
	Code string `json:"code"`
}

type apiTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

type label struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Color      int64  `json:"color"`
	ItemOrder  int64  `json:"item_order"`
	IsDeleted  int64  `json:"is_deleted"`
	IsFavorite int64  `json:"is_favorite"`
}

type project struct {
	ID             int64  `json:"id"`
	LegacyID       int64  `json:"legacy_id"`
	Name           string `json:"name"`
	Color          int64  `json:"color"`
	ParentID       int64  `json:"parent_id"`
	ChildOrder     int64  `json:"child_order"`
	Collapsed      int64  `json:"collapsed"`
	Shared         bool   `json:"shared"`
	LegacyParentID int64  `json:"legacy_parent_id"`
	SyncID         int64  `json:"sync_id"`
	IsDeleted      int64  `json:"is_deleted"`
	IsArchived     int64  `json:"is_archived"`
	IsFavorite     int64  `json:"is_favorite"`
}

type dueDate struct {
	Date        string      `json:"date"`
	Timezone    interface{} `json:"timezone"`
	String      string      `json:"string"`
	Lang        string      `json:"lang"`
	IsRecurring bool        `json:"is_recurring"`
}

type item struct {
	ID              int64       `json:"id"`
	LegacyID        int64       `json:"legacy_id"`
	UserID          int64       `json:"user_id"`
	ProjectID       int64       `json:"project_id"`
	LegacyProjectID int64       `json:"legacy_project_id"`
	Content         string      `json:"content"`
	Priority        int64       `json:"priority"`
	Due             *dueDate    `json:"due"`
	ParentID        int64       `json:"parent_id"`
	LegacyParentID  int64       `json:"legacy_parent_id"`
	ChildOrder      int64       `json:"child_order"`
	SectionID       int64       `json:"section_id"`
	DayOrder        int64       `json:"day_order"`
	Collapsed       int64       `json:"collapsed"`
	Children        interface{} `json:"children"`
	Labels          []int64     `json:"labels"`
	AddedByUID      int64       `json:"added_by_uid"`
	AssignedByUID   int64       `json:"assigned_by_uid"`
	ResponsibleUID  int64       `json:"responsible_uid"`
	Checked         int64       `json:"checked"`
	InHistory       int64       `json:"in_history"`
	IsDeleted       int64       `json:"is_deleted"`
	DateAdded       time.Time   `json:"date_added"`
	HasMoreNotes    bool        `json:"has_more_notes"`
	DateCompleted   time.Time   `json:"date_completed"`
}

type doneItem struct {
	CompletedDate time.Time `json:"completed_date"`
	Content       string    `json:"content"`
	ID            int64     `json:"id"`
	ProjectID     int64     `json:"project_id"`
	TaskID        int64     `json:"task_id"`
	UserID        int       `json:"user_id"`
}

type doneItemSync struct {
	Items []*doneItem `json:"items"`
}

type fileAttachment struct {
	FileType    string `json:"file_type"`
	FileName    string `json:"file_name"`
	FileSize    int64  `json:"file_size"`
	FileURL     string `json:"file_url"`
	UploadState string `json:"upload_state"`
}

type note struct {
	ID              int64           `json:"id"`
	LegacyID        int64           `json:"legacy_id"`
	PostedUID       int64           `json:"posted_uid"`
	ProjectID       int64           `json:"project_id"`
	LegacyProjectID int64           `json:"legacy_project_id"`
	ItemID          int64           `json:"item_id"`
	LegacyItemID    int64           `json:"legacy_item_id"`
	Content         string          `json:"content"`
	FileAttachment  *fileAttachment `json:"file_attachment"`
	UidsToNotify    []int64         `json:"uids_to_notify"`
	IsDeleted       int64           `json:"is_deleted"`
	Posted          time.Time       `json:"posted"`
}

type projectNote struct {
	Content        string          `json:"content"`
	FileAttachment *fileAttachment `json:"file_attachment"`
	ID             int64           `json:"id"`
	IsDeleted      int64           `json:"is_deleted"`
	Posted         time.Time       `json:"posted"`
	PostedUID      int64           `json:"posted_uid"`
	ProjectID      int64           `json:"project_id"`
	UidsToNotify   []int64         `json:"uids_to_notify"`
}

type reminder struct {
	ID        int64    `json:"id"`
	NotifyUID int64    `json:"notify_uid"`
	ItemID    int64    `json:"item_id"`
	Service   string   `json:"service"`
	Type      string   `json:"type"`
	Due       *dueDate `json:"due"`
	MmOffset  int64    `json:"mm_offset"`
	IsDeleted int64    `json:"is_deleted"`
}

type section struct {
	ID           int64     `json:"id"`
	DateAdded    time.Time `json:"date_added"`
	IsDeleted    bool      `json:"is_deleted"`
	Name         string    `json:"name"`
	ProjectID    int64     `json:"project_id"`
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

var todoistColors = map[int64]string{}

func init() {
	todoistColors = make(map[int64]string, 19)
	// The todoists colors are static, taken from https://developer.todoist.com/sync/v8/#colors
	todoistColors = map[int64]string{
		30: "b8256f",
		31: "db4035",
		32: "ff9933",
		33: "fad000",
		34: "afb83b",
		35: "7ecc49",
		36: "299438",
		37: "6accbc",
		38: "158fad",
		39: "14aaf5",
		40: "96c3eb",
		41: "4073ff",
		42: "884dff",
		43: "af38eb",
		44: "eb96eb",
		45: "e05194",
		46: "ff8d85",
		47: "808080",
		48: "b8b8b8",
		49: "ccac93",
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
	return "https://todoist.com/oauth/authorize" +
		"?client_id=" + config.MigrationTodoistClientID.GetString() +
		"&scope=data:read" +
		"&state=" + utils.MakeRandomString(32)
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

func convertTodoistToVikunja(sync *sync, doneItems map[int64]*doneItem) (fullVikunjaHierachie []*models.NamespaceWithListsAndTasks, err error) {

	newNamespace := &models.NamespaceWithListsAndTasks{
		Namespace: models.Namespace{
			Title: "Migrated from todoist",
		},
	}

	// A map for all vikunja lists with the project id they're coming from as key
	lists := make(map[int64]*models.ListWithTasksAndBuckets, len(sync.Projects))

	// A map for all vikunja tasks with the todoist task id as key to find them easily and add more data
	tasks := make(map[int64]*models.TaskWithComments, len(sync.Items))

	// A map for all vikunja labels with the todoist id as key to find them easier
	labels := make(map[int64]*models.Label, len(sync.Labels))

	for _, p := range sync.Projects {
		list := &models.ListWithTasksAndBuckets{
			List: models.List{
				Title:      p.Name,
				HexColor:   todoistColors[p.Color],
				IsArchived: p.IsArchived == 1,
			},
		}

		lists[p.ID] = list

		newNamespace.Lists = append(newNamespace.Lists, list)
	}

	sort.Slice(sync.Sections, func(i, j int) bool {
		return sync.Sections[i].SectionOrder < sync.Sections[j].SectionOrder
	})

	for _, section := range sync.Sections {
		if section.IsDeleted || section.ProjectID == 0 {
			continue
		}

		lists[section.ProjectID].Buckets = append(lists[section.ProjectID].Buckets, &models.Bucket{
			ID:      section.ID,
			Title:   section.Name,
			Created: section.DateAdded,
		})
	}

	for _, label := range sync.Labels {
		labels[label.ID] = &models.Label{
			Title:    label.Name,
			HexColor: todoistColors[label.Color],
		}
	}

	for _, i := range sync.Items {
		task := &models.TaskWithComments{
			Task: models.Task{
				Title:    i.Content,
				Created:  i.DateAdded.In(config.GetTimeZone()),
				Done:     i.Checked == 1,
				BucketID: i.SectionID,
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
		for _, lID := range i.Labels {
			task.Labels = append(task.Labels, labels[lID])
		}

		tasks[i.ID] = task

		lists[i.ProjectID].Tasks = append(lists[i.ProjectID].Tasks, task)
	}

	// If the parenId of a task is not 0, create a task relation
	// We're looping again here to make sure we have seem all tasks before and have them in our map
	for _, i := range sync.Items {
		if i.ParentID == 0 {
			continue
		}

		if _, exists := tasks[i.ParentID]; !exists {
			log.Debugf("[Todoist Migration] Could not find task %d in tasks map while trying to get resolve subtasks for task %d", i.ParentID, i.ID)
			continue
		}

		// Prevent all those nil errors
		if tasks[i.ParentID].RelatedTasks == nil {
			tasks[i.ParentID].RelatedTasks = make(models.RelatedTaskMap)
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
			log.Debugf("[Todoist Migration] Could not find task %d for note %d", n.ItemID, n.ID)
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

	// Project Notes -> List Descriptions
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
			log.Debugf("Could not find task %d for reminder %d while trying to resolve reminders", r.ItemID, r.ID)
			continue
		}

		date, err := parseDate(r.Due.Date)
		if err != nil {
			return nil, err
		}

		tasks[r.ItemID].Reminders = append(tasks[r.ItemID].Reminders, date.In(config.GetTimeZone()))
	}

	return []*models.NamespaceWithListsAndTasks{
		newNamespace,
	}, err
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
		"token":          []string{token},
		"sync_token":     []string{"*"},
		"resource_types": []string{"[\"all\"]"},
	}
	resp, err := migration.DoPost("https://api.todoist.com/sync/v8/sync", form)
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
	// Get all done tasks
	resp, err = migration.DoPost("https://api.todoist.com/sync/v8/completed/get_all", form)
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

	doneItems := make(map[int64]*doneItem, len(completedSyncResponse.Items))
	for _, i := range completedSyncResponse.Items {
		if _, has := doneItems[i.TaskID]; has {
			// Only set the newest completion date
			continue
		}
		doneItems[i.TaskID] = i
	}

	log.Debugf("[Todoist Migration] Got %d done items for user %d", len(completedSyncResponse.Items), u.ID)
	log.Debugf("[Todoist Migration] Getting archived projects for user %d", u.ID)

	// Get all archived projects
	resp, err = migration.DoPost("https://api.todoist.com/sync/v8/projects/get_archived", form)
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

	// Project data is not included in the regular sync for archived projects so we need to get all of those by hand
	//https://api.todoist.com/sync/v8/projects/get_data\?project_id\=2269005399
	for _, p := range archivedProjects {
		resp, err = migration.DoPost("https://api.todoist.com/sync/v8/projects/get_data?project_id="+strconv.FormatInt(p.ID, 10), form)
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
