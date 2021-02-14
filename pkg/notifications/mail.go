// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2021 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package notifications

import "code.vikunja.io/api/pkg/mail"

// Mail is a mail message
type Mail struct {
	from       string
	to         string
	subject    string
	actionText string
	actionURL  string
	greeting   string
	introLines []string
	outroLines []string
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

// Line adds a line of text to the mail
func (m *Mail) Line(line string) *Mail {
	if m.actionURL == "" {
		m.introLines = append(m.introLines, line)
		return m
	}

	m.outroLines = append(m.outroLines, line)

	return m
}

// SendMail passes the notification to the mailing queue for sending
func SendMail(m *Mail) error {
	opts, err := RenderMail(m)
	if err != nil {
		return err
	}

	mail.SendMail(opts)

	return nil
}
