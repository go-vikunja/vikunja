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
	"regexp"
	"strings"
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

const mailTemplateConversationalPlain = `
{{ if .HeaderLinePlain }}{{ .HeaderLinePlain }}
{{ end }}{{ range $line := .IntroLines}}
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
    <meta name="viewport" content="width=device-width">
    <meta name="color-scheme" content="light dark">
    <meta name="supported-color-schemes" content="light dark">
    <style>
        :root {
            color-scheme: light dark;
        }
        @media (prefers-color-scheme: dark) {
            .email-card {
                box-shadow: 0.3em 0.3em 0.8em rgba(0,0,0,0.3) !important;
                -webkit-box-shadow: 0.3em 0.3em 0.8em rgba(0,0,0,0.3) !important;
            }
            .email-button {
                box-shadow: 0 3px 6px rgba(0,0,0,0.2), 0 2px 4px rgba(0,0,0,0.15) !important;
                -webkit-box-shadow: 0 3px 6px rgba(0,0,0,0.2), 0 2px 4px rgba(0,0,0,0.15) !important;
            }
        }
    </style>
</head>
<body style="width: 100%; padding: 0; margin: 0; background: #f3f4f6">
<div style="width: 100%; font-family: 'Open Sans', sans-serif; Text-rendering: optimizeLegibility">
    <div style="width: 600px; margin: 0 auto; Text-align: justify;">
        <h1 style="font-size: 30px; Text-align: center;">
            <img src="cid:logo.png" style="height: 75px;" alt="Vikunja"/>
        </h1>
        <div class="email-card" style="border: 1px solid #dbdbdb; -webkit-box-shadow: 0.3em 0.3em 0.8em #e6e6e6; box-shadow: 0.3em 0.3em 0.8em #e6e6e6; color: #4a4a4a; padding: 5px 25px; border-radius: 3px; background: #fff;">
<p>
	{{ .Greeting }}
</p>

{{ range $line := .IntroLinesHTML}}
	{{ $line }}
{{ end }}

{{ if .ActionURL }}
	<a class="email-button" href="{{ .ActionURL }}" title="{{ .ActionText }}"
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

const mailTemplateConversationalHTML = `
<!doctype html>
<html style="width: 100%; height: 100%; padding: 0; margin: 0;">
<head>
    <meta name="viewport" content="width=device-width">
    <meta charset="utf-8">
</head>
<body style="width: 100%; padding: 0; margin: 0; background: #f6f8fa; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Noto Sans', Helvetica, Arial, sans-serif;">
<div style="margin: 0 auto; background: #ffffff;">

    {{ if .HeaderLineHTML }}
    <div style="padding: 12px 20px 0; color: #57606a; font-size: 14px; line-height: 1.5;">
        {{ .HeaderLineHTML }}
    </div>
    {{ end }}

    {{ if or .IntroLinesHTML .OutroLinesHTML }}
    <div style="padding-left: 20px; color: #24292f; font-size: 14px; line-height: 1.5;">

        {{ range $line := .IntroLinesHTML}}
            {{ $line }}
        {{ end }}

        {{ range $line := .OutroLinesHTML}}
            {{ $line }}
        {{ end }}
    </div>
    {{ end }}

    {{ if or .ActionURL .FooterLinesHTML }}
    <div style="padding: 4px 20px 8px 20px; border-top: 1px solid #d1d9e0; padding-top: 6px; font-size: 12px">
        {{ if .ActionURL }}
        <a href="{{ .ActionURL }}" style="color: #0969da; text-decoration: none; font-weight: 500; font-size: 12px;">
            {{ .ActionText }} →
        </a>
        {{ end }}
    	<div style="padding-top: 6px; color: #656d76;">
    	    {{ range $line := .FooterLinesHTML }}
    	        {{ $line }}
    	    {{ end }}
    	</div>
    </div>
    {{ end }}
</div>
</body>
</html>
`

//go:embed logo.png
var logo embed.FS

func convertLinesToHTML(lines []*mailLine) (linesHTML []templatehtml.HTML, err error) {
	p := bluemonday.UGCPolicy()
	// Allow data URI images for inline avatars in mentions
	p.AllowDataURIImages()
	// Allow style attribute on img and div elements for avatar and layout styling
	p.AllowAttrs("style").OnElements("img", "div")
	// Allow specific CSS properties for avatar styling
	p.AllowStyles("border-radius", "vertical-align", "margin-right").OnElements("img")
	// Allow padding styles on div elements for content spacing
	p.AllowStyles("padding-top", "margin-bottom").OnElements("div")

	for _, line := range lines {
		if line.isHTML {
			sanitized := p.Sanitize(line.Text)
			if trimmed := strings.TrimSpace(sanitized); trimmed != "" && !startsWithBlockElement(trimmed) {
				sanitized = "<p>" + sanitized + "</p>"
			}
			// #nosec G203 -- the html is sanitized
			linesHTML = append(linesHTML, templatehtml.HTML(ensurePMargins(sanitized)))
			continue
		}

		md := []byte(line.Text)
		var buf bytes.Buffer
		err = goldmark.Convert(md, &buf)
		if err != nil {
			return nil, err
		}
		// #nosec G203 -- the html is sanitized
		linesHTML = append(linesHTML, templatehtml.HTML(ensurePMargins(p.Sanitize(buf.String()))))
	}

	return
}

// sanitizeLinesToHTML sanitizes lines without wrapping in <p> tags or adding margins.
// Used for footer lines and other content that should not have paragraph styling.
func sanitizeLinesToHTML(lines []*mailLine) (linesHTML []templatehtml.HTML, err error) {
	p := bluemonday.UGCPolicy()
	p.AllowDataURIImages()
	p.AllowAttrs("style").OnElements("img", "div")
	p.AllowStyles("border-radius", "vertical-align", "margin-right").OnElements("img")
	p.AllowStyles("padding-top", "margin-bottom").OnElements("div")

	for _, line := range lines {
		if line.isHTML {
			// #nosec G203 -- the html is sanitized
			linesHTML = append(linesHTML, templatehtml.HTML(p.Sanitize(line.Text)))
			continue
		}

		md := []byte(line.Text)
		var buf bytes.Buffer
		err = goldmark.Convert(md, &buf)
		if err != nil {
			return nil, err
		}
		sanitized := p.Sanitize(buf.String())
		// Strip <p> wrapping added by goldmark since the template already provides a container
		sanitized = rePTagOpen.ReplaceAllString(sanitized, "")
		sanitized = strings.ReplaceAll(sanitized, "</p>", "")
		sanitized = strings.TrimSpace(sanitized)
		// #nosec G203 -- the html is sanitized
		linesHTML = append(linesHTML, templatehtml.HTML(sanitized))
	}

	return
}

var rePTagOpen = regexp.MustCompile(`<p[^>]*>`)

func startsWithBlockElement(html string) bool {
	lower := strings.ToLower(html)
	for _, tag := range []string{"<p", "<div", "<h1", "<h2", "<h3", "<h4", "<h5", "<h6", "<ul", "<ol", "<li", "<table", "<blockquote", "<pre", "<hr"} {
		if strings.HasPrefix(lower, tag) {
			return true
		}
	}
	return false
}

var (
	reLinks    = regexp.MustCompile(`<a[^>]+href="([^"]*)"[^>]*>([^<]*)</a>`)
	reHTMLTags = regexp.MustCompile(`<[^>]+>`)
	rePTag     = regexp.MustCompile(`<p(?:\s[^>]*)?>`)
)

const pMarginStyle = `style="margin-top: 10px; margin-bottom: 10px;"`

// ensurePMargins replaces all <p> and <p ...> tags with a version
// that has fixed 10px top/bottom margins, ensuring consistent spacing
// across email clients.
func ensurePMargins(html string) string {
	return rePTag.ReplaceAllString(html, "<p "+pMarginStyle+">")
}

// convertLinesToPlain converts mail lines to plain text, stripping HTML from lines marked as HTML.
func convertLinesToPlain(lines []*mailLine) []*mailLine {
	plain := make([]*mailLine, 0, len(lines))
	for _, line := range lines {
		if !line.isHTML {
			plain = append(plain, line)
			continue
		}

		text := line.Text
		// Convert <a href="url">text</a> to "text (url)"
		text = reLinks.ReplaceAllString(text, "$2 ($1)")
		// Strip remaining HTML tags
		text = reHTMLTags.ReplaceAllString(text, "")
		// Clean up HTML entities
		text = strings.ReplaceAll(text, "&gt;", ">")
		text = strings.ReplaceAll(text, "&lt;", "<")
		text = strings.ReplaceAll(text, "&amp;", "&")
		text = strings.TrimSpace(text)

		if text != "" {
			plain = append(plain, &mailLine{Text: text})
		}
	}
	return plain
}

// RenderMail takes a precomposed mail message and renders it into a ready to send mail.Opts object
func RenderMail(m *Mail, lang string) (mailOpts *mail.Opts, err error) {

	var htmlContent bytes.Buffer
	var plainContent bytes.Buffer

	// Select template based on conversational flag
	var plainTemplate, htmlTemplate string
	if m.conversational {
		plainTemplate = mailTemplateConversationalPlain
		htmlTemplate = mailTemplateConversationalHTML
	} else {
		plainTemplate = mailTemplatePlain
		htmlTemplate = mailTemplateHTML
	}

	plain, err := templatetext.New("mail-plain").Parse(plainTemplate)
	if err != nil {
		return nil, err
	}

	html, err := templatehtml.New("mail-html").Parse(htmlTemplate)
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
	if m.conversational {
		data["IntroLines"] = convertLinesToPlain(m.introLines)
		data["OutroLines"] = convertLinesToPlain(m.outroLines)
		if m.headerLine != nil {
			plainHeaders := convertLinesToPlain([]*mailLine{m.headerLine})
			if len(plainHeaders) > 0 {
				data["HeaderLinePlain"] = plainHeaders[0].Text
			}
		}
	} else {
		data["IntroLines"] = m.introLines
		data["OutroLines"] = m.outroLines
	}
	data["FooterLines"] = m.footerLines
	data["ActionText"] = m.actionText
	data["ActionURL"] = m.actionURL
	data["Boundary"] = boundary
	data["FrontendURL"] = config.ServicePublicURL.GetString()
	data["CopyURLText"] = i18n.T(lang, "notifications.common.copy_url")

	if m.headerLine != nil {
		p := bluemonday.UGCPolicy()
		p.AllowDataURIImages()
		p.AllowAttrs("style").OnElements("img", "div")
		p.AllowStyles("border-radius", "vertical-align", "margin-right").OnElements("img")
		// #nosec G203 -- the html is sanitized
		data["HeaderLineHTML"] = templatehtml.HTML(p.Sanitize(m.headerLine.Text))
	}

	data["IntroLinesHTML"], err = convertLinesToHTML(m.introLines)
	if err != nil {
		return nil, err
	}

	data["OutroLinesHTML"], err = convertLinesToHTML(m.outroLines)
	if err != nil {
		return nil, err
	}

	data["FooterLinesHTML"], err = sanitizeLinesToHTML(m.footerLines)
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
		ThreadID:    m.threadID,
		EmbedFS: map[string]*embed.FS{
			"logo.png": &logo,
		},
	}

	return mailOpts, nil
}
