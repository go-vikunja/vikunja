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

	err := cron.Schedule("0 * * * *", func() { checkForExpiringAPITokensAt(time.Now()) })
	if err != nil {
		log.Fatalf("Could not register API token expiry check cron: %s", err)
	}
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

	ownerIDs := make([]int64, 0, len(tokens))
	for _, token := range tokens {
		ownerIDs = append(ownerIDs, token.OwnerID)
	}

	owners, err := user.GetUsersByIDs(s, ownerIDs)
	if err != nil {
		log.Errorf(logPrefix+"Error getting token owners: %s", err)
		return
	}

	botOwners, err := getBotOwners(s, owners)
	if err != nil {
		log.Errorf(logPrefix+"Error getting bot owners: %s", err)
		return
	}

	for _, token := range tokens {
		owner, exists := owners[token.OwnerID]
		if !exists {
			continue
		}

		for _, r := range tokenExpiryRecipients(owner, botOwners) {
			if err := sendTokenExpiryNotification(s, r, token, oneDay); err != nil {
				log.Errorf(logPrefix+"Error sending notification for token %d to user %d: %s", token.ID, r.user.ID, err)
			}
		}
	}

	if err := s.Commit(); err != nil {
		log.Errorf(logPrefix+"Error committing session: %s", err)
	}
}

// tokenExpiryRecipient is a human to notify; bot is set when the token belongs to a bot.
type tokenExpiryRecipient struct {
	user *user.User
	bot  *user.User
}

// Bots never receive notifications (ShouldNotify is false), so a bot token is
// routed to a human instead. Instance bots without an owner will need their own branch here.
func tokenExpiryRecipients(owner *user.User, botOwners map[int64]*user.User) []tokenExpiryRecipient {
	if !owner.IsBot() {
		return []tokenExpiryRecipient{{user: owner}}
	}

	human, exists := botOwners[owner.BotOwnerID]
	if !exists {
		return nil
	}

	return []tokenExpiryRecipient{{user: human, bot: owner}}
}

func getBotOwners(s *xorm.Session, owners map[int64]*user.User) (map[int64]*user.User, error) {
	botOwnerIDs := []int64{}
	for _, u := range owners {
		if u.IsBot() {
			botOwnerIDs = append(botOwnerIDs, u.BotOwnerID)
		}
	}

	return user.GetUsersByIDs(s, botOwnerIDs)
}

// Sends only the most urgent notification: 1-day if within 24h, otherwise 7-day.
func sendTokenExpiryNotification(s *xorm.Session, r tokenExpiryRecipient, token *APIToken, oneDay time.Time) error {
	if token.ExpiresAt.Before(oneDay) || token.ExpiresAt.Equal(oneDay) {
		return sendTokenExpiryNotificationIfNew(s, r.user, &APITokenExpiringDayNotification{
			User:  r.user,
			Token: token,
			Bot:   r.bot,
		})
	}

	return sendTokenExpiryNotificationIfNew(s, r.user, &APITokenExpiringWeekNotification{
		User:  r.user,
		Token: token,
		Bot:   r.bot,
	})
}

// sendTokenExpiryNotificationIfNew checks whether a notification with the same
// name and subject (token ID) has already been sent for this user. If not, it
// sends the notification (both email and DB).
func sendTokenExpiryNotificationIfNew(s *xorm.Session, u *user.User, n notifications.NotificationWithSubject) error {
	existing, err := notifications.GetNotificationsForNameAndUser(s, u.ID, n.Name(), n.SubjectID())
	if err != nil {
		return err
	}

	if len(existing) > 0 {
		return nil
	}

	return notifications.Notify(u, n, s)
}
