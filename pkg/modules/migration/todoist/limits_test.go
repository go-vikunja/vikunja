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

package todoist

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/modules/migration"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/require"
)

// Completed tasks paginate forever, one new task per page.
func fakeTodoist(t *testing.T, syncBody string) {
	t.Helper()
	prevAllowNonRoutable := config.OutgoingRequestsAllowNonRoutableIPs.GetBool()
	config.OutgoingRequestsAllowNonRoutableIPs.Set("true")

	var page atomic.Int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/oauth/access_token":
			_, _ = w.Write([]byte(`{"access_token":"token"}`))
		case r.URL.Path == "/api/v1/sync":
			_, _ = w.Write([]byte(syncBody))
		case r.URL.Path == "/api/v1/tasks/completed":
			n := page.Add(1)
			_, _ = fmt.Fprintf(w, `{"items":[{"task_id":"done%d"}],"next_cursor":"c%d"}`, n, n)
		case strings.HasPrefix(r.URL.Path, "/api/v1/tasks/"):
			_, _ = w.Write([]byte(`{"id":"x","content":"done"}`))
		default:
			_, _ = w.Write([]byte(`{"results":[]}`))
		}
	}))

	prevToken, prevAPI := oauthTokenURL, apiPrefix
	oauthTokenURL, apiPrefix = srv.URL+"/oauth/access_token", srv.URL+"/api/v1/"
	t.Cleanup(func() {
		srv.Close()
		oauthTokenURL, apiPrefix = prevToken, prevAPI
		config.OutgoingRequestsAllowNonRoutableIPs.Set(prevAllowNonRoutable)
	})
}

func TestMigrateStopsAtItemLimit(t *testing.T) {
	fakeTodoist(t, `{"items":[{"id":"1"},{"id":"2"}]}`)
	config.MigrationMaxItems.Set(5)
	t.Cleanup(func() { config.MigrationMaxItems.Set(100000) })

	err := (&Migration{Code: "code"}).Migrate(&user.User{ID: 1})
	var limitErr *migration.ErrImportItemLimitExceeded
	require.ErrorAs(t, err, &limitErr)
}

func TestMigrateRejectsOversizedResponse(t *testing.T) {
	fakeTodoist(t, `{"items":[],"padding":"`+strings.Repeat("x", 2048)+`"}`)
	prevMax := migration.MaxResponseBytes
	migration.MaxResponseBytes = 1024
	t.Cleanup(func() { migration.MaxResponseBytes = prevMax })

	err := (&Migration{Code: "code"}).Migrate(&user.User{ID: 1})
	require.ErrorIs(t, err, migration.ErrResponseTooLarge)
}
