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

package commands

import (
	"bytes"
	"strings"
	"testing"
	"text/template"

	"code.vikunja.io/veans/internal/config"
)

func TestPrimeTemplate_RendersAnchors(t *testing.T) {
	data := primeContext{
		Server:            "https://vikunja.example.com",
		ProjectID:         42,
		ProjectTitle:      "Test Project",
		ProjectIdentifier: "PROJ",
		ViewID:            7,
		Buckets:           config.Buckets{Todo: 11, InProgress: 12, InReview: 13, Done: 14, Scrapped: 15},
		BotUsername:       "bot-myrepo",
		TaskIDExample:     "PROJ-1",
	}
	tpl, err := template.New("prime").Parse(promptTemplate)
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		t.Fatal(err)
	}
	out := buf.String()

	mustContain := []string{
		"<EXTREMELY_IMPORTANT>",
		"</EXTREMELY_IMPORTANT>",
		"bot-myrepo",
		"Test Project",
		"PROJ-1",
		"Refs:",
		"veans claim",
		"veans list --ready",
		"--description-replace-old",
		"Todo",
		"In Progress",
		"In Review",
		"Done",
		"Scrapped",
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("rendered prompt missing %q", s)
		}
	}

	// Buckets show concrete IDs.
	for _, want := range []string{"`11`", "`12`", "`13`", "`14`", "`15`"} {
		if !strings.Contains(out, want) {
			t.Errorf("bucket id %s not present in output", want)
		}
	}
}

func TestPrimeTemplate_NoIdentifierFallback(t *testing.T) {
	data := primeContext{
		ProjectTitle:      "No Ident",
		ProjectIdentifier: "",
		BotUsername:       "bot-x",
		TaskIDExample:     "#1",
		Server:            "https://vikunja.example.com",
		ProjectID:         1,
		ViewID:            1,
	}
	tpl, _ := template.New("prime").Parse(promptTemplate)
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "no identifier") {
		t.Errorf("expected fallback copy when project has no identifier; got:\n%s", out)
	}
	if !strings.Contains(out, "#NN") {
		t.Errorf("expected #NN format mention in fallback")
	}
}
