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

package websocket

import (
	"sync"

	"code.vikunja.io/api/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

var hubMetricsOnce sync.Once

func registerHubMetrics() {
	hubMetricsOnce.Do(func() {
		metrics.GetRegistry().MustRegister(
			prometheus.NewGaugeFunc(prometheus.GaugeOpts{
				Name: "vikunja_websocket_connections",
				Help: "Open WebSocket connections on this process, including pending authentication.",
			}, func() float64 {
				h := GetHub()
				if h == nil {
					return 0
				}
				h.mu.RLock()
				defer h.mu.RUnlock()
				return float64(len(h.all))
			}),
			prometheus.NewCounterFunc(prometheus.CounterOpts{
				Name: "vikunja_websocket_resyncs_total",
				Help: "WebSocket disconnects requested for resynchronization or shutdown.",
			}, func() float64 {
				h := GetHub()
				if h == nil {
					return 0
				}
				h.mu.RLock()
				defer h.mu.RUnlock()
				return float64(h.resyncs)
			}),
		)
	})
}
