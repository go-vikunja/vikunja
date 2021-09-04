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

package microsofttodo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/migration"
	"code.vikunja.io/api/pkg/user"
)

const apiScopes = `tasks.read tasks.read.shared`

type Migration struct {
	Code string `json:"code"`
}

type apiTokenResponse struct {
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	ExpiresIn    int    `json:"expires_in"`
	ExtExpiresIn int    `json:"ext_expires_in"`
	AccessToken  string `json:"access_token"`
}

type task struct {
	OdataEtag            string            `json:"@odata.etag"`
	Importance           string            `json:"importance"`
	IsReminderOn         bool              `json:"isReminderOn"`
	Status               string            `json:"status"`
	Title                string            `json:"title"`
	CreatedDateTime      time.Time         `json:"createdDateTime"`
	LastModifiedDateTime time.Time         `json:"lastModifiedDateTime"`
	ID                   string            `json:"id"`
	Body                 *body             `json:"body"`
	DueDateTime          *dateTimeTimeZone `json:"dueDateTime"`
	Recurrence           *recurrence       `json:"recurrence"`
	ReminderDateTime     *dateTimeTimeZone `json:"reminderDateTime"`
	CompletedDateTime    *dateTimeTimeZone `json:"completedDateTime"`
}
type dateTimeTimeZone struct {
	DateTime string `json:"dateTime"`
	TimeZone string `json:"timeZone"`
}
type body struct {
	Content     string `json:"content"`
	ContentType string `json:"contentType"`
}
type pattern struct {
	Type           string   `json:"type"`
	Interval       int64    `json:"interval"`
	Month          int64    `json:"month"`
	DayOfMonth     int64    `json:"dayOfMonth"`
	DaysOfWeek     []string `json:"daysOfWeek"`
	FirstDayOfWeek string   `json:"firstDayOfWeek"`
	Index          string   `json:"index"`
}
type taskRange struct {
	Type                string `json:"type"`
	StartDate           string `json:"startDate"`
	EndDate             string `json:"endDate"`
	RecurrenceTimeZone  string `json:"recurrenceTimeZone"`
	NumberOfOccurrences int    `json:"numberOfOccurrences"`
}
type recurrence struct {
	Pattern *pattern   `json:"pattern"`
	Range   *taskRange `json:"range"`
}

type tasksResponse struct {
	OdataContext string  `json:"@odata.context"`
	Value        []*task `json:"value"`
}

type list struct {
	ID                string  `json:"id"`
	OdataEtag         string  `json:"@odata.etag"`
	DisplayName       string  `json:"displayName"`
	IsOwner           bool    `json:"isOwner"`
	IsShared          bool    `json:"isShared"`
	WellknownListName string  `json:"wellknownListName"`
	Tasks             []*task `json:"-"` // This field does not exist in the api, we're just using it to return a structure with everything at once
}

type listsResponse struct {
	OdataContext string  `json:"@odata.context"`
	Value        []*list `json:"value"`
}

func (dtt *dateTimeTimeZone) toTime() (t time.Time, err error) {
	loc, err := time.LoadLocation(dtt.TimeZone)
	if err != nil {
		return t, err
	}

	return time.ParseInLocation(time.RFC3339Nano, dtt.DateTime+"Z", loc)
}

// AuthURL returns the url users need to authenticate against
// @Summary Get the auth url from Microsoft Todo
// @Description Returns the auth url where the user needs to get its auth code. This code can then be used to migrate everything from Microsoft Todo to Vikunja.
// @tags migration
// @Produce json
// @Security JWTKeyAuth
// @Success 200 {object} handler.AuthURL "The auth url."
// @Failure 500 {object} models.Message "Internal server error"
// @Router /migration/microsoft-todo/auth [get]
func (m *Migration) AuthURL() string {
	return "https://login.microsoftonline.com/common/oauth2/v2.0/authorize" +
		"?client_id=" + config.MigrationMicrosoftTodoClientID.GetString() +
		"&response_type=code" +
		"&redirect_uri=" + config.MigrationMicrosoftTodoRedirectURL.GetString() +
		"&response_mode=query" +
		"&scope=" + apiScopes
}

// Name is used to get the name of the Microsoft Todo migration - we're using the docs here to annotate the status route.
// @Summary Get migration status
// @Description Returns if the current user already did the migation or not. This is useful to show a confirmation message in the frontend if the user is trying to do the same migration again.
// @tags migration
// @Produce json
// @Security JWTKeyAuth
// @Success 200 {object} migration.Status "The migration status"
// @Failure 500 {object} models.Message "Internal server error"
// @Router /migration/microsoft-todo/status [get]
func (m *Migration) Name() string {
	return "microsoft-todo"
}

func getMicrosoftGraphAuthToken(code string) (accessToken string, err error) {

	form := url.Values{
		"client_id":     []string{config.MigrationMicrosoftTodoClientID.GetString()},
		"client_secret": []string{config.MigrationMicrosoftTodoClientSecret.GetString()},
		"scope":         []string{apiScopes},
		"code":          []string{code},
		"redirect_uri":  []string{config.MigrationMicrosoftTodoRedirectURL.GetString()},
		"grant_type":    []string{"authorization_code"},
	}
	resp, err := migration.DoPost("https://login.microsoftonline.com/common/oauth2/v2.0/token", form)
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

func makeAuthenticatedGetRequest(token, urlPart string, v interface{}) error {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://graph.microsoft.com/v1.0/me/todo/"+urlPart, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)

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
		return fmt.Errorf("Microsoft Graph API Error: Status Code: %d, Response was: %s", resp.StatusCode, buf.String())
	}

	// If the response is an empty json array, we need to exit here, otherwise this breaks the json parser since it
	// expects a null for an empty slice
	str := buf.String()
	if str == "[]" {
		return nil
	}

	return json.Unmarshal(buf.Bytes(), v)
}

func getMicrosoftTodoData(token string) (microsoftTodoData []*list, err error) {

	microsoftTodoData = []*list{}

	lists := &listsResponse{}
	err = makeAuthenticatedGetRequest(token, "lists", lists)
	if err != nil {
		log.Errorf("[Microsoft Todo Migration] Could not get lists: %s", err)
		return
	}

	log.Debugf("[Microsoft Todo Migration] Got %d lists", len(lists.Value))

	for _, list := range lists.Value {
		tasksResponse := &tasksResponse{}
		err = makeAuthenticatedGetRequest(token, "lists/"+list.ID+"/tasks", tasksResponse)
		if err != nil {
			log.Errorf("[Microsoft Todo Migration] Could not get tasks for list %s: %s", list.ID, err)
			return
		}

		log.Debugf("[Microsoft Todo Migration] Got %d tasks for list %s", len(tasksResponse.Value), list.ID)

		list.Tasks = tasksResponse.Value

		microsoftTodoData = append(microsoftTodoData, list)
	}

	log.Debugf("[Microsoft Todo Migration] Got all tasks for %d lists", len(lists.Value))

	return
}

func convertMicrosoftTodoData(todoData []*list) (vikunjsStructure []*models.NamespaceWithListsAndTasks, err error) {

	// One namespace with all lists
	vikunjsStructure = []*models.NamespaceWithListsAndTasks{
		{
			Namespace: models.Namespace{
				Title: "Migrated from Microsoft Todo",
			},
			Lists: []*models.ListWithTasksAndBuckets{},
		},
	}

	log.Debugf("[Microsoft Todo Migration] Converting %d lists", len(todoData))

	for _, l := range todoData {

		log.Debugf("[Microsoft Todo Migration] Converting list %s", l.ID)

		// Lists only with title
		list := &models.ListWithTasksAndBuckets{
			List: models.List{
				Title: l.DisplayName,
			},
		}

		log.Debugf("[Microsoft Todo Migration] Converting %d tasks", len(l.Tasks))

		for _, t := range l.Tasks {

			log.Debugf("[Microsoft Todo Migration] Converting task %s", t.ID)

			task := &models.Task{
				Title: t.Title,
				Done:  t.Status == "completed",
			}

			// Done Status
			if task.Done {
				log.Debugf("[Microsoft Todo Migration] Converting done at for task %s", t.ID)
				task.DoneAt, err = t.CompletedDateTime.toTime()
				if err != nil {
					return
				}
			}

			// Description
			if t.Body != nil && t.Body.ContentType == "text" {
				task.Description = t.Body.Content
			}

			// Priority
			switch t.Importance {
			case "low":
				task.Priority = 1
			case "normal":
				task.Priority = 2
			case "high":
				task.Priority = 3
			default:
				task.Priority = 0
			}

			// Reminders
			if t.ReminderDateTime != nil {
				log.Debugf("[Microsoft Todo Migration] Converting reminder for task %s", t.ID)
				reminder, err := t.ReminderDateTime.toTime()
				if err != nil {
					return nil, err
				}

				task.Reminders = []time.Time{reminder}
			}

			// Due Date
			if t.DueDateTime != nil {
				log.Debugf("[Microsoft Todo Migration] Converting due date for task %s", t.ID)
				dueDate, err := t.DueDateTime.toTime()
				if err != nil {
					return nil, err
				}

				task.DueDate = dueDate
			}

			// Repeating
			if t.Recurrence != nil && t.Recurrence.Pattern != nil {
				log.Debugf("[Microsoft Todo Migration] Converting recurring pattern for task %s", t.ID)
				switch t.Recurrence.Pattern.Type {
				case "daily":
					task.RepeatAfter = t.Recurrence.Pattern.Interval * 60 * 60 * 24
				case "weekly":
					task.RepeatAfter = t.Recurrence.Pattern.Interval * 60 * 60 * 24 * 7
				case "monthly":
					task.RepeatAfter = t.Recurrence.Pattern.Interval * 60 * 60 * 24 * 30
				case "yearly":
					task.RepeatAfter = t.Recurrence.Pattern.Interval * 60 * 60 * 24 * 365
				}
			}

			list.Tasks = append(list.Tasks, &models.TaskWithComments{Task: *task})
			log.Debugf("[Microsoft Todo Migration] Done converted %d tasks", len(l.Tasks))
		}

		vikunjsStructure[0].Lists = append(vikunjsStructure[0].Lists, list)
		log.Debugf("[Microsoft Todo Migration] Done converting list %s", l.ID)
	}

	return
}

// Migrate gets all tasks from Microsoft Todo for a user and puts them into vikunja
// @Summary Migrate all lists, tasks etc. from Microsoft Todo
// @Description Migrates all tasklinsts, tasks, notes and reminders from Microsoft Todo to Vikunja.
// @tags migration
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param migrationCode body microsofttodo.Migration true "The auth token previously obtained from the auth url. See the docs for /migration/microsoft-todo/auth."
// @Success 200 {object} models.Message "A message telling you everything was migrated successfully."
// @Failure 500 {object} models.Message "Internal server error"
// @Router /migration/microsoft-todo/migrate [post]
func (m *Migration) Migrate(user *user.User) (err error) {

	log.Debugf("[Microsoft Todo Migration] Start Microsoft Todo migration for user %d", user.ID)
	log.Debugf("[Microsoft Todo Migration] Getting Microsoft Graph api token")

	token, err := getMicrosoftGraphAuthToken(m.Code)
	if err != nil {
		log.Debugf("[Microsoft Todo Migration] Error getting auth token: %s", err)
		return
	}

	log.Debugf("[Microsoft Todo Migration] Got Microsoft Graph api token")
	log.Debugf("[Microsoft Todo Migration] Retrieving Microsoft Todo data")

	todoData, err := getMicrosoftTodoData(token)
	if err != nil {
		log.Debugf("[Microsoft Todo Migration] Error getting Microsoft Todo data: %s", err)
		return
	}

	log.Debugf("[Microsoft Todo Migration] Got Microsoft Todo data")
	log.Debugf("[Microsoft Todo Migration] Start converting Microsoft Todo data")

	vikunjaStructure, err := convertMicrosoftTodoData(todoData)
	if err != nil {
		log.Debugf("[Microsoft Todo Migration] Error converting Microsoft Todo data: %s", err)
		return
	}

	log.Debugf("[Microsoft Todo Migration] Done converting Microsoft Todo data")
	log.Debugf("[Microsoft Todo Migration] Creating new structure")

	err = migration.InsertFromStructure(vikunjaStructure, user)
	if err != nil {
		log.Debugf("[Microsoft Todo Migration] Error while creating new structure: %s", err)
		return
	}

	log.Debugf("[Microsoft Todo Migration] Created new structure")
	log.Debugf("[Microsoft Todo Migration] Microsoft Todo migration done for user %d", user.ID)

	return
}
