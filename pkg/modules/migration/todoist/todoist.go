// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2020 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package todoist

import (
	"bytes"
	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/migration"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
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
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Color      int    `json:"color"`
	ItemOrder  int    `json:"item_order"`
	IsDeleted  int    `json:"is_deleted"`
	IsFavorite int    `json:"is_favorite"`
}

type project struct {
	ID             int    `json:"id"`
	LegacyID       int    `json:"legacy_id"`
	Name           string `json:"name"`
	Color          int    `json:"color"`
	ParentID       int    `json:"parent_id"`
	ChildOrder     int    `json:"child_order"`
	Collapsed      int    `json:"collapsed"`
	Shared         bool   `json:"shared"`
	LegacyParentID int    `json:"legacy_parent_id"`
	SyncID         int    `json:"sync_id"`
	IsDeleted      int    `json:"is_deleted"`
	IsArchived     int    `json:"is_archived"`
	IsFavorite     int    `json:"is_favorite"`
}

type dueDate struct {
	Date        string      `json:"date"`
	Timezone    interface{} `json:"timezone"`
	String      string      `json:"string"`
	Lang        string      `json:"lang"`
	IsRecurring bool        `json:"is_recurring"`
}

type item struct {
	ID              int         `json:"id"`
	LegacyID        int         `json:"legacy_id"`
	UserID          int         `json:"user_id"`
	ProjectID       int         `json:"project_id"`
	LegacyProjectID int         `json:"legacy_project_id"`
	Content         string      `json:"content"`
	Priority        int         `json:"priority"`
	Due             *dueDate    `json:"due"`
	ParentID        int         `json:"parent_id"`
	LegacyParentID  int         `json:"legacy_parent_id"`
	ChildOrder      int         `json:"child_order"`
	SectionID       int         `json:"section_id"`
	DayOrder        int         `json:"day_order"`
	Collapsed       int         `json:"collapsed"`
	Children        interface{} `json:"children"`
	Labels          []int       `json:"labels"`
	AddedByUID      int         `json:"added_by_uid"`
	AssignedByUID   int         `json:"assigned_by_uid"`
	ResponsibleUID  int         `json:"responsible_uid"`
	Checked         int         `json:"checked"`
	InHistory       int         `json:"in_history"`
	IsDeleted       int         `json:"is_deleted"`
	DateAdded       time.Time   `json:"date_added"`
	HasMoreNotes    bool        `json:"has_more_notes"`
	DateCompleted   time.Time   `json:"date_completed"`
}

type fileAttachment struct {
	FileType    string `json:"file_type"`
	FileName    string `json:"file_name"`
	FileSize    int    `json:"file_size"`
	FileURL     string `json:"file_url"`
	UploadState string `json:"upload_state"`
}

type note struct {
	ID              int             `json:"id"`
	LegacyID        int             `json:"legacy_id"`
	PostedUID       int             `json:"posted_uid"`
	ProjectID       int             `json:"project_id"`
	LegacyProjectID int             `json:"legacy_project_id"`
	ItemID          int             `json:"item_id"`
	LegacyItemID    int             `json:"legacy_item_id"`
	Content         string          `json:"content"`
	FileAttachment  *fileAttachment `json:"file_attachment"`
	UidsToNotify    []int           `json:"uids_to_notify"`
	IsDeleted       int             `json:"is_deleted"`
	Posted          time.Time       `json:"posted"`
}

type projectNote struct {
	Content        string          `json:"content"`
	FileAttachment *fileAttachment `json:"file_attachment"`
	ID             int64           `json:"id"`
	IsDeleted      int             `json:"is_deleted"`
	Posted         time.Time       `json:"posted"`
	PostedUID      int             `json:"posted_uid"`
	ProjectID      int             `json:"project_id"`
	UidsToNotify   []int           `json:"uids_to_notify"`
}

type reminder struct {
	ID        int      `json:"id"`
	NotifyUID int      `json:"notify_uid"`
	ItemID    int      `json:"item_id"`
	Service   string   `json:"service"`
	Type      string   `json:"type"`
	Due       *dueDate `json:"due"`
	MmOffset  int      `json:"mm_offset"`
	IsDeleted int      `json:"is_deleted"`
}

type sync struct {
	Projects     []*project     `json:"projects"`
	Items        []*item        `json:"items"`
	Labels       []*label       `json:"labels"`
	Notes        []*note        `json:"notes"`
	ProjectNotes []*projectNote `json:"project_notes"`
	Reminders    []*reminder    `json:"reminders"`
}

var todoistColors = map[int]string{}

func init() {
	todoistColors = make(map[int]string, 19)
	// The todoists colors are static, taken from https://developer.todoist.com/sync/v8/#colors
	todoistColors = map[int]string{
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

func doPost(url string, form url.Values) (resp *http.Response, err error) {
	req, err := http.NewRequest("POST", url, strings.NewReader(form.Encode()))
	if err != nil {
		return
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	hc := http.Client{}
	return hc.Do(req)
}

func convertTodoistToVikunja(sync *sync) (fullVikunjaHierachie []*models.NamespaceWithLists, err error) {

	newNamespace := &models.NamespaceWithLists{
		Namespace: models.Namespace{
			Title: "Migrated from todoist",
		},
	}

	// A map for all vikunja lists with the project id they're coming from as key
	lists := make(map[int]*models.List, len(sync.Projects))

	// A map for all vikunja tasks with the todoist task id as key to find them easily and add more data
	tasks := make(map[int]*models.Task, len(sync.Items))

	// A map for all vikunja labels with the todoist id as key to find them easier
	labels := make(map[int]*models.Label, len(sync.Labels))

	for _, p := range sync.Projects {
		list := &models.List{
			Title:      p.Name,
			HexColor:   todoistColors[p.Color],
			IsArchived: p.IsArchived == 1,
		}

		lists[p.ID] = list

		newNamespace.Lists = append(newNamespace.Lists, list)
	}

	for _, label := range sync.Labels {
		labels[label.ID] = &models.Label{
			Title:    label.Name,
			HexColor: todoistColors[label.Color],
		}
	}

	for _, i := range sync.Items {
		task := &models.Task{
			Title:   i.Content,
			Created: i.DateAdded.In(config.GetTimeZone()),
			Done:    i.Checked == 1,
		}

		// Only try to parse the task done at date if the task is actually done
		// Sometimes weired things happen if we try to parse nil dates.
		if task.Done {
			task.DoneAt = i.DateCompleted.In(config.GetTimeZone())
		}

		// Todoist priorities only range from 1 (lowest) and max 4 (highest), so we need to make slight adjustments
		if i.Priority > 1 {
			task.Priority = int64(i.Priority)
		}

		// Put the due date together
		if i.Due != nil {
			dueDate, err := time.Parse("2006-01-02", i.Due.Date)
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

		// Prevent all those nil errors
		if tasks[i.ParentID].RelatedTasks == nil {
			tasks[i.ParentID].RelatedTasks = make(models.RelatedTaskMap)
		}

		tasks[i.ParentID].RelatedTasks[models.RelationKindSubtask] = append(tasks[i.ParentID].RelatedTasks[models.RelationKindSubtask], tasks[i.ID])
	}

	// Task Notes -> Task Descriptions
	// FIXME: Should be comments
	for _, n := range sync.Notes {
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
			resp, err := http.Get(n.FileAttachment.FileURL)
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()
			buf := &bytes.Buffer{}
			_, err = buf.ReadFrom(resp.Body)
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

		var err error
		var date time.Time
		date, err = time.Parse("2006-01-02T15:04:05Z", r.Due.Date)
		if err != nil {
			date, err = time.Parse("2006-01-02T15:04:05", r.Due.Date)
		}
		if err != nil {
			date, err = time.Parse("2006-01-02", r.Due.Date)
		}
		if err != nil {
			return nil, err
		}

		tasks[r.ItemID].Reminders = append(tasks[r.ItemID].Reminders, date.In(config.GetTimeZone()))
	}

	return []*models.NamespaceWithLists{
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
	resp, err := doPost("https://todoist.com/oauth/access_token", form)
	if err != nil {
		return
	}

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
	resp, err := doPost("https://api.todoist.com/sync/v8/sync", form)
	if err != nil {
		return
	}

	syncResponse := &sync{}
	err = json.NewDecoder(resp.Body).Decode(syncResponse)
	if err != nil {
		return
	}

	log.Debugf("[Todoist Migration] Got all todoist user data for user %d", u.ID)
	log.Debugf("[Todoist Migration] Start converting data for user %d", u.ID)

	fullVikunjaHierachie, err := convertTodoistToVikunja(syncResponse)
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
