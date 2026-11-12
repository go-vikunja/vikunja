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

package cron

import (
	"errors"
	"time"

	"code.vikunja.io/api/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"xorm.io/builder"
	"xorm.io/xorm/schemas"
)

func (r *Runner) registerLag(name string) error {
	err := metrics.GetRegistry().Register(prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name:        "vikunja_cron_lag_seconds",
		Help:        "Seconds since the last committed scheduling watermark.",
		ConstLabels: prometheus.Labels{"job": name},
	}, func() float64 {
		if r.engine.Dialect().URI().DBType == schemas.POSTGRES {
			var row struct{ Lag float64 }
			found, err := r.engine.SQL("SELECT GREATEST(0, EXTRACT(EPOCH FROM ((clock_timestamp() AT TIME ZONE 'UTC') - last_run_until))) AS lag FROM cron_state WHERE job = ?", name).Get(&row)
			if err != nil || !found {
				return 0
			}
			return row.Lag
		}
		state := State{}
		found, err := r.engine.Where(builder.Eq{"job": name}).Get(&state)
		if err != nil || !found {
			return 0
		}
		return max(0, time.Since(state.LastRunUntil).Seconds())
	}))
	var registered prometheus.AlreadyRegisteredError
	if errors.As(err, &registered) {
		return nil
	}
	return err
}
