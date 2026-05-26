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

package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSuggestedBotUsername(t *testing.T) {
	cases := map[string]string{
		"/home/user/myrepo": "bot-myrepo",
		"/tmp/My Project":   "bot-my-project",
		"/x/Hello_World":    "bot-hello-world",
		"/x/CRAZY---Name!!": "bot-crazy-name",
		"/x/.dotted":        "bot-dotted",
	}
	for in, want := range cases {
		if got := SuggestedBotUsername(in); got != want {
			t.Errorf("%s: got %q, want %q", in, got, want)
		}
	}
}

func TestFormatTaskID(t *testing.T) {
	withIdent := &Config{ProjectIdentifier: "PROJ"}
	if got := withIdent.FormatTaskID(7); got != "PROJ-7" {
		t.Errorf("got %q want PROJ-7", got)
	}
	noIdent := &Config{}
	if got := noIdent.FormatTaskID(7); got != "#7" {
		t.Errorf("got %q want #7", got)
	}
}

func TestFindAndLoadRoundtrip(t *testing.T) {
	dir := t.TempDir()
	deeper := filepath.Join(dir, "a", "b", "c")
	if err := os.MkdirAll(deeper, 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := &Config{
		Server:            "https://example.com",
		ProjectID:         42,
		ProjectIdentifier: "PROJ",
		ViewID:            7,
		Buckets:           Buckets{Todo: 1, InProgress: 2, InReview: 3, Done: 4, Scrapped: 5},
		Bot:               Bot{Username: "bot-test", UserID: 99},
	}
	if err := cfg.SaveAs(filepath.Join(dir, Filename)); err != nil {
		t.Fatal(err)
	}

	// Find from the deeper directory should walk up.
	found, err := Find(deeper)
	if err != nil {
		t.Fatalf("Find: %v", err)
	}
	if !strings.HasSuffix(found, Filename) {
		t.Fatalf("found path %q does not end in %s", found, Filename)
	}
	loaded, err := Load(found)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.ProjectID != 42 || loaded.Bot.Username != "bot-test" {
		t.Fatalf("unexpected reload shape: %+v", loaded)
	}
}

func TestFindMissing(t *testing.T) {
	dir := t.TempDir()
	if _, err := Find(dir); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
