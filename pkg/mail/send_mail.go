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

package mail

import (
	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"
	"gopkg.in/gomail.v2"
)

// Opts holds infos for a mail
type Opts struct {
	From        string
	To          string
	Subject     string
	Message     string
	HTMLMessage string
	ContentType ContentType
	Boundary    string
	Headers     []*header
}

// ContentType represents mail content types
type ContentType int

// Enumerate all the team rights
const (
	ContentTypePlain ContentType = iota
	ContentTypeHTML
	ContentTypeMultipart
)

type header struct {
	Field   string
	Content string
}

// SendTestMail sends a test mail to a receipient.
// It works without a queue.
func SendTestMail(opts *Opts) error {
	if config.MailerHost.GetString() == "" {
		log.Warning("Mailer seems to be not configured! Please see the config docs for more details.")
		return nil
	}

	d := getDialer()
	s, err := d.Dial()
	if err != nil {
		return err
	}
	defer s.Close()

	m := sendMail(opts)

	return gomail.Send(s, m)
}

func sendMail(opts *Opts) *gomail.Message {
	m := gomail.NewMessage()
	if opts.From == "" {
		opts.From = config.MailerFromEmail.GetString()
	}
	m.SetHeader("From", opts.From)
	m.SetHeader("To", opts.To)
	m.SetHeader("Subject", opts.Subject)
	for _, h := range opts.Headers {
		m.SetHeader(h.Field, h.Content)
	}

	switch opts.ContentType {
	case ContentTypePlain:
		m.SetBody("text/plain", opts.Message)
	case ContentTypeHTML:
		m.SetBody("text/html", opts.Message)
	case ContentTypeMultipart:
		m.SetBody("text/plain", opts.Message)
		m.AddAlternative("text/html", opts.HTMLMessage)
	}
	return m
}

// SendMail puts a mail in the queue
func SendMail(opts *Opts) {
	if isUnderTest {
		sentMails = append(sentMails, opts)
		return
	}

	m := sendMail(opts)
	Queue <- m
}
