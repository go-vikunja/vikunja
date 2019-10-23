//   Copyright (c) 2019 Vikunja and contributors.
//
//   This program is free software: you can redistribute it and/or modify
//   it under the terms of the GNU Lesser General Public License as published by
//   the Free Software Foundation, either version 3 of the License, or
//   (at your option) any later version.
//
//   This program is distributed in the hope that it will be useful,
//   but WITHOUT ANY WARRANTY; without even the implied warranty of
//   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//   GNU Lesser General Public License for more details.
//
//   You should have received a copy of the GNU Lesser General Public License
//   along with this program.  If not, see <http://www.gnu.org/licenses/>.
//

package handler

import (
	"code.vikunja.io/web"
	"github.com/op/go-logging"
)

// Config contains the config for the web handler
type Config struct {
	AuthProvider    *web.Auths
	LoggingProvider *logging.Logger
	MaxItemsPerPage int
}

var config *Config

func init() {
	config = &Config{}
}

// SetAuthProvider sets the auth provider in config
func SetAuthProvider(provider *web.Auths) {
	config.AuthProvider = provider
}

// SetLoggingProvider sets the logging provider in the config
func SetLoggingProvider(logger *logging.Logger) {
	config.LoggingProvider = logger
}

// SetMaxItemsPerPage sets the max number of items per page in the config
func SetMaxItemsPerPage(maxItemsPerPage int) {
	config.MaxItemsPerPage = maxItemsPerPage
}
