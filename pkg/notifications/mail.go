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

package notifications

import (
	"fmt"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/i18n"
	"code.vikunja.io/api/pkg/mail"
)

// Mail is a mail message
type Mail struct {
	from        string
	to          string
	subject     string
	actionText  string
	actionURL   string
	greeting    string
	introLines  []*mailLine
	outroLines  []*mailLine
	footerLines []*mailLine
}

type mailLine struct {
	Text   string
	isHTML bool
}

// NewMail creates a new mail object with a default greeting
func NewMail() *Mail {
	return &Mail{}
}

// From sets the from name and email address
func (m *Mail) From(from string) *Mail {
	m.from = from
	return m
}

// To sets the recipient of the mail message
func (m *Mail) To(to string) *Mail {
	m.to = to
	return m
}

// Subject sets the subject of the mail message
func (m *Mail) Subject(subject string) *Mail {
	m.subject = subject
	return m
}

// Greeting sets the greeting of the mail message
func (m *Mail) Greeting(greeting string) *Mail {
	m.greeting = greeting
	return m
}

// Action sets any action a mail might have
func (m *Mail) Action(text, url string) *Mail {
	m.actionText = text
	m.actionURL = url
	return m
}

// Line adds a line of Text to the mail
func (m *Mail) Line(line string) *Mail {
	return m.appendLine(line, false)
}

func (m *Mail) FooterLine(line string) *Mail {
	m.footerLines = append(m.footerLines, &mailLine{
		Text: line,
	})
	return m
}

func (m *Mail) IncludeLinkToSettings(lang string) *Mail {
	link := config.ServicePublicURL.GetString() + "user/settings/general"
	m.FooterLine(fmt.Sprintf(i18n.T(lang, "notifications.common.actions.change_notification_settings_link"), link))
	return m
}

func (m *Mail) HTML(line string) *Mail {
	return m.appendLine(line, true)
}

func (m *Mail) appendLine(line string, isHTML bool) *Mail {
	if m.actionURL == "" {
		m.introLines = append(m.introLines, &mailLine{
			Text:   line,
			isHTML: isHTML,
		})
		return m
	}

	m.outroLines = append(m.outroLines, &mailLine{
		Text:   line,
		isHTML: isHTML,
	})

	return m
}

// SendMail passes the notification to the mailing queue for sending
func SendMail(m *Mail, lang string) error {
	opts, err := RenderMail(m, lang)
	if err != nil {
		return err
	}

	mail.SendMail(opts)

	return nil
}
