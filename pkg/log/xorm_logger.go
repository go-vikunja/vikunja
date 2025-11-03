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

package log

import (
	"fmt"
	"log/slog"

	"xorm.io/xorm/log"
)

// XormLogger holds an implementation of the xorm logger interface.
type XormLogger struct {
	logger  *slog.Logger
	showSQL bool
}

// NewXormLogger creates and initializes a new xorm logger
func NewXormLogger(configLogEnabled bool, configLogDatabase string, configLogDatabaseLevel string, configLogFormat string) *XormLogger {
	handler, _ := makeLogHandler(configLogEnabled, configLogDatabase, configLogDatabaseLevel, configLogFormat)

	xormLogger := &XormLogger{
		logger: slog.New(handler).With("component", "database"),
	}

	xormLogger.showSQL = true

	return xormLogger
}

// Debug logs a debug string
func (x *XormLogger) Debug(v ...interface{}) {
	x.logger.Debug(fmt.Sprint(v...))
}

// Debugf logs a debug string
func (x *XormLogger) Debugf(format string, v ...interface{}) {
	x.logger.Debug(fmt.Sprintf(format, v...))
}

// Error logs a debug string
func (x *XormLogger) Error(v ...interface{}) {
	x.logger.Error(fmt.Sprint(v...))
}

// Errorf logs a debug string
func (x *XormLogger) Errorf(format string, v ...interface{}) {
	x.logger.Error(fmt.Sprintf(format, v...))
}

// Info logs an info string
func (x *XormLogger) Info(v ...interface{}) {
	x.logger.Info(fmt.Sprint(v...))
}

// Infof logs an info string
func (x *XormLogger) Infof(format string, v ...interface{}) {
	x.logger.Info(fmt.Sprintf(format, v...))
}

// Warn logs a warning string
func (x *XormLogger) Warn(v ...interface{}) {
	x.logger.Warn(fmt.Sprint(v...))
}

// Warnf logs a warning string
func (x *XormLogger) Warnf(format string, v ...interface{}) {
	x.logger.Warn(fmt.Sprintf(format, v...))
}

// Level returns the current set log level
func (x *XormLogger) Level() log.LogLevel {
	return log.LOG_DEBUG
}

// SetLevel sets the log level
func (x *XormLogger) SetLevel(_ log.LogLevel) {
}

// ShowSQL sets whether to show the log level or not
func (x *XormLogger) ShowSQL(show ...bool) {
	if len(show) > 0 {
		x.showSQL = show[0]
	}
}

// IsShowSQL returns if sql queries should be shown
func (x *XormLogger) IsShowSQL() bool {
	return x.showSQL
}
