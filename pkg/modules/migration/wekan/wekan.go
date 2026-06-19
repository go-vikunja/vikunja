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

package wekan

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"sort"
	"time"

	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/migration"
	"code.vikunja.io/api/pkg/user"

	"github.com/yuin/goldmark"
)

// wekanBoard represents the top-level WeKan board JSON export.
type wekanBoard struct {
	ID     string       `json:"_id"`
	Title  string       `json:"title"`
	Labels []wekanLabel `json:"labels"`
	Lists  []wekanList  `json:"lists"`
	Cards  []wekanCard  `json:"cards"`
	// These are flat arrays at root level in the export, linked by IDs.
	Checklists     []wekanChecklist     `json:"checklists"`
	ChecklistItems []wekanChecklistItem `json:"checklistItems"`
	Comments       []wekanComment       `json:"comments"`
	Attachments    []wekanAttachment    `json:"attachments"`
}

type wekanLabel struct {
	ID    string `json:"_id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

type wekanList struct {
	ID       string  `json:"_id"`
	Title    string  `json:"title"`
	Sort     float64 `json:"sort"`
	Archived bool    `json:"archived"`
}

type wekanCard struct {
	ID               string     `json:"_id"`
	Title            string     `json:"title"`
	Description      string     `json:"description"`
	ListID           string     `json:"listId"`
	LabelIDs         []string   `json:"labelIds"`
	Sort             float64    `json:"sort"`
	Archived         bool       `json:"archived"`
	StartAt          *time.Time `json:"startAt"`
	DueAt            *time.Time `json:"dueAt"`
	EndAt            *time.Time `json:"endAt"`
	CreatedAt        *time.Time `json:"createdAt"`
	DateLastActivity *time.Time `json:"dateLastActivity"`
	ParentID         string     `json:"parentId"`
}

type wekanChecklist struct {
	ID     string  `json:"_id"`
	CardID string  `json:"cardId"`
	Title  string  `json:"title"`
	Sort   float64 `json:"sort"`
}

type wekanChecklistItem struct {
	ID          string  `json:"_id"`
	ChecklistID string  `json:"checklistId"`
	CardID      string  `json:"cardId"`
	Title       string  `json:"title"`
	Sort        float64 `json:"sort"`
	IsFinished  bool    `json:"isFinished"`
}

type wekanComment struct {
	ID        string     `json:"_id"`
	Text      string     `json:"text"`
	CreatedAt *time.Time `json:"createdAt"`
	CardID    string     `json:"cardId"`
}

type wekanAttachment struct {
	ID     string `json:"_id"`
	CardID string `json:"cardId"`
	File   string `json:"file"` // base64-encoded file contents
	Name   string `json:"name"`
	Type   string `json:"type"` // MIME type
}

// wekanColorMap maps WeKan label color names to hex values.
// Values sourced from WeKan's client/components/cards/labels.css.
var wekanColorMap = map[string]string{
	"white":         "ffffff",
	"green":         "3cb500",
	"yellow":        "fad900",
	"orange":        "ff9f19",
	"red":           "eb4646",
	"purple":        "a632db",
	"blue":          "0079bf",
	"sky":           "00c2e0",
	"lime":          "51e898",
	"pink":          "ff78cb",
	"black":         "4d4d4d",
	"silver":        "c0c0c0",
	"peachpuff":     "ffdab9",
	"crimson":       "dc143c",
	"plum":          "dda0dd",
	"darkgreen":     "006400",
	"slateblue":     "6a5acd",
	"magenta":       "ff00ff",
	"gold":          "ffd700",
	"navy":          "000080",
	"gray":          "808080",
	"saddlebrown":   "8b4513",
	"paleturquoise": "afeeee",
	"mistyrose":     "ffe4e1",
	"indigo":        "4b0082",
}

func convertMarkdownToHTML(input string) (string, error) {
	var buf bytes.Buffer
	err := goldmark.Convert([]byte(input), &buf)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func convertWekanToVikunja(board *wekanBoard) []*models.ProjectWithTasksAndBuckets {
	// Build lookup maps
	labelsByID := make(map[string]wekanLabel, len(board.Labels))
	for _, l := range board.Labels {
		labelsByID[l.ID] = l
	}

	// Build checklist items grouped by checklist ID
	checklistItemsByChecklistID := make(map[string][]wekanChecklistItem)
	for _, item := range board.ChecklistItems {
		checklistItemsByChecklistID[item.ChecklistID] = append(
			checklistItemsByChecklistID[item.ChecklistID], item,
		)
	}
	for id := range checklistItemsByChecklistID {
		items := checklistItemsByChecklistID[id]
		sort.Slice(items, func(i, j int) bool {
			return items[i].Sort < items[j].Sort
		})
	}

	// Build checklists grouped by card ID
	checklistsByCardID := make(map[string][]wekanChecklist)
	for _, cl := range board.Checklists {
		checklistsByCardID[cl.CardID] = append(checklistsByCardID[cl.CardID], cl)
	}
	for id := range checklistsByCardID {
		cls := checklistsByCardID[id]
		sort.Slice(cls, func(i, j int) bool {
			return cls[i].Sort < cls[j].Sort
		})
	}

	// Build comments grouped by card ID
	commentsByCardID := make(map[string][]wekanComment)
	for _, c := range board.Comments {
		commentsByCardID[c.CardID] = append(commentsByCardID[c.CardID], c)
	}

	// Build attachments grouped by card ID
	attachmentsByCardID := make(map[string][]wekanAttachment)
	for _, a := range board.Attachments {
		attachmentsByCardID[a.CardID] = append(attachmentsByCardID[a.CardID], a)
	}

	// Create buckets from lists, maintaining sort order
	sortedLists := make([]wekanList, len(board.Lists))
	copy(sortedLists, board.Lists)
	sort.Slice(sortedLists, func(i, j int) bool {
		return sortedLists[i].Sort < sortedLists[j].Sort
	})

	buckets := make([]*models.Bucket, 0, len(sortedLists))
	listIDToBucketID := make(map[string]int64)
	for i, l := range sortedLists {
		bucketID := int64(i + 1)
		listIDToBucketID[l.ID] = bucketID
		buckets = append(buckets, &models.Bucket{
			ID:    bucketID,
			Title: l.Title,
		})
	}

	// Convert cards to tasks
	tasks := make([]*models.TaskWithComments, 0, len(board.Cards))
	for _, card := range board.Cards {
		task := &models.TaskWithComments{
			Task: models.Task{
				Title:    card.Title,
				Position: card.Sort,
				Done:     card.Archived,
				BucketID: listIDToBucketID[card.ListID],
			},
		}

		if card.Description != "" {
			html, err := convertMarkdownToHTML(card.Description)
			if err != nil {
				log.Errorf("[WeKan migration] Error converting description to HTML for card %s: %s", card.ID, err.Error())
				task.Description = card.Description
			} else {
				task.Description = html
			}
		}

		if card.StartAt != nil {
			task.StartDate = *card.StartAt
		}
		if card.DueAt != nil {
			task.DueDate = *card.DueAt
		}

		// Labels
		for _, labelID := range card.LabelIDs {
			label, exists := labelsByID[labelID]
			if !exists {
				continue
			}

			title := label.Name
			if title == "" {
				title = label.Color
			}

			hexColor := wekanColorMap[label.Color]

			task.Labels = append(task.Labels, &models.Label{
				Title:    title,
				HexColor: hexColor,
			})
		}

		// Checklists → append to description as HTML task list
		// This follows the same pattern as the Trello importer.
		if checklists, ok := checklistsByCardID[card.ID]; ok {
			for _, cl := range checklists {
				items, hasItems := checklistItemsByChecklistID[cl.ID]
				if !hasItems || len(items) == 0 {
					continue
				}
				task.Description += "\n\n<h2> " + cl.Title + "</h2>\n\n" + `<ul data-type="taskList">`
				for _, item := range items {
					task.Description += "\n"
					if item.IsFinished {
						task.Description += `<li data-checked="true" data-type="taskItem"><label><input type="checkbox" checked="checked"><span></span></label><div><p>` + item.Title + `</p></div></li>`
					} else {
						task.Description += `<li data-checked="false" data-type="taskItem"><label><input type="checkbox"><span></span></label><div><p>` + item.Title + `</p></div></li>`
					}
				}
				task.Description += "</ul>"
			}
		}

		// Comments
		if comments, ok := commentsByCardID[card.ID]; ok {
			for _, c := range comments {
				commentText := c.Text
				commentHTML, err := convertMarkdownToHTML(c.Text)
				if err != nil {
					log.Errorf("[WeKan migration] Error converting comment to HTML for card %s: %s", card.ID, err.Error())
				} else {
					commentText = commentHTML
				}
				tc := &models.TaskComment{
					Comment: commentText,
				}
				if c.CreatedAt != nil {
					tc.Created = *c.CreatedAt
					tc.Updated = *c.CreatedAt
				}
				task.Comments = append(task.Comments, tc)
			}
		}

		// Attachments
		if attachments, ok := attachmentsByCardID[card.ID]; ok {
			for _, a := range attachments {
				decoded, err := base64.StdEncoding.DecodeString(a.File)
				if err != nil {
					log.Errorf("[WeKan migration] Error decoding attachment %s on card %s: %s", a.ID, card.ID, err.Error())
					continue
				}
				task.Attachments = append(task.Attachments, &models.TaskAttachment{
					File: &files.File{
						Name:        a.Name,
						Mime:        a.Type,
						Size:        uint64(len(decoded)),
						FileContent: decoded,
					},
				})
			}
		}

		tasks = append(tasks, task)
	}

	title := board.Title
	if title == "" {
		title = "Imported WeKan Board"
	}

	project := &models.ProjectWithTasksAndBuckets{
		Project: models.Project{
			Title: title,
		},
		Tasks:   tasks,
		Buckets: buckets,
	}

	return []*models.ProjectWithTasksAndBuckets{project}
}

func parseWekanJSON(r io.Reader) (*wekanBoard, error) {
	var board wekanBoard
	decoder := json.NewDecoder(r)
	err := decoder.Decode(&board)
	if err != nil {
		return nil, err
	}
	return &board, nil
}

// Migrator is the WeKan migration struct.
type Migrator struct{}

// Name is used to identify the wekan migration.
// @Summary Get migration status
// @Description Returns if the current user already did the migration or not. This is useful to show a confirmation message in the frontend if the user is trying to do the same migration again.
// @tags migration
// @Produce json
// @Security JWTKeyAuth
// @Success 200 {object} migration.Status "The migration status"
// @Failure 500 {object} models.Message "Internal server error"
// @Router /migration/wekan/status [get]
func (m *Migrator) Name() string {
	return "wekan"
}

// Migrate takes a WeKan board JSON export and imports it into Vikunja.
// @Summary Import all projects, tasks etc. from a WeKan board export
// @Description Imports all projects, tasks, labels, checklists, comments, and attachments from a WeKan board JSON export into Vikunja.
// @tags migration
// @Accept x-www-form-urlencoded
// @Produce json
// @Security JWTKeyAuth
// @Param import formData string true "The WeKan board JSON export file."
// @Success 200 {object} models.Message "A message telling you everything was migrated successfully."
// @Failure 500 {object} models.Message "Internal server error"
// @Router /migration/wekan/migrate [put]
func (m *Migrator) Migrate(user *user.User, file io.ReaderAt, size int64) error {
	if size == 0 {
		return &migration.ErrFileIsEmpty{}
	}

	fr := io.NewSectionReader(file, 0, size)

	board, err := parseWekanJSON(fr)
	if err != nil {
		return err
	}

	if board.Title == "" && len(board.Cards) == 0 {
		return &migration.ErrFileIsEmpty{}
	}

	vikunjaData := convertWekanToVikunja(board)

	return migration.InsertFromStructure(vikunjaData, user)
}
