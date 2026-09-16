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
	"crypto/subtle"
	"net/http"
	"net/http/pprof"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/metrics"
	"code.vikunja.io/api/pkg/models"
	auth2 "code.vikunja.io/api/pkg/modules/auth"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func setupMetrics(a *echo.Group) {
	if !config.MetricsEnabled.GetBool() {
		return
	}

	metrics.InitMetrics()

	r := a.Group("/metrics")
	if auth := metricsBasicAuth(); auth != nil {
		r.Use(auth)
	}

	r.GET("", echo.WrapHandler(promhttp.HandlerFor(metrics.GetRegistry(), promhttp.HandlerOpts{})))
}

func metricsBasicAuth() echo.MiddlewareFunc {
	if config.MetricsUsername.GetString() == "" || config.MetricsPassword.GetString() == "" {
		return nil
	}
	return middleware.BasicAuth(func(_ *echo.Context, username, password string) (bool, error) {
		if subtle.ConstantTimeCompare([]byte(username), []byte(config.MetricsUsername.GetString())) == 1 &&
			subtle.ConstantTimeCompare([]byte(password), []byte(config.MetricsPassword.GetString())) == 1 {
			return true, nil
		}
		return false, nil
	})
}

// Go's profiler at its canonical path (pprof.Index derives the profile name from it), gated like the metrics
// endpoint. Explicit handlers rather than the net/http/pprof blank import, which would register them on
// http.DefaultServeMux unconditionally.
func setupPprof(e *echo.Echo) {
	if !config.MetricsEnabled.GetBool() || !config.MetricsPprof.GetBool() {
		return
	}

	r := e.Group("/debug/pprof")
	if auth := metricsBasicAuth(); auth != nil {
		r.Use(auth)
	}

	r.GET("/cmdline", echo.WrapHandler(http.HandlerFunc(pprof.Cmdline)))
	r.GET("/profile", echo.WrapHandler(http.HandlerFunc(pprof.Profile)))
	r.GET("/symbol", echo.WrapHandler(http.HandlerFunc(pprof.Symbol)))
	r.POST("/symbol", echo.WrapHandler(http.HandlerFunc(pprof.Symbol)))
	r.GET("/trace", echo.WrapHandler(http.HandlerFunc(pprof.Trace)))
	r.GET("/", echo.WrapHandler(http.HandlerFunc(pprof.Index)))
	r.GET("/*", echo.WrapHandler(http.HandlerFunc(pprof.Index)))
}

func setupMetricsMiddleware(a *echo.Group) {
	if !config.MetricsEnabled.GetBool() {
		return
	}

	a.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			// Unauthenticated routes share this middleware and have no user to count.
			if auth2.HasAuthInContext(c) {
				if err := updateActiveUsersFromContext(c); err != nil {
					log.Error(err)
				}
			}
			return next(c)
		}
	})
}

// updateActiveUsersFromContext updates the currently active users in redis
func updateActiveUsersFromContext(c *echo.Context) (err error) {
	auth, err := auth2.GetAuthFromClaims(c)
	if err != nil {
		return
	}

	if _, is := auth.(*models.LinkSharing); is {
		return metrics.SetLinkShareActive(auth)
	}

	return metrics.SetUserActive(auth)
}
