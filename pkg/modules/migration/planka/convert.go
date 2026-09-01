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

import (
	"bytes"
	"errors"
	"fmt"
	"maps"
	"math"
	"sort"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/richtext"
)

// labelColors maps Planka label color names (server/api/models/Label.js COLORS) to the hex values
// the Planka client renders them with (client/src/styles.module.scss). Gradients use their first stop.
var labelColors = map[string]string{
	"muddy-grey":      "69655a",
	"autumn-leafs":    "c9b037",
	"morning-sky":     "52b9d5",
	"antique-blue":    "6c99bb",
	"egg-yellow":      "f9c423",
	"desert-sand":     "fad371",
	"dark-granite":    "8b8680",
	"fresh-salad":     "ced85e",
	"lagoon-blue":     "109dc0",
	"midnight-blue":   "0a63a0",
	"light-orange":    "fdae5f",
	"pumpkin-orange":  "ed9223",
	"light-concrete":  "afb0a4",
	"sunny-grass":     "beca02",
	"navy-blue":       "1d7299",
	"lilac-eyes":      "406cbd",
	"apricot-red":     "fc736c",
	"orange-peel":     "de692f",
	"bright-moss":     "96b352",
	"deep-ocean":      "004c70",
	"summer-sky":      "5d9cec",
	"berry-red":       "e83855",
	"light-cocoa":     "a85540",
	"grey-stone":      "aab2bd",
	"tank-green":      "8aa177",
	"coral-green":     "2b6a6c",
	"sugar-plum":      "7e86c7",
	"pink-tulip":      "e34f7c",
	"shady-rust":      "87564a",
	"wet-rock":        "83949b",
	"wet-moss":        "4a8753",
	"turquoise-sea":   "00858a",
	"lavender-fields": "b287bd",
	"piggy-red":       "f97394",
	"light-mud":       "c7a57a",
	"gun-metal":       "4f6573",
	"modern-green":    "77ce87",
	"french-coast":    "00b4b1",
	"sweet-lilac":     "975298",
	"red-burgundy":    "ad5f7d",
	"silver-glint":    "adadad",
	"pirate-gold":     "b47e11",
}

// attachmentDownloader fetches the content of a file attachment. Kept as a function so the
// conversion stays free of HTTP.
type attachmentDownloader func(a *plankaAttachment) (*bytes.Buffer, error)

// converter carries the id counters that must be unique across the whole import.
type converter struct {
	data         *plankaData
	download     attachmentDownloader
	projectID    int64
	bucketID     int64
	attachmentID int64
}

func convertPlankaToVikunja(data *plankaData, download attachmentDownloader) ([]*models.ProjectWithTasksAndBuckets, error) {
	c := &converter{data: data, download: download}
	root := c.newProject("Migrated from Planka", "", nil)
	result := []*models.ProjectWithTasksAndBuckets{root}

	for _, p := range data.Projects {
		description := ""
		if p.Description != nil {
			description = *p.Description
		}

		if len(p.Boards) <= 1 {
			project := c.newProject(p.Name, description, models.Ptr(root.ID))
			if len(p.Boards) == 1 {
				if err := c.convertBoard(p.Boards[0], project); err != nil {
					return nil, err
				}
			}
			result = append(result, project)
			continue
		}

		parent := c.newProject(p.Name, description, models.Ptr(root.ID))
		result = append(result, parent)
		for _, b := range p.Boards {
			child := c.newProject(b.Name, "", models.Ptr(parent.ID))
			if err := c.convertBoard(b, child); err != nil {
				return nil, err
			}
			result = append(result, child)
		}
	}

	return result, nil
}

func (c *converter) newProject(title, description string, parentID *int64) *models.ProjectWithTasksAndBuckets {
	c.projectID++
	return &models.ProjectWithTasksAndBuckets{
		Project: models.Project{
			ID:              c.projectID,
			ParentProjectID: parentID,
			Title:           title,
			Description:     description,
		},
	}
}

func listTypeOrder(t string) int {
	switch t {
	case listTypeActive:
		return 0
	case listTypeClosed:
		return 1
	default:
		return 2
	}
}

func positionOrInf(p *float64) float64 {
	if p == nil {
		return math.Inf(1)
	}
	return *p
}

func timeOrZero(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}

func (c *converter) convertBoard(bd *plankaBoardData, project *models.ProjectWithTasksAndBuckets) error {
	log.Debugf("[Planka Migration] Converting board %s with %d cards", bd.ID, len(bd.Cards))

	labelsByID := map[string]plankaLabel{}
	for _, l := range bd.Labels {
		labelsByID[l.ID] = l
	}
	labelIDsByCard := map[string][]string{}
	for _, cl := range bd.CardLabels {
		labelIDsByCard[cl.CardID] = append(labelIDsByCard[cl.CardID], cl.LabelID)
	}
	taskListsByCard := map[string][]plankaTaskList{}
	for _, tl := range bd.TaskLists {
		taskListsByCard[tl.CardID] = append(taskListsByCard[tl.CardID], tl)
	}
	tasksByList := map[string][]plankaTask{}
	for _, t := range bd.Tasks {
		tasksByList[t.TaskListID] = append(tasksByList[t.TaskListID], t)
	}
	attachmentsByCard := map[string][]plankaAttachment{}
	for _, a := range bd.Attachments {
		attachmentsByCard[a.CardID] = append(attachmentsByCard[a.CardID], a)
	}
	valuesByCard := map[string][]plankaCustomFieldValue{}
	for _, v := range bd.CustomFieldValues {
		valuesByCard[v.CardID] = append(valuesByCard[v.CardID], v)
	}
	groupsByID := map[string]plankaCustomFieldGroup{}
	for _, g := range bd.CustomFieldGroups {
		groupsByID[g.ID] = g
	}
	fieldsByID := map[string]plankaCustomField{}
	maps.Copy(fieldsByID, c.data.BaseCustomFields)
	for _, f := range bd.CustomFields {
		fieldsByID[f.ID] = f
	}
	cardsByList := map[string][]plankaCard{}
	for _, card := range bd.Cards {
		cardsByList[card.ListID] = append(cardsByList[card.ListID], card)
	}

	lists := make([]plankaList, 0, len(bd.Lists))
	for _, l := range bd.Lists {
		if l.Type != listTypeTrash {
			lists = append(lists, l)
		}
	}
	sort.SliceStable(lists, func(i, j int) bool {
		if listTypeOrder(lists[i].Type) != listTypeOrder(lists[j].Type) {
			return listTypeOrder(lists[i].Type) < listTypeOrder(lists[j].Type)
		}
		return positionOrInf(lists[i].Position) < positionOrInf(lists[j].Position)
	})

	for _, list := range lists {
		c.bucketID++
		bucket := &models.Bucket{
			ID:    c.bucketID,
			Title: listTitle(list),
		}
		project.Buckets = append(project.Buckets, bucket)

		cards := cardsByList[list.ID]
		sort.SliceStable(cards, func(i, j int) bool {
			if positionOrInf(cards[i].Position) != positionOrInf(cards[j].Position) {
				return positionOrInf(cards[i].Position) < positionOrInf(cards[j].Position)
			}
			return timeOrZero(cards[i].CreatedAt).Before(timeOrZero(cards[j].CreatedAt))
		})

		for _, card := range cards {
			task := &models.TaskWithComments{
				Task: models.Task{
					Title:    card.Name,
					BucketID: bucket.ID,
					Done:     card.IsClosed || list.Type == listTypeClosed || list.Type == listTypeArchive,
					Created:  timeOrZero(card.CreatedAt),
				},
			}
			if card.DueDate != nil {
				task.DueDate = *card.DueDate
			}

			var md strings.Builder
			if card.Description != nil {
				md.WriteString(*card.Description)
			}
			md.WriteString(customFieldsMarkdown(valuesByCard[card.ID], groupsByID, fieldsByID, c.data.BaseCustomFieldGroups))
			md.WriteString(checklistsMarkdown(taskListsByCard[card.ID], tasksByList))
			md.WriteString(linkAttachmentsMarkdown(attachmentsByCard[card.ID]))

			description, err := richtext.MarkdownToHTML(md.String())
			if err != nil {
				return fmt.Errorf("converting description of card %s: %w", card.ID, err)
			}
			task.Description = description

			for _, labelID := range labelIDsByCard[card.ID] {
				label, has := labelsByID[labelID]
				if !has {
					continue
				}
				task.Labels = append(task.Labels, &models.Label{
					Title:    labelTitle(label),
					HexColor: labelColors[label.Color],
				})
			}

			for i := range attachmentsByCard[card.ID] {
				a := &attachmentsByCard[card.ID][i]
				if a.Type == attachmentTypeLink {
					continue
				}
				buf, err := c.download(a)
				if err != nil {
					// Budget failures abort; ordinary download failures remain skippable.
					var budgetErr *ErrImportBudgetExceeded
					if errors.As(err, &budgetErr) {
						return err
					}
					log.Errorf("[Planka Migration] Could not download attachment %s of card %s, skipping: %s", a.ID, card.ID, err)
					continue
				}
				c.attachmentID++
				task.Attachments = append(task.Attachments, &models.TaskAttachment{
					ID: c.attachmentID,
					File: &files.File{
						Name:        a.Name,
						Mime:        a.Data.MimeType,
						Size:        uint64(buf.Len()), //nolint:gosec // buffer length is never negative
						Created:     timeOrZero(a.CreatedAt),
						FileContent: buf.Bytes(),
					},
				})
				if card.CoverAttachmentID != nil && *card.CoverAttachmentID == a.ID {
					task.CoverImageAttachmentID = c.attachmentID
				}
			}

			for _, comment := range bd.Comments[card.ID] {
				text := comment.Text
				if comment.UserID != nil && *comment.UserID != c.data.CurrentUserID {
					text = "*" + c.userName(*comment.UserID) + "*:\n\n" + text
				}
				html, err := richtext.MarkdownToHTML(text)
				if err != nil {
					return fmt.Errorf("converting comment %s of card %s: %w", comment.ID, card.ID, err)
				}
				task.Comments = append(task.Comments, &models.TaskComment{
					Comment: html,
					Created: timeOrZero(comment.CreatedAt),
					Updated: timeOrZero(comment.CreatedAt),
				})
			}

			project.Tasks = append(project.Tasks, task)
		}
	}

	return nil
}

func (c *converter) userName(id string) string {
	if u, has := c.data.Users[id]; has && u.Name != "" {
		return u.Name
	}
	return "Unknown user"
}

func listTitle(l plankaList) string {
	if l.Name != nil && *l.Name != "" {
		return *l.Name
	}
	if l.Type == "" {
		return "List"
	}
	return strings.ToUpper(l.Type[:1]) + l.Type[1:]
}

func labelTitle(l plankaLabel) string {
	if l.Name != nil && *l.Name != "" {
		return *l.Name
	}
	return l.Color
}

func checklistsMarkdown(taskLists []plankaTaskList, tasksByList map[string][]plankaTask) string {
	sort.SliceStable(taskLists, func(i, j int) bool { return taskLists[i].Position < taskLists[j].Position })

	var md strings.Builder
	for _, tl := range taskLists {
		tasks := tasksByList[tl.ID]
		if len(tasks) == 0 {
			continue
		}
		sort.SliceStable(tasks, func(i, j int) bool { return tasks[i].Position < tasks[j].Position })

		md.WriteString("\n\n## " + tl.Name + "\n\n")
		for _, t := range tasks {
			if t.IsCompleted {
				md.WriteString("- [x] ")
			} else {
				md.WriteString("- [ ] ")
			}
			md.WriteString(t.Name + "\n")
		}
	}
	return md.String()
}

func linkAttachmentsMarkdown(attachments []plankaAttachment) string {
	var md strings.Builder
	for _, a := range attachments {
		if a.Type != attachmentTypeLink || a.Data.URL == "" {
			continue
		}
		if md.Len() == 0 {
			md.WriteString("\n\n## Links\n\n")
		}
		name := a.Name
		if name == "" {
			name = a.Data.URL
		}
		md.WriteString("- [" + name + "](" + a.Data.URL + ")\n")
	}
	return md.String()
}

func customFieldsMarkdown(values []plankaCustomFieldValue, groups map[string]plankaCustomFieldGroup, fields map[string]plankaCustomField, baseGroups map[string]plankaBaseCustomFieldGroup) string {
	if len(values) == 0 {
		return ""
	}

	valuesByGroup := map[string][]plankaCustomFieldValue{}
	groupIDs := []string{}
	for _, v := range values {
		if v.Content == "" {
			continue
		}
		if _, has := fields[v.CustomFieldID]; !has {
			log.Debugf("[Planka Migration] Skipping value of unknown custom field %s", v.CustomFieldID)
			continue
		}
		if _, seen := valuesByGroup[v.CustomFieldGroupID]; !seen {
			groupIDs = append(groupIDs, v.CustomFieldGroupID)
		}
		valuesByGroup[v.CustomFieldGroupID] = append(valuesByGroup[v.CustomFieldGroupID], v)
	}
	sort.SliceStable(groupIDs, func(i, j int) bool {
		return groups[groupIDs[i]].Position < groups[groupIDs[j]].Position
	})

	var md strings.Builder
	for _, gid := range groupIDs {
		group := groups[gid]
		vals := valuesByGroup[gid]
		sort.SliceStable(vals, func(i, j int) bool {
			return fields[vals[i].CustomFieldID].Position < fields[vals[j].CustomFieldID].Position
		})

		title := ""
		if group.Name != nil {
			title = *group.Name
		}
		if title == "" && group.BaseCustomFieldGroupID != nil {
			title = baseGroups[*group.BaseCustomFieldGroupID].Name
		}
		if title == "" {
			title = "Custom fields"
		}

		md.WriteString("\n\n## " + title + "\n\n| Field | Value |\n| --- | --- |\n")
		for _, v := range vals {
			md.WriteString("| " + escapeTableCell(fields[v.CustomFieldID].Name) + " | " + escapeTableCell(v.Content) + " |\n")
		}
	}
	return md.String()
}

func escapeTableCell(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "|", "\\|"), "\n", " ")
}
