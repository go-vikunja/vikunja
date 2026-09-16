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

package planka

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/files"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPlankaJobBudget covers all job budgets from GHSA-wq92-8x3r-fm38.
func TestPlankaJobBudget(t *testing.T) {
	assertBudgetErr := func(t *testing.T, err error) {
		t.Helper()
		require.Error(t, err)
		var budgetErr *ErrImportBudgetExceeded
		require.ErrorAs(t, err, &budgetErr, "expected ErrImportBudgetExceeded, got %v", err)
	}

	newBudgetClient := func(t *testing.T) (*client, *fakePlanka) {
		t.Helper()
		f, srv := newFake(t)
		f.validAPIKey = "key"
		f.fixtures = plankaFixtures

		c, err := newClient(srv.URL)
		require.NoError(t, err)
		require.NoError(t, c.login(t.Context(), "key", "", ""))
		return c, f
	}

	t.Run("aggregate response budget stops the fetch", func(t *testing.T) {
		c, _ := newBudgetClient(t)
		c.budget.maxAggregateBytes = 100

		_, err := fetchAll(c)
		assertBudgetErr(t, err)
	})

	t.Run("a single response is capped at 4 MiB", func(t *testing.T) {
		padding := strings.Repeat("x", 5*1024*1024)
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"items":["` + padding + `"]}`))
		}))
		t.Cleanup(srv.Close)

		c, err := newClient(srv.URL)
		require.NoError(t, err)

		out := &struct {
			Items []string `json:"items"`
		}{}
		err = c.get("/api/projects", nil, out)
		assertBudgetErr(t, err)
	})

	t.Run("an oversized child response aborts the board fetch with a budget error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			if r.URL.Path == "/api/boards/1" {
				_, _ = w.Write([]byte(`{"item":{"id":"1"},"included":{"lists":[{"id":"2","type":"archive"}]}}`))
				return
			}
			_, _ = w.Write([]byte(`{"items":[],"padding":"` + strings.Repeat("x", 512) + `"}`))
		}))
		t.Cleanup(srv.Close)

		c, err := newClient(srv.URL)
		require.NoError(t, err)
		c.budget.maxResponseBytes = 256

		_, err = fetchBoard(c, "1")
		assertBudgetErr(t, err)
	})

	t.Run("aggregate response bytes are enforced when decoding fails", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"items":["` + strings.Repeat("x", 256) + `"`))
		}))
		t.Cleanup(srv.Close)

		c, err := newClient(srv.URL)
		require.NoError(t, err)
		c.budget.maxResponseBytes = 1024
		c.budget.maxAggregateBytes = 64

		err = c.get("/api/projects", nil, &projectsResponse{})
		assertBudgetErr(t, err)
	})

	t.Run("decoded entities are bounded", func(t *testing.T) {
		c, _ := newBudgetClient(t)
		c.budget.maxEntities = 3

		_, err := fetchAll(c)
		assertBudgetErr(t, err)
	})

	t.Run("entity budget rejects a response before allocating endpoint slices", func(t *testing.T) {
		body := `{"items":[{"id":"1"},{"id":"2"}],"included":{}}`
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(body))
		}))
		t.Cleanup(srv.Close)

		c, err := newClient(srv.URL)
		require.NoError(t, err)
		c.budget.maxEntities = 1
		out := &projectsResponse{}

		err = c.get("/api/projects", nil, out)
		assertBudgetErr(t, err)
		assert.Empty(t, out.Items)
		assert.Zero(t, c.budget.entities)
	})

	for _, tc := range []struct {
		name          string
		boardResponse string
		childPath     string
		childResponse string
	}{
		{
			name:          "archived card budget exhaustion aborts the board fetch",
			boardResponse: `{"item":{"id":"1"},"included":{"lists":[{"id":"2","type":"archive"}]}}`,
			childPath:     "/api/lists/2/cards",
			childResponse: `{"items":[{"id":"3","listChangedAt":"2024-01-01T00:00:00Z"}],"included":{}}`,
		},
		{
			name:          "comment budget exhaustion aborts the board fetch",
			boardResponse: `{"item":{"id":"1"},"included":{"cards":[{"id":"3","commentsTotal":1}]}}`,
			childPath:     "/api/cards/3/comments",
			childResponse: `{"items":[{"id":"4"}],"included":{}}`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			requestedPath := make(chan string, 1)
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if r.URL.Path == "/api/boards/1" {
					_, _ = w.Write([]byte(tc.boardResponse))
					return
				}
				requestedPath <- r.URL.Path
				_, _ = w.Write([]byte(tc.childResponse))
			}))
			t.Cleanup(srv.Close)

			c, err := newClient(srv.URL)
			require.NoError(t, err)
			c.budget.maxEntities = 1

			_, err = fetchBoard(c, "1")
			assertBudgetErr(t, err)
			select {
			case path := <-requestedPath:
				assert.Equal(t, tc.childPath, path)
			case <-time.After(time.Second):
				t.Fatal("timed out waiting for child request")
			}
		})
	}

	t.Run("outbound request attempts are bounded, including retries", func(t *testing.T) {
		attempts := 0
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			attempts++
			w.WriteHeader(http.StatusInternalServerError)
		}))
		t.Cleanup(srv.Close)

		c, err := newClient(srv.URL)
		require.NoError(t, err)
		c.budget.maxRequests = 2

		_, err = fetchAll(c)
		assertBudgetErr(t, err)
		assert.LessOrEqual(t, attempts, 2, "the request budget must stop outbound attempts")
	})

	t.Run("retained attachment bytes are bounded", func(t *testing.T) {
		c, _ := newBudgetClient(t)
		c.budget.maxAttachmentBytes = 4 * 1024 * 1024

		buf, err := c.download("7", "big.bin")
		require.NoError(t, err)
		assert.Equal(t, 3*1024*1024, buf.Len())

		_, err = c.download("7", "big.bin")
		assertBudgetErr(t, err)
	})

	t.Run("ordinary attachment failures are still skipped, not budget errors", func(t *testing.T) {
		c, _ := newBudgetClient(t)

		_, err := c.download("missing", "404.bin")
		require.Error(t, err)
		var budgetErr *ErrImportBudgetExceeded
		assert.NotErrorAs(t, err, &budgetErr, "a missing attachment is an ordinary failure, got %v", err)
	})

	t.Run("oversized attachment attempts consume the aggregate budget", func(t *testing.T) {
		oldMax := config.FilesMaxSize.GetString()
		config.FilesMaxSize.Set("1MB")
		require.NoError(t, config.SetMaxFileSizeMBytesFromString("1MB"))
		t.Cleanup(func() {
			config.FilesMaxSize.Set(oldMax)
			_ = config.SetMaxFileSizeMBytesFromString(oldMax)
		})

		c, _ := newBudgetClient(t)
		c.budget.maxAttachmentBytes = 2 * 1024 * 1024

		_, err := c.download("7", "big.bin")
		var tooLarge files.ErrFileIsTooLarge
		require.ErrorAs(t, err, &tooLarge)
		assert.Positive(t, c.budget.attachmentBytes)

		_, err = c.download("7", "big.bin")
		assertBudgetErr(t, err)
		assert.NotErrorAs(t, err, &tooLarge)
	})

	t.Run("a complete small import still works within the budget", func(t *testing.T) {
		c, _ := newBudgetClient(t)

		data, err := fetchAll(c)
		require.NoError(t, err)
		require.NotEmpty(t, data.Projects)
		assert.LessOrEqual(t, c.budget.responseBytes, c.budget.maxAggregateBytes)
		assert.LessOrEqual(t, c.budget.entities, c.budget.maxEntities)
		assert.Positive(t, c.budget.entities)
	})
}
