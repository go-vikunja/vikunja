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
	"strconv"

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
		count, _ := GetCount(key)
		return float64(count)
	}))
	if err != nil {
		log.Criticalf("Could not register metrics for %s: %s", key, err)
	}
}

// InitMetrics Initializes the metrics
func InitMetrics() {
	GetRegistry()

	registerPromMetric(ProjectCountKey, "The number of projects on this instance")
	registerPromMetric(UserCountKey, "The total number of shares on this instance")
	registerPromMetric(TaskCountKey, "The total number of tasks on this instance")
	registerPromMetric(TeamCountKey, "The total number of teams on this instance")
	registerPromMetric(FilesCountKey, "The total number of files on this instance")
	registerPromMetric(AttachmentsCountKey, "The total number of attachments on this instance")

	setupActiveUsersMetric()
	setupActiveLinkSharesMetric()
}

// GetCount returns the current count from keyvalue
func GetCount(key string) (count int64, err error) {
	cnt, exists, err := keyvalue.Get(key)
	if err != nil {
		return 0, err
	}
	if !exists {
		return 0, nil
	}

	if s, is := cnt.(string); is {
		count, err = strconv.ParseInt(s, 10, 64)
	} else {
		count = cnt.(int64)
	}

	return
}

// SetCount sets the project count to a given value
func SetCount(count int64, key string) error {
	return keyvalue.Put(key, count)
}
