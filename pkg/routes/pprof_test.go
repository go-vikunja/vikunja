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

package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"code.vikunja.io/api/pkg/config"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
)

func TestSetupPprof(t *testing.T) {
	config.InitDefaultConfig()
	get := func(t *testing.T, e *echo.Echo, path string, auth bool) int {
		t.Helper()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		if auth {
			req.SetBasicAuth("m", "s")
		}
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		return rec.Code
	}

	t.Run("off by default", func(t *testing.T) {
		config.MetricsEnabled.Set(true)
		config.MetricsPprof.Set(false)
		e := echo.New()
		setupPprof(e)
		assert.Equal(t, http.StatusNotFound, get(t, e, "/debug/pprof/cmdline", false))
	})
	t.Run("needs metrics enabled", func(t *testing.T) {
		config.MetricsEnabled.Set(false)
		config.MetricsPprof.Set(true)
		e := echo.New()
		setupPprof(e)
		assert.Equal(t, http.StatusNotFound, get(t, e, "/debug/pprof/cmdline", false))
	})
	t.Run("serves profiles behind the metrics basic auth", func(t *testing.T) {
		config.MetricsEnabled.Set(true)
		config.MetricsPprof.Set(true)
		config.MetricsUsername.Set("m")
		config.MetricsPassword.Set("s")
		defer config.MetricsUsername.Set("")
		defer config.MetricsPassword.Set("")
		e := echo.New()
		setupPprof(e)
		assert.Equal(t, http.StatusUnauthorized, get(t, e, "/debug/pprof/cmdline", false))
		assert.Equal(t, http.StatusOK, get(t, e, "/debug/pprof/cmdline", true))
		assert.Equal(t, http.StatusOK, get(t, e, "/debug/pprof/", true))
		assert.Equal(t, http.StatusOK, get(t, e, "/debug/pprof/goroutine", true))
	})
}
