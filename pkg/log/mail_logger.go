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

	maillog "github.com/wneessen/go-mail/log"
)

type MailLogger struct {
	logger *slog.Logger
	level  maillog.Level
}

// NewMailLogger creates and initializes a new mail logger
func NewMailLogger(configLogEnabled bool, configLogMail string, configLogMailLevel string, configLogFormat string) maillog.Logger {
	lvl := strings.ToUpper(configLogMailLevel)
	var level slog.Level
	switch lvl {
	case "ERROR":
		level = slog.LevelError
	case "WARNING":
		level = slog.LevelWarn
	case "DEBUG":
		level = slog.LevelDebug
	default:
		level = slog.LevelInfo
	}

	var writer = io.Discard
	if configLogEnabled && configLogMail != "off" {
		writer = GetLogWriter(configLogMail, "mail")
	}

	handler := CreateHandler(writer, level, configLogFormat)

	mailLogger := &MailLogger{
		logger: slog.New(handler).With("component", "mail"),
	}

	switch {
	case level <= slog.LevelDebug:
		mailLogger.level = maillog.LevelDebug
	case level <= slog.LevelInfo:
		mailLogger.level = maillog.LevelInfo
	case level <= slog.LevelWarn:
		mailLogger.level = maillog.LevelWarn
	default:
		mailLogger.level = maillog.LevelError
	}

	return mailLogger
}

func (m *MailLogger) Debugf(l maillog.Log) {
	if m.level >= maillog.LevelDebug {
		m.logger.Debug(fmt.Sprintf(l.Format, l.Messages...))
	}
}

func (m *MailLogger) Infof(l maillog.Log) {
	if m.level >= maillog.LevelInfo {
		m.logger.Info(fmt.Sprintf(l.Format, l.Messages...))
	}
}

func (m *MailLogger) Warnf(l maillog.Log) {
	if m.level >= maillog.LevelWarn {
		m.logger.Warn(fmt.Sprintf(l.Format, l.Messages...))
	}
}

func (m *MailLogger) Errorf(l maillog.Log) {
	m.logger.Error(fmt.Sprintf(l.Format, l.Messages...))
}
