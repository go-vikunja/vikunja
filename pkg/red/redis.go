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

package red

import (
	"code.vikunja.io/api/pkg/log"
	"github.com/go-redis/redis"
	"github.com/spf13/viper"
)

var r *redis.Client

// SetRedis initializes a redis connection
func init() {
	if !viper.GetBool("redis.enabled") {
		return
	}

	if viper.GetString("redis.host") == "" {
		log.Log.Fatal("No redis host provided.")
	}

	r = redis.NewClient(&redis.Options{
		Addr:     viper.GetString("redis.host"),
		Password: viper.GetString("redis.password"),
		DB:       viper.GetInt("redis.db"),
	})

	err := r.Ping().Err()
	if err != nil {
		log.Log.Fatal(err.Error())
	}
}

// GetRedis returns a pointer to a redis client
func GetRedis() *redis.Client {
	return r
}
