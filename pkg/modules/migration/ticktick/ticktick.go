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

package ticktick

import (
	"encoding/csv"
	"errors"
	"io"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/log"

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/migration"
	"code.vikunja.io/api/pkg/user"
)

const timeISO = "2006-01-02T15:04:05-0700"

type Migrator struct {
}

type tickTickTask struct {
	FolderName    string
	ListName      string
	Title         string
	Tags          []string
	Content       string
	IsChecklist   bool
	StartDate     time.Time
	DueDate       time.Time
	Reminder      time.Duration
	Repeat        string
	Priority      int
	Status        string
	CreatedTime   time.Time
	CompletedTime time.Time
	Order         float64
	TaskID        int64
	ParentID      int64
}

// Copied from https://stackoverflow.com/a/57617885
var durationRegex = regexp.MustCompile(`P([\d\.]+Y)?([\d\.]+M)?([\d\.]+D)?T?([\d\.]+H)?([\d\.]+M)?([\d\.]+?S)?`)

// ParseDuration converts a ISO8601 duration into a time.Duration
func parseDuration(str string) time.Duration {
	matches := durationRegex.FindStringSubmatch(str)

	if len(matches) == 0 {
		return 0
	}

	years := parseDurationPart(matches[1], time.Hour*24*365)
	months := parseDurationPart(matches[2], time.Hour*24*30)
	days := parseDurationPart(matches[3], time.Hour*24)
	hours := parseDurationPart(matches[4], time.Hour)
	minutes := parseDurationPart(matches[5], time.Second*60)
	seconds := parseDurationPart(matches[6], time.Second)

	return years + months + days + hours + minutes + seconds
}

func parseDurationPart(value string, unit time.Duration) time.Duration {
	if len(value) != 0 {
		if parsed, err := strconv.ParseFloat(value[:len(value)-1], 64); err == nil {
			return time.Duration(float64(unit) * parsed)
		}
	}
	return 0
}

func convertTickTickToVikunja(tasks []*tickTickTask) (result []*models.NamespaceWithListsAndTasks) {
	namespace := &models.NamespaceWithListsAndTasks{
		Namespace: models.Namespace{
			Title: "Migrated from TickTick",
		},
		Lists: []*models.ListWithTasksAndBuckets{},
	}

	lists := make(map[string]*models.ListWithTasksAndBuckets)
	for _, t := range tasks {
		_, has := lists[t.ListName]
		if !has {
			lists[t.ListName] = &models.ListWithTasksAndBuckets{
				List: models.List{
					Title: t.ListName,
				},
			}
		}

		labels := make([]*models.Label, 0, len(t.Tags))
		for _, tag := range t.Tags {
			labels = append(labels, &models.Label{
				Title: tag,
			})
		}

		task := &models.TaskWithComments{
			Task: models.Task{
				ID:          t.TaskID,
				Title:       t.Title,
				Description: t.Content,
				StartDate:   t.StartDate,
				EndDate:     t.DueDate,
				DueDate:     t.DueDate,
				Reminders: []time.Time{
					t.DueDate.Add(t.Reminder * -1),
				},
				Done:     t.Status == "1",
				DoneAt:   t.CompletedTime,
				Position: t.Order,
				Labels:   labels,
			},
		}

		if t.ParentID != 0 {
			task.RelatedTasks = map[models.RelationKind][]*models.Task{
				models.RelationKindParenttask: {{ID: t.ParentID}},
			}
		}

		lists[t.ListName].Tasks = append(lists[t.ListName].Tasks, task)
	}

	for _, l := range lists {
		namespace.Lists = append(namespace.Lists, l)
	}

	sort.Slice(namespace.Lists, func(i, j int) bool {
		return namespace.Lists[i].Title < namespace.Lists[j].Title
	})

	return []*models.NamespaceWithListsAndTasks{namespace}
}

// Name is used to get the name of the ticktick migration - we're using the docs here to annotate the status route.
// @Summary Get migration status
// @Description Returns if the current user already did the migation or not. This is useful to show a confirmation message in the frontend if the user is trying to do the same migration again.
// @tags migration
// @Produce json
// @Security JWTKeyAuth
// @Success 200 {object} migration.Status "The migration status"
// @Failure 500 {object} models.Message "Internal server error"
// @Router /migration/ticktick/status [get]
func (m *Migrator) Name() string {
	return "ticktick"
}

// Migrate takes a ticktick export, parses it and imports everything in it into Vikunja.
// @Summary Import all lists, tasks etc. from a TickTick backup export
// @Description Imports all projects, tasks, notes, reminders, subtasks and files from a TickTick backup export into Vikunja.
// @tags migration
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param import formData string true "The TickTick backup csv file."
// @Success 200 {object} models.Message "A message telling you everything was migrated successfully."
// @Failure 500 {object} models.Message "Internal server error"
// @Router /migration/ticktick/migrate [post]
func (m *Migrator) Migrate(user *user.User, file io.ReaderAt, size int64) error {
	fr := io.NewSectionReader(file, 0, size)
	r := csv.NewReader(fr)

	allTasks := []*tickTickTask{}
	line := 0
	for {

		record, err := r.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			log.Debugf("[TickTick Migration] CSV parse error: %s", err)
		}

		line++
		if line <= 4 {
			continue
		}

		priority, err := strconv.Atoi(record[10])
		if err != nil {
			return err
		}
		order, err := strconv.ParseFloat(record[14], 64)
		if err != nil {
			return err
		}
		taskID, err := strconv.ParseInt(record[21], 10, 64)
		if err != nil {
			return err
		}
		parentID, err := strconv.ParseInt(record[21], 10, 64)
		if err != nil {
			return err
		}

		reminder := parseDuration(record[8])

		t := &tickTickTask{
			ListName:    record[1],
			Title:       record[2],
			Tags:        strings.Split(record[3], ", "),
			Content:     record[4],
			IsChecklist: record[5] == "Y",
			Reminder:    reminder,
			Repeat:      record[9],
			Priority:    priority,
			Status:      record[11],
			Order:       order,
			TaskID:      taskID,
			ParentID:    parentID,
		}

		if record[6] != "" {
			t.StartDate, err = time.Parse(timeISO, record[6])
			if err != nil {
				return err
			}
		}
		if record[7] != "" {
			t.DueDate, err = time.Parse(timeISO, record[7])
			if err != nil {
				return err
			}
		}
		if record[12] != "" {
			t.StartDate, err = time.Parse(timeISO, record[12])
			if err != nil {
				return err
			}
		}
		if record[13] != "" {
			t.CompletedTime, err = time.Parse(timeISO, record[13])
			if err != nil {
				return err
			}
		}

		allTasks = append(allTasks, t)
	}

	vikunjaTasks := convertTickTickToVikunja(allTasks)

	return migration.InsertFromStructure(vikunjaTasks, user)
}
