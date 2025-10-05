# Plan for showing tasks due today in reminder emails

## Overview
Extend the daily overdue reminder email so it can also list tasks that are due later the same day. The user can switch this behaviour on or off in their settings. The mail will contain two sections:

1. **Overdue tasks** – tasks whose due date/time passed.
2. **Due today** – tasks due later today (in the user's timezone).

A task appearing in both categories must only be listed as overdue.

---

## Backend
1. **Configuration & model changes**
   - Add config key `defaultsettings.today_tasks_reminders_enabled` with a default (initially `false` to preserve current behaviour).
   - Create DB migration adding column `today_tasks_reminders_enabled` to `users` table with index and default pulled from config.
   - Extend structs (`pkg/user/user.go`, `pkg/routes/api/v1/user_settings.go`, `pkg/user/user_create.go`) with field `TodayTasksRemindersEnabled` / JSON `today_tasks_reminders_enabled`.
   - Expose field in API routes; update Swagger docs.

2. **Collecting tasks**
   - Rename `getUndoneOverdueTasks` to something like `getTasksForDailyReminder` and extend it to also fetch tasks due later today.
   - Query tasks with due dates up to end‑of‑day across time zones (≈ now + 38h) and categorise per user into overdue vs due today using their timezone and current reminder time.
   - Only include "due today" section when the user has `TodayTasksRemindersEnabled` set.

3. **Notifications**
   - Replace `UndoneTasksOverdueNotification`/`UndoneTaskOverdueNotification` with a new notification struct (e.g. `DailyTasksReminderNotification`) holding two task lists: overdue and due today.
   - Build email with two sections and appropriate headings, handling cases where only one of the sections has tasks.
   - Add translation strings in `pkg/i18n/lang/en.json` for the new subject line and section titles/messages.

4. **Cron job**
   - Adjust `RegisterOverdueReminderCron` to call the new task collector and send the new notification type. The cron should trigger when the user’s configured reminder time is reached.

5. **Tests**
   - Extend fixtures with a task that is due later on the same day.
   - Update `pkg/models/task_overdue_reminder_test.go` to assert both overdue and due‑today categorisation and that due‑today tasks are only included for users with the new setting enabled.

---

## Frontend
1. **User settings model**
   - Add `todayTasksRemindersEnabled` to `IUserSettings` and `UserSettingsModel` with default `false`.

2. **Settings UI**
   - In `frontend/src/views/user/settings/General.vue`, expose separate checkboxes for overdue reminders and "Include tasks due today in reminder email" (translation key `user.settings.general.todayReminders`).
   - Show the reminder time input when either overdue or today reminders are enabled.

3. **Translations**
   - Add the new label to `frontend/src/i18n/lang/en.json` and placeholders in other languages.

4. **Store/Service**
   - Ensure the settings store and `UserSettingsService` send/receive the new field (interfaces handle most of it).
   - Add/adjust tests or Cypress spec to cover toggling the setting.

---

## Email & localisation
- Update backend translation strings (`pkg/i18n/lang/en.json`) for:
  - Subject when only due‑today tasks exist.
  - Section headers "Overdue tasks" and "Tasks due today".
  - Introductory texts for each section.
- Ensure translation keys are referenced in notification builder.

---

## Summary
The change adds a new user preference to include tasks due today in the daily reminder. Backend collects due‑today tasks alongside overdue ones and sends a combined email. Frontend exposes independent toggles for both overdue and today reminders. Tests and translations verify the new behaviour.
