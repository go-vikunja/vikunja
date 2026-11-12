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
	"errors"
	"strings"
	"time"
)

func validateRuntimeConfig() error {
	if JobsRetention.GetDuration() <= 24*time.Hour {
		return errors.New("jobs.retention must exceed 24h")
	}
	for _, queue := range []string{"mail", "bulk-mail", "notifications", "webhooks"} {
		prefix := "jobs.queues." + queue + "."
		if Key(prefix+"concurrency").GetInt() < 1 || Key(prefix+"lease").GetDuration() < 30*time.Second || Key(prefix+"timeout").GetDuration() <= 0 {
			return errors.New("job queues require positive concurrency/deadline and a lease of at least 30s")
		}
	}
	if EventsMaxConcurrentHandlers.GetInt() < 1 {
		return errors.New("events.maxconcurrenthandlers must be positive")
	}
	duration, err := time.ParseDuration(ServiceShutdownTimeout.GetString())
	if err != nil || duration <= 0 {
		return errors.New("service.shutdowntimeout must be a positive duration")
	}
	if WebsocketRedisEnabled.GetBool() {
		if !RedisEnabled.GetBool() {
			return errors.New("websocket.redis.enabled requires redis.enabled")
		}
		if strings.TrimSpace(RedisPrefix.GetString()) == "" {
			return errors.New("websocket.redis.enabled requires a nonempty redis.prefix")
		}
	}
	if !ServiceClusterMode.GetBool() {
		return nil
	}
	if ServiceSecret.GetString() == "" {
		return errors.New("service.clustermode requires an explicitly configured shared service.secret")
	}
	if DatabaseType.GetString() != "postgres" {
		return errors.New("service.clustermode requires database.type postgres")
	}
	if !RedisEnabled.GetBool() {
		return errors.New("service.clustermode requires redis.enabled")
	}
	if KeyvalueType.GetString() != "redis" {
		return errors.New("service.clustermode requires keyvalue.type redis")
	}
	if store := RateLimitStore.GetString(); store != "redis" && store != "keyvalue" {
		return errors.New("service.clustermode requires ratelimit.store redis or keyvalue")
	}
	if !WebsocketRedisEnabled.GetBool() {
		return errors.New("service.clustermode requires websocket.redis.enabled")
	}
	if FilesType.GetString() != "s3" && (FilesType.GetString() != "local" || !FilesShared.GetBool()) {
		return errors.New("service.clustermode requires files.type s3, or local with files.shared on a shared filesystem")
	}
	if ServiceEnableMigrations.GetBool() {
		return errors.New("service.clustermode requires service.enablemigrations=false; run vikunja migrate once before rollout")
	}
	if AutoTLSEnabled.GetBool() {
		return errors.New("service.clustermode requires autotls.enabled=false; terminate TLS at the load balancer")
	}
	return nil
}
