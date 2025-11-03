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

package red

import (
	"context"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"
	"github.com/redis/go-redis/v9"
)

var r *redis.Client

// InitRedis initializes a redis connection
func InitRedis() {
	if r != nil {
		return
	}

	if !config.RedisEnabled.GetBool() {
		return
	}

	if config.RedisHost.GetString() == "" {
		log.Fatal("No redis host provided.")
	}

	r = redis.NewClient(&redis.Options{
		Addr:     config.RedisHost.GetString(),
		Password: config.RedisPassword.GetString(),
		DB:       config.RedisDB.GetInt(),
	})

	err := r.Ping(context.Background()).Err()
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Debug("Redis initialized")
}

// GetRedis returns a pointer to a redis client
func GetRedis() *redis.Client {
	return r
}
