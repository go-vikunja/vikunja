# User-Level Webhooks Implementation Plan

**Goal:** Allow users to configure webhook URLs in their user settings that receive task reminder and overdue events across all projects, reusing the existing project-level webhook infrastructure (HMAC signing, basic auth, proxy, retry).

**Architecture:** Extend the existing `webhooks` table with a nullable `user_id` column (making `project_id` nullable too). Add a `User` field to the existing reminder/overdue events so the `WebhookListener` can look up user-level webhooks. Add a `TasksOverdueEvent` batch event. Extract the frontend webhook form into a shared component.

**Tech Stack:** Go (XORM, Watermill events), Vue 3 + TypeScript, Bulma CSS

**What main already provides:**
- `TaskReminderFiredEvent` and `TaskOverdueEvent` event types (with `Task` + `Project`, but NO `User` field)
- Both registered with `RegisterEventForWebhook()` in `RegisterListeners()`
- Both cron jobs dispatch events and gate email on config checks
- `getTasksWithRemindersDueAndTheirUsers(s, now)` and `getUndoneOverdueTasks(s, now)` — no `cond` parameter

---

### Task 1: Clean up PR branch — revert files to main-branch state

This task removes the separate user-webhook-settings infrastructure from the current PR branch. The useful frontend scaffolding from the PR is preserved.

**Keep from PR (do not touch):**
- `frontend/src/router/index.ts` — webhooks route entry
- `frontend/src/views/user/Settings.vue` — webhooks nav item
- `frontend/src/stores/config.ts` — `webhooksEnabled` field
- `frontend/src/views/user/settings/Webhooks.vue` — page shell (reworked in Task 10)
- `frontend/src/i18n/lang/en.json` — webhook i18n strings (adapted in Task 10)
- `frontend/src/models/userSettings.ts` — webhook fields (harmless, may be useful later)

**Delete (not reusable — wrong data model or duplicated infrastructure):**
- `pkg/models/user_webhook_setting.go`
- `pkg/notifications/webhook.go`
- `pkg/migration/20260128170701.go`
- `pkg/routes/api/v1/user_webhook_settings.go`
- `frontend/src/models/userWebhookSetting.ts`
- `frontend/src/modelTypes/IUserWebhookSetting.ts`
- `frontend/src/services/userWebhookSettings.ts`

**Revert to main (changes not needed with the new approach):**
- `pkg/notifications/notification.go`
- `pkg/user/user.go`
- `pkg/models/notifications.go`
- `pkg/models/models.go`

**Step 1: Delete the files that are no longer needed**

```bash
rm pkg/models/user_webhook_setting.go
rm pkg/notifications/webhook.go
rm pkg/migration/20260128170701.go
rm pkg/routes/api/v1/user_webhook_settings.go
rm frontend/src/models/userWebhookSetting.ts
rm frontend/src/modelTypes/IUserWebhookSetting.ts
rm frontend/src/services/userWebhookSettings.ts
```

**Step 2: Revert modified files to main-branch state**

```bash
git checkout main -- pkg/notifications/notification.go
git checkout main -- pkg/user/user.go
git checkout main -- pkg/models/notifications.go
git checkout main -- pkg/models/models.go
git checkout main -- pkg/models/listeners.go
```

Note: Do NOT revert `task_reminder.go`, `task_overdue_reminder.go`, `events.go`, or `webhooks.go` — those already have the correct main-branch state after the rebase.

**Step 3: Revert user settings route registration**

In `pkg/routes/routes.go`, remove the user webhook settings routes (lines 382-387):

```go
// Remove these lines:
// Webhook settings
u.GET("/settings/webhooks", apiv1.GetUserWebhookSettings)
u.GET("/settings/webhooks/types", apiv1.GetAvailableWebhookNotificationTypes)
u.GET("/settings/webhooks/:type", apiv1.GetUserWebhookSettingByType)
u.PUT("/settings/webhooks/:type", apiv1.CreateOrUpdateUserWebhookSetting)
u.DELETE("/settings/webhooks/:type", apiv1.DeleteUserWebhookSetting)
```

**Step 4: Verify it compiles**

```bash
mage build
```

**Step 5: Run existing tests**

```bash
mage test:feature
```

**Step 6: Commit**

```bash
git add -A
git commit -m "refactor: remove separate user webhook settings system

Removes the user_webhook_settings table, notification-level webhook
delivery code, and related API routes. This will be replaced by
extending the existing project-level webhook infrastructure."
```

---

### Task 2: Database migration — add user_id to webhooks table

**Files:**
- Create: `pkg/migration/<timestamp>.go` (use `mage dev:make-migration` to get correct timestamp)
- Modify: `pkg/models/webhooks.go`

**Step 1: Create the migration**

```bash
mage dev:make-migration WebhookUserID
```

Then edit the generated file to contain:

```go
package migration

import (
	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "<generated-timestamp>",
		Description: "Add user_id to webhooks table and make project_id nullable",
		Migrate: func(tx *xorm.Engine) error {
			exists, err := columnExists(tx, "webhooks", "user_id")
			if err != nil {
				return err
			}
			if !exists {
				if _, err = tx.Exec("ALTER TABLE webhooks ADD COLUMN user_id bigint NULL"); err != nil {
					return err
				}
			}

			if _, err = tx.Exec("CREATE INDEX IF NOT EXISTS IDX_webhooks_user_id ON webhooks (user_id)"); err != nil {
				_ = err
			}

			return nil
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
```

**Step 2: Update the Webhook struct in `pkg/models/webhooks.go`**

Add the `UserID` field and make `ProjectID` nullable. The current struct on main (`pkg/models/webhooks.go:47-73`) has `ProjectID` as `xorm:"bigint not null index"`. Change to:

```go
	ProjectID int64 `xorm:"bigint null index" json:"project_id" param:"project"`
	// The user ID if this is a user-level webhook (mutually exclusive with ProjectID)
	UserID int64 `xorm:"bigint null index" json:"user_id"`
```

**Step 3: Add validation to Webhook.Create**

At the top of the `Create` method (currently at `pkg/models/webhooks.go:121`), add:

```go
	// Validate that exactly one of ProjectID or UserID is set
	if w.ProjectID == 0 && w.UserID == 0 {
		return InvalidFieldError([]string{"project_id", "user_id"})
	}
	if w.ProjectID != 0 && w.UserID != 0 {
		return InvalidFieldError([]string{"project_id", "user_id"})
	}
```

**Step 4: Verify it compiles and tests pass**

```bash
mage build && mage test:feature
```

**Step 5: Commit**

```bash
git add pkg/migration/<timestamp>.go pkg/models/webhooks.go
git commit -m "feat: add user_id column to webhooks table

Adds nullable user_id to support user-level webhooks alongside
project-level ones. Application-level validation ensures exactly
one of project_id or user_id is set."
```

---

### Task 3: Add User field to existing events, add TasksOverdueEvent

Main already has `TaskReminderFiredEvent` and `TaskOverdueEvent` but they only carry `Task` + `Project`. We need to add a `User` field so `WebhookListener` can extract the target user ID. We also add `TasksOverdueEvent` (batch) and a `Reminder` field to `TaskReminderFiredEvent`.

**Files:**
- Modify: `pkg/models/events.go`

**Step 1: Update existing event structs and add TasksOverdueEvent**

In `pkg/models/events.go`, find `TaskReminderFiredEvent` (currently around line 170) and update:

```go
// TaskReminderFiredEvent represents an event where a task reminder has fired
type TaskReminderFiredEvent struct {
	Task     *Task         `json:"task"`
	User     *user.User    `json:"user"`
	Project  *Project      `json:"project"`
	Reminder *TaskReminder `json:"reminder"`
}
```

Find `TaskOverdueEvent` and update:

```go
// TaskOverdueEvent represents an event where a task is overdue
type TaskOverdueEvent struct {
	Task    *Task      `json:"task"`
	User    *user.User `json:"user"`
	Project *Project   `json:"project"`
}
```

Add `TasksOverdueEvent` after `TaskOverdueEvent`:

```go
// TasksOverdueEvent represents an event where multiple tasks are overdue for a user
type TasksOverdueEvent struct {
	Tasks    []*Task            `json:"tasks"`
	User     *user.User         `json:"user"`
	Projects map[int64]*Project `json:"projects"`
}

// Name defines the name for TasksOverdueEvent
func (t *TasksOverdueEvent) Name() string {
	return "tasks.overdue"
}
```

**Step 2: Verify it compiles**

```bash
mage build
```

**Step 3: Commit**

```bash
git add pkg/models/events.go
git commit -m "feat: add User field to reminder/overdue events, add TasksOverdueEvent

Extends TaskReminderFiredEvent and TaskOverdueEvent with a User field
so user-level webhooks can look up the target user. Adds Reminder
field to TaskReminderFiredEvent. Adds TasksOverdueEvent for batch
overdue notifications."
```

---

### Task 4: Add user-directed event tracking

Main already registers `TaskReminderFiredEvent` and `TaskOverdueEvent` with `RegisterEventForWebhook`. We need to add tracking for which events are user-directed (for user-level webhook lookups) and register `TasksOverdueEvent`.

**Files:**
- Modify: `pkg/models/webhooks.go` (add `userDirectedWebhookEvents` map and helpers)
- Modify: `pkg/models/listeners.go` (change registration calls, add `TasksOverdueEvent`)

**Step 1: Add user-directed events tracking in `pkg/models/webhooks.go`**

In the `init()` function (currently at `pkg/models/webhooks.go:82-85`), add initialization:

```go
var userDirectedWebhookEvents map[string]bool

func init() {
	availableWebhookEvents = make(map[string]bool)
	availableWebhookEventsLock = &sync.Mutex{}
	userDirectedWebhookEvents = make(map[string]bool)
}
```

Add new functions after `GetAvailableWebhookEvents()` (currently at line 97):

```go
// RegisterUserDirectedEventForWebhook registers an event as both a webhook event and a user-directed event
func RegisterUserDirectedEventForWebhook(event events.Event) {
	RegisterEventForWebhook(event)
	availableWebhookEventsLock.Lock()
	defer availableWebhookEventsLock.Unlock()
	userDirectedWebhookEvents[event.Name()] = true
}

// IsUserDirectedEvent returns whether an event name is user-directed
func IsUserDirectedEvent(eventName string) bool {
	availableWebhookEventsLock.Lock()
	defer availableWebhookEventsLock.Unlock()
	return userDirectedWebhookEvents[eventName]
}

// GetUserDirectedWebhookEvents returns a sorted list of user-directed webhook event names
func GetUserDirectedWebhookEvents() []string {
	availableWebhookEventsLock.Lock()
	defer availableWebhookEventsLock.Unlock()

	evts := []string{}
	for e := range userDirectedWebhookEvents {
		evts = append(evts, e)
	}
	sort.Strings(evts)
	return evts
}
```

**Step 2: Update registrations in `pkg/models/listeners.go`**

In `RegisterListeners()`, change the two existing reminder event registrations from `RegisterEventForWebhook` to `RegisterUserDirectedEventForWebhook`, and add `TasksOverdueEvent`. Find (currently around line 97-98):

```go
		RegisterEventForWebhook(&TaskReminderFiredEvent{})
		RegisterEventForWebhook(&TaskOverdueEvent{})
```

Replace with:

```go
		RegisterUserDirectedEventForWebhook(&TaskReminderFiredEvent{})
		RegisterUserDirectedEventForWebhook(&TaskOverdueEvent{})
		RegisterUserDirectedEventForWebhook(&TasksOverdueEvent{})
```

**Step 3: Verify it compiles**

```bash
mage build
```

**Step 4: Commit**

```bash
git add pkg/models/webhooks.go pkg/models/listeners.go
git commit -m "feat: add user-directed webhook event tracking

Adds RegisterUserDirectedEventForWebhook, IsUserDirectedEvent, and
GetUserDirectedWebhookEvents. Re-registers reminder/overdue events
as user-directed. Registers new TasksOverdueEvent."
```

---

### Task 5: Extend WebhookListener to handle user-level webhooks

**Files:**
- Modify: `pkg/models/listeners.go` (add `getUserIDFromAnyEvent`, `reloadUserInEvent`, extend `Handle`)

**Step 1: Add getUserIDFromAnyEvent function**

After `getProjectIDFromAnyEvent` (currently around line 891), add:

```go
func getUserIDFromAnyEvent(eventPayload map[string]interface{}) int64 {
	if u, has := eventPayload["user"]; has {
		userMap, ok := u.(map[string]interface{})
		if !ok {
			return 0
		}
		if userID, has := userMap["id"]; has {
			return getIDAsInt64(userID)
		}
	}

	return 0
}
```

**Step 2: Add reloadUserInEvent function**

After `reloadAssigneeInEvent`, add:

```go
func reloadUserInEvent(s *xorm.Session, event map[string]interface{}) error {
	u, has := event["user"]
	if !has || u == nil {
		return nil
	}

	userMap, ok := u.(map[string]interface{})
	if !ok {
		return nil
	}

	userID := getIDAsInt64(userMap["id"])
	if userID <= 0 {
		return nil
	}

	fullUser, err := user.GetUserByID(s, userID)
	if err != nil && !user.IsErrUserDoesNotExist(err) {
		return err
	}
	if err == nil {
		event["user"] = fullUser
	}

	return nil
}
```

Add a call to it in `reloadEventData`, after the `reloadAssigneeInEvent` call:

```go
	err = reloadUserInEvent(s, event)
	if err != nil {
		return nil, doerID, err
	}
```

**Step 3: Extend WebhookListener.Handle()**

The current `Handle()` method returns early when `projectID == 0`. We need to allow user-directed events to proceed even without a project ID. Replace the `Handle` method with:

```go
func (wl *WebhookListener) Handle(msg *message.Message) (err error) {
	var event map[string]interface{}
	err = json.Unmarshal(msg.Payload, &event)
	if err != nil {
		return err
	}

	s := db.NewSession()
	defer s.Close()

	projectID := getProjectIDFromAnyEvent(event)
	isUserDirected := IsUserDirectedEvent(wl.EventName)

	// For non-user-directed events, we need a project ID
	if projectID == 0 && !isUserDirected {
		log.Debugf("event %s does not contain a project id, not handling webhook", wl.EventName)
		return nil
	}

	// Look up project-level webhooks
	matchingWebhooks := []*Webhook{}
	if projectID > 0 {
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
		err = s.Where("project_id IS NOT NULL").
			In("project_id", projectIDs).
			Find(&ws)
		if err != nil {
			return err
		}

		for _, w := range ws {
			for _, e := range w.Events {
				if e == wl.EventName {
					matchingWebhooks = append(matchingWebhooks, w)
					break
				}
			}
		}
	}

	// Look up user-level webhooks for user-directed events
	if isUserDirected {
		userID := getUserIDFromAnyEvent(event)
		if userID > 0 {
			userWebhooks := []*Webhook{}
			err = s.Where("user_id = ? AND project_id IS NULL", userID).
				Find(&userWebhooks)
			if err != nil {
				return err
			}

			for _, w := range userWebhooks {
				for _, e := range w.Events {
					if e == wl.EventName {
						matchingWebhooks = append(matchingWebhooks, w)
						break
					}
				}
			}
		}
	}

	if len(matchingWebhooks) == 0 {
		log.Debugf("Did not find any webhook for the %s event, not sending", wl.EventName)
		return nil
	}

	var doerID int64
	event, doerID, err = reloadEventData(s, event, projectID)
	if err != nil {
		return err
	}

	for _, webhook := range matchingWebhooks {
		if _, has := event["project"]; !has && webhook.ProjectID > 0 {
			project, err := GetProjectSimpleByID(s, webhook.ProjectID)
			if err != nil && !IsErrProjectDoesNotExist(err) {
				log.Errorf("Could not load project for webhook %d: %s", webhook.ID, err)
			}
			if project != nil {
				err = project.ReadOne(s, &user.User{ID: doerID})
				if err != nil && !IsErrProjectDoesNotExist(err) {
					log.Errorf("Could not load project for webhook %d: %s", webhook.ID, err)
				}
				if err == nil {
					event["project"] = project
				}
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
```

**Step 4: Verify it compiles**

```bash
mage build
```

**Step 5: Commit**

```bash
git add pkg/models/listeners.go
git commit -m "feat: extend WebhookListener to handle user-level webhooks

Adds getUserIDFromAnyEvent and reloadUserInEvent. For user-directed
events, Handle() now also queries webhooks by user_id in addition
to project-level webhooks."
```

---

### Task 6: Update permission checks for user-level webhooks

**Files:**
- Modify: `pkg/models/webhooks_permissions.go`

**Step 1: Update permission methods to handle both project and user webhooks**

Replace the contents of `pkg/models/webhooks_permissions.go`:

```go
package models

import (
	"code.vikunja.io/api/pkg/web"
	"xorm.io/xorm"
)

func (w *Webhook) CanRead(s *xorm.Session, a web.Auth) (bool, int, error) {
	// User-level webhook: user owns it
	if w.UserID > 0 {
		return w.UserID == a.GetID(), int(PermissionRead), nil
	}

	// Project-level webhook: delegate to project
	p := &Project{ID: w.ProjectID}
	return p.CanRead(s, a)
}

func (w *Webhook) CanDelete(s *xorm.Session, a web.Auth) (bool, error) {
	return w.canDoWebhook(s, a)
}

func (w *Webhook) CanUpdate(s *xorm.Session, a web.Auth) (bool, error) {
	return w.canDoWebhook(s, a)
}

func (w *Webhook) CanCreate(s *xorm.Session, a web.Auth) (bool, error) {
	return w.canDoWebhook(s, a)
}

func (w *Webhook) canDoWebhook(s *xorm.Session, a web.Auth) (bool, error) {
	_, isShareAuth := a.(*LinkSharing)
	if isShareAuth {
		return false, nil
	}

	// User-level webhook: user owns it or is creating new
	if w.UserID > 0 || w.ProjectID == 0 {
		return w.UserID == 0 || w.UserID == a.GetID(), nil
	}

	// Project-level webhook: delegate to project
	p := &Project{ID: w.ProjectID}
	return p.CanUpdate(s, a)
}
```

**Step 2: Verify it compiles**

```bash
mage build
```

**Step 3: Commit**

```bash
git add pkg/models/webhooks_permissions.go
git commit -m "feat: update webhook permissions for user-level webhooks

Permission methods now handle both project-level and user-level
webhooks. User-level webhooks are owned by the user who created them."
```

---

### Task 7: Update cron jobs to include User in event payloads and add TasksOverdueEvent

Main already dispatches events from both cron jobs, but without the `User` field. We need to update the dispatch calls to include `User`, add `Reminder` to the reminder event, add `TasksOverdueEvent` dispatch, and re-introduce the `cond` parameter so all users are fetched when webhooks are enabled.

**Files:**
- Modify: `pkg/models/task_reminder.go`
- Modify: `pkg/models/task_overdue_reminder.go`
- Modify: `pkg/models/notifications.go` (add `TaskReminder` field to `ReminderDueNotification`)

**Step 1: Add TaskReminder to ReminderDueNotification**

In `pkg/models/notifications.go`, add `TaskReminder` to `ReminderDueNotification`:

```go
type ReminderDueNotification struct {
	User         *user.User    `json:"user,omitempty"`
	Task         *Task         `json:"task"`
	Project      *Project      `json:"project"`
	TaskReminder *TaskReminder `json:"reminder"`
}
```

**Step 2: Update task_reminder.go**

Add `cond builder.Cond` parameter to `getTasksWithRemindersDueAndTheirUsers`:

```go
func getTasksWithRemindersDueAndTheirUsers(s *xorm.Session, now time.Time, cond builder.Cond) (reminderNotifications []*ReminderDueNotification, err error) {
```

Change the hardcoded filter call (currently `getTaskUsersForTasks(s, taskIDs, builder.Eq{"users.email_reminders_enabled": true})`) to:

```go
	usersWithReminders, err := getTaskUsersForTasks(s, taskIDs, cond)
```

Update the notification construction to include `TaskReminder`:

```go
				reminderNotifications = append(reminderNotifications, &ReminderDueNotification{
					User:         u.User,
					Task:         u.Task,
					Project:      projects[u.Task.ProjectID],
					TaskReminder: r,
				})
```

Replace `RegisterReminderCron` (currently dispatches events without `User`):

```go
func RegisterReminderCron() {
	webhookEnabled := config.WebhooksEnabled.GetBool()
	emailEnabled := config.ServiceEnableEmailReminders.GetBool() && config.MailerEnabled.GetBool()

	if !emailEnabled && !webhookEnabled {
		return
	}

	if !emailEnabled {
		log.Info("Mailer is disabled, not sending reminders per mail")
	}

	tz := config.GetTimeZone()
	log.Debugf("[Task Reminder Cron] Timezone is %s", tz)

	err := cron.Schedule("* * * * *", func() {
		s := db.NewSession()
		defer s.Close()

		now := time.Now()

		// When only email is enabled, filter to email-enabled users for efficiency.
		// When webhooks are enabled, we need all users so the event system can
		// look up matching webhooks.
		var cond builder.Cond
		if emailEnabled && !webhookEnabled {
			cond = builder.Eq{"users.email_reminders_enabled": true}
		}

		reminders, err := getTasksWithRemindersDueAndTheirUsers(s, now, cond)
		if err != nil {
			log.Errorf("[Task Reminder Cron] Could not get tasks with reminders in the next minute: %s", err)
			return
		}

		if len(reminders) == 0 {
			return
		}

		log.Debugf("[Task Reminder Cron] Sending %d reminders", len(reminders))

		for _, n := range reminders {
			if emailEnabled && n.User.EmailRemindersEnabled {
				err = notifications.Notify(n.User, n)
				if err != nil {
					log.Errorf("[Task Reminder Cron] Could not notify user %d: %s", n.User.ID, err)
					return
				}
			}

			if webhookEnabled {
				err = events.Dispatch(&TaskReminderFiredEvent{
					Task:     n.Task,
					User:     n.User,
					Project:  n.Project,
					Reminder: n.TaskReminder,
				})
				if err != nil {
					log.Errorf("[Task Reminder Cron] Could not dispatch reminder event for task %d: %s", n.Task.ID, err)
				}
			}

			log.Debugf("[Task Reminder Cron] Sent reminder for task %d to user %d", n.Task.ID, n.User.ID)
		}
	})
	if err != nil {
		log.Fatalf("Could not register reminder cron: %s", err)
	}
}
```

**Step 3: Update task_overdue_reminder.go**

Add `cond builder.Cond` parameter to `getUndoneOverdueTasks`:

```go
func getUndoneOverdueTasks(s *xorm.Session, now time.Time, cond builder.Cond) (usersWithTasks map[int64]*userWithTasks, err error) {
```

Change the hardcoded filter call to:

```go
	users, err := getTaskUsersForTasks(s, taskIDs, cond)
```

Replace `RegisterOverdueReminderCron`:

```go
func RegisterOverdueReminderCron() {
	webhookEnabled := config.WebhooksEnabled.GetBool()
	emailEnabled := config.ServiceEnableEmailReminders.GetBool() && config.MailerEnabled.GetBool()

	if !emailEnabled && !webhookEnabled {
		return
	}

	if !emailEnabled {
		log.Info("Mailer is disabled, not sending overdue reminders per mail")
	}

	err := cron.Schedule("* * * * *", func() {
		s := db.NewSession()
		defer s.Close()

		now := time.Now()

		var cond builder.Cond
		if emailEnabled && !webhookEnabled {
			cond = builder.Eq{"users.overdue_tasks_reminders_enabled": true}
		}

		uts, err := getUndoneOverdueTasks(s, now, cond)
		if err != nil {
			log.Errorf("[Undone Overdue Tasks Reminder] Could not get undone overdue tasks in the next minute: %s", err)
			return
		}

		if len(uts) == 0 {
			return
		}

		log.Debugf("[Undone Overdue Tasks Reminder] Sending reminders to %d users", len(uts))

		taskIDs := []int64{}
		for _, ut := range uts {
			for _, t := range ut.tasks {
				taskIDs = append(taskIDs, t.ID)
			}
		}

		projects, err := GetProjectsMapSimpleByTaskIDs(s, taskIDs)
		if err != nil {
			log.Errorf("[Undone Overdue Tasks Reminder] Could not get projects for tasks: %s", err)
			return
		}

		for _, ut := range uts {
			// Send email notification (existing behavior)
			if emailEnabled && ut.user.OverdueTasksRemindersEnabled {
				var n notifications.Notification = &UndoneTasksOverdueNotification{
					User:     ut.user,
					Tasks:    ut.tasks,
					Projects: projects,
				}

				if len(ut.tasks) == 1 {
					for _, t := range ut.tasks {
						n = &UndoneTaskOverdueNotification{
							User:    ut.user,
							Task:    t,
							Project: projects[t.ProjectID],
						}
					}
				}

				err = notifications.Notify(ut.user, n)
				if err != nil {
					log.Errorf("[Undone Overdue Tasks Reminder] Could not notify user %d: %s", ut.user.ID, err)
					return
				}
			}

			// Dispatch webhook events
			if webhookEnabled {
				// Per-task events
				for _, t := range ut.tasks {
					err = events.Dispatch(&TaskOverdueEvent{
						Task:    t,
						User:    ut.user,
						Project: projects[t.ProjectID],
					})
					if err != nil {
						log.Errorf("[Undone Overdue Tasks Reminder] Could not dispatch overdue event for task %d: %s", t.ID, err)
					}
				}

				// Batch event
				err = events.Dispatch(&TasksOverdueEvent{
					Tasks:    mapToSlice(ut.tasks),
					User:     ut.user,
					Projects: projects,
				})
				if err != nil {
					log.Errorf("[Undone Overdue Tasks Reminder] Could not dispatch batch overdue event for user %d: %s", ut.user.ID, err)
				}
			}

			log.Debugf("[Undone Overdue Tasks Reminder] Sent reminder for %d tasks to user %d", len(ut.tasks), ut.user.ID)
		}
	})
	if err != nil {
		log.Fatalf("Could not register undone overdue tasks reminder cron: %s", err)
	}
}

func mapToSlice(m map[int64]*Task) []*Task {
	tasks := make([]*Task, 0, len(m))
	for _, t := range m {
		tasks = append(tasks, t)
	}
	return tasks
}
```

**Step 4: Update tests**

In `pkg/models/task_reminder_test.go`, update calls to match new signature:

```go
// Change:
notifications, err := getTasksWithRemindersDueAndTheirUsers(s, now)
// To:
notifications, err := getTasksWithRemindersDueAndTheirUsers(s, now, builder.Eq{"users.email_reminders_enabled": true})
```

In `pkg/models/task_overdue_reminder_test.go`, update calls:

```go
// Change:
tasks, err := getUndoneOverdueTasks(s, now)
// To:
tasks, err := getUndoneOverdueTasks(s, now, builder.Eq{"users.overdue_tasks_reminders_enabled": true})
```

Add `"xorm.io/builder"` to imports in both test files.

**Step 5: Verify it compiles and tests pass**

```bash
mage build && mage test:feature
```

**Step 6: Commit**

```bash
git add pkg/models/task_reminder.go pkg/models/task_reminder_test.go pkg/models/task_overdue_reminder.go pkg/models/task_overdue_reminder_test.go pkg/models/notifications.go
git commit -m "feat: include User in webhook event payloads, add TasksOverdueEvent dispatch

Updates cron jobs to include User in event payloads for user-level
webhook lookups. Adds cond parameter to user-fetching functions so
all users are found when webhooks are enabled. Dispatches
TasksOverdueEvent batch events. Gates email on user preference."
```

---

### Task 8: Add API routes for user-level webhooks

**Files:**
- Create: `pkg/routes/api/v1/user_webhooks.go`
- Modify: `pkg/routes/routes.go` (add route registration)

**Step 1: Create the API handler file**

Create `pkg/routes/api/v1/user_webhooks.go`:

```go
package v1

import (
	"net/http"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"github.com/labstack/echo/v5"
)

// GetUserWebhooks returns all webhook targets for the current user
// @Summary Get all user-level webhook targets
// @Description Get all webhook targets configured for the current user (not project-specific).
// @tags webhooks
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Success 200 {array} models.Webhook "The list of webhook targets"
// @Failure 500 {object} models.Message "Internal server error"
// @Router /user/settings/webhooks [get]
func GetUserWebhooks(c *echo.Context) error {
	u, err := user.GetCurrentUser(c)
	if err != nil {
		return err
	}

	s := db.NewSession()
	defer s.Close()

	ws := []*models.Webhook{}
	err = s.Where("user_id = ?", u.ID).Find(&ws)
	if err != nil {
		return err
	}

	// Strip secrets from response
	for _, w := range ws {
		w.Secret = ""
	}

	return c.JSON(http.StatusOK, ws)
}

// CreateUserWebhook creates a new user-level webhook target
// @Summary Create a user-level webhook target
// @Description Create a webhook target for the current user that receives events across all projects.
// @tags webhooks
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param webhook body models.Webhook true "The webhook target"
// @Success 200 {object} models.Webhook "The created webhook target"
// @Failure 400 {object} web.HTTPError "Invalid webhook"
// @Failure 500 {object} models.Message "Internal server error"
// @Router /user/settings/webhooks [put]
func CreateUserWebhook(c *echo.Context) error {
	u, err := user.GetCurrentUser(c)
	if err != nil {
		return err
	}

	w := &models.Webhook{}
	if err := c.Bind(w); err != nil {
		return err
	}

	// Force user-level webhook
	w.UserID = u.ID
	w.ProjectID = 0

	s := db.NewSession()
	defer s.Close()

	if err := s.Begin(); err != nil {
		return err
	}

	err = w.Create(s, u)
	if err != nil {
		_ = s.Rollback()
		return err
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return err
	}

	return c.JSON(http.StatusOK, w)
}

// UpdateUserWebhook updates a user-level webhook target's events
// @Summary Update a user-level webhook target
// @Description Update the events for a user-level webhook target.
// @tags webhooks
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "Webhook ID"
// @Success 200 {object} models.Webhook "The updated webhook target"
// @Failure 404 {object} web.HTTPError "Webhook not found"
// @Failure 500 {object} models.Message "Internal server error"
// @Router /user/settings/webhooks/{id} [post]
func UpdateUserWebhook(c *echo.Context) error {
	u, err := user.GetCurrentUser(c)
	if err != nil {
		return err
	}

	w := &models.Webhook{}
	if err := c.Bind(w); err != nil {
		return err
	}

	s := db.NewSession()
	defer s.Close()

	// Verify webhook belongs to user
	existing := &models.Webhook{}
	has, err := s.Where("id = ? AND user_id = ?", w.ID, u.ID).Get(existing)
	if err != nil {
		return err
	}
	if !has {
		return echo.ErrNotFound
	}

	if err := s.Begin(); err != nil {
		return err
	}

	err = w.Update(s, u)
	if err != nil {
		_ = s.Rollback()
		return err
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return err
	}

	return c.JSON(http.StatusOK, w)
}

// DeleteUserWebhook deletes a user-level webhook target
// @Summary Delete a user-level webhook target
// @Description Delete a user-level webhook target.
// @tags webhooks
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "Webhook ID"
// @Success 200 {object} models.Message "Successfully deleted"
// @Failure 404 {object} web.HTTPError "Webhook not found"
// @Failure 500 {object} models.Message "Internal server error"
// @Router /user/settings/webhooks/{id} [delete]
func DeleteUserWebhook(c *echo.Context) error {
	u, err := user.GetCurrentUser(c)
	if err != nil {
		return err
	}

	s := db.NewSession()
	defer s.Close()

	webhookID := c.Param("id")

	// Verify webhook belongs to user
	existing := &models.Webhook{}
	has, err := s.Where("id = ? AND user_id = ?", webhookID, u.ID).Get(existing)
	if err != nil {
		return err
	}
	if !has {
		return echo.ErrNotFound
	}

	if err := s.Begin(); err != nil {
		return err
	}

	err = existing.Delete(s, u)
	if err != nil {
		_ = s.Rollback()
		return err
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return err
	}

	return c.JSON(http.StatusOK, &models.Message{Message: "Successfully deleted."})
}

// GetUserDirectedWebhookEvents returns events available for user-level webhooks
// @Summary Get available user-directed webhook events
// @Description Get all webhook events that can be used with user-level webhook targets.
// @tags webhooks
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Success 200 {array} string "The list of user-directed webhook events"
// @Router /user/settings/webhooks/events [get]
func GetUserDirectedWebhookEvents(c *echo.Context) error {
	return c.JSON(http.StatusOK, models.GetUserDirectedWebhookEvents())
}
```

**Step 2: Register routes in `pkg/routes/routes.go`**

In the user routes section (around line 380), add the new routes (the old ones from the PR were removed in Task 1):

```go
	// User-level webhooks
	if config.WebhooksEnabled.GetBool() {
		u.GET("/settings/webhooks", apiv1.GetUserWebhooks)
		u.GET("/settings/webhooks/events", apiv1.GetUserDirectedWebhookEvents)
		u.PUT("/settings/webhooks", apiv1.CreateUserWebhook)
		u.POST("/settings/webhooks/:id", apiv1.UpdateUserWebhook)
		u.DELETE("/settings/webhooks/:id", apiv1.DeleteUserWebhook)
	}
```

Note: The `/events` route must be registered BEFORE `/:id` to avoid the router matching "events" as an ID.

**Step 3: Verify it compiles**

```bash
mage build
```

**Step 4: Commit**

```bash
git add pkg/routes/api/v1/user_webhooks.go pkg/routes/routes.go
git commit -m "feat: add API routes for user-level webhooks

Adds CRUD endpoints under /user/settings/webhooks for managing
user-level webhook targets, plus an events endpoint returning
only user-directed event types."
```

---

### Task 9: Extract shared frontend webhook component

**Files:**
- Create: `frontend/src/components/misc/WebhookManager.vue`
- Modify: `frontend/src/views/project/settings/ProjectSettingsWebhooks.vue`

**Step 1: Create the shared component**

Create `frontend/src/components/misc/WebhookManager.vue`. This extracts the create form and webhook list table from `ProjectSettingsWebhooks.vue` into a reusable component.

Props:
- `webhooks: IWebhook[]` — list of existing webhooks
- `availableEvents: string[]` — events to show as checkboxes
- `loading: boolean` — loading state

Events:
- `create(webhook: IWebhook)` — emitted when user submits the create form
- `delete(webhookId: number)` — emitted when user confirms deletion

The component contains the create form (target URL, secret, basic auth, event checkboxes) and the table of existing webhooks with delete buttons. It does NOT handle API calls — the parent does.

```vue
<script lang="ts" setup>
import {ref, watch} from 'vue'
import {useI18n} from 'vue-i18n'

import type {IWebhook} from '@/modelTypes/IWebhook'
import WebhookModel from '@/models/webhook'
import BaseButton from '@/components/base/BaseButton.vue'
import FancyCheckbox from '@/components/input/FancyCheckbox.vue'
import FormField from '@/components/input/FormField.vue'
import Expandable from '@/components/base/Expandable.vue'
import User from '@/components/misc/User.vue'
import {formatDateShort} from '@/helpers/time/formatDate'
import {isValidHttpUrl} from '@/helpers/isValidHttpUrl'

defineOptions({name: 'WebhookManager'})

const props = defineProps<{
	webhooks: IWebhook[]
	availableEvents: string[]
	loading?: boolean
}>()

const emit = defineEmits<{
	create: [webhook: IWebhook]
	delete: [webhookId: number]
}>()

const {t} = useI18n({useScope: 'global'})

const showNewForm = ref(false)
const showBasicAuth = ref(false)
const newWebhook = ref(new WebhookModel())
const newWebhookEvents = ref<Record<string, boolean>>({})

function initEvents(events: string[]) {
	newWebhookEvents.value = Object.fromEntries(
		events.map(event => [event, false]),
	)
}

watch(() => props.availableEvents, (events) => {
	if (events) initEvents(events)
}, {immediate: true})

const webhookTargetUrlValid = ref(true)
const selectedEventsValid = ref(true)
const showDeleteModal = ref(false)
const webhookIdToDelete = ref<number>()

function validateTargetUrl() {
	webhookTargetUrlValid.value = isValidHttpUrl(newWebhook.value.targetUrl)
}

function getSelectedEventsArray() {
	return Object.entries(newWebhookEvents.value)
		.filter(([, use]) => use)
		.map(([event]) => event)
}

function validateSelectedEvents() {
	const events = getSelectedEventsArray()
	selectedEventsValid.value = events.length > 0
}

function create() {
	validateTargetUrl()
	if (!webhookTargetUrlValid.value) {
		return
	}

	const selectedEvents = getSelectedEventsArray()
	newWebhook.value.events = selectedEvents

	validateSelectedEvents()
	if (!selectedEventsValid.value) {
		return
	}

	emit('create', newWebhook.value)
	newWebhook.value = new WebhookModel()
	showNewForm.value = false
}

function confirmDelete(webhookId: number) {
	webhookIdToDelete.value = webhookId
	showDeleteModal.value = true
}

function doDelete() {
	if (webhookIdToDelete.value) {
		emit('delete', webhookIdToDelete.value)
	}
	showDeleteModal.value = false
}
</script>

<template>
	<div>
		<XButton
			v-if="!(webhooks?.length === 0 || showNewForm)"
			icon="plus"
			class="mbe-4"
			@click="showNewForm = true"
		>
			{{ $t('project.webhooks.create') }}
		</XButton>

		<div
			v-if="webhooks?.length === 0 || showNewForm"
			class="p-4"
		>
			<FormField
				id="targetUrl"
				v-model="newWebhook.targetUrl"
				:label="$t('project.webhooks.targetUrl')"
				required
				:placeholder="$t('project.webhooks.targetUrl')"
				:error="webhookTargetUrlValid ? null : $t('project.webhooks.targetUrlInvalid')"
				@focusout="validateTargetUrl"
			/>
			<div class="field">
				<label
					class="label"
					for="secret"
				>
					{{ $t('project.webhooks.secret') }}
				</label>
				<div class="control">
					<input
						id="secret"
						v-model="newWebhook.secret"
						class="input"
					>
				</div>
				<p class="help">
					{{ $t('project.webhooks.secretHint') }}
					<BaseButton href="https://vikunja.io/docs/webhooks/">
						{{ $t('project.webhooks.secretDocs') }}
					</BaseButton>
				</p>
			</div>
			<BaseButton
				class="mbe-2 has-text-primary"
				@click="showBasicAuth = !showBasicAuth"
			>
				{{ $t('project.webhooks.basicauthlink') }}
			</BaseButton>
			<Expandable
				:open="showBasicAuth"
				class="content"
			>
				<div class="field">
					<label
						class="label"
						for="basicauthuser"
					>
						{{ $t('project.webhooks.basicauthuser') }}
					</label>
					<div class="control">
						<input
							id="basicauthuser"
							v-model="newWebhook.basicauthuser"
							class="input"
						>
					</div>
				</div>
				<div class="field">
					<label
						class="label"
						for="basicauthpassword"
					>
						{{ $t('project.webhooks.basicauthpassword') }}
					</label>
					<div class="control">
						<input
							id="basicauthpassword"
							v-model="newWebhook.basicauthpassword"
							class="input"
						>
					</div>
				</div>
			</Expandable>
			<div class="field">
				<label
					class="label"
					for="events"
				>
					{{ $t('project.webhooks.events') }}
				</label>
				<p class="help">
					{{ $t('project.webhooks.eventsHint') }}
				</p>
				<div class="control">
					<FancyCheckbox
						v-for="event in availableEvents"
						:key="event"
						v-model="newWebhookEvents[event]"
						class="available-events-check"
						@update:modelValue="validateSelectedEvents"
					>
						{{ event }}
					</FancyCheckbox>
				</div>
				<p
					v-if="!selectedEventsValid"
					class="help is-danger"
				>
					{{ $t('project.webhooks.mustSelectEvents') }}
				</p>
			</div>
			<XButton
				icon="plus"
				@click="create"
			>
				{{ $t('project.webhooks.create') }}
			</XButton>
		</div>

		<table
			v-if="webhooks?.length > 0"
			class="table has-actions is-striped is-hoverable is-fullwidth"
		>
			<thead>
				<tr>
					<th>{{ $t('project.webhooks.targetUrl') }}</th>
					<th>{{ $t('project.webhooks.events') }}</th>
					<th>{{ $t('misc.created') }}</th>
					<th>{{ $t('misc.createdBy') }}</th>
					<th />
				</tr>
			</thead>
			<tbody>
				<tr
					v-for="w in webhooks"
					:key="w.id"
				>
					<td>{{ w.targetUrl }}</td>
					<td>{{ w.events.join(', ') }}</td>
					<td>{{ formatDateShort(w.created) }}</td>
					<td>
						<User
							:avatar-size="25"
							:user="w.createdBy"
						/>
					</td>

					<td class="actions">
						<XButton
							danger
							icon="trash-alt"
							@click="() => confirmDelete(w.id)"
						/>
					</td>
				</tr>
			</tbody>
		</table>

		<Modal
			:enabled="showDeleteModal"
			@close="showDeleteModal = false"
			@submit="doDelete()"
		>
			<template #header>
				<span>{{ $t('project.webhooks.delete') }}</span>
			</template>

			<template #text>
				<p>{{ $t('project.webhooks.deleteText') }}</p>
			</template>
		</Modal>
	</div>
</template>

<style lang="scss" scoped>
.available-events-check {
	margin-inline-end: .5rem;
	inline-size: 12.5rem;
}
</style>
```

**Step 2: Rework `ProjectSettingsWebhooks.vue` to use the shared component**

```vue
<script lang="ts" setup>
import {ref, computed, watchEffect} from 'vue'
import {useRoute} from 'vue-router'
import {useI18n} from 'vue-i18n'
import {useTitle} from '@vueuse/core'

import ProjectService from '@/services/project'
import ProjectModel from '@/models/project'
import type {IProject} from '@/modelTypes/IProject'
import type {IWebhook} from '@/modelTypes/IWebhook'

import CreateEdit from '@/components/misc/CreateEdit.vue'
import WebhookManager from '@/components/misc/WebhookManager.vue'

import {useBaseStore} from '@/stores/base'
import WebhookService from '@/services/webhook'
import {success} from '@/message'

defineOptions({name: 'ProjectSettingWebhooks'})

const {t} = useI18n({useScope: 'global'})

const project = ref<IProject>()
useTitle(t('project.webhooks.title'))

async function loadProject(projectId: number) {
	const projectService = new ProjectService()
	const newProject = await projectService.get(new ProjectModel({id: projectId}))
	await useBaseStore().handleSetCurrentProject({project: newProject})
	project.value = newProject
	await loadWebhooks()
}

const route = useRoute()
const projectId = computed(() => route.params.projectId !== undefined
	? parseInt(route.params.projectId as string)
	: undefined,
)

watchEffect(() => projectId.value !== undefined && loadProject(projectId.value))

const webhooks = ref<IWebhook[]>([])
const webhookService = new WebhookService()
const availableEvents = ref<string[]>([])
const loading = ref(false)

async function loadWebhooks() {
	loading.value = true
	try {
		webhooks.value = await webhookService.getAll({projectId: project.value.id})
		availableEvents.value = await webhookService.getAvailableEvents()
	} finally {
		loading.value = false
	}
}

async function handleCreate(webhook: IWebhook) {
	webhook.projectId = project.value.id
	const created = await webhookService.create(webhook)
	webhooks.value.push(created)
}

async function handleDelete(webhookId: number) {
	await webhookService.delete({
		id: webhookId,
		projectId: project.value.id,
	})
	success({message: t('project.webhooks.deleteSuccess')})
	await loadWebhooks()
}
</script>

<template>
	<CreateEdit
		:title="$t('project.webhooks.title')"
		:has-primary-action="false"
		:wide="true"
	>
		<WebhookManager
			:webhooks="webhooks"
			:available-events="availableEvents"
			:loading="loading"
			@create="handleCreate"
			@delete="handleDelete"
		/>
	</CreateEdit>
</template>
```

**Step 3: Verify frontend builds**

```bash
cd frontend && pnpm build:dev
```

**Step 4: Commit**

```bash
git add frontend/src/components/misc/WebhookManager.vue frontend/src/views/project/settings/ProjectSettingsWebhooks.vue
git commit -m "refactor: extract webhook form into shared WebhookManager component

Extracts the webhook creation form and list table from
ProjectSettingsWebhooks into a reusable WebhookManager component
that will be shared with the user-level webhooks page."
```

---

### Task 10: Create user webhook settings frontend page

**Files:**
- Modify: `frontend/src/views/user/settings/Webhooks.vue` (rewrite using shared component)
- Modify: `frontend/src/services/webhook.ts` (add user-scoped service)
- Modify: `frontend/src/models/webhook.ts` (add userId field)
- Modify: `frontend/src/modelTypes/IWebhook.ts` (add userId field)
- Keep: `frontend/src/router/index.ts` (already has webhooks route from PR)
- Keep: `frontend/src/views/user/Settings.vue` (already has webhooks nav item from PR)
- Keep: `frontend/src/stores/config.ts` (already has webhooksEnabled from PR)
- Modify: `frontend/src/i18n/lang/en.json` (simplify webhook i18n keys)

**Step 1: Add userId to the webhook model**

In `frontend/src/modelTypes/IWebhook.ts`, add `userId`:

```typescript
export interface IWebhook extends IAbstract {
	id: number
	projectId: number
	userId: number
	secret: string
	basicauthuser: string
	basicauthpassword: string
	targetUrl: string
	events: string[]
	createdBy: IUser

	created: Date
	updated: Date
}
```

In `frontend/src/models/webhook.ts`, add `userId = 0` after `projectId`:

```typescript
export default class WebhookModel extends AbstractModel<IWebhook> implements IWebhook {
	id = 0
	projectId = 0
	userId = 0
	secret = ''
	// ... rest unchanged
```

**Step 2: Create a user webhook service**

Add to `frontend/src/services/webhook.ts`:

```typescript
export class UserWebhookService extends AbstractService<IWebhook> {
	constructor() {
		super({
			getAll: '/user/settings/webhooks',
			create: '/user/settings/webhooks',
			update: '/user/settings/webhooks/{id}',
			delete: '/user/settings/webhooks/{id}',
		})
	}

	modelFactory(data) {
		return new WebhookModel(data)
	}

	async getAvailableEvents(): Promise<string[]> {
		const cancel = this.setLoading()

		try {
			const response = await this.http.get('/user/settings/webhooks/events')
			return response.data
		} finally {
			cancel()
		}
	}
}
```

**Step 3: Rewrite the Webhooks.vue page**

Replace `frontend/src/views/user/settings/Webhooks.vue`:

```vue
<script lang="ts" setup>
import {ref, onMounted} from 'vue'
import {useI18n} from 'vue-i18n'

import Card from '@/components/misc/Card.vue'
import WebhookManager from '@/components/misc/WebhookManager.vue'

import {useTitle} from '@/composables/useTitle'
import {success} from '@/message'
import {UserWebhookService} from '@/services/webhook'
import type {IWebhook} from '@/modelTypes/IWebhook'

defineOptions({name: 'UserSettingsWebhooks'})

const {t} = useI18n({useScope: 'global'})
useTitle(() => `${t('user.settings.webhooks.title')} - ${t('user.settings.title')}`)

const service = new UserWebhookService()
const webhooks = ref<IWebhook[]>([])
const availableEvents = ref<string[]>([])
const loading = ref(false)

async function loadWebhooks() {
	loading.value = true
	try {
		webhooks.value = await service.getAll()
		availableEvents.value = await service.getAvailableEvents()
	} finally {
		loading.value = false
	}
}

async function handleCreate(webhook: IWebhook) {
	const created = await service.create(webhook)
	webhooks.value.push(created)
}

async function handleDelete(webhookId: number) {
	await service.delete({id: webhookId})
	success({message: t('project.webhooks.deleteSuccess')})
	await loadWebhooks()
}

onMounted(() => {
	loadWebhooks()
})
</script>

<template>
	<Card
		:title="$t('user.settings.webhooks.title')"
		:loading="loading"
	>
		<p class="mb-4">
			{{ $t('user.settings.webhooks.description') }}
		</p>

		<WebhookManager
			:webhooks="webhooks"
			:available-events="availableEvents"
			:loading="loading"
			@create="handleCreate"
			@delete="handleDelete"
		/>
	</Card>
</template>
```

**Step 4: Simplify i18n strings**

In `frontend/src/i18n/lang/en.json`, replace the verbose `user.settings.webhooks` block with:

```json
"webhooks": {
  "title": "Webhook Notifications",
  "description": "Configure webhook URLs to receive POST requests when reminder or overdue events fire. These webhooks receive events from all your projects."
}
```

Keep the existing `project.webhooks.*` keys unchanged — those are used by the shared WebhookManager component and the project settings page.

**Step 5: Verify frontend builds and lints**

```bash
cd frontend && pnpm build:dev && pnpm lint && pnpm typecheck
```

**Step 6: Commit**

```bash
git add frontend/src/views/user/settings/Webhooks.vue frontend/src/services/webhook.ts frontend/src/models/webhook.ts frontend/src/modelTypes/IWebhook.ts frontend/src/i18n/lang/en.json
git commit -m "feat: add user-level webhooks settings page

Reworks the Webhooks.vue page to use the shared WebhookManager
component backed by a UserWebhookService that calls the
/user/settings/webhooks API endpoints."
```

---

### Task 11: Final integration test and cleanup

**Step 1: Run the full backend test suite**

```bash
mage test:feature
```

**Step 2: Run frontend tests**

```bash
cd frontend && pnpm test:unit
```

**Step 3: Run linters**

```bash
mage lint:fix
cd frontend && pnpm lint:fix && pnpm lint:styles:fix
```

**Step 4: Manual smoke test**

Start the dev server and verify:
1. Project webhook settings still work (create, list, delete)
2. Project webhook settings now show reminder events in the event list
3. User webhook settings page loads, shows only user-directed events
4. Can create a user-level webhook with URL, secret, events
5. Can delete a user-level webhook

**Step 5: Commit any lint fixes**

```bash
git add -A
git commit -m "chore: lint fixes"
```

**Step 6: Format Go code**

```bash
mage fmt
```
