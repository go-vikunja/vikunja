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

package client

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"code.vikunja.io/veans/internal/output"
)

func writeTempFile(t *testing.T, name, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestUploadTaskAttachment_SendsMultipartFile(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v2/tasks/9/attachments" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		f, hdr, err := r.FormFile("files")
		if err != nil {
			t.Fatalf("FormFile: %v", err)
		}
		defer f.Close()
		body, _ := io.ReadAll(f)
		if hdr.Filename != "log.txt" || string(body) != "hello" {
			t.Errorf("got file %q with %q", hdr.Filename, body)
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"errors":null,"success":[{"id":5,"task_id":9,"file":{"id":1,"name":"log.txt","size":5}}]}`))
	}))
	defer srv.Close()

	a, err := New(srv.URL, "tk").UploadTaskAttachment(context.Background(), 9, writeTempFile(t, "log.txt", "hello"))
	if err != nil {
		t.Fatalf("UploadTaskAttachment: %v", err)
	}
	if a.ID != 5 || a.File == nil || a.File.Name != "log.txt" {
		t.Fatalf("unexpected attachment %+v", a)
	}
}

func TestUploadTaskAttachment_PerFileErrorFails(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"errors":[{"code":4013,"message":"file too large"}],"success":null}`))
	}))
	defer srv.Close()

	_, err := New(srv.URL, "tk").UploadTaskAttachment(context.Background(), 9, writeTempFile(t, "big.bin", "x"))
	var oe *output.Error
	if !errors.As(err, &oe) || oe.Code != output.CodeValidation {
		t.Fatalf("expected VALIDATION_ERROR, got %v", err)
	}
}
