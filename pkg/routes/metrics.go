// Copyright 2019 Vikunja and contriubtors. All rights reserved.
//
// This file is part of Vikunja.
//
// Vikunja is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Vikunja is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Vikunja.  If not, see <https://www.gnu.org/licenses/>.

package routes

import (
	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/metrics"
	"code.vikunja.io/api/pkg/models"
	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func setupMetrics(a *echo.Group) {
	if !config.ServiceEnableMetrics.GetBool() {
		return
	}

	if !config.RedisEnabled.GetBool() {
		log.Fatal("You have to enable redis in order to use metrics")
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
			models.User{},
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

	// init active users, sometimes we'll have garbage from previous runs in redis instead
	if err := metrics.SetActiveUsers([]*metrics.ActiveUser{}); err != nil {
		log.Fatalf("Could not set initial count for active users, error was %s", err)
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
			if err := models.UpdateActiveUsersFromContext(c); err != nil {
				log.Error(err)
				return next(c)
			}
			return next(c)
		}
	})
}
