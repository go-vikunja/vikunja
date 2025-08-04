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

package user

import (
	"time"

	"code.vikunja.io/api/pkg/cron"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/notifications"

	"xorm.io/builder"
	"xorm.io/xorm"
)

func RegisterDeletionNotificationCron() {
	err := cron.Schedule("0 * * * *", notifyUsersScheduledForDeletion)
	if err != nil {
		log.Errorf("Could not register deletion cron: %s", err.Error())
	}
}

func notifyUsersScheduledForDeletion() {
	s := db.NewSession()
	users := []*User{}
	err := s.Where(builder.NotNull{"deletion_scheduled_at"}).
		Find(&users)
	if err != nil {
		log.Errorf("Could not get users scheduled for deletion: %s", err)
		return
	}

	if len(users) == 0 {
		return
	}

	log.Debugf("Found %d users scheduled for deletion to notify", len(users))

	for _, user := range users {
		if time.Since(user.DeletionLastReminderSent) < time.Hour*24 {
			continue
		}

		var number = 2
		if user.DeletionLastReminderSent.IsZero() {
			number = 3
		}
		if user.DeletionScheduledAt.Sub(user.DeletionLastReminderSent) < time.Hour*24 {
			number = 1
		}

		log.Debugf("Notifying user %d of the deletion of their account...", user.ID)

		err = notifications.Notify(user, &AccountDeletionNotification{
			User:               user,
			NotificationNumber: number,
		})
		if err != nil {
			log.Errorf("Could not notify user %d of their deletion: %s", user.ID, err)
			continue
		}

		user.DeletionLastReminderSent = time.Now()
		_, err = s.Where("id = ?", user.ID).
			Cols("deletion_last_reminder_sent").
			Update(user)
		if err != nil {
			log.Errorf("Could update user %d last deletion reminder sent date: %s", user.ID, err)
		}
	}
}

// RequestDeletion creates a user deletion confirm token and sends a notification to the user
func RequestDeletion(s *xorm.Session, user *User) (err error) {
	token, err := generateToken(s, user, TokenAccountDeletion)
	if err != nil {
		return err
	}

	return notifications.Notify(user, &AccountDeletionConfirmNotification{
		User:         user,
		ConfirmToken: token.Token,
	})
}

// ConfirmDeletion ConformDeletion checks a token and schedules the user for deletion
func ConfirmDeletion(s *xorm.Session, user *User, token string) (err error) {
	tk, err := getToken(s, token, TokenAccountDeletion)
	if err != nil {
		return err
	}

	if tk == nil {
		return ErrInvalidDeletionToken{
			Token: token,
		}
	}

	if tk.UserID != user.ID {
		return ErrTokenUserMismatch{
			TokenUserID: tk.UserID,
			UserID:      user.ID,
		}
	}

	err = removeTokens(s, user, TokenAccountDeletion)
	if err != nil {
		return err
	}

	return ScheduleDeletion(s, user)
}

func ScheduleDeletion(s *xorm.Session, user *User) (err error) {
	user.DeletionScheduledAt = time.Now().Add(3 * 24 * time.Hour)
	_, err = s.Where("id = ?", user.ID).
		Cols("deletion_scheduled_at").
		Update(user)
	return err
}

// CancelDeletion cancels the deletion of a user
func CancelDeletion(s *xorm.Session, user *User) (err error) {
	user.DeletionScheduledAt = time.Time{}
	user.DeletionLastReminderSent = time.Time{}
	_, err = s.Where("id = ?", user.ID).
		Cols("deletion_scheduled_at", "deletion_last_reminder_sent").
		Update(user)
	return
}
