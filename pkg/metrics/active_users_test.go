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
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/modules/keyvalue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	log.InitLogger()
	config.InitDefaultConfig()
	os.Exit(m.Run())
}

type testAuth struct{ id int64 }

func (a *testAuth) GetID() int64 { return a.id }

func TestActiveUsers(t *testing.T) {
	keyvalue.InitStorage()

	require.NoError(t, SetUserActive(&testAuth{id: 1}))
	require.NoError(t, SetUserActive(&testAuth{id: 2}))
	require.NoError(t, keyvalue.PutWithTTL(activeUsersKeyPrefix+"3", time.Now(), time.Nanosecond))
	require.NoError(t, SetLinkShareActive(&testAuth{id: 1}))

	assert.Equal(t, 2, countActive(activeUsersKeyPrefix))
	assert.Equal(t, 1, countActive(activeLinkSharesKeyPrefix))

	keys, err := keyvalue.ListKeys(activeUsersKeyPrefix)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{
		activeUsersKeyPrefix + "1",
		activeUsersKeyPrefix + "2",
	}, keys)
}

func TestActiveUsersConcurrent(t *testing.T) {
	keyvalue.InitStorage()

	for i := 100; i < 110; i++ {
		require.NoError(t, keyvalue.PutWithTTL(activeUsersKeyPrefix+strconv.Itoa(i), time.Now(), time.Nanosecond))
	}

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func(id int64) {
			defer wg.Done()
			assert.NoError(t, SetUserActive(&testAuth{id: id}))
		}(int64(i))
		go func() {
			defer wg.Done()
			countActive(activeUsersKeyPrefix)
		}()
	}
	wg.Wait()
	assert.Equal(t, 50, countActive(activeUsersKeyPrefix))
}
