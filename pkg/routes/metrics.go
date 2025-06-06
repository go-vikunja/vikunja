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

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/metrics"
	"code.vikunja.io/api/pkg/models"
	auth2 "code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/user"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func setupMetrics(a *echo.Group) {
	if !config.MetricsEnabled.GetBool() {
		return
	}

	metrics.InitMetrics()

	type countable struct {
		Key  string
		Type interface{}
	}

	for _, c := range []countable{
		{
			metrics.ProjectCountKey,
			models.Project{},
		},
		{
			metrics.UserCountKey,
			user.User{},
		},
		{
			metrics.TaskCountKey,
			models.Task{},
		},
		{
			metrics.TeamCountKey,
			models.Team{},
		},
		{
			metrics.FilesCountKey,
			files.File{},
		},
		{
			metrics.AttachmentsCountKey,
			models.TaskAttachment{},
		},
	} {
		// Set initial totals
		total, err := models.GetTotalCount(c.Type)
		if err != nil {
			log.Fatalf("Could not get initial count for %v, error was %s", c.Type, err)
		}
		if err := metrics.SetCount(total, c.Key); err != nil {
			log.Fatalf("Could not set initial count for %v, error was %s", c.Type, err)
		}
	}

	r := a.Group("/metrics")

	if config.MetricsUsername.GetString() != "" && config.MetricsPassword.GetString() != "" {
		r.Use(middleware.BasicAuth(func(username, password string, _ echo.Context) (bool, error) {
			if subtle.ConstantTimeCompare([]byte(username), []byte(config.MetricsUsername.GetString())) == 1 &&
				subtle.ConstantTimeCompare([]byte(password), []byte(config.MetricsPassword.GetString())) == 1 {
				return true, nil
			}
			return false, nil
		}))
	}

	r.GET("", echo.WrapHandler(promhttp.HandlerFor(metrics.GetRegistry(), promhttp.HandlerOpts{})))
}

func setupMetricsMiddleware(a *echo.Group) {
	if !config.MetricsEnabled.GetBool() {
		return
	}

	a.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			// Update currently active users
			if err := updateActiveUsersFromContext(c); err != nil {
				log.Error(err)
				return next(c)
			}
			return next(c)
		}
	})
}

// updateActiveUsersFromContext updates the currently active users in redis
func updateActiveUsersFromContext(c echo.Context) (err error) {
	auth, err := auth2.GetAuthFromClaims(c)
	if err != nil {
		return
	}

	if _, is := auth.(*models.LinkSharing); is {
		return metrics.SetLinkShareActive(auth)
	}

	return metrics.SetUserActive(auth)
}
