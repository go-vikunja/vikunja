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

func TestTaskURL(t *testing.T) {
	cases := []struct {
		name string
		cfg  Config
		want string
	}{
		{"frontend url wins", Config{Server: "https://api.example.com", FrontendURL: "https://vikunja.example.com/"}, "https://vikunja.example.com/tasks/42"},
		{"falls back to server", Config{Server: "https://vikunja.example.com"}, "https://vikunja.example.com/tasks/42"},
		{"keeps subpath", Config{FrontendURL: "https://example.com/vikunja/"}, "https://example.com/vikunja/tasks/42"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.cfg.TaskURL(42); got != tc.want {
				t.Errorf("got %q want %q", got, tc.want)
			}
		})
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

func TestSave_SymlinkOutsideDirRefused(t *testing.T) {
	base := t.TempDir()
	repo := filepath.Join(base, "repo")
	if err := os.Mkdir(repo, 0o755); err != nil {
		t.Fatal(err)
	}
	victim := filepath.Join(base, "victim.yml")
	const original = "server: https://x\nproject_id: 1\n"
	if err := os.WriteFile(victim, []byte(original), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(repo, Filename)
	if err := os.Symlink(victim, link); err != nil {
		t.Fatal(err)
	}

	c, err := Load(link)
	if err != nil {
		t.Fatal(err)
	}
	c.FrontendURL = "https://changed"
	if err := c.Save(); err == nil {
		t.Fatal("expected an error writing through a symlink escaping the config dir")
	}
	if buf, _ := os.ReadFile(victim); string(buf) != original {
		t.Fatalf("victim changed: %q", buf)
	}
}

func TestSave_InDirSymlinkWorks(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "conf"), 0o755); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(dir, "conf", "veans.yml")
	if err := os.WriteFile(target, []byte("server: https://x\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, Filename)
	if err := os.Symlink(filepath.Join("conf", "veans.yml"), link); err != nil {
		t.Fatal(err)
	}

	c, err := Load(link)
	if err != nil {
		t.Fatal(err)
	}
	c.FrontendURL = "https://changed"
	if err := c.Save(); err != nil {
		t.Fatal(err)
	}
	saved, err := Load(target)
	if err != nil {
		t.Fatal(err)
	}
	if saved.FrontendURL != "https://changed" {
		t.Fatalf("frontend_url = %q", saved.FrontendURL)
	}
}
