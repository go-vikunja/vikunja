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
	"io"
	"log/slog"
	"strings"

	"xorm.io/xorm/log"
)

// XormFmt defines the prefix for xorm logging strings
const XormFmt = "[DATABASE]"

// XormLogger holds an implementation of the xorm logger interface.
type XormLogger struct {
	logger  *slog.Logger
	level   log.LogLevel
	showSQL bool
}

// NewXormLogger creates and initializes a new xorm logger
func NewXormLogger(configLogEnabled bool, configLogDatabase string, configLogDatabaseLevel string, configLogFormat string) *XormLogger {
	lvl := strings.ToUpper(configLogDatabaseLevel)
	var level slog.Level
	switch lvl {
	case "CRITICAL", "ERROR":
		level = slog.LevelError
	case "WARNING":
		level = slog.LevelWarn
	case "NOTICE", "INFO":
		level = slog.LevelInfo
	case "DEBUG":
		level = slog.LevelDebug
	default:
		level = slog.LevelInfo
	}

	var writer = io.Discard
	if configLogEnabled && configLogDatabase != "off" {
		writer = GetLogWriter(configLogDatabase, "database")
	}

	handler := CreateHandler(writer, level, configLogFormat)

	xormLogger := &XormLogger{
		logger: slog.New(handler).With("component", "database"),
	}

	switch {
	case level <= slog.LevelDebug:
		xormLogger.level = log.LOG_DEBUG
	case level <= slog.LevelInfo:
		xormLogger.level = log.LOG_INFO
	case level <= slog.LevelWarn:
		xormLogger.level = log.LOG_WARNING
	default:
		xormLogger.level = log.LOG_ERR
	}

	xormLogger.showSQL = true

	return xormLogger
}

// Debug logs a debug string
func (x *XormLogger) Debug(v ...interface{}) {
	if x.level <= log.LOG_DEBUG {
		x.logger.Debug(fmt.Sprint(v...))
	}
}

// Debugf logs a debug string
func (x *XormLogger) Debugf(format string, v ...interface{}) {
	if x.level <= log.LOG_DEBUG {
		x.logger.Debug(fmt.Sprintf(format, v...))
	}
}

// Error logs a debug string
func (x *XormLogger) Error(v ...interface{}) {
	if x.level <= log.LOG_ERR {
		x.logger.Error(fmt.Sprint(v...))
	}
}

// Errorf logs a debug string
func (x *XormLogger) Errorf(format string, v ...interface{}) {
	if x.level <= log.LOG_ERR {
		x.logger.Error(fmt.Sprintf(format, v...))
	}
}

// Info logs an info string
func (x *XormLogger) Info(v ...interface{}) {
	if x.level <= log.LOG_INFO {
		x.logger.Info(fmt.Sprint(v...))
	}
}

// Infof logs an info string
func (x *XormLogger) Infof(format string, v ...interface{}) {
	if x.level <= log.LOG_INFO {
		x.logger.Info(fmt.Sprintf(format, v...))
	}
}

// Warn logs a warning string
func (x *XormLogger) Warn(v ...interface{}) {
	if x.level <= log.LOG_WARNING {
		x.logger.Warn(fmt.Sprint(v...))
	}
}

// Warnf logs a warning string
func (x *XormLogger) Warnf(format string, v ...interface{}) {
	if x.level <= log.LOG_WARNING {
		x.logger.Warn(fmt.Sprintf(format, v...))
	}
}

// Level returns the current set log level
func (x *XormLogger) Level() log.LogLevel {
	return x.level
}

// SetLevel sets the log level
func (x *XormLogger) SetLevel(l log.LogLevel) {
	x.level = l
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
