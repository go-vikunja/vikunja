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
	"sync"
	"testing"
	"testing/synctest"
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
	synctest.Test(t, func(t *testing.T) {
		keyvalue.InitStorage()

		require.NoError(t, SetUserActive(&testAuth{id: 1}))
		require.NoError(t, SetUserActive(&testAuth{id: 2}))
		require.NoError(t, SetLinkShareActive(&testAuth{id: 1}))

		assert.Equal(t, 2, countActive(activeUsersKeyPrefix))
		assert.Equal(t, 1, countActive(activeLinkSharesKeyPrefix))

		time.Sleep(20 * time.Second)
		require.NoError(t, SetUserActive(&testAuth{id: 2}))
		time.Sleep(10*time.Second - time.Nanosecond)
		assert.Equal(t, 2, countActive(activeUsersKeyPrefix))
		assert.Equal(t, 1, countActive(activeLinkSharesKeyPrefix))

		time.Sleep(time.Nanosecond)
		assert.Equal(t, 1, countActive(activeUsersKeyPrefix))
		assert.Zero(t, countActive(activeLinkSharesKeyPrefix))

		keys, err := keyvalue.ListKeys(activeUsersKeyPrefix)
		require.NoError(t, err)
		assert.Equal(t, []string{activeUsersKeyPrefix + "2"}, keys)

		time.Sleep(20 * time.Second)
		assert.Zero(t, countActive(activeUsersKeyPrefix))
	})
}

func TestActiveUsersConcurrent(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		keyvalue.InitStorage()

		for i := int64(100); i < 110; i++ {
			require.NoError(t, SetUserActive(&testAuth{id: i}))
		}
		time.Sleep(30 * time.Second)

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
	})
}
