package mail

import (
	"bytes"
	"code.vikunja.io/api/pkg/utils"
	"github.com/labstack/gommon/log"
	"github.com/spf13/viper"
	"gopkg.in/gomail.v2"
	"text/template"
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
	m.SetHeader("From", viper.GetString("mailer.fromemail"))
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

	t := &Template{
		Templates: template.Must(template.ParseGlob("templates/mail/*.tmpl")),
	}

	boundary := "np" + utils.MakeRandomString(13)

	data["Boundary"] = boundary
	data["FrontendURL"] = viper.GetString("service.frontendurl")

	if err := t.Templates.ExecuteTemplate(&htmlContent, tpl+".html.tmpl", data); err != nil {
		log.Error(3, "Template: %v", err)
		return
	}

	if err := t.Templates.ExecuteTemplate(&plainContent, tpl+".plain.tmpl", data); err != nil {
		log.Error(3, "Template: %v", err)
		return
	}

	opts := &Opts{
		To:          to,
		Subject:     subject,
		Message:     plainContent.String(),
		HTMLMessage: htmlContent.String(),
		ContentType: ContentTypeMultipart,
		Boundary:    boundary,
		Headers:     []*header{{Field: "MIME-Version", Content: "1.0"}},
	}

	SendMail(opts)
}
