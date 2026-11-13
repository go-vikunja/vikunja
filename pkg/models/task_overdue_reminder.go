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

package models

import (
	"slices"
	"strconv"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/events"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/cron"

	"code.vikunja.io/api/pkg/log"

	"code.vikunja.io/api/pkg/user"

	"xorm.io/builder"
	"xorm.io/xorm"
)

func overdueSlot(window cron.Window, timezone, clock string) (time.Time, error) {
	if timezone == "" {
		timezone = config.GetTimeZone().String()
	}
	location, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Time{}, err
	}
	if clock == "" {
		clock = config.DefaultSettingsOverdueTaskRemindersTime.GetString()
	}
	if clock == "" {
		clock = "09:00"
	}
	daily, err := time.Parse("15:04", clock)
	if err != nil {
		return time.Time{}, err
	}
	local := window.To.In(location)
	slot := time.Date(local.Year(), local.Month(), local.Day(), daily.Hour(), daily.Minute(), 0, 0, location)
	if slot.After(window.To) {
		local = local.AddDate(0, 0, -1)
		slot = time.Date(local.Year(), local.Month(), local.Day(), daily.Hour(), daily.Minute(), 0, 0, location)
	}
	if !slot.After(window.From) {
		return time.Time{}, nil
	}
	return slot, nil
}

func getOverdueUsersInWindow(s *xorm.Session, window cron.Window, cond builder.Cond) (map[int64]*userWithTasks, error) {
	eligible := builder.And(
		builder.Eq{"users.status": user.StatusActive},
		builder.Or(builder.IsNull{"users.bot_owner_id"}, builder.Eq{"users.bot_owner_id": 0}),
	)
	if cond != nil {
		eligible = eligible.And(cond)
	}
	users, err := user.GetUsersByCond(s, eligible)
	if err != nil {
		return nil, err
	}
	result := make(map[int64]*userWithTasks)
	for _, u := range users {
		slot, err := overdueSlot(window, u.Timezone, u.OverdueTasksRemindersTime)
		if err != nil {
			log.Warningf("Invalid overdue reminder time for user %d: %v", u.ID, err)
			continue
		}
		if !slot.IsZero() {
			result[u.ID] = &userWithTasks{user: u, assigned: make(map[int64]*Task), followed: make(map[int64]*Task), slot: slot}
		}
	}
	return result, nil
}

func getOverdueTasksInWindow(s *xorm.Session, window cron.Window, cond builder.Cond) (map[int64]*userWithTasks, error) {
	users, err := getOverdueUsersInWindow(s, window, cond)
	if err != nil {
		return nil, err
	}
	for id, ut := range users {
		if err := ut.loadOverdueTasks(s); err != nil {
			return nil, err
		}
		if len(ut.assigned)+len(ut.followed) == 0 {
			delete(users, id)
		}
	}
	return users, nil
}

func (ut *userWithTasks) loadOverdueTasks(s *xorm.Session) error {
	const batchSize = 200
	var after int64
	for {
		var tasks []*Task
		err := s.Select("tasks.*").Join("INNER", "projects", "projects.id = tasks.project_id").
			Where(builder.Gt{"tasks.id": after}).And(builder.Lt{"tasks.due_date": ut.slot}).
			And(builder.Eq{"tasks.done": false, "projects.is_archived": false}).
			And(builder.IsNull{"tasks.deleted_at"}).And(builder.IsNull{"projects.deleted_at"}).
			And(`tasks.id IN (
				SELECT entity_id FROM subscriptions WHERE entity_type = ? AND user_id = ? AND muted = ?
				UNION
				SELECT t.id FROM tasks t
				JOIN project_ancestors pa ON pa.project_id = t.project_id
				JOIN subscriptions sub ON sub.entity_id = pa.ancestor_id
				WHERE sub.entity_type = ? AND sub.user_id = ? AND sub.muted = ?
			)`, SubscriptionEntityTask, ut.user.ID, false, SubscriptionEntityProject, ut.user.ID, false).
			OrderBy("tasks.id").Limit(batchSize).Find(&tasks)
		if err != nil {
			return err
		}
		if len(tasks) == 0 {
			return nil
		}
		after = tasks[len(tasks)-1].ID
		ids := make([]int64, 0, len(tasks))
		for _, task := range tasks {
			ids = append(ids, task.ID)
		}
		subs, err := GetSubscriptionsForEntitiesAndUser(s, SubscriptionEntityTask, ids, ut.user)
		if err != nil {
			return err
		}
		// The user-scoped subscription resolver does not check current read access.
		subs, err = filterSubscriptionsByReadPermission(s, SubscriptionEntityTask, subs)
		if err != nil {
			return err
		}
		var assignees []*TaskAssginee
		if err := s.In("task_id", ids).Where("user_id = ?", ut.user.ID).Find(&assignees); err != nil {
			return err
		}
		assigned := make(map[int64]bool, len(assignees))
		for _, a := range assignees {
			assigned[a.TaskID] = true
		}
		for _, task := range tasks {
			if len(subs[task.ID]) == 0 {
				continue
			}
			if assigned[task.ID] {
				ut.assigned[task.ID] = task
			} else {
				ut.followed[task.ID] = task
			}
		}
		if len(tasks) < batchSize {
			return nil
		}
	}
}

type userWithTasks struct {
	slot     time.Time
	user     *user.User
	assigned map[int64]*Task
	followed map[int64]*Task
}

func (ut *userWithTasks) allTasks() []*Task {
	return append(mapToSlice(ut.assigned), mapToSlice(ut.followed)...)
}

func (ut *userWithTasks) projects(s *xorm.Session) (map[int64]*Project, error) {
	seen := make(map[int64]bool)
	var ids []int64
	for _, task := range ut.allTasks() {
		if !seen[task.ProjectID] {
			seen[task.ProjectID] = true
			ids = append(ids, task.ProjectID)
		}
	}
	projects := make(map[int64]*Project)
	for batch := range slices.Chunk(ids, 200) {
		if err := s.In("id", batch).Find(&projects); err != nil {
			return nil, err
		}
	}
	return projects, nil
}

func RegisterOverdueReminderCron() {
	if err := cron.Register("overdue-reminders", "* * * * *", enqueueOverdueReminders); err != nil {
		log.Fatalf("Register overdue reminders: %v", err)
	}
}

func mapToSlice(m map[int64]*Task) []*Task {
	tasks := make([]*Task, 0, len(m))
	for _, t := range m {
		tasks = append(tasks, t)
	}
	return tasks
}

// hasNonWebhookListener catches listeners registered by plugins.
func hasNonWebhookListener(eventName string) bool {
	return slices.ContainsFunc(events.ListenersFor(eventName), func(l events.Listener) bool {
		_, isWebhook := l.(*WebhookListener)
		return !isWebhook
	})
}

// overdueEventsWithListeners lets the cron skip dispatching events nobody handles,
// since each dispatched event costs the webhook listener a db session.
func overdueEventsWithListeners(s *xorm.Session) (single, batch bool, err error) {
	singleName := (&TaskOverdueEvent{}).Name()
	batchName := (&TasksOverdueEvent{}).Name()

	single = hasNonWebhookListener(singleName)
	batch = hasNonWebhookListener(batchName)
	if single && batch {
		return
	}

	webhooks := []*Webhook{}
	err = s.Cols("events").Find(&webhooks)
	if err != nil {
		return
	}

	for _, w := range webhooks {
		single = single || slices.Contains(w.Events, singleName)
		batch = batch || slices.Contains(w.Events, batchName)
	}
	return
}

// Eligibility is batched per recipient; the listener still checks current creator access.
func overdueEventsForRecipient(s *xorm.Session, ut *userWithTasks) (singleTasks []*Task, batch bool, err error) {
	singleName := (&TaskOverdueEvent{}).Name()
	batchName := (&TasksOverdueEvent{}).Name()
	allSingle := hasNonWebhookListener(singleName)
	batch = hasNonWebhookListener(batchName)
	tasks := ut.allTasks()
	if len(tasks) == 0 || allSingle && batch {
		return tasks, batch, nil
	}

	descendants := make(map[int64]map[int64]bool)
	projectIDs := []int64{}
	for _, task := range tasks {
		if descendants[task.ProjectID] == nil {
			descendants[task.ProjectID] = map[int64]bool{task.ProjectID: true}
			projectIDs = append(projectIDs, task.ProjectID)
		}
	}
	ancestors := []ProjectAncestor{}
	if err := s.Where("project_id IN (" + overdueProjectIDList(projectIDs) + ")").Find(&ancestors); err != nil {
		return nil, false, err
	}
	for _, ancestor := range ancestors {
		if descendants[ancestor.AncestorID] == nil {
			descendants[ancestor.AncestorID] = make(map[int64]bool)
		}
		descendants[ancestor.AncestorID][ancestor.ProjectID] = true
	}
	hookProjectIDs := make([]int64, 0, len(descendants))
	for id := range descendants {
		hookProjectIDs = append(hookProjectIDs, id)
	}
	hooks := []*Webhook{}
	err = s.Cols("project_id", "user_id", "events").Where(builder.Or(
		builder.Expr("project_id IN ("+overdueProjectIDList(hookProjectIDs)+")"),
		builder.And(builder.Eq{"user_id": ut.user.ID}, builder.Or(builder.IsNull{"project_id"}, builder.Eq{"project_id": 0})),
	)).Find(&hooks)
	if err != nil {
		return nil, false, err
	}
	eligibleProjects := make(map[int64]bool)
	for _, hook := range hooks {
		if hook.ProjectID > 0 {
			if slices.Contains(hook.Events, singleName) {
				for id := range descendants[hook.ProjectID] {
					eligibleProjects[id] = true
				}
			}
			continue
		}
		allSingle = allSingle || slices.Contains(hook.Events, singleName)
		batch = batch || slices.Contains(hook.Events, batchName)
	}
	for _, task := range tasks {
		if allSingle || eligibleProjects[task.ProjectID] {
			singleTasks = append(singleTasks, task)
		}
	}
	return singleTasks, batch, nil
}

// Numeric IDs avoid bind limits for recipients spanning many projects.
func overdueProjectIDList(ids []int64) string {
	if len(ids) == 0 {
		return "NULL"
	}
	var result strings.Builder
	for i, id := range ids {
		if i > 0 {
			result.WriteByte(',')
		}
		result.WriteString(strconv.FormatInt(id, 10))
	}
	return result.String()
}
