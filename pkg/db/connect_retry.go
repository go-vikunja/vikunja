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

package db

import (
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"
	"xorm.io/xorm"
)

func pingWithRetry(ping func() error, intervalMs, timeoutMs int) error {
	if intervalMs <= 0 {
		return ping()
	}

	if timeoutMs <= 0 {
		timeoutMs = 60000
	}

	deadline := time.Now().Add(time.Duration(timeoutMs) * time.Millisecond)
	var err error

	for attempt := 1; ; attempt++ {
		err = ping()
		if err == nil {
			if attempt > 1 {
				log.Infof("Connected to database after %d attempts", attempt)
			}
			return nil
		}

		if time.Now().After(deadline) {
			return err
		}

		log.Warningf("Could not connect to db (attempt %d), retrying in %dms: %v", attempt, intervalMs, err)
		time.Sleep(time.Duration(intervalMs) * time.Millisecond)
	}
}

func pingEngineWithRetry(engine *xorm.Engine) error {
	return pingWithRetry(
		engine.Ping,
		config.DatabaseConnectRetryInterval.GetInt(),
		config.DatabaseConnectRetryTimeout.GetInt(),
	)
}
