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

package metrics

import (
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/modules/keyvalue"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	ProjectCountKey     = `project_count`
	UserCountKey        = `user_count`
	TaskCountKey        = `task_count`
	TeamCountKey        = `team_count`
	FilesCountKey       = `files_count`
	AttachmentsCountKey = `attachments_count`
)

// countCacheTTL is how long a cached entity count is served before it is recomputed
// from the database. The counts are inherently approximate (Prometheus samples them),
// so a short staleness window is fine and keeps the cache self-healing — a missed
// InvalidateCount call costs at most this much staleness, never a permanent drift.
const countCacheTTL = 30 * time.Second

// countTables maps each count metric key to the database table it counts.
var countTables = map[string]string{
	ProjectCountKey:     "projects",
	UserCountKey:        "users",
	TaskCountKey:        "tasks",
	TeamCountKey:        "teams",
	FilesCountKey:       "files",
	AttachmentsCountKey: "task_attachments",
}

var registry *prometheus.Registry

func GetRegistry() *prometheus.Registry {
	if registry == nil {
		registry = prometheus.NewRegistry()
		registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
		registry.MustRegister(collectors.NewGoCollector())
	}

	return registry
}

func registerPromMetric(key, description string) {
	err := registry.Register(promauto.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "vikunja_" + key,
		Help: description,
	}, func() float64 {
		count, err := GetCount(key)
		if err != nil {
			log.Errorf("Could not get count for metric %s: %s", key, err)
		}
		return float64(count)
	}))
	if err != nil {
		log.Criticalf("Could not register metrics for %s: %s", key, err)
	}
}

// InitMetrics Initializes the metrics
func InitMetrics() {
	GetRegistry()

	registerPromMetric(ProjectCountKey, "The total number of projects on this instance")
	registerPromMetric(UserCountKey, "The total number of users on this instance")
	registerPromMetric(TaskCountKey, "The total number of tasks on this instance")
	registerPromMetric(TeamCountKey, "The total number of teams on this instance")
	registerPromMetric(FilesCountKey, "The total number of files on this instance")
	registerPromMetric(AttachmentsCountKey, "The total number of attachments on this instance")

	setupActiveUsersMetric()
	setupActiveLinkSharesMetric()
}

// GetCount returns the current count for the given metric key. The value is counted
// directly from the database and cached for countCacheTTL, so repeated scrapes don't
// hit the database on every request.
func GetCount(key string) (int64, error) {
	return keyvalue.RememberFor(key, countCacheTTL, func() (int64, error) {
		return countFromDatabase(key)
	})
}

// countFromDatabase runs a COUNT(*) for the table backing the given metric key.
func countFromDatabase(key string) (int64, error) {
	table, has := countTables[key]
	if !has {
		return 0, nil
	}

	s := db.NewSession()
	defer s.Close()

	query := s.Table(table)
	if key == TaskCountKey {
		// Exclude soft-deleted tasks; no bean here, so the xorm deleted tag doesn't apply
		query = query.Where("deleted_at IS NULL")
	}

	return query.Count()
}

// InvalidateCount drops the cached count for a key so the next read recomputes it from
// the database. Use it where instant freshness is worth the extra COUNT(*); everywhere
// else the countCacheTTL keeps the value reasonably up to date on its own.
func InvalidateCount(key string) error {
	return keyvalue.Del(key)
}
