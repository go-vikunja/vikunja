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

package jobs

import (
	"code.vikunja.io/api/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"xorm.io/xorm"
)

type metricLabel struct{ event, queue, kind string }

var jobCounters = func() map[string]*prometheus.CounterVec {
	counters := make(map[string]*prometheus.CounterVec)
	for name, help := range map[string]string{
		"enqueued":        "Jobs inserted and committed, excluding duplicate live keys.",
		"completed":       "Job completions committed by the owning attempt.",
		"failed_attempts": "Handler failures recorded by the owning attempt.",
		"rescued":         "Expired claims successfully rescued.",
		"parked":          "Jobs parked after exhausting their attempt budget.",
	} {
		counter := prometheus.NewCounterVec(prometheus.CounterOpts{Name: "vikunja_jobs_" + name + "_total", Help: help}, []string{"queue", "kind"})
		metrics.GetRegistry().MustRegister(counter)
		counters[name] = counter
	}
	return counters
}()

func countAfterCommit(s *xorm.Session, event, queue, kind string, count int64) {
	if count == 0 {
		return
	}
	label := metricLabel{event, queue, kind}
	if !s.IsInTx() {
		jobCounters[event].WithLabelValues(queue, kind).Add(float64(count))
		return
	}
	pending := pendingFor(s)
	pending.Lock()
	pending.counts[label] += count
	pending.Unlock()
}
