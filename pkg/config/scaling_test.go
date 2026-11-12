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

package config

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestHorizontalScalingDefaults(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)
	InitDefaultConfig()

	assert.True(t, Key("service.enablescheduler").GetBool())
	assert.True(t, Key("service.enablemigrations").GetBool())
	assert.False(t, Key("service.clustermode").GetBool())
	assert.False(t, Key("websocket.redis.enabled").GetBool())
	assert.Equal(t, 16, Key("events.maxconcurrenthandlers").GetInt())
	assert.True(t, Key("jobs.queues.reminder-expansion.enabled").GetBool())
	assert.Equal(t, 2, Key("jobs.queues.reminder-expansion.concurrency").GetInt())
	assert.Equal(t, 5, Key("jobs.queues.reminder-expansion.maxattempts").GetInt())
	assert.Equal(t, "vikunja", Key("redis.prefix").GetString())
}

func TestClusterConfiguration(t *testing.T) {
	cases := []struct {
		name      string
		change    func()
		wantError string
	}{
		{name: "valid"},
		{name: "scheduler on every replica", change: func() { ServiceEnableScheduler.Set(true) }},
		{name: "mysql", change: func() { DatabaseType.Set("mysql") }, wantError: "database.type"},
		{name: "missing secret", change: func() { ServiceSecret.Set("") }, wantError: "service.secret"},
		{name: "sqlite", change: func() { DatabaseType.Set("sqlite") }, wantError: "database.type"},
		{name: "local keys", change: func() { KeyvalueType.Set("memory") }, wantError: "keyvalue.type"},
		{name: "local limiter", change: func() { RateLimitStore.Set("memory") }, wantError: "ratelimit.store"},
		{name: "redis disabled", change: func() { RedisEnabled.Set(false) }, wantError: "redis.enabled"},
		{name: "local sockets", change: func() { WebsocketRedisEnabled.Set(false) }, wantError: "websocket.redis.enabled"},
		{name: "local files", change: func() { FilesType.Set("local") }, wantError: "files.shared"},
		{name: "shared files", change: func() { FilesType.Set("local"); FilesShared.Set(true) }},
		{name: "auto migration", change: func() { ServiceEnableMigrations.Set(true) }, wantError: "service.enablemigrations"},
		{name: "auto tls", change: func() { AutoTLSEnabled.Set(true) }, wantError: "autotls.enabled"},
		{name: "empty namespace", change: func() { RedisPrefix.Set("") }, wantError: "redis.prefix"},
		{name: "invalid concurrency", change: func() { EventsMaxConcurrentHandlers.Set(0) }, wantError: "events.maxconcurrenthandlers"},
		{name: "invalid reminder expansion concurrency", change: func() { Key("jobs.queues.reminder-expansion.concurrency").Set(0) }, wantError: "concurrency"},
		{name: "zero attempts", change: func() { Key("jobs.queues.mail.maxattempts").Set(0) }, wantError: "maxattempts"},
		{name: "negative attempts", change: func() { Key("jobs.queues.webhook-delivery.maxattempts").Set(-1) }, wantError: "maxattempts"},
		{name: "short retry delay", change: func() { Key("jobs.queues.bulk-mail.retryinitialdelay").Set("500ms") }, wantError: "retryinitialdelay"},
		{name: "retry cap below initial", change: func() { Key("jobs.queues.mail.retrymaxdelay").Set("1s") }, wantError: "retrymaxdelay"},
		{name: "invalid shutdown", change: func() { ServiceShutdownTimeout.Set("invalid") }, wantError: "service.shutdowntimeout"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			viper.Reset()
			t.Cleanup(viper.Reset)
			initDefaultConfig()
			ServiceClusterMode.Set(true)
			ServiceEnableScheduler.Set(false)
			ServiceSecret.Set("test-cluster-secret")
			ServiceEnableMigrations.Set(false)
			DatabaseType.Set("postgres")
			RedisEnabled.Set(true)
			KeyvalueType.Set("redis")
			RateLimitStore.Set("redis")
			WebsocketRedisEnabled.Set(true)
			FilesType.Set("s3")
			if tc.change != nil {
				tc.change()
			}
			err := validateRuntimeConfig()
			if tc.wantError == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tc.wantError)
			}
		})
	}
}
