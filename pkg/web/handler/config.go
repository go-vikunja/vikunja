// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package handler

import (
	"github.com/op/go-logging"
)

// Config contains the config for the web handler
type Config struct {
	LoggingProvider *logging.Logger
}

var config *Config

func init() {
	config = &Config{}
}

// SetLoggingProvider sets the logging provider in the config
func SetLoggingProvider(logger *logging.Logger) {
	config.LoggingProvider = logger
}
