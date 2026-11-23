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
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"code.vikunja.io/veans/internal/output"
)

// TestRunCreate_PartialFailureNamesCreatedTask reproduces a bot token without
// the bucket-move scope: the task exists, so the error must name it and the
// steps that already succeeded.
func TestRunCreate_PartialFailureNamesCreatedTask(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v2/projects/7/tasks":
			_ = json.NewEncoder(w).Encode(map[string]any{"id": 50, "index": 5})
		case r.Method == http.MethodPut && r.URL.Path == "/api/v2/projects/7/views/1/buckets/11/tasks":
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 11, "message": "missing, malformed, expired or otherwise invalid token provided",
			})
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			http.Error(w, "unexpected", http.StatusInternalServerError)
		}
	}))
	t.Cleanup(srv.Close)
	rt := newTestRuntime(srv.URL)
	rt.cfg.ProjectIdentifier = "SF"

	_, err := runCreate(context.Background(), rt, "probe", &createFlags{statusName: "in-progress"})

	var oe *output.Error
	if !errors.As(err, &oe) {
		t.Fatalf("want *output.Error, got %v", err)
	}
	if oe.Code != output.CodeAuth {
		t.Errorf("code = %s, want %s", oe.Code, output.CodeAuth)
	}
	if oe.Task != "SF-5" || oe.Failed != "set bucket" || !reflect.DeepEqual(oe.Completed, []string{"create"}) {
		t.Errorf("task=%q failed=%q completed=%v", oe.Task, oe.Failed, oe.Completed)
	}
	if !strings.Contains(oe.Message, "SF-5") || !strings.Contains(oe.Message, "veans login") {
		t.Errorf("message should name the task and the fix: %s", oe.Message)
	}
}
