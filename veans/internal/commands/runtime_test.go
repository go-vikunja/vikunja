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
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"code.vikunja.io/veans/internal/client"
	"code.vikunja.io/veans/internal/config"
	"code.vikunja.io/veans/internal/output"
)

func TestBackfillFrontendURL(t *testing.T) {
	cases := []struct {
		name        string
		frontendURL string
		want        func(server string) string
	}{
		{"uses /info", "https://vikunja.example.com/", func(string) string { return "https://vikunja.example.com/" }},
		{"falls back to server when unset", "", func(server string) string { return server }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/v2/info" {
					t.Errorf("unexpected request %s", r.URL.Path)
				}
				_, _ = w.Write([]byte(`{"frontend_url":"` + tc.frontendURL + `"}`))
			}))
			defer srv.Close()

			path := filepath.Join(t.TempDir(), config.Filename)
			cfg := &config.Config{Server: srv.URL, ProjectID: 1}
			if err := cfg.SaveAs(path); err != nil {
				t.Fatal(err)
			}
			rt := &runtime{cfg: cfg, client: client.New(srv.URL, "")}

			rt.backfillFrontendURL(context.Background())

			want := tc.want(srv.URL)
			if cfg.FrontendURL != want {
				t.Errorf("in memory: got %q want %q", cfg.FrontendURL, want)
			}
			saved, err := config.Load(path)
			if err != nil {
				t.Fatal(err)
			}
			if saved.FrontendURL != want {
				t.Errorf("on disk: got %q want %q", saved.FrontendURL, want)
			}
		})
	}
}

func TestBackfillFrontendURL_InfoFailsLeavesConfig(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	path := filepath.Join(t.TempDir(), config.Filename)
	cfg := &config.Config{Server: srv.URL, ProjectID: 1}
	if err := cfg.SaveAs(path); err != nil {
		t.Fatal(err)
	}
	rt := &runtime{cfg: cfg, client: client.New(srv.URL, "")}

	rt.backfillFrontendURL(context.Background())

	if cfg.FrontendURL != "" {
		t.Errorf("got %q, want empty so the next run retries", cfg.FrontendURL)
	}
}

func TestBackfillFrontendURL_DoesNotFollowSymlinkOutOfRepo(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"frontend_url":"https://vikunja.example.com/"}`))
	}))
	defer srv.Close()

	base := t.TempDir()
	repo := filepath.Join(base, "repo")
	if err := os.Mkdir(repo, 0o755); err != nil {
		t.Fatal(err)
	}
	victim := filepath.Join(base, "victim.yml")
	original := "server: " + srv.URL + "\nproject_id: 1\n"
	if err := os.WriteFile(victim, []byte(original), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(repo, config.Filename)
	if err := os.Symlink(victim, link); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load(link)
	if err != nil {
		t.Fatal(err)
	}
	rt := &runtime{cfg: cfg, client: client.New(srv.URL, "")}

	rt.backfillFrontendURL(context.Background())

	if buf, _ := os.ReadFile(victim); string(buf) != original {
		t.Fatalf("backfill wrote through the symlink: %q", buf)
	}
}

func TestResolveTaskID_BareNumberNotFoundExplainsIndex(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"items":[],"total_pages":1}`))
	}))
	t.Cleanup(srv.Close)
	rt := newTestRuntime(srv.URL)
	rt.cfg.ProjectIdentifier = "CL"

	_, err := rt.resolveTaskID(context.Background(), "25787")
	var oe *output.Error
	if !errors.As(err, &oe) || oe.Code != output.CodeNotFound {
		t.Fatalf("want NOT_FOUND, got %v", err)
	}
	for _, want := range []string{"CL-25787", "project index", "--id 25787"} {
		if !strings.Contains(oe.Message, want) {
			t.Errorf("message %q lacks %q", oe.Message, want)
		}
	}
}

func TestTaskID_ByIDChecksProject(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		project := 7
		if r.URL.Path == "/api/v2/tasks/2" {
			project = 8
		}
		_, _ = w.Write(fmt.Appendf(nil, `{"id":1,"project_id":%d}`, project))
	}))
	t.Cleanup(srv.Close)
	rt := newTestRuntime(srv.URL)

	if id, err := rt.taskID(context.Background(), nil, 1); err != nil || id != 1 {
		t.Fatalf("own project: id=%d err=%v", id, err)
	}
	if _, err := rt.taskID(context.Background(), nil, 2); err == nil {
		t.Fatal("task from another project should be rejected")
	}
}

func TestAddTaskIDFlag_Args(t *testing.T) {
	cases := []struct {
		args    []string
		wantErr bool
	}{
		{[]string{"SF-1"}, false},
		{[]string{"--id", "25787"}, false},
		{[]string{"SF-1", "--id", "25787"}, true},
		{nil, true},
		{[]string{"--id", "0"}, true},
	}
	for _, tc := range cases {
		cmd := &cobra.Command{Use: "x", RunE: func(*cobra.Command, []string) error { return nil }}
		var id int64
		cmd.Args = addTaskIDFlag(cmd, &id)
		cmd.SetArgs(tc.args)
		cmd.SilenceUsage, cmd.SilenceErrors = true, true
		if err := cmd.Execute(); (err != nil) != tc.wantErr {
			t.Errorf("%v: err=%v, wantErr=%v", tc.args, err, tc.wantErr)
		}
	}
}
