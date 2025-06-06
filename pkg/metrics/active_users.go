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
	"sync"
	"time"

	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/modules/keyvalue"
	"code.vikunja.io/api/pkg/web"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const secondsUntilInactive = 30
const activeUsersKey = `active_users`
const activeLinkSharesKey = `active_link_shares`

// ActiveAuthenticable defines an active user or link share
type ActiveAuthenticable struct {
	ID       int64
	LastSeen time.Time
}

type activeUsersMap map[int64]*ActiveAuthenticable

type ActiveUsers struct {
	users activeUsersMap
	mutex *sync.Mutex
}

var activeUsers *ActiveUsers

type activeLinkSharesMap map[int64]*ActiveAuthenticable

type ActiveLinkShares struct {
	shares activeLinkSharesMap
	mutex  *sync.Mutex
}

var activeLinkShares *ActiveLinkShares

func init() {
	activeUsers = &ActiveUsers{
		users: make(map[int64]*ActiveAuthenticable),
		mutex: &sync.Mutex{},
	}
	activeLinkShares = &ActiveLinkShares{
		shares: make(map[int64]*ActiveAuthenticable),
		mutex:  &sync.Mutex{},
	}
}

func setupActiveUsersMetric() {
	err := registry.Register(promauto.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "vikunja_active_users",
		Help: "The number of users active within the last 30 seconds",
	}, func() float64 {
		allActiveUsers := activeUsersMap{}
		_, err := keyvalue.GetWithValue(activeUsersKey, &allActiveUsers)
		if err != nil {
			log.Error(err.Error())
			return 0
		}
		if allActiveUsers == nil {
			return 0
		}
		count := 0
		for _, u := range allActiveUsers {
			if time.Since(u.LastSeen) < secondsUntilInactive*time.Second {
				count++
			}
		}
		return float64(count)
	}))
	if err != nil {
		log.Criticalf("Could not register metrics for currently active shares: %s", err)
	}
}

func setupActiveLinkSharesMetric() {
	err := registry.Register(promauto.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "vikunja_active_link_shares",
		Help: "The number of link shares active within the last 30 seconds. Similar to vikunja_active_users.",
	}, func() float64 {
		allActiveLinkShares := activeLinkSharesMap{}
		_, err := keyvalue.GetWithValue(activeLinkSharesKey, &allActiveLinkShares)
		if err != nil {
			log.Error(err.Error())
			return 0
		}
		if allActiveLinkShares == nil {
			return 0
		}
		count := 0
		for _, u := range allActiveLinkShares {
			if time.Since(u.LastSeen) < secondsUntilInactive*time.Second {
				count++
			}
		}
		return float64(count)
	}))
	if err != nil {
		log.Criticalf("Could not register metrics for currently active link shares: %s", err)
	}
}

// SetUserActive sets a user as active and pushes it to keyvalue
func SetUserActive(a web.Auth) (err error) {
	activeUsers.mutex.Lock()
	defer activeUsers.mutex.Unlock()
	activeUsers.users[a.GetID()] = &ActiveAuthenticable{
		ID:       a.GetID(),
		LastSeen: time.Now(),
	}

	return keyvalue.Put(activeUsersKey, activeUsers.users)
}

// SetLinkShareActive sets a user as active and pushes it to keyvalue
func SetLinkShareActive(a web.Auth) (err error) {
	activeLinkShares.mutex.Lock()
	defer activeLinkShares.mutex.Unlock()
	activeLinkShares.shares[a.GetID()] = &ActiveAuthenticable{
		ID:       a.GetID(),
		LastSeen: time.Now(),
	}

	return keyvalue.Put(activeLinkSharesKey, activeLinkShares.shares)
}
