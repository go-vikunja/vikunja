//  Vikunja is a todo-list application to facilitate your life.
//  Copyright 2018 Vikunja and contributors. All rights reserved.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU General Public License as published by
//  the Free Software Foundation, either version 3 of the License, or
//  (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU General Public License for more details.
//
//  You should have received a copy of the GNU General Public License
//  along with this program.  If not, see <https://www.gnu.org/licenses/>.

package metrics

import (
	"bytes"
	"code.vikunja.io/api/pkg/log"
	"encoding/gob"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"time"
)

// SecondsUntilInactive defines the seconds until a user is considered inactive
const SecondsUntilInactive = 60

// ActiveUsersKey is the key used to store active users in redis
const ActiveUsersKey = `activeusers`

// ActiveUser defines an active user
type ActiveUser struct {
	UserID   int64
	LastSeen time.Time
}

func init() {
	promauto.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "vikunja_active_users",
		Help: "The currently active users on this node",
	}, func() float64 {

		allActiveUsers, err := GetActiveUsers()
		if err != nil {
			log.Error(err.Error())
		}
		activeUsersCount := 0
		for _, u := range allActiveUsers {
			if time.Since(u.LastSeen) < SecondsUntilInactive*time.Second {
				activeUsersCount++
			}
		}
		return float64(activeUsersCount)
	})
}

// GetActiveUsers returns the active users from redis
func GetActiveUsers() (users []*ActiveUser, err error) {

	activeUsersR, err := r.Get(ActiveUsersKey).Bytes()
	if err != nil {
		if err.Error() == "redis: nil" {
			return users, nil
		}
		return
	}

	var b bytes.Buffer
	_, err = b.Write(activeUsersR)
	if err != nil {
		return nil, err
	}
	d := gob.NewDecoder(&b)
	if err := d.Decode(&users); err != nil {
		return nil, err
	}
	return
}

// SetActiveUsers sets the active users from redis
func SetActiveUsers(users []*ActiveUser) (err error) {
	var b bytes.Buffer
	e := gob.NewEncoder(&b)
	if err := e.Encode(users); err != nil {
		return err
	}

	return r.Set(ActiveUsersKey, b.Bytes(), 0).Err()
}
