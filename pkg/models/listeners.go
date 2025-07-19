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
	"context"
	"encoding/json"
	"strconv"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/metrics"
	"code.vikunja.io/api/pkg/modules/keyvalue"
	"code.vikunja.io/api/pkg/notifications"
	"code.vikunja.io/api/pkg/user"

	"github.com/ThreeDotsLabs/watermill/message"
	"xorm.io/builder"
	"xorm.io/xorm"
)

// RegisterListeners registers all event listeners
func RegisterListeners() {
	if config.MetricsEnabled.GetBool() {
		events.RegisterListener((&ProjectCreatedEvent{}).Name(), &IncreaseProjectCounter{})
		events.RegisterListener((&ProjectDeletedEvent{}).Name(), &DecreaseProjectCounter{})
		events.RegisterListener((&TaskCreatedEvent{}).Name(), &IncreaseTaskCounter{})
		events.RegisterListener((&TaskDeletedEvent{}).Name(), &DecreaseTaskCounter{})
		events.RegisterListener((&TeamDeletedEvent{}).Name(), &DecreaseTeamCounter{})
		events.RegisterListener((&TeamCreatedEvent{}).Name(), &IncreaseTeamCounter{})
		events.RegisterListener((&TaskAttachmentCreatedEvent{}).Name(), &IncreaseAttachmentCounter{})
		events.RegisterListener((&TaskAttachmentDeletedEvent{}).Name(), &DecreaseAttachmentCounter{})
	}
	events.RegisterListener((&TaskCommentCreatedEvent{}).Name(), &SendTaskCommentNotification{})
	events.RegisterListener((&TaskAssigneeCreatedEvent{}).Name(), &SendTaskAssignedNotification{})
	events.RegisterListener((&TaskDeletedEvent{}).Name(), &SendTaskDeletedNotification{})
	events.RegisterListener((&ProjectCreatedEvent{}).Name(), &SendProjectCreatedNotification{})
	events.RegisterListener((&TeamMemberAddedEvent{}).Name(), &SendTeamMemberAddedNotification{})
	events.RegisterListener((&TaskCommentUpdatedEvent{}).Name(), &HandleTaskCommentEditMentions{})
	events.RegisterListener((&TaskCreatedEvent{}).Name(), &HandleTaskCreateMentions{})
	events.RegisterListener((&TaskUpdatedEvent{}).Name(), &HandleTaskUpdatedMentions{})
	events.RegisterListener((&UserDataExportRequestedEvent{}).Name(), &HandleUserDataExport{})
	events.RegisterListener((&TaskCommentCreatedEvent{}).Name(), &HandleTaskUpdateLastUpdated{})
	events.RegisterListener((&TaskCommentUpdatedEvent{}).Name(), &HandleTaskUpdateLastUpdated{})
	events.RegisterListener((&TaskCommentDeletedEvent{}).Name(), &HandleTaskUpdateLastUpdated{})
	events.RegisterListener((&TaskAssigneeCreatedEvent{}).Name(), &HandleTaskUpdateLastUpdated{})
	events.RegisterListener((&TaskAssigneeDeletedEvent{}).Name(), &HandleTaskUpdateLastUpdated{})
	events.RegisterListener((&TaskAttachmentCreatedEvent{}).Name(), &HandleTaskUpdateLastUpdated{})
	events.RegisterListener((&TaskAttachmentDeletedEvent{}).Name(), &HandleTaskUpdateLastUpdated{})
	events.RegisterListener((&TaskRelationCreatedEvent{}).Name(), &HandleTaskUpdateLastUpdated{})
	events.RegisterListener((&TaskRelationDeletedEvent{}).Name(), &HandleTaskUpdateLastUpdated{})
	events.RegisterListener((&TaskCreatedEvent{}).Name(), &UpdateTaskInSavedFilterViews{})
	events.RegisterListener((&TaskUpdatedEvent{}).Name(), &UpdateTaskInSavedFilterViews{})
	if config.TypesenseEnabled.GetBool() {
		events.RegisterListener((&TaskDeletedEvent{}).Name(), &RemoveTaskFromTypesense{})
		events.RegisterListener((&TaskCreatedEvent{}).Name(), &AddTaskToTypesense{})
		events.RegisterListener((&TaskUpdatedEvent{}).Name(), &UpdateTaskInTypesense{})
		events.RegisterListener((&TaskPositionsRecalculatedEvent{}).Name(), &UpdateTaskPositionsInTypesense{})
	}
	if config.WebhooksEnabled.GetBool() {
		RegisterEventForWebhook(&TaskCreatedEvent{})
		RegisterEventForWebhook(&TaskUpdatedEvent{})
		RegisterEventForWebhook(&TaskDeletedEvent{})
		RegisterEventForWebhook(&TaskAssigneeCreatedEvent{})
		RegisterEventForWebhook(&TaskAssigneeDeletedEvent{})
		RegisterEventForWebhook(&TaskCommentCreatedEvent{})
		RegisterEventForWebhook(&TaskCommentUpdatedEvent{})
		RegisterEventForWebhook(&TaskCommentDeletedEvent{})
		RegisterEventForWebhook(&TaskAttachmentCreatedEvent{})
		RegisterEventForWebhook(&TaskAttachmentDeletedEvent{})
		RegisterEventForWebhook(&TaskRelationCreatedEvent{})
		RegisterEventForWebhook(&TaskRelationDeletedEvent{})
		RegisterEventForWebhook(&ProjectUpdatedEvent{})
		RegisterEventForWebhook(&ProjectDeletedEvent{})
		RegisterEventForWebhook(&ProjectSharedWithUserEvent{})
		RegisterEventForWebhook(&ProjectSharedWithTeamEvent{})
	}
}

//////
// Task Events

// IncreaseTaskCounter  represents a listener
type IncreaseTaskCounter struct {
}

// Name defines the name for the IncreaseTaskCounter listener
func (s *IncreaseTaskCounter) Name() string {
	return "task.counter.increase"
}

// Handle is executed when the event IncreaseTaskCounter listens on is fired
func (s *IncreaseTaskCounter) Handle(_ *message.Message) (err error) {
	return keyvalue.IncrBy(metrics.TaskCountKey, 1)
}

// DecreaseTaskCounter  represents a listener
type DecreaseTaskCounter struct {
}

// Name defines the name for the DecreaseTaskCounter listener
func (s *DecreaseTaskCounter) Name() string {
	return "task.counter.decrease"
}

// Handle is executed when the event DecreaseTaskCounter listens on is fired
func (s *DecreaseTaskCounter) Handle(_ *message.Message) (err error) {
	return keyvalue.DecrBy(metrics.TaskCountKey, 1)
}

func notifyMentionedUsers(sess *xorm.Session, task *Task, text string, n notifications.NotificationWithSubject) (users map[int64]*user.User, err error) {
	users, err = FindMentionedUsersInText(sess, text)
	if err != nil {
		return
	}

	if len(users) == 0 {
		return
	}

	log.Debugf("Processing %d mentioned users for text %d", len(users), n.SubjectID())

	var notified int
	for _, u := range users {
		can, _, err := task.CanRead(sess, u)
		if err != nil {
			return users, err
		}

		if !can {
			continue
		}

		// Don't notify a user if they were already notified
		dbn, err := notifications.GetNotificationsForNameAndUser(sess, u.ID, n.Name(), n.SubjectID())
		if err != nil {
			return users, err
		}

		if len(dbn) > 0 {
			continue
		}

		err = notifications.Notify(u, n)
		if err != nil {
			return users, err
		}
		notified++
	}

	log.Debugf("Notified %d mentioned users for text %d", notified, n.SubjectID())

	return
}

// SendTaskCommentNotification  represents a listener
type SendTaskCommentNotification struct {
}

// Name defines the name for the SendTaskCommentNotification listener
func (s *SendTaskCommentNotification) Name() string {
	return "task.comment.notification.send"
}

// Handle is executed when the event SendTaskCommentNotification listens on is fired
func (s *SendTaskCommentNotification) Handle(msg *message.Message) (err error) {
	event := &TaskCommentCreatedEvent{}
	err = json.Unmarshal(msg.Payload, event)
	if err != nil {
		return err
	}

	sess := db.NewSession()
	defer sess.Close()

	n := &TaskCommentNotification{
		Doer:      event.Doer,
		Task:      event.Task,
		Comment:   event.Comment,
		Mentioned: true,
	}
	mentionedUsers, err := notifyMentionedUsers(sess, event.Task, event.Comment.Comment, n)
	if err != nil {
		return err
	}

	subscribers, err := GetSubscriptionsForEntity(sess, SubscriptionEntityTask, event.Task.ID)
	if err != nil {
		return err
	}

	log.Debugf("Sending task comment notifications to %d subscribers for task %d", len(subscribers), event.Task.ID)

	for _, subscriber := range subscribers {
		if subscriber.UserID == event.Doer.ID {
			continue
		}

		if _, has := mentionedUsers[subscriber.UserID]; has {
			continue
		}

		n := &TaskCommentNotification{
			Doer:    event.Doer,
			Task:    event.Task,
			Comment: event.Comment,
		}
		err = notifications.Notify(subscriber.User, n)
		if err != nil {
			return
		}
	}

	return
}

// HandleTaskCommentEditMentions  represents a listener
type HandleTaskCommentEditMentions struct {
}

// Name defines the name for the HandleTaskCommentEditMentions listener
func (s *HandleTaskCommentEditMentions) Name() string {
	return "handle.task.comment.edit.mentions"
}

// Handle is executed when the event HandleTaskCommentEditMentions listens on is fired
func (s *HandleTaskCommentEditMentions) Handle(msg *message.Message) (err error) {
	event := &TaskCommentUpdatedEvent{}
	err = json.Unmarshal(msg.Payload, event)
	if err != nil {
		return err
	}

	sess := db.NewSession()
	defer sess.Close()

	n := &TaskCommentNotification{
		Doer:      event.Doer,
		Task:      event.Task,
		Comment:   event.Comment,
		Mentioned: true,
	}
	_, err = notifyMentionedUsers(sess, event.Task, event.Comment.Comment, n)
	return err
}

// SendTaskAssignedNotification  represents a listener
type SendTaskAssignedNotification struct {
}

// Name defines the name for the SendTaskAssignedNotification listener
func (s *SendTaskAssignedNotification) Name() string {
	return "task.assigned.notification.send"
}

// Handle is executed when the event SendTaskAssignedNotification listens on is fired
func (s *SendTaskAssignedNotification) Handle(msg *message.Message) (err error) {
	event := &TaskAssigneeCreatedEvent{}
	err = json.Unmarshal(msg.Payload, event)
	if err != nil {
		return err
	}

	sess := db.NewSession()
	defer sess.Close()

	subscribers, err := GetSubscriptionsForEntity(sess, SubscriptionEntityTask, event.Task.ID)
	if err != nil {
		return err
	}

	log.Debugf("Sending task assigned notifications to %d subscribers for task %d", len(subscribers), event.Task.ID)

	task, err := GetTaskByIDSimple(sess, event.Task.ID)
	if err != nil {
		return err
	}

	notifiedUsers := make(map[int64]bool)

	for _, subscriber := range subscribers {
		if subscriber.UserID == event.Doer.ID {
			continue
		}

		if notifiedUsers[subscriber.UserID] {
			// Users may be subscribed to the task and the project itself, which leads to double notifications
			continue
		}

		n := &TaskAssignedNotification{
			Doer:     event.Doer,
			Task:     &task,
			Assignee: event.Assignee,
			Target:   subscriber.User,
		}
		err = notifications.Notify(subscriber.User, n)
		if err != nil {
			return
		}

		notifiedUsers[subscriber.UserID] = true
	}

	return nil
}

// SendTaskDeletedNotification  represents a listener
type SendTaskDeletedNotification struct {
}

// Name defines the name for the SendTaskDeletedNotification listener
func (s *SendTaskDeletedNotification) Name() string {
	return "task.deleted.notification.send"
}

// Handle is executed when the event SendTaskDeletedNotification listens on is fired
func (s *SendTaskDeletedNotification) Handle(msg *message.Message) (err error) {
	event := &TaskDeletedEvent{}
	err = json.Unmarshal(msg.Payload, event)
	if err != nil {
		return err
	}

	sess := db.NewSession()
	defer sess.Close()

	var subscribers []*SubscriptionWithUser
	subscribers, err = GetSubscriptionsForEntity(sess, SubscriptionEntityTask, event.Task.ID)
	// If the task does not exist and no one has explicitly subscribed to it, we won't find any subscriptions for it.
	// Hence, we need to check for subscriptions to the parent project manually.
	if err != nil && (IsErrTaskDoesNotExist(err) || IsErrProjectDoesNotExist(err)) {
		subscribers, err = GetSubscriptionsForEntity(sess, SubscriptionEntityProject, event.Task.ProjectID)
	}
	if err != nil {
		return err
	}

	log.Debugf("Sending task deleted notifications to %d subscribers for task %d", len(subscribers), event.Task.ID)

	for _, subscriber := range subscribers {
		if subscriber.UserID == event.Doer.ID {
			continue
		}

		n := &TaskDeletedNotification{
			Doer: event.Doer,
			Task: event.Task,
		}
		err = notifications.Notify(subscriber.User, n)
		if err != nil {
			return
		}
	}

	return nil
}

// HandleTaskCreateMentions  represents a listener
type HandleTaskCreateMentions struct {
}

// Name defines the name for the HandleTaskCreateMentions listener
func (s *HandleTaskCreateMentions) Name() string {
	return "task.created.mentions"
}

// Handle is executed when the event HandleTaskCreateMentions listens on is fired
func (s *HandleTaskCreateMentions) Handle(msg *message.Message) (err error) {
	event := &TaskCreatedEvent{}
	err = json.Unmarshal(msg.Payload, event)
	if err != nil {
		return err
	}

	sess := db.NewSession()
	defer sess.Close()

	n := &UserMentionedInTaskNotification{
		Task:  event.Task,
		Doer:  event.Doer,
		IsNew: true,
	}
	_, err = notifyMentionedUsers(sess, event.Task, event.Task.Description, n)
	return err
}

// HandleTaskUpdatedMentions  represents a listener
type HandleTaskUpdatedMentions struct {
}

// Name defines the name for the HandleTaskUpdatedMentions listener
func (s *HandleTaskUpdatedMentions) Name() string {
	return "task.updated.mentions"
}

// Handle is executed when the event HandleTaskUpdatedMentions listens on is fired
func (s *HandleTaskUpdatedMentions) Handle(msg *message.Message) (err error) {
	event := &TaskUpdatedEvent{}
	err = json.Unmarshal(msg.Payload, event)
	if err != nil {
		return err
	}

	sess := db.NewSession()
	defer sess.Close()

	n := &UserMentionedInTaskNotification{
		Task:  event.Task,
		Doer:  event.Doer,
		IsNew: false,
	}

	_, err = notifyMentionedUsers(sess, event.Task, event.Task.Description, n)
	return err
}

// HandleTaskUpdateLastUpdated  represents a listener
type HandleTaskUpdateLastUpdated struct {
}

// Name defines the name for the HandleTaskUpdateLastUpdated listener
func (s *HandleTaskUpdateLastUpdated) Name() string {
	return "handle.task.update.last.updated"
}

// Handle is executed when the event HandleTaskUpdateLastUpdated listens on is fired
func (s *HandleTaskUpdateLastUpdated) Handle(msg *message.Message) (err error) {
	// Using a map here allows us to plug this listener to all kinds of task events
	event := map[string]interface{}{}
	err = json.Unmarshal(msg.Payload, &event)
	if err != nil {
		return err
	}

	task, is := event["task"].(map[string]interface{})
	if !is {
		log.Errorf("Event payload does not contain task")
		return
	}

	taskID, is := task["id"]
	if !is {
		log.Errorf("Event payload does not contain a valid task ID")
		return
	}

	var taskIDInt int64
	switch v := taskID.(type) {
	case int64:
		taskIDInt = v
	case int:
		taskIDInt = int64(v)
	case int32:
		taskIDInt = int64(v)
	case float64:
		taskIDInt = int64(v)
	case float32:
		taskIDInt = int64(v)
	default:
		log.Errorf("Event payload does not contain a valid task ID")
		return
	}

	sess := db.NewSession()
	defer sess.Close()

	return updateTaskLastUpdated(sess, &Task{ID: taskIDInt})
}

// RemoveTaskFromTypesense represents a listener
type RemoveTaskFromTypesense struct {
}

// Name defines the name for the RemoveTaskFromTypesense listener
func (s *RemoveTaskFromTypesense) Name() string {
	return "typesense.task.remove"
}

// Handle is executed when the event RemoveTaskFromTypesense listens on is fired
func (s *RemoveTaskFromTypesense) Handle(msg *message.Message) (err error) {
	event := &TaskDeletedEvent{}
	err = json.Unmarshal(msg.Payload, event)
	if err != nil {
		return err
	}

	log.Debugf("[Typesense Sync] Removing task %d from Typesense", event.Task.ID)

	_, err = typesenseClient.
		Collection("tasks").
		Document(strconv.FormatInt(event.Task.ID, 10)).
		Delete(context.Background())
	return err
}

// AddTaskToTypesense  represents a listener
type AddTaskToTypesense struct {
}

// Name defines the name for the AddTaskToTypesense listener
func (l *AddTaskToTypesense) Name() string {
	return "typesense.task.add"
}

// Handle is executed when the event AddTaskToTypesense listens on is fired
func (l *AddTaskToTypesense) Handle(msg *message.Message) (err error) {
	event := &TaskCreatedEvent{}
	err = json.Unmarshal(msg.Payload, event)
	if err != nil {
		return err
	}

	log.Debugf("New task %d created, adding to typesense…", event.Task.ID)

	s := db.NewSession()
	defer s.Close()

	task := make(map[int64]*Task, 1)
	task[event.Task.ID] = event.Task // Will be filled with all data by the Typesense connector

	return reindexTasksInTypesense(s, task)
}

// UpdateTaskInTypesense  represents a listener
type UpdateTaskInTypesense struct {
}

// Name defines the name for the UpdateTaskInTypesense listener
func (l *UpdateTaskInTypesense) Name() string {
	return "typesense.task.update"
}

// Handle is executed when the event UpdateTaskInTypesense listens on is fired
func (l *UpdateTaskInTypesense) Handle(msg *message.Message) (err error) {
	event := &TaskUpdatedEvent{}
	err = json.Unmarshal(msg.Payload, event)
	if err != nil {
		return err
	}

	s := db.NewSession()
	defer s.Close()

	task := make(map[int64]*Task, 1)
	task[event.Task.ID] = event.Task // Will be filled with all data by the Typesense connector

	return reindexTasksInTypesense(s, task)
}

// UpdateTaskPositionsInTypesense  represents a listener
type UpdateTaskPositionsInTypesense struct {
}

// Name defines the name for the UpdateTaskPositionsInTypesense listener
func (l *UpdateTaskPositionsInTypesense) Name() string {
	return "typesense.task.position.update"
}

// Handle is executed when the event UpdateTaskPositionsInTypesense listens on is fired
func (l *UpdateTaskPositionsInTypesense) Handle(msg *message.Message) (err error) {
	event := &TaskPositionsRecalculatedEvent{}
	err = json.Unmarshal(msg.Payload, event)
	if err != nil {
		return err
	}

	taskIDs := []int64{}
	for _, position := range event.NewTaskPositions {
		taskIDs = append(taskIDs, position.TaskID)
	}

	s := db.NewSession()
	defer s.Close()

	tasks, err := GetTasksSimpleByIDs(s, taskIDs)

	taskMap := make(map[int64]*Task, 1)
	for _, task := range tasks {
		taskMap[task.ID] = task
	}

	return reindexTasksInTypesense(s, taskMap)
}

// IncreaseAttachmentCounter  represents a listener
type IncreaseAttachmentCounter struct {
}

// Name defines the name for the IncreaseAttachmentCounter listener
func (s *IncreaseAttachmentCounter) Name() string {
	return "increase.attachment.counter"
}

// Handle is executed when the event IncreaseAttachmentCounter listens on is fired
func (s *IncreaseAttachmentCounter) Handle(_ *message.Message) (err error) {
	return keyvalue.IncrBy(metrics.AttachmentsCountKey, 1)
}

// DecreaseAttachmentCounter  represents a listener
type DecreaseAttachmentCounter struct {
}

// Name defines the name for the DecreaseAttachmentCounter listener
func (s *DecreaseAttachmentCounter) Name() string {
	return "decrease.attachment.counter"
}

// Handle is executed when the event DecreaseAttachmentCounter listens on is fired
func (s *DecreaseAttachmentCounter) Handle(_ *message.Message) (err error) {
	return keyvalue.DecrBy(metrics.AttachmentsCountKey, 1)
}

// UpdateTaskInSavedFilterViews  represents a listener
type UpdateTaskInSavedFilterViews struct {
}

// Name defines the name for the UpdateTaskInSavedFilterViews listener
func (l *UpdateTaskInSavedFilterViews) Name() string {
	return "task.set.saved.filter.views"
}

// Handle is executed when the event UpdateTaskInSavedFilterViews listens on is fired
func (l *UpdateTaskInSavedFilterViews) Handle(msg *message.Message) (err error) {
	event := &TaskCreatedEvent{}
	err = json.Unmarshal(msg.Payload, event)
	if err != nil {
		return err
	}

	// This operation is potentially very resource-heavy, because we don't know if a task is included
	// in a filter until we evaluate that filter. We need to evaluate each filter individually - since
	// there can be many filters, this can take a while to execute.
	// For this reason, we do this in an asynchronous event listener.

	s := db.NewSession()
	defer s.Close()

	// Get all saved filters with a manual kanban view
	kanbanFilterViews := []*ProjectView{}
	err = s.Where("project_id < 0 and view_kind = ? and bucket_configuration_mode = ?", ProjectViewKindKanban, BucketConfigurationModeManual).
		Find(&kanbanFilterViews)
	if err != nil {
		return err
	}

	filterIDs := []int64{}
	for _, view := range kanbanFilterViews {
		filterIDs = append(filterIDs, GetSavedFilterIDFromProjectID(view.ProjectID))
	}

	filters := map[int64]*SavedFilter{}
	err = s.In("id", filterIDs).Find(&filters)
	if err != nil {
		return err
	}

	var fallbackTimezone string
	if event.Doer != nil {
		var u *user.User
		u, err = user.GetUserByID(s, event.Doer.GetID())
		if err == nil {
			fallbackTimezone = u.Timezone
			// When a link share triggered this event, the user id will be 0, and thus this fails.
			// Only passing the value along when the user was retrieved successfully ensures the whole handler
			// does not fail because of that.
			// When the fallback is empty, it will be handled later anyhow.
		}
	}

	taskBuckets := []*TaskBucket{}
	taskPositions := []*TaskPosition{}

	viewIDToCleanUp := []int64{}

	for _, view := range kanbanFilterViews {
		filter, exists := filters[GetSavedFilterIDFromProjectID(view.ProjectID)]
		if !exists {
			log.Debugf("Did not find filter for view %d", view.ID)
			continue
		}

		taskBucket, taskPosition, err := addTaskToFilter(s, filter, view, fallbackTimezone, event.Task)
		if err != nil {
			if IsErrInvalidFilterExpression(err) ||
				IsErrInvalidTaskFilterValue(err) ||
				IsErrInvalidTaskFilterConcatinator(err) ||
				IsErrInvalidTaskFilterComparator(err) ||
				IsErrInvalidTaskField(err) {
				log.Debugf("Invalid filter expression for view %d, expression: %v", view.ID, view.Filter)
				continue
			}

			return err
		}

		if taskBucket != nil && taskPosition != nil {
			taskBuckets = append(taskBuckets, taskBucket)
			taskPositions = append(taskPositions, taskPosition)
			viewIDToCleanUp = append(viewIDToCleanUp, view.ID)
		}
	}

	if len(taskBuckets) > 0 || len(taskPositions) > 0 {
		_, err = s.And(
			builder.Eq{"task_id": event.Task.ID},
			builder.In("project_view_id", viewIDToCleanUp),
		).
			Delete(&TaskBucket{})
		if err != nil {
			return
		}
		_, err = s.And(
			builder.Eq{"task_id": event.Task.ID},
			builder.In("project_view_id", viewIDToCleanUp),
		).
			Delete(&TaskPosition{})
		if err != nil {
			return
		}
		_, err = s.Insert(taskBuckets)
		if err != nil {
			return
		}
		_, err = s.Insert(taskPositions)
		if err != nil {
			return
		}

		task := make(map[int64]*Task, 1)
		task[event.Task.ID] = event.Task // Will be filled with all data by the Typesense connector

		return reindexTasksInTypesense(s, task)
	}

	return nil
}

///////
// Project Event Listeners

type IncreaseProjectCounter struct {
}

func (s *IncreaseProjectCounter) Name() string {
	return "project.counter.increase"
}

func (s *IncreaseProjectCounter) Handle(_ *message.Message) (err error) {
	return keyvalue.IncrBy(metrics.ProjectCountKey, 1)
}

type DecreaseProjectCounter struct {
}

func (s *DecreaseProjectCounter) Name() string {
	return "project.counter.decrease"
}

func (s *DecreaseProjectCounter) Handle(_ *message.Message) (err error) {
	return keyvalue.DecrBy(metrics.ProjectCountKey, 1)
}

// SendProjectCreatedNotification  represents a listener
type SendProjectCreatedNotification struct {
}

// Name defines the name for the SendProjectCreatedNotification listener
func (s *SendProjectCreatedNotification) Name() string {
	return "send.project.created.notification"
}

// Handle is executed when the event SendProjectCreatedNotification listens on is fired
func (s *SendProjectCreatedNotification) Handle(msg *message.Message) (err error) {
	event := &ProjectCreatedEvent{}
	err = json.Unmarshal(msg.Payload, event)
	if err != nil {
		return err
	}

	sess := db.NewSession()
	defer sess.Close()

	subscribers, err := GetSubscriptionsForEntity(sess, SubscriptionEntityProject, event.Project.ID)
	if err != nil {
		return err
	}

	log.Debugf("Sending project created notifications to %d subscribers for project %d", len(subscribers), event.Project.ID)

	for _, subscriber := range subscribers {
		if subscriber.UserID == event.Doer.ID {
			continue
		}

		n := &ProjectCreatedNotification{
			Doer:    event.Doer,
			Project: event.Project,
		}
		err = notifications.Notify(subscriber.User, n)
		if err != nil {
			return
		}
	}

	return nil
}

// WebhookListener represents a listener
type WebhookListener struct {
	EventName string
}

// Name defines the name for the WebhookListener listener
func (wl *WebhookListener) Name() string {
	return "webhook.listener"
}

type WebhookPayload struct {
	EventName string      `json:"event_name"`
	Time      time.Time   `json:"time"`
	Data      interface{} `json:"data"`
}

func getIDAsInt64(id interface{}) int64 {
	switch v := id.(type) {
	case int64:
		return v
	case float64:
		return int64(v)
	}
	return id.(int64)
}

func getProjectIDFromAnyEvent(eventPayload map[string]interface{}) int64 {
	if task, has := eventPayload["task"]; has {
		t := task.(map[string]interface{})
		if projectID, has := t["project_id"]; has {
			return getIDAsInt64(projectID)
		}
	}

	if project, has := eventPayload["project"]; has {
		t := project.(map[string]interface{})
		if projectID, has := t["id"]; has {
			return getIDAsInt64(projectID)
		}
	}

	return 0
}

func reloadEventData(s *xorm.Session, event map[string]interface{}, projectID int64) (eventWithData map[string]interface{}, doerID int64, err error) {
	// Load event data again so that it is always populated in the webhook payload
	if doer, has := event["doer"]; has {
		d := doer.(map[string]interface{})
		if rawDoerID, has := d["id"]; has {
			doerID = getIDAsInt64(rawDoerID)
			if doerID > 0 {
				fullDoer, err := user.GetUserByID(s, doerID)
				if err != nil && !user.IsErrUserDoesNotExist(err) {
					return nil, 0, err
				}
				if err == nil {
					event["doer"] = fullDoer
				}
			}
		}
	}

	if task, has := event["task"]; has && doerID != 0 {
		t := task.(map[string]interface{})
		if taskID, has := t["id"]; has {
			id := getIDAsInt64(taskID)
			fullTask := Task{
				ID: id,
				Expand: []TaskCollectionExpandable{
					TaskCollectionExpandBuckets,
				},
			}
			err = fullTask.ReadOne(s, &user.User{ID: doerID})
			if err != nil && !IsErrTaskDoesNotExist(err) {
				return
			}
			if err == nil {
				event["task"] = fullTask
			}
		}
	}

	if _, has := event["project"]; has && doerID != 0 {
		project := &Project{ID: projectID}
		err = project.ReadOne(s, &user.User{ID: doerID})
		if err != nil && !IsErrProjectDoesNotExist(err) {
			return
		}
		if err == nil {
			event["project"] = project
		}
	}

	return event, doerID, nil
}

// Handle is executed when the event WebhookListener listens on is fired
func (wl *WebhookListener) Handle(msg *message.Message) (err error) {
	var event map[string]interface{}
	err = json.Unmarshal(msg.Payload, &event)
	if err != nil {
		return err
	}

	projectID := getProjectIDFromAnyEvent(event)
	if projectID == 0 {
		log.Debugf("event %s does not contain a project id, not handling webhook", wl.EventName)
		return nil
	}

	s := db.NewSession()
	defer s.Close()

	parents, err := GetAllParentProjects(s, projectID)
	if err != nil {
		return err
	}

	projectIDs := make([]int64, 0, len(parents)+1)
	projectIDs = append(projectIDs, projectID)

	for _, p := range parents {
		projectIDs = append(projectIDs, p.ID)
	}

	ws := []*Webhook{}
	err = s.In("project_id", projectIDs).
		Find(&ws)
	if err != nil {
		return err
	}

	matchingWebhooks := []*Webhook{}
	for _, w := range ws {
		for _, e := range w.Events {
			if e == wl.EventName {
				matchingWebhooks = append(matchingWebhooks, w)
				break
			}
		}
	}

	if len(matchingWebhooks) == 0 {
		log.Debugf("Did not find any webhook for the %s event for project %d, not sending", wl.EventName, projectID)
		return nil
	}

	var doerID int64
	event, doerID, err = reloadEventData(s, event, projectID)
	if err != nil {
		return err
	}

	for _, webhook := range matchingWebhooks {

		if _, has := event["project"]; !has {
			project := &Project{ID: webhook.ProjectID}
			err = project.ReadOne(s, &user.User{ID: doerID})
			if err != nil && !IsErrProjectDoesNotExist(err) {
				log.Errorf("Could not load project for webhook %d: %s", webhook.ID, err)
			}
			if err == nil {
				event["project"] = project
			}
		}

		err = webhook.sendWebhookPayload(&WebhookPayload{
			EventName: wl.EventName,
			Time:      time.Now(),
			Data:      event,
		})
		if err != nil {
			return err
		}
	}

	return
}

///////
// Team Events

// IncreaseTeamCounter  represents a listener
type IncreaseTeamCounter struct {
}

// Name defines the name for the IncreaseTeamCounter listener
func (s *IncreaseTeamCounter) Name() string {
	return "team.counter.increase"
}

// Handle is executed when the event IncreaseTeamCounter listens on is fired
func (s *IncreaseTeamCounter) Handle(_ *message.Message) (err error) {
	return keyvalue.IncrBy(metrics.TeamCountKey, 1)
}

// DecreaseTeamCounter  represents a listener
type DecreaseTeamCounter struct {
}

// Name defines the name for the DecreaseTeamCounter listener
func (s *DecreaseTeamCounter) Name() string {
	return "team.counter.decrease"
}

// Handle is executed when the event DecreaseTeamCounter listens on is fired
func (s *DecreaseTeamCounter) Handle(_ *message.Message) (err error) {
	return keyvalue.DecrBy(metrics.TeamCountKey, 1)
}

// SendTeamMemberAddedNotification  represents a listener
type SendTeamMemberAddedNotification struct {
}

// Name defines the name for the SendTeamMemberAddedNotification listener
func (s *SendTeamMemberAddedNotification) Name() string {
	return "team.member.added.notification"
}

// Handle is executed when the event SendTeamMemberAddedNotification listens on is fired
func (s *SendTeamMemberAddedNotification) Handle(msg *message.Message) (err error) {
	event := &TeamMemberAddedEvent{}
	err = json.Unmarshal(msg.Payload, event)
	if err != nil {
		return err
	}

	// Don't notify the user themselves
	if event.Doer.ID == event.Member.ID {
		return nil
	}

	return notifications.Notify(event.Member, &TeamMemberAddedNotification{
		Member: event.Member,
		Doer:   event.Doer,
		Team:   event.Team,
	})
}

// HandleUserDataExport  represents a listener
type HandleUserDataExport struct {
}

// Name defines the name for the HandleUserDataExport listener
func (s *HandleUserDataExport) Name() string {
	return "handle.user.data.export"
}

// Handle is executed when the event HandleUserDataExport listens on is fired
func (s *HandleUserDataExport) Handle(msg *message.Message) (err error) {
	event := &UserDataExportRequestedEvent{}
	err = json.Unmarshal(msg.Payload, event)
	if err != nil {
		return err
	}

	log.Debugf("Starting to export user data for user %d...", event.User.ID)

	sess := db.NewSession()
	defer sess.Close()
	err = sess.Begin()
	if err != nil {
		return
	}

	err = ExportUserData(sess, event.User)
	if err != nil {
		_ = sess.Rollback()
		return
	}

	log.Debugf("Done exporting user data for user %d...", event.User.ID)

	err = sess.Commit()
	return err
}
