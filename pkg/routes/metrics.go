// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2020 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package routes

import (
	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/metrics"
	"code.vikunja.io/api/pkg/models"
	auth2 "code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/user"
	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func setupMetrics(a *echo.Group) {
	if !config.ServiceEnableMetrics.GetBool() {
		return
	}

	metrics.InitMetrics()

	type countable struct {
		Rediskey string
		Type     interface{}
	}

	for _, c := range []countable{
		{
			metrics.ListCountKey,
			models.List{},
		},
		{
			metrics.UserCountKey,
			user.User{},
		},
		{
			metrics.NamespaceCountKey,
			models.Namespace{},
		},
		{
			metrics.TaskCountKey,
			models.Task{},
		},
		{
			metrics.TeamCountKey,
			models.Team{},
		},
	} {
		// Set initial totals
		total, err := models.GetTotalCount(c.Type)
		if err != nil {
			log.Fatalf("Could not get initial count for %v, error was %s", c.Type, err)
		}
		if err := metrics.SetCount(total, c.Rediskey); err != nil {
			log.Fatalf("Could not set initial count for %v, error was %s", c.Type, err)
		}
	}

	a.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
}

func setupMetricsMiddleware(a *echo.Group) {
	if !config.ServiceEnableMetrics.GetBool() {
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

	return metrics.SetUserActive(auth)
}
