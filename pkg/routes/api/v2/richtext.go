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

package apiv2

import (
	"context"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/humaecho5"
	"code.vikunja.io/api/pkg/richtext"

	"github.com/danielgtaylor/huma/v2"
	"github.com/labstack/echo/v5"
)

const (
	// "markdown" converts rich-text fields on read and write; anything else keeps HTML.
	richTextFormatQuery  = "format"
	richTextFormatHeader = "X-Vikunja-Format"
	markdownFormat       = "markdown"
)

// requestWantsMarkdown reports whether the request asked for markdown. The per-op
// `format` query field on the input structs only documents the param; the value is
// read here so this also catches the X-Vikunja-Format header — the only channel
// that survives AutoPatch's PATCH re-dispatch (it strips the query).
func requestWantsMarkdown(ctx context.Context) bool {
	ec, ok := ctx.Value(humaecho5.EchoContextKey).(*echo.Context)
	if !ok {
		return false
	}
	return ec.QueryParam(richTextFormatQuery) == markdownFormat ||
		ec.Request().Header.Get(richTextFormatHeader) == markdownFormat
}

// richTextFormatAPIDescription documents the cross-cutting markdown behavior at
// the top of the OpenAPI spec (Scalar renders it on the docs landing page).
const richTextFormatAPIDescription = "## Rich-text fields\n\n" +
	"Descriptions (task, project, label, team, saved filter) and task comments are stored as HTML. " +
	"Add `?format=markdown` to read and write them as GFM Markdown instead; on write it is converted " +
	"to HTML and `@mentions` resolved to existing users. On `PATCH`, send the `X-Vikunja-Format: markdown` " +
	"header instead (merge-patch drops query parameters). CalDAV always exchanges task descriptions as " +
	"Markdown.\n\n" +
	"Writing is lossy: Markdown can't express every HTML construct (e.g. underline), so a field you send " +
	"as Markdown is stored as its converted HTML — formatting Markdown can't represent is dropped. Omit a " +
	"field, or use `format=html`, to leave it untouched (note a full `PUT` and `PATCH` round-trip the " +
	"whole resource, so send `format=html` unless you actually edited the rich-text fields). Unknown " +
	"`@mentions` stay as plain text."

// stripPatchFormatQuery removes the `format` query param AutoPatch copies onto
// each synthesised PATCH. The query doesn't survive AutoPatch's re-dispatch, so
// advertising it on PATCH would be a trap (markdown silently stored as HTML);
// PATCH uses the X-Vikunja-Format header instead. Call after EnableAutoPatch.
func stripPatchFormatQuery(api huma.API) {
	for _, item := range api.OpenAPI().Paths {
		if item == nil || item.Patch == nil {
			continue
		}
		kept := item.Patch.Parameters[:0]
		for _, p := range item.Patch.Parameters {
			if p.Name == richTextFormatQuery && p.In == "query" {
				continue
			}
			kept = append(kept, p)
		}
		item.Patch.Parameters = kept
	}
}

// convertToMarkdown converts the given HTML fields to Markdown in place when the
// request asked for markdown. Read handlers call it on returned fields; write
// handlers after persisting, to echo back in the requested format. Best effort: a
// conversion error leaves the HTML untouched.
func convertToMarkdown(ctx context.Context, fields ...*string) {
	if !requestWantsMarkdown(ctx) {
		return
	}
	for _, field := range fields {
		if field == nil {
			continue
		}
		if md, err := richtext.HTMLToMarkdown(*field); err == nil {
			*field = md
		}
	}
}

// convertTasksToMarkdown converts each task's description plus any expanded
// rich-text children (comments, related tasks) to markdown when requested. Dedups
// by field pointer so a task reachable twice (e.g. as another's relation) isn't
// converted twice — a second HTML→markdown pass would escape the markdown.
func convertTasksToMarkdown(ctx context.Context, tasks ...*models.Task) {
	if !requestWantsMarkdown(ctx) {
		return
	}
	seen := map[*string]struct{}{}
	toMarkdown := func(field *string) {
		if field == nil {
			return
		}
		if _, done := seen[field]; done {
			return
		}
		seen[field] = struct{}{}
		if md, err := richtext.HTMLToMarkdown(*field); err == nil {
			*field = md
		}
	}
	for _, task := range tasks {
		if task == nil {
			continue
		}
		toMarkdown(&task.Description)
		for _, comment := range task.Comments {
			if comment != nil {
				toMarkdown(&comment.Comment)
			}
		}
		for _, related := range task.RelatedTasks {
			for _, rel := range related {
				if rel != nil {
					toMarkdown(&rel.Description)
				}
			}
		}
	}
}

// convertToHTML converts the given Markdown fields to canonical HTML in place,
// rebuilding @mentions, when the request asked for markdown (no-op otherwise).
// Write handlers call it on the request body before persisting.
func convertToHTML(ctx context.Context, fields ...*string) error {
	if !requestWantsMarkdown(ctx) {
		return nil
	}

	s := db.NewSession()
	defer s.Close()

	for _, field := range fields {
		if field == nil {
			continue
		}
		htmlDesc, err := richtext.MarkdownToHTMLWithMentions(s, *field)
		if err != nil {
			return err
		}
		*field = htmlDesc
	}
	return nil
}
