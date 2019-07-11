//  Vikunja is a todo-list application to facilitate your life.
//  Copyright 2018 Vikunja and contributors. All rights reserved.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU General Public License as published by
//  the Free Software Foundation, either version 3 of the License, or
//  (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU General Public License for more details.
//
//  You should have received a copy of the GNU General Public License
//  along with this program.  If not, see <https://www.gnu.org/licenses/>.

package mail

import (
	"bytes"
	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/static"
	"code.vikunja.io/api/pkg/utils"
	"github.com/shurcooL/httpfs/html/vfstemplate"
	"gopkg.in/gomail.v2"
	"html/template"
)

// Opts holds infos for a mail
type Opts struct {
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

// SendMail puts a mail in the queue
func SendMail(opts *Opts) {
	m := gomail.NewMessage()
	m.SetHeader("From", config.MailerFromEmail.GetString())
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

	Queue <- m
}

// Template holds a pointer about a template
type Template struct {
	Templates *template.Template
}

// SendMailWithTemplate parses a template and sends it via mail
func SendMailWithTemplate(to, subject, tpl string, data map[string]interface{}) {
	var htmlContent bytes.Buffer
	var plainContent bytes.Buffer

	t, err := vfstemplate.ParseGlob(static.Templates, nil, "*.tmpl")
	if err != nil {
		log.Log.Errorf("SendMailWithTemplate: ParseGlob: %v", err)
		return
	}

	boundary := "np" + utils.MakeRandomString(13)

	data["Boundary"] = boundary
	data["FrontendURL"] = config.ServiceFrontendurl.GetString()

	if err := t.ExecuteTemplate(&htmlContent, tpl+".html.tmpl", data); err != nil {
		log.Log.Errorf("ExecuteTemplate: %v", err)
		return
	}

	if err := t.ExecuteTemplate(&plainContent, tpl+".plain.tmpl", data); err != nil {
		log.Log.Errorf("ExecuteTemplate: %v", err)
		return
	}

	opts := &Opts{
		To:          to,
		Subject:     subject,
		Message:     plainContent.String(),
		HTMLMessage: htmlContent.String(),
		ContentType: ContentTypeMultipart,
		Boundary:    boundary,
	}

	SendMail(opts)
}
