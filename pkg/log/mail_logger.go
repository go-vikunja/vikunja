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
	"github.com/wneessen/go-mail/log"
)

type MailLogger struct {
	logger *logging.Logger
	level  log.Level
}

const mailFormat = `%{color}%{time:` + time.RFC3339Nano + `}: %{level}` + "\t" + `▶ [MAIL] %{id:03x}%{color:reset} %{message}`
const mailLogModule = `vikunja_mail`

// NewMailLogger creates and initializes a new mail logger
func NewMailLogger(configLogEnabled bool, configLogMail string, configLogMailLevel string) log.Logger {
	lvl := strings.ToUpper(configLogMailLevel)
	level, err := logging.LogLevel(lvl)
	if err != nil {
		Criticalf("Error setting mail log level %s: %s", lvl, err.Error())
	}

	mailLogger := &MailLogger{
		logger: logging.MustGetLogger(mailLogModule),
	}

	var backend logging.Backend
	backend = &NoopBackend{}
	if configLogEnabled && configLogMail != "off" {
		logBackend := logging.NewLogBackend(GetLogWriter(configLogMail, "mail"), "", 0)
		backend = logging.NewBackendFormatter(logBackend, logging.MustStringFormatter(mailFormat+"\n"))
	}

	backendLeveled := logging.AddModuleLevel(backend)
	backendLeveled.SetLevel(level, mailLogModule)

	mailLogger.logger.SetBackend(backendLeveled)

	switch level {
	case logging.CRITICAL:
	case logging.ERROR:
		mailLogger.level = log.LevelError
	case logging.WARNING:
		mailLogger.level = log.LevelWarn
	case logging.NOTICE:
	case logging.INFO:
		mailLogger.level = log.LevelInfo
	case logging.DEBUG:
		mailLogger.level = log.LevelDebug
	default:
		mailLogger.level = 0
	}

	return mailLogger
}

func (m *MailLogger) Debugf(l log.Log) {
	m.logger.Debugf(l.Format, l.Messages...)
}

func (m *MailLogger) Infof(l log.Log) {
	m.logger.Infof(l.Format, l.Messages...)
}

func (m *MailLogger) Warnf(l log.Log) {
	m.logger.Warningf(l.Format, l.Messages...)
}

func (m *MailLogger) Errorf(l log.Log) {
	m.logger.Errorf(l.Format, l.Messages...)
}
