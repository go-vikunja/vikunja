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

package planka

import "time"

// Planka ids are string snowflakes. Only fields the importer uses are declared.

type plankaUser struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
}

type plankaProject struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

type plankaBoard struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	ProjectID string  `json:"projectId"`
	Position  float64 `json:"position"`
}

const (
	listTypeActive  = "active"
	listTypeClosed  = "closed"
	listTypeArchive = "archive"
	listTypeTrash   = "trash"
)

type plankaList struct {
	ID       string   `json:"id"`
	Name     *string  `json:"name"`
	Type     string   `json:"type"`
	Position *float64 `json:"position"`
}

type plankaCard struct {
	ID                string     `json:"id"`
	Name              string     `json:"name"`
	Description       *string    `json:"description"`
	DueDate           *time.Time `json:"dueDate"`
	IsClosed          bool       `json:"isClosed"`
	CommentsTotal     int        `json:"commentsTotal"`
	Position          *float64   `json:"position"`
	ListID            string     `json:"listId"`
	CoverAttachmentID *string    `json:"coverAttachmentId"`
	ListChangedAt     *time.Time `json:"listChangedAt"`
	CreatedAt         *time.Time `json:"createdAt"`
}

type plankaLabel struct {
	ID    string  `json:"id"`
	Name  *string `json:"name"`
	Color string  `json:"color"`
}

type plankaCardLabel struct {
	CardID  string `json:"cardId"`
	LabelID string `json:"labelId"`
}

type plankaTaskList struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	CardID   string  `json:"cardId"`
	Position float64 `json:"position"`
}

type plankaTask struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	IsCompleted bool    `json:"isCompleted"`
	TaskListID  string  `json:"taskListId"`
	Position    float64 `json:"position"`
}

const attachmentTypeLink = "link"

type plankaAttachment struct {
	ID        string     `json:"id"`
	Type      string     `json:"type"`
	Name      string     `json:"name"`
	CardID    string     `json:"cardId"`
	CreatedAt *time.Time `json:"createdAt"`
	Data      struct {
		URL      string `json:"url"`
		MimeType string `json:"mimeType"`
	} `json:"data"`
}

type plankaBaseCustomFieldGroup struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type plankaCustomFieldGroup struct {
	ID                     string  `json:"id"`
	Name                   *string `json:"name"`
	Position               float64 `json:"position"`
	BaseCustomFieldGroupID *string `json:"baseCustomFieldGroupId"`
}

type plankaCustomField struct {
	ID                     string  `json:"id"`
	Name                   string  `json:"name"`
	Position               float64 `json:"position"`
	CustomFieldGroupID     *string `json:"customFieldGroupId"`
	BaseCustomFieldGroupID *string `json:"baseCustomFieldGroupId"`
}

type plankaCustomFieldValue struct {
	CardID             string `json:"cardId"`
	CustomFieldGroupID string `json:"customFieldGroupId"`
	CustomFieldID      string `json:"customFieldId"`
	Content            string `json:"content"`
}

type plankaComment struct {
	ID        string     `json:"id"`
	Text      string     `json:"text"`
	CardID    string     `json:"cardId"`
	UserID    *string    `json:"userId"`
	CreatedAt *time.Time `json:"createdAt"`
}

// projectsResponse is GET /api/projects.
type projectsResponse struct {
	Items    []plankaProject `json:"items"`
	Included struct {
		Boards                []plankaBoard                `json:"boards"`
		BaseCustomFieldGroups []plankaBaseCustomFieldGroup `json:"baseCustomFieldGroups"`
		CustomFields          []plankaCustomField          `json:"customFields"`
	} `json:"included"`
}

// boardIncluded is the "included" part of GET /api/boards/:id and GET /api/lists/:id/cards.
type boardIncluded struct {
	Users             []plankaUser             `json:"users"`
	Labels            []plankaLabel            `json:"labels"`
	Lists             []plankaList             `json:"lists"`
	Cards             []plankaCard             `json:"cards"`
	CardLabels        []plankaCardLabel        `json:"cardLabels"`
	TaskLists         []plankaTaskList         `json:"taskLists"`
	Tasks             []plankaTask             `json:"tasks"`
	Attachments       []plankaAttachment       `json:"attachments"`
	CustomFieldGroups []plankaCustomFieldGroup `json:"customFieldGroups"`
	CustomFields      []plankaCustomField      `json:"customFields"`
	CustomFieldValues []plankaCustomFieldValue `json:"customFieldValues"`
}

// boardResponse is GET /api/boards/:id. Cards of archive/trash lists are not part of it.
type boardResponse struct {
	Item     plankaBoard   `json:"item"`
	Included boardIncluded `json:"included"`
}

// listCardsResponse is GET /api/lists/:id/cards.
type listCardsResponse struct {
	Items    []plankaCard  `json:"items"`
	Included boardIncluded `json:"included"`
}

// commentsResponse is GET /api/cards/:id/comments.
type commentsResponse struct {
	Items    []plankaComment `json:"items"`
	Included struct {
		Users []plankaUser `json:"users"`
	} `json:"included"`
}

// plankaBoardData is a board with everything the importer needs, merged from the board payload,
// the paged archive-list cards and the per-card comments.
type plankaBoardData struct {
	plankaBoard
	boardIncluded
	Comments map[string][]plankaComment
}

type plankaProjectData struct {
	plankaProject
	Boards []*plankaBoardData
}

// plankaData is everything fetched from a Planka instance for one user.
type plankaData struct {
	CurrentUserID         string
	Users                 map[string]plankaUser
	BaseCustomFieldGroups map[string]plankaBaseCustomFieldGroup
	BaseCustomFields      map[string]plankaCustomField
	Projects              []*plankaProjectData
}
