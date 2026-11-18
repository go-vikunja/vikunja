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
	"encoding/json"
	"strings"
	"testing"

	"code.vikunja.io/veans/internal/client"
	"code.vikunja.io/veans/internal/config"
)

func renderRuntime() *runtime {
	return &runtime{cfg: &config.Config{Server: "https://v.example", FrontendURL: "https://v.example/"}}
}

func taskWithRelated() *client.Task {
	return &client.Task{
		ID: 1, Title: "t", Identifier: "SF-1", Description: "<p>own</p>",
		RelatedTasks: map[string][]*client.Task{
			"parenttask": {{ID: 2, Identifier: "SF-2", Description: "<p>huge</p>"}},
		},
	}
}

func TestParseFields(t *testing.T) {
	fs, err := parseFields(" identifier, url ,")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(fs, ",") != "identifier,url" {
		t.Fatalf("got %v", fs)
	}
	if fs, err := parseFields(""); err != nil || fs != nil {
		t.Fatalf("empty: %v %v", fs, err)
	}
	if _, err := parseFields("identifier,bogus"); err == nil || !strings.Contains(err.Error(), "bogus") {
		t.Fatalf("want unknown-field error, got %v", err)
	}
}

func TestPrintTask_FieldsKeepOrderAndNullForAbsent(t *testing.T) {
	var buf bytes.Buffer
	fs, _ := parseFields("url,identifier,due_date")
	if err := renderRuntime().printTask(&buf, fs, taskWithRelated()); err != nil {
		t.Fatal(err)
	}
	want := `{"url":"https://v.example/tasks/1","identifier":"SF-1","due_date":null}`
	if got := strings.TrimSpace(buf.String()); got != want {
		t.Fatalf("got  %s\nwant %s", got, want)
	}
}

// Without --fields the task prints whole, except nested related-task
// descriptions, whose URL must not be the first "url" in the output either.
func TestPrintTask_DefaultStripsNestedDescriptions(t *testing.T) {
	var buf bytes.Buffer
	if err := renderRuntime().printTask(&buf, nil, taskWithRelated()); err != nil {
		t.Fatal(err)
	}
	var got client.Task
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Description != "<p>own</p>" {
		t.Errorf("top-level description changed: %q", got.Description)
	}
	parent := got.RelatedTasks["parenttask"][0]
	if parent.Description != "" || parent.URL != "https://v.example/tasks/2" {
		t.Errorf("nested task: description=%q url=%q", parent.Description, parent.URL)
	}
	if strings.Contains(buf.String(), "huge") {
		t.Errorf("nested description leaked: %s", buf.String())
	}
}

func TestPrintTasks_FieldsAndEmpty(t *testing.T) {
	rt := renderRuntime()
	var buf bytes.Buffer
	if err := rt.printTasks(&buf, nil, nil); err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(buf.String()); got != "[]" {
		t.Fatalf("empty list should print [], got %s", got)
	}
	buf.Reset()
	fs, _ := parseFields("identifier")
	if err := rt.printTasks(&buf, fs, []*client.Task{taskWithRelated()}); err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(buf.String()); got != `[{"identifier":"SF-1"}]` {
		t.Fatalf("got %s", got)
	}
}
