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
	"strings"
	"time"

	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/modules/keyvalue"
	"code.vikunja.io/api/pkg/web"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const secondsUntilInactive = 30
const activeUsersKeyPrefix = `active_users:`
const activeLinkSharesKeyPrefix = `active_link_shares:`

func registerActiveMetric(name, help, prefix string) {
	// pre-prefix key from older versions, never read again
	legacyKey := strings.TrimSuffix(prefix, ":")
	if err := keyvalue.Del(legacyKey); err != nil {
		log.Errorf("Could not remove legacy key %s: %s", legacyKey, err)
	}

	err := registry.Register(promauto.NewGaugeFunc(prometheus.GaugeOpts{
		Name: name,
		Help: help,
	}, func() float64 {
		return float64(countActive(prefix))
	}))
	if err != nil {
		log.Criticalf("Could not register metrics for %s: %s", name, err)
	}
}

func countActive(prefix string) int {
	keys, err := keyvalue.ListKeys(prefix)
	if err != nil {
		log.Errorf("Could not list keys with prefix %s: %s", prefix, err)
		return 0
	}

	return len(keys)
}

func setActive(prefix string, a web.Auth) error {
	return keyvalue.PutWithTTL(prefix+strconv.FormatInt(a.GetID(), 10), true, secondsUntilInactive*time.Second)
}

func SetUserActive(a web.Auth) error {
	return setActive(activeUsersKeyPrefix, a)
}

func SetLinkShareActive(a web.Auth) error {
	return setActive(activeLinkSharesKeyPrefix, a)
}
