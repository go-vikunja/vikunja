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
	"bytes"
	"embed"
	templatehtml "html/template"
	templatetext "text/template"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/i18n"
	"code.vikunja.io/api/pkg/mail"
	"code.vikunja.io/api/pkg/utils"

	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
)

const mailTemplatePlain = `
{{ .Greeting }}
{{ range $line := .IntroLines}}
{{ $line.Text }}
{{ end }}
{{ if .ActionURL }}{{ .ActionText }}:
{{ .ActionURL }}{{end}}
{{ range $line := .OutroLines}}
{{ $line.Text }}
{{ end }}
{{ range $line := .FooterLines}}
{{ $line.Text }}
{{ end }}`

const mailTemplateHTML = `
<!doctype html>
<html style="width: 100%; height: 100%; padding: 0; margin: 0;">
<head>
    <meta name="viewport" content="width: display-width;">
</head>
<body style="width: 100%; padding: 0; margin: 0; background: #f3f4f6">
<div style="width: 100%; font-family: 'Open Sans', sans-serif; Text-rendering: optimizeLegibility">
    <div style="width: 600px; margin: 0 auto; Text-align: justify;">
        <h1 style="font-size: 30px; Text-align: center;">
            <img src="cid:logo.png" style="height: 75px;" alt="Vikunja"/>
        </h1>
        <div style="border: 1px solid #dbdbdb; -webkit-box-shadow: 0.3em 0.3em 0.8em #e6e6e6; box-shadow: 0.3em 0.3em 0.8em #e6e6e6; color: #4a4a4a; padding: 5px 25px; border-radius: 3px; background: #fff;">
<p>
	{{ .Greeting }}
</p>

{{ range $line := .IntroLinesHTML}}
	{{ $line }}
{{ end }}

{{ if .ActionURL }}
	<a href="{{ .ActionURL }}" title="{{ .ActionText }}"
		style="position: relative;Text-decoration:none;display: block;border-radius: 4px;cursor: pointer;padding-bottom: 8px;padding-left: 14px;padding-right: 14px;padding-top: 8px;width:280px;margin:10px auto;Text-align: center;white-space: nowrap;border: 0;Text-transform: uppercase;font-size: 14px;font-weight: 700;-webkit-box-shadow: 0 3px 6px rgba(107,114,128,.12),0 2px 4px rgba(107,114,128,.1);box-shadow: 0 3px 6px rgba(107,114,128,.12),0 2px 4px rgba(107,114,128,.1);background-color: #1973ff;border-color: transparent;color: #fff;">
		{{ .ActionText }}
	</a>
{{end}}

{{ range $line := .OutroLinesHTML}}
	{{ $line }}
{{ end }}

{{ if .ActionURL }}
	<div style="color: #9CA3AF;font-size:12px;border-top: 1px solid #dbdbdb;margin-top:20px;padding-top:20px;">
	<p>
		{{ .CopyURLText }}<br/>
		{{ .ActionURL }}
	</p>
{{ range $line := .FooterLinesHTML}}
	{{ $line }}
	{{ end }}
	</div>
{{ else }}
{{ if .FooterLinesHTML }}
	<div style="color: #9CA3AF;font-size:12px;border-top: 1px solid #dbdbdb;margin-top:20px;padding-top:20px;">
		{{ range $line := .FooterLinesHTML }}
			{{ $line }}
		{{ end }}
	</div>
{{ end }}
{{ end }}
</div>
</div>
</div>
</body>
</html>
`

//go:embed logo.png
var logo embed.FS

func convertLinesToHTML(lines []*mailLine) (linesHTML []templatehtml.HTML, err error) {
	p := bluemonday.UGCPolicy()

	for _, line := range lines {
		if line.isHTML {
			// #nosec G203 -- the html is sanitized
			linesHTML = append(linesHTML, templatehtml.HTML(p.Sanitize(line.Text)))
			continue
		}

		md := []byte(templatehtml.HTMLEscapeString(line.Text))
		var buf bytes.Buffer
		err = goldmark.Convert(md, &buf)
		if err != nil {
			return nil, err
		}
		// #nosec G203 -- the html is sanitized
		linesHTML = append(linesHTML, templatehtml.HTML(p.Sanitize(buf.String())))
	}

	return
}

// RenderMail takes a precomposed mail message and renders it into a ready to send mail.Opts object
func RenderMail(m *Mail, lang string) (mailOpts *mail.Opts, err error) {

	var htmlContent bytes.Buffer
	var plainContent bytes.Buffer

	plain, err := templatetext.New("mail-plain").Parse(mailTemplatePlain)
	if err != nil {
		return nil, err
	}

	html, err := templatehtml.New("mail-plain").Parse(mailTemplateHTML)
	if err != nil {
		return nil, err
	}

	boundaryStr, err := utils.CryptoRandomString(13)
	if err != nil {
		return nil, err
	}
	boundary := "np" + boundaryStr

	data := make(map[string]interface{})

	data["Greeting"] = m.greeting
	data["IntroLines"] = m.introLines
	data["OutroLines"] = m.outroLines
	data["FooterLines"] = m.footerLines
	data["ActionText"] = m.actionText
	data["ActionURL"] = m.actionURL
	data["Boundary"] = boundary
	data["FrontendURL"] = config.ServicePublicURL.GetString()
	data["CopyURLText"] = i18n.T(lang, "notifications.common.copy_url")

	data["IntroLinesHTML"], err = convertLinesToHTML(m.introLines)
	if err != nil {
		return nil, err
	}

	data["OutroLinesHTML"], err = convertLinesToHTML(m.outroLines)
	if err != nil {
		return nil, err
	}

	data["FooterLinesHTML"], err = convertLinesToHTML(m.footerLines)
	if err != nil {
		return nil, err
	}

	err = plain.Execute(&plainContent, data)
	if err != nil {
		return nil, err
	}
	err = html.Execute(&htmlContent, data)
	if err != nil {
		return nil, err
	}

	mailOpts = &mail.Opts{
		From:        m.from,
		To:          m.to,
		Subject:     m.subject,
		ContentType: mail.ContentTypeMultipart,
		Message:     plainContent.String(),
		HTMLMessage: htmlContent.String(),
		Boundary:    boundary,
		EmbedFS: map[string]*embed.FS{
			"logo.png": &logo,
		},
	}

	return mailOpts, nil
}
