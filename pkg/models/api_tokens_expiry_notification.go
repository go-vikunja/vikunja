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
	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/i18n"
	"code.vikunja.io/api/pkg/notifications"
	"code.vikunja.io/api/pkg/user"
)

// APITokenExpiringWeekNotification is sent 7 days before an API token expires.
type APITokenExpiringWeekNotification struct {
	User  *user.User `json:"user"`
	Token *APIToken  `json:"api_token"`
}

func (n *APITokenExpiringWeekNotification) ToMail(lang string) *notifications.Mail {
	return notifications.NewMail().
		Subject(i18n.T(lang, "notifications.api_token.expiring.week.subject", n.Token.Title)).
		Greeting(i18n.T(lang, "notifications.greeting", n.User.GetName())).
		Line(i18n.T(lang, "notifications.api_token.expiring.week.message", n.Token.Title, n.Token.ExpiresAt.Format("2006-01-02"))).
		Action(i18n.T(lang, "notifications.api_token.expiring.action"), config.ServicePublicURL.GetString()+"user/settings/api-tokens").
		Line(i18n.T(lang, "notifications.common.have_nice_day"))
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
}

func (n *APITokenExpiringDayNotification) ToMail(lang string) *notifications.Mail {
	return notifications.NewMail().
		Subject(i18n.T(lang, "notifications.api_token.expiring.day.subject", n.Token.Title)).
		Greeting(i18n.T(lang, "notifications.greeting", n.User.GetName())).
		Line(i18n.T(lang, "notifications.api_token.expiring.day.message", n.Token.Title, n.Token.ExpiresAt.Format("2006-01-02"))).
		Action(i18n.T(lang, "notifications.api_token.expiring.action"), config.ServicePublicURL.GetString()+"user/settings/api-tokens").
		Line(i18n.T(lang, "notifications.common.have_nice_day"))
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
