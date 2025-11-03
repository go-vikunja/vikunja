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

	maillog "github.com/wneessen/go-mail/log"
)

type MailLogger struct {
	logger *slog.Logger
}

// NewMailLogger creates and initializes a new mail logger
func NewMailLogger(configLogEnabled bool, configLogMail string, configLogMailLevel string, configLogFormat string) maillog.Logger {
	handler, _ := makeLogHandler(configLogEnabled, configLogMail, configLogMailLevel, configLogFormat)

	mailLogger := &MailLogger{
		logger: slog.New(handler).With("component", "mail"),
	}

	return mailLogger
}

func (m *MailLogger) Debugf(l maillog.Log) {
	m.logger.Debug(fmt.Sprintf(l.Format, l.Messages...))
}

func (m *MailLogger) Infof(l maillog.Log) {
	m.logger.Info(fmt.Sprintf(l.Format, l.Messages...))
}

func (m *MailLogger) Warnf(l maillog.Log) {
	m.logger.Warn(fmt.Sprintf(l.Format, l.Messages...))
}

func (m *MailLogger) Errorf(l maillog.Log) {
	m.logger.Error(fmt.Sprintf(l.Format, l.Messages...))
}
