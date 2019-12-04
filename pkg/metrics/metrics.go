// Vikunja is a todo-list application to facilitate your life.
// Copyright 2018-2019 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package metrics

import (
	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/red"
	"github.com/go-redis/redis"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var r *redis.Client

const (
	// ListCountKey is the name of the key in which we save the list count
	ListCountKey = `listcount`

	// UserCountKey is the name of the key we use to store total users in redis
	UserCountKey = `usercount`

	// NamespaceCountKey is the name of the key we use to store the amount of total namespaces in redis
	NamespaceCountKey = `namespacecount`

	// TaskCountKey is the name of the key we use to store the amount of total tasks in redis
	TaskCountKey = `taskcount`

	// TeamCountKey is the name of the key we use to store the amount of total teams in redis
	TeamCountKey = `teamcount`
)

// InitMetrics Initializes the metrics
func InitMetrics() {
	r = red.GetRedis()

	// Register total list count metric
	promauto.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "vikunja_list_count",
		Help: "The number of lists on this instance",
	}, func() float64 {
		count, _ := GetCount(ListCountKey)
		return float64(count)
	})

	// Register total user count metric
	promauto.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "vikunja_user_count",
		Help: "The total number of users on this instance",
	}, func() float64 {
		count, _ := GetCount(UserCountKey)
		return float64(count)
	})

	// Register total Namespaces count metric
	promauto.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "vikunja_namespcae_count",
		Help: "The total number of namespaces on this instance",
	}, func() float64 {
		count, _ := GetCount(NamespaceCountKey)
		return float64(count)
	})

	// Register total Tasks count metric
	promauto.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "vikunja_task_count",
		Help: "The total number of tasks on this instance",
	}, func() float64 {
		count, _ := GetCount(TaskCountKey)
		return float64(count)
	})

	// Register total user count metric
	promauto.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "vikunja_team_count",
		Help: "The total number of teams on this instance",
	}, func() float64 {
		count, _ := GetCount(TeamCountKey)
		return float64(count)
	})
}

// GetCount returns the current count from redis
func GetCount(key string) (count int64, err error) {
	count, err = r.Get(key).Int64()
	if err != nil && err.Error() != "redis: nil" {
		return
	}
	err = nil

	return
}

// SetCount sets the list count to a given value
func SetCount(count int64, key string) error {
	return r.Set(key, count, 0).Err()
}

// UpdateCount updates a count with a given amount
func UpdateCount(update int64, key string) {
	if !config.ServiceEnableMetrics.GetBool() {
		return
	}
	oldtotal, err := GetCount(key)
	if err != nil {
		log.Error(err.Error())
	}

	err = SetCount(oldtotal+update, key)
	if err != nil {
		log.Error(err.Error())
	}
}
