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

package initialize

import (
	"context"
	"fmt"

	"code.vikunja.io/api/pkg/cron"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/health"
	"code.vikunja.io/api/pkg/jobs"
	"code.vikunja.io/api/pkg/license"
	"code.vikunja.io/api/pkg/plugins"
	ws "code.vikunja.io/api/pkg/websocket"
)

// Shutdown is called after HTTP requests have drained.
func Shutdown(ctx context.Context) error {
	health.BeginShutdown()
	if hub := ws.GetHub(); hub != nil {
		defer hub.Shutdown()
	}
	defer license.Shutdown() // License validation remains active until shutdown.

	if err := cron.StopWithContext(ctx); err != nil {
		return fmt.Errorf("drain scheduled jobs: %w", err)
	}
	pluginsDone := make(chan struct{})
	go func() {
		plugins.Shutdown()
		close(pluginsDone)
	}()
	select {
	case <-pluginsDone:
	case <-ctx.Done():
		return fmt.Errorf("stop plugins: %w", ctx.Err())
	}
	if err := events.Shutdown(ctx); err != nil {
		return fmt.Errorf("drain events: %w", err)
	}

	return jobs.Shutdown(ctx)
}
