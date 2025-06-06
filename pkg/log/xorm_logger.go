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
	"strings"
	"time"

	"github.com/op/go-logging"
	"xorm.io/xorm/log"
)

// XormFmt defines the format for xorm logging strings
const XormFmt = `%{color}%{time:` + time.RFC3339Nano + `}: %{level}` + "\t" + `▶ [DATABASE] %{id:03x}%{color:reset} %{message}`

const xormLogModule = `vikunja_database`

// XormLogger holds an implementation of the xorm logger interface.
type XormLogger struct {
	logger  *logging.Logger
	level   log.LogLevel
	showSQL bool
}

// NewXormLogger creates and initializes a new xorm logger
func NewXormLogger(configLogEnabled bool, configLogDatabase string, configLogDatabaseLevel string) *XormLogger {
	lvl := strings.ToUpper(configLogDatabaseLevel)
	level, err := logging.LogLevel(lvl)
	if err != nil {
		Criticalf("Error setting database log level %s: %s", lvl, err.Error())
	}

	xormLogger := &XormLogger{
		logger: logging.MustGetLogger(xormLogModule),
	}

	var backend logging.Backend
	backend = &NoopBackend{}
	if configLogEnabled && configLogDatabase != "off" {
		logBackend := logging.NewLogBackend(GetLogWriter(configLogDatabase, "database"), "", 0)
		backend = logging.NewBackendFormatter(logBackend, logging.MustStringFormatter(XormFmt+"\n"))
	}

	backendLeveled := logging.AddModuleLevel(backend)
	backendLeveled.SetLevel(level, xormLogModule)

	xormLogger.logger.SetBackend(backendLeveled)

	switch level {
	case logging.CRITICAL:
	case logging.ERROR:
		xormLogger.level = log.LOG_ERR
	case logging.WARNING:
		xormLogger.level = log.LOG_WARNING
	case logging.NOTICE:
	case logging.INFO:
		xormLogger.level = log.LOG_INFO
	case logging.DEBUG:
		xormLogger.level = log.LOG_DEBUG
	default:
		xormLogger.level = log.LOG_OFF
	}

	xormLogger.showSQL = true

	return xormLogger
}

// Debug logs a debug string
func (x *XormLogger) Debug(v ...interface{}) {
	x.logger.Debug(v...)
}

// Debugf logs a debug string
func (x *XormLogger) Debugf(format string, v ...interface{}) {
	x.logger.Debugf(format, v...)
}

// Error logs a debug string
func (x *XormLogger) Error(v ...interface{}) {
	x.logger.Error(v...)
}

// Errorf logs a debug string
func (x *XormLogger) Errorf(format string, v ...interface{}) {
	x.logger.Errorf(format, v...)
}

// Info logs an info string
func (x *XormLogger) Info(v ...interface{}) {
	x.logger.Info(v...)
}

// Infof logs an info string
func (x *XormLogger) Infof(format string, v ...interface{}) {
	x.logger.Infof(format, v...)
}

// Warn logs a warning string
func (x *XormLogger) Warn(v ...interface{}) {
	x.logger.Warning(v...)
}

// Warnf logs a warning string
func (x *XormLogger) Warnf(format string, v ...interface{}) {
	x.logger.Warningf(format, v...)
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
