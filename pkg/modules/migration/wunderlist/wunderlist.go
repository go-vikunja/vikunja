// Vikunja is a todo-list application to facilitate your life.
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

package wunderlist

import (
	"bytes"
	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/migration"
	"code.vikunja.io/api/pkg/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// Migration represents the implementation of the migration for wunderlist
type Migration struct {
	// Code is the code used to get a user api token
	Code string `query:"code" json:"code"`
}

// This represents all necessary fields for getting an api token for the wunderlist api from a code
type wunderlistAuthRequest struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Code         string `json:"code"`
}

type wunderlistAuthToken struct {
	AccessToken string `json:"access_token"`
}

type task struct {
	ID          int       `json:"id"`
	AssigneeID  int       `json:"assignee_id"`
	CreatedAt   time.Time `json:"created_at"`
	CreatedByID int       `json:"created_by_id"`
	DueDate     string    `json:"due_date"`
	ListID      int       `json:"list_id"`
	Revision    int       `json:"revision"`
	Starred     bool      `json:"starred"`
	Title       string    `json:"title"`
	Completed   bool      `json:"completed"`
	CompletedAt time.Time `json:"completed_at"`
}

type list struct {
	ID        int       `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Title     string    `json:"title"`
	ListType  string    `json:"list_type"`
	Type      string    `json:"type"`
	Revision  int       `json:"revision"`

	Migrated bool `json:"-"`
}

type folder struct {
	ID                 int       `json:"id"`
	Title              string    `json:"title"`
	ListIds            []int     `json:"list_ids"`
	CreatedAt          time.Time `json:"created_at"`
	CreatedByRequestID string    `json:"created_by_request_id"`
	UpdatedAt          time.Time `json:"updated_at"`
	Type               string    `json:"type"`
	Revision           int       `json:"revision"`
}

type note struct {
	ID        int       `json:"id"`
	TaskID    int       `json:"task_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Revision  int       `json:"revision"`
}

type file struct {
	ID             int       `json:"id"`
	URL            string    `json:"url"`
	TaskID         int       `json:"task_id"`
	ListID         int       `json:"list_id"`
	UserID         int       `json:"user_id"`
	FileName       string    `json:"file_name"`
	ContentType    string    `json:"content_type"`
	FileSize       int       `json:"file_size"`
	LocalCreatedAt time.Time `json:"local_created_at"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	Type           string    `json:"type"`
	Revision       int       `json:"revision"`
}

type reminder struct {
	ID        int       `json:"id"`
	Date      time.Time `json:"date"`
	TaskID    int       `json:"task_id"`
	Revision  int       `json:"revision"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type subtask struct {
	ID          int       `json:"id"`
	TaskID      int       `json:"task_id"`
	CreatedAt   time.Time `json:"created_at"`
	CreatedByID int       `json:"created_by_id"`
	Revision    int       `json:"revision"`
	Title       string    `json:"title"`
}

type wunderlistContents struct {
	tasks     []*task
	lists     []*list
	folders   []*folder
	notes     []*note
	files     []*file
	reminders []*reminder
	subtasks  []*subtask
}

func convertListForFolder(listID int, list *list, content *wunderlistContents) (*models.List, error) {

	l := &models.List{
		Title:   list.Title,
		Created: list.CreatedAt.Unix(),
	}

	// Find all tasks belonging to this list and put them in
	for _, t := range content.tasks {
		if t.ListID == listID {
			newTask := &models.Task{
				Text:    t.Title,
				Created: t.CreatedAt.Unix(),
				Done:    t.Completed,
			}

			// Set Done At
			if newTask.Done {
				newTask.DoneAtUnix = t.CompletedAt.Unix()
			}

			// Parse the due date
			if t.DueDate != "" {
				dueDate, err := time.Parse("2006-01-02", t.DueDate)
				if err != nil {
					return nil, err
				}
				newTask.DueDateUnix = dueDate.Unix()
			}

			// Find related notes
			for _, n := range content.notes {
				if n.TaskID == t.ID {
					newTask.Description = n.Content
				}
			}

			// Attachments
			for _, f := range content.files {
				if f.TaskID == t.ID {
					// Download the attachment and put it in the file
					resp, err := http.Get(f.URL)
					if err != nil {
						return nil, err
					}
					defer resp.Body.Close()
					buf := &bytes.Buffer{}
					_, err = buf.ReadFrom(resp.Body)
					if err != nil {
						return nil, err
					}

					newTask.Attachments = append(newTask.Attachments, &models.TaskAttachment{
						File: &files.File{
							Name:        f.FileName,
							Mime:        f.ContentType,
							Size:        uint64(f.FileSize),
							Created:     f.CreatedAt,
							CreatedUnix: f.CreatedAt.Unix(),
							// We directly pass the file contents here to have a way to link the attachment to the file later.
							// Because we don't have an ID for our task at this point of the migration, we cannot just throw all
							// attachments in a slice and do the work of downloading and properly storing them later.
							FileContent: buf.Bytes(),
						},
						Created: f.CreatedAt.Unix(),
					})
				}
			}

			// Subtasks
			for _, s := range content.subtasks {
				if s.TaskID == t.ID {
					if newTask.RelatedTasks[models.RelationKindSubtask] == nil {
						newTask.RelatedTasks = make(models.RelatedTaskMap)
					}
					newTask.RelatedTasks[models.RelationKindSubtask] = append(newTask.RelatedTasks[models.RelationKindSubtask], &models.Task{
						Text: s.Title,
					})
				}
			}

			// Reminders
			for _, r := range content.reminders {
				if r.TaskID == t.ID {
					newTask.RemindersUnix = append(newTask.RemindersUnix, r.Date.Unix())
				}
			}

			l.Tasks = append(l.Tasks, newTask)
		}
	}
	return l, nil
}

func convertWunderlistToVikunja(content *wunderlistContents) (fullVikunjaHierachie []*models.NamespaceWithLists, err error) {

	// Make a map from the list with the key being list id for easier handling
	listMap := make(map[int]*list, len(content.lists))
	for _, l := range content.lists {
		listMap[l.ID] = l
	}

	// First, we look through all folders and create namespaces for them.
	for _, folder := range content.folders {
		namespace := &models.NamespaceWithLists{
			Namespace: models.Namespace{
				Name:    folder.Title,
				Created: folder.CreatedAt.Unix(),
				Updated: folder.UpdatedAt.Unix(),
			},
		}

		// Then find all lists for that folder
		for _, listID := range folder.ListIds {
			if list, exists := listMap[listID]; exists {
				l, err := convertListForFolder(listID, list, content)
				if err != nil {
					return nil, err
				}
				namespace.Lists = append(namespace.Lists, l)
				// And mark the list as migrated so we don't iterate over it again
				list.Migrated = true
			}
		}

		// And then finally put the namespace (which now has all the details) back in the full array.
		fullVikunjaHierachie = append(fullVikunjaHierachie, namespace)
	}

	// At the end, loop over all lists which don't belong to a namespace and put them in a default namespace
	if len(listMap) > 0 {
		newNamespace := &models.NamespaceWithLists{
			Namespace: models.Namespace{
				Name: "Migrated from wunderlist",
			},
		}

		for _, list := range listMap {

			if list.Migrated {
				continue
			}

			l, err := convertListForFolder(list.ID, list, content)
			if err != nil {
				return nil, err
			}
			newNamespace.Lists = append(newNamespace.Lists, l)
		}

		fullVikunjaHierachie = append(fullVikunjaHierachie, newNamespace)
	}

	return
}

func makeAuthGetRequest(token *wunderlistAuthToken, urlPart string, v interface{}, urlParams url.Values) error {
	req, err := http.NewRequest(http.MethodGet, "https://a.wunderlist.com/api/v1/"+urlPart, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Access-Token", token.AccessToken)
	req.Header.Set("X-Client-ID", config.MigrationWunderlistClientID.GetString())
	req.URL.RawQuery = urlParams.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	buf := &bytes.Buffer{}
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode > 399 {
		return fmt.Errorf("wunderlist API Error: Status Code: %d, Response was: %s", resp.StatusCode, buf.String())
	}

	// If the response is an empty json array, we need to exit here, otherwise this breaks the json parser since it
	// expects a null for an empty slice
	str := buf.String()
	if str == "[]" {
		return nil
	}

	return json.Unmarshal(buf.Bytes(), v)
}

// Migrate migrates a user's wunderlist lists, tasks, etc.
// @Summary Migrate all lists, tasks etc. from wunderlist
// @Description Migrates all folders, lists, tasks, notes, reminders, subtasks and files from wunderlist to vikunja.
// @tags migration
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param migrationCode body wunderlist.Migration true "The auth code previously obtained from the auth url. See the docs for /migration/wunderlist/auth."
// @Success 200 {object} models.Message "A message telling you everything was migrated successfully."
// @Failure 500 {object} models.Message "Internal server error"
// @Router /migration/wunderlist/migrate [post]
func (w *Migration) Migrate(user *models.User) (err error) {

	log.Debugf("[Wunderlist migration] Starting wunderlist migration for user %d", user.ID)

	// Struct init
	wContent := &wunderlistContents{
		tasks:     []*task{},
		lists:     []*list{},
		folders:   []*folder{},
		notes:     []*note{},
		files:     []*file{},
		reminders: []*reminder{},
		subtasks:  []*subtask{},
	}

	// 0. Get api token from oauth user token
	authRequest := wunderlistAuthRequest{
		ClientID:     config.MigrationWunderlistClientID.GetString(),
		ClientSecret: config.MigrationWunderlistClientSecret.GetString(),
		Code:         w.Code,
	}
	jsonAuth, err := json.Marshal(authRequest)
	if err != nil {
		return
	}
	resp, err := http.Post("https://www.wunderlist.com/oauth/access_token", "application/json", bytes.NewBuffer(jsonAuth))
	if err != nil {
		return
	}

	authToken := &wunderlistAuthToken{}
	err = json.NewDecoder(resp.Body).Decode(authToken)
	if err != nil {
		return
	}

	log.Debugf("[Wunderlist migration] Start getting all data from wunderlist for user %d", user.ID)

	// 1. Get all folders
	err = makeAuthGetRequest(authToken, "folders", &wContent.folders, nil)
	if err != nil {
		return
	}

	// 2. Get all lists
	err = makeAuthGetRequest(authToken, "lists", &wContent.lists, nil)
	if err != nil {
		return
	}

	for _, l := range wContent.lists {

		listQueryParam := url.Values{"list_id": []string{strconv.Itoa(l.ID)}}

		// 3. Get all tasks for each list
		tasks := []*task{}
		err = makeAuthGetRequest(authToken, "tasks", &tasks, listQueryParam)
		if err != nil {
			return
		}
		wContent.tasks = append(wContent.tasks, tasks...)

		// 3. Get all done tasks for each list
		doneTasks := []*task{}
		err = makeAuthGetRequest(authToken, "tasks", &doneTasks, url.Values{"list_id": []string{strconv.Itoa(l.ID)}, "completed": []string{"true"}})
		if err != nil {
			return
		}
		wContent.tasks = append(wContent.tasks, doneTasks...)

		// 4. Get all notes for all lists
		notes := []*note{}
		err = makeAuthGetRequest(authToken, "notes", &notes, listQueryParam)
		if err != nil {
			return
		}
		wContent.notes = append(wContent.notes, notes...)

		// 5. Get all files for all lists
		fils := []*file{}
		err = makeAuthGetRequest(authToken, "files", &fils, listQueryParam)
		if err != nil {
			return
		}
		wContent.files = append(wContent.files, fils...)

		// 6. Get all reminders for all lists
		reminders := []*reminder{}
		err = makeAuthGetRequest(authToken, "reminders", &reminders, listQueryParam)
		if err != nil {
			return
		}
		wContent.reminders = append(wContent.reminders, reminders...)

		// 7. Get all subtasks for all lists
		subtasks := []*subtask{}
		err = makeAuthGetRequest(authToken, "subtasks", &subtasks, listQueryParam)
		if err != nil {
			return
		}
		wContent.subtasks = append(wContent.subtasks, subtasks...)
	}

	log.Debugf("[Wunderlist migration] Got all data from wunderlist for user %d", user.ID)
	log.Debugf("[Wunderlist migration] Migrating data to vikunja format for user %d", user.ID)

	// Convert + Insert everything
	fullVikunjaHierachie, err := convertWunderlistToVikunja(wContent)
	if err != nil {
		return
	}

	log.Debugf("[Wunderlist migration] Done migrating data to vikunja format for user %d", user.ID)
	log.Debugf("[Wunderlist migration] Insert data into db for user %d", user.ID)

	err = migration.InsertFromStructure(fullVikunjaHierachie, user)

	log.Debugf("[Wunderlist migration] Done inserting data into db for user %d", user.ID)
	log.Debugf("[Wunderlist migration] Wunderlist migration for user %d done", user.ID)

	return err
}

// AuthURL returns the url users need to authenticate against
// @Summary Get the auth url from wunderlist
// @Description Returns the auth url where the user needs to get its auth code. This code can then be used to migrate everything from wunderlist to Vikunja.
// @tags migration
// @Produce json
// @Security JWTKeyAuth
// @Success 200 {object} handler.AuthURL "The auth url."
// @Failure 500 {object} models.Message "Internal server error"
// @Router /migration/wunderlist/auth [get]
func (w *Migration) AuthURL() string {
	return "https://www.wunderlist.com/oauth/authorize?client_id=" +
		config.MigrationWunderlistClientID.GetString() +
		"&redirect_uri=" +
		config.MigrationWunderlistRedirectURL.GetString() +
		"&state=" + utils.MakeRandomString(32)
}
