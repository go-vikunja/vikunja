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
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/cron"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/notifications"
	"code.vikunja.io/api/pkg/user"

	"xorm.io/builder"
	"xorm.io/xorm"
)

// RegisterAPITokenExpiryCheckCron registers the cron job that checks for
// expiring API tokens and notifies their owners.
func RegisterAPITokenExpiryCheckCron() {
	if !config.MailerEnabled.GetBool() {
		return
	}

	err := cron.Schedule("0 * * * *", checkForExpiringAPITokens)
	if err != nil {
		log.Fatalf("Could not register API token expiry check cron: %s", err)
	}
}

func checkForExpiringAPITokens() {
	checkForExpiringAPITokensAt(time.Now())
}

func checkForExpiringAPITokensAt(now time.Time) {
	const logPrefix = "[API Token Expiry Check] "

	oneDay := now.Add(24 * time.Hour)
	sevenDays := now.Add(7 * 24 * time.Hour)

	s := db.NewSession()
	defer s.Close()

	// Find all tokens expiring within the next 7 days that haven't expired yet
	var tokens []*APIToken
	err := s.Where(
		builder.Gt{"expires_at": now},
	).And(
		builder.Lte{"expires_at": sevenDays},
	).Find(&tokens)
	if err != nil {
		log.Errorf(logPrefix+"Error getting expiring tokens: %s", err)
		return
	}

	if len(tokens) == 0 {
		return
	}

	log.Debugf(logPrefix+"Found %d tokens expiring within 7 days", len(tokens))

	// Collect unique owner IDs and fetch users
	ownerIDs := make([]int64, 0, len(tokens))
	for _, token := range tokens {
		ownerIDs = append(ownerIDs, token.OwnerID)
	}

	users, err := user.GetUsersByIDs(s, ownerIDs)
	if err != nil {
		log.Errorf(logPrefix+"Error getting token owners: %s", err)
		return
	}

	for _, token := range tokens {
		u, exists := users[token.OwnerID]
		if !exists {
			continue
		}

		// Determine which thresholds apply
		expiresWithinOneDay := token.ExpiresAt.Before(oneDay) || token.ExpiresAt.Equal(oneDay)

		if expiresWithinOneDay {
			if err := sendTokenExpiryNotificationIfNew(s, u, token, &APITokenExpiringDayNotification{
				User:  u,
				Token: token,
			}); err != nil {
				log.Errorf(logPrefix+"Error sending 1-day notification for token %d: %s", token.ID, err)
			}
		}

		// Always check the 7-day notification (token is within 7 days by the query)
		if err := sendTokenExpiryNotificationIfNew(s, u, token, &APITokenExpiringWeekNotification{
			User:  u,
			Token: token,
		}); err != nil {
			log.Errorf(logPrefix+"Error sending 7-day notification for token %d: %s", token.ID, err)
		}
	}
}

// sendTokenExpiryNotificationIfNew checks whether a notification with the same
// name and subject (token ID) has already been sent for this user. If not, it
// sends the notification (both email and DB).
func sendTokenExpiryNotificationIfNew(s *xorm.Session, u *user.User, _ *APIToken, n notifications.NotificationWithSubject) error {
	existing, err := notifications.GetNotificationsForNameAndUser(s, u.ID, n.Name(), n.SubjectID())
	if err != nil {
		return err
	}

	if len(existing) > 0 {
		return nil
	}

	return notifications.Notify(u, n, s)
}
