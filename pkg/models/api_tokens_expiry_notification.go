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
	"code.vikunja.io/api/pkg/i18n"
	"code.vikunja.io/api/pkg/notifications"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/utils"
)

func init() {
	notifications.Register(func() notifications.PersistedNotification { return &APITokenExpiringWeekNotification{} })
	notifications.Register(func() notifications.PersistedNotification { return &APITokenExpiringDayNotification{} })
}

// Bots never receive notifications, so an expiring bot token is sent to a human
// (User) with Bot set to the token's owner so the mail can name it.
func apiTokenExpiryTitle(lang, period string, token *APIToken, bot *user.User) string {
	if bot != nil {
		return i18n.T(lang, "notifications.api_token.expiring."+period+".bot_subject", token.Title, bot.Username)
	}
	return i18n.T(lang, "notifications.api_token.expiring."+period+".subject", token.Title)
}

func apiTokenExpiryMail(lang, period string, recipient *user.User, token *APIToken, bot *user.User) *notifications.Mail {
	expires := token.ExpiresAt.Format("2006-01-02")
	in := utils.HumanizeDuration(time.Until(token.ExpiresAt), lang)
	mail := notifications.NewMail().Greeting(i18n.T(lang, "notifications.greeting", recipient.GetName()))
	if bot != nil {
		mail.Line(i18n.T(lang, "notifications.api_token.expiring."+period+".bot_message", notifications.EscapeMarkdown(token.Title), notifications.EscapeMarkdown(bot.Username), expires, in)).
			Action(i18n.T(lang, "notifications.api_token.expiring.bot_action"), config.ServicePublicURL.GetString()+"user/settings/bots")
	} else {
		mail.Line(i18n.T(lang, "notifications.api_token.expiring."+period+".message", notifications.EscapeMarkdown(token.Title), expires, in)).
			Action(i18n.T(lang, "notifications.api_token.expiring.action"), config.ServicePublicURL.GetString()+"user/settings/api-tokens")
	}
	return mail.Line(i18n.T(lang, "notifications.common.have_nice_day"))
}

// APITokenExpiringWeekNotification is sent 7 days before an API token expires.
type APITokenExpiringWeekNotification struct {
	User  *user.User `json:"user"`
	Token *APIToken  `json:"api_token"`
	Bot   *user.User `json:"bot,omitempty"`
}

func (n *APITokenExpiringWeekNotification) ToTitle(lang string) string {
	return apiTokenExpiryTitle(lang, "week", n.Token, n.Bot)
}

func (n *APITokenExpiringWeekNotification) ToMail(lang string) *notifications.Mail {
	return apiTokenExpiryMail(lang, "week", n.User, n.Token, n.Bot)
}

func (n *APITokenExpiringWeekNotification) ToDB() any {
	return n
}

func (n *APITokenExpiringWeekNotification) Name() string {
	return "api_token.expiring.week"
}

func (n *APITokenExpiringWeekNotification) SubjectID() int64 {
	return n.Token.ID
}

// APITokenExpiringDayNotification is sent 1 day before an API token expires.
type APITokenExpiringDayNotification struct {
	User  *user.User `json:"user"`
	Token *APIToken  `json:"api_token"`
	Bot   *user.User `json:"bot,omitempty"`
}

func (n *APITokenExpiringDayNotification) ToTitle(lang string) string {
	return apiTokenExpiryTitle(lang, "day", n.Token, n.Bot)
}

func (n *APITokenExpiringDayNotification) ToMail(lang string) *notifications.Mail {
	return apiTokenExpiryMail(lang, "day", n.User, n.Token, n.Bot)
}

func (n *APITokenExpiringDayNotification) ToDB() any {
	return n
}

func (n *APITokenExpiringDayNotification) Name() string {
	return "api_token.expiring.day"
}

func (n *APITokenExpiringDayNotification) SubjectID() int64 {
	return n.Token.ID
}
