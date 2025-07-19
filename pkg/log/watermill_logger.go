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

	"github.com/ThreeDotsLabs/watermill"
)

type WatermillLogger struct {
	logger *slog.Logger
}

// NewWatermillLogger creates and initializes a new watermill logger
func NewWatermillLogger(configLogEnabled bool, configLogEvents string, configLogEventsLevel string, configLogFormat string) *WatermillLogger {
	handler, _ := makeLogHandler(configLogEnabled, configLogEvents, configLogEventsLevel, configLogFormat)

	watermillLogger := &WatermillLogger{
		logger: slog.New(handler).With("component", "events"),
	}

	return watermillLogger
}

func concatFields(fields watermill.LogFields) string {
	full := ""

	for key, val := range fields {
		full += fmt.Sprintf("%s=%v, ", key, val)
	}

	if full != "" {
		full = full[:len(full)-2]
	}

	return full
}

func (w *WatermillLogger) Error(msg string, err error, fields watermill.LogFields) {
	w.logger.Error(fmt.Sprintf("%s: %s, %s", msg, err, concatFields(fields)))
}

func (w *WatermillLogger) Info(msg string, fields watermill.LogFields) {
	w.logger.Info(fmt.Sprintf("%s, %s", msg, concatFields(fields)))
}

func (w *WatermillLogger) Debug(msg string, fields watermill.LogFields) {
	w.logger.Debug(fmt.Sprintf("%s, %s", msg, concatFields(fields)))
}

func (w *WatermillLogger) Trace(msg string, fields watermill.LogFields) {
	w.logger.Debug(fmt.Sprintf("%s, %s", msg, concatFields(fields)))
}

func (w *WatermillLogger) With(_ watermill.LogFields) watermill.LoggerAdapter {
	return w
}
