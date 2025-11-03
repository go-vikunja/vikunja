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
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type EchoLogger struct {
	logger *slog.Logger
	writer io.Writer
}

// NewEchoLogger creates and initializes a new echo logger
func NewEchoLogger(configLogEnabled bool, configLogEcho string, configLogFormat string) echo.Logger {
	handler, writer := makeLogHandler(configLogEnabled, configLogEcho, "DEBUG", configLogFormat)

	echoLogger := &EchoLogger{
		logger: slog.New(handler).With("component", "http"),
		writer: writer,
	}

	return echoLogger
}

func (e *EchoLogger) Output() io.Writer {
	return e.writer
}

func (e *EchoLogger) SetOutput(_ io.Writer) {
}

func (e *EchoLogger) Prefix() string {
	return "http"
}

func (e *EchoLogger) SetPrefix(_ string) {
}

func (e *EchoLogger) Level() log.Lvl {
	return log.DEBUG
}

func (e *EchoLogger) SetLevel(_ log.Lvl) {
}

func (e *EchoLogger) SetHeader(_ string) {
}

func (e *EchoLogger) Print(i ...interface{}) {
	e.logger.Info(fmt.Sprint(i...))
}

func (e *EchoLogger) Printf(format string, args ...interface{}) {
	e.logger.Info(fmt.Sprintf(format, args...))
}

func (e *EchoLogger) Printj(j log.JSON) {
	if b, err := json.Marshal(j); err == nil {
		e.logger.Info(string(b))
	}
}

func (e *EchoLogger) Debug(i ...interface{}) {
	e.logger.Debug(fmt.Sprint(i...))
}

func (e *EchoLogger) Debugf(format string, args ...interface{}) {
	e.logger.Debug(fmt.Sprintf(format, args...))
}

func (e *EchoLogger) Debugj(j log.JSON) {
	if b, err := json.Marshal(j); err == nil {
		e.logger.Debug(string(b))
	}
}

func (e *EchoLogger) Info(i ...interface{}) {
	e.logger.Info(fmt.Sprint(i...))
}

func (e *EchoLogger) Infof(format string, args ...interface{}) {
	e.logger.Info(fmt.Sprintf(format, args...))
}

func (e *EchoLogger) Infoj(j log.JSON) {
	if b, err := json.Marshal(j); err == nil {
		e.logger.Info(string(b))
	}
}

func (e *EchoLogger) Warn(i ...interface{}) {
	e.logger.Warn(fmt.Sprint(i...))
}

func (e *EchoLogger) Warnf(format string, args ...interface{}) {
	e.logger.Warn(fmt.Sprintf(format, args...))
}

func (e *EchoLogger) Warnj(j log.JSON) {
	if b, err := json.Marshal(j); err == nil {
		e.logger.Warn(string(b))
	}
}

func (e *EchoLogger) Error(i ...interface{}) {
	e.logger.Error(fmt.Sprint(i...))
}

func (e *EchoLogger) Errorf(format string, args ...interface{}) {
	e.logger.Error(fmt.Sprintf(format, args...))
}

func (e *EchoLogger) Errorj(j log.JSON) {
	if b, err := json.Marshal(j); err == nil {
		e.logger.Error(string(b))
	}
}

func (e *EchoLogger) Fatal(i ...interface{}) {
	e.logger.Error(fmt.Sprint(i...))
	os.Exit(1)
}

func (e *EchoLogger) Fatalj(j log.JSON) {
	if b, err := json.Marshal(j); err == nil {
		e.logger.Error(string(b))
	}
	os.Exit(1)
}

func (e *EchoLogger) Fatalf(format string, args ...interface{}) {
	e.logger.Error(fmt.Sprintf(format, args...))
	os.Exit(1)
}

func (e *EchoLogger) Panic(i ...interface{}) {
	msg := fmt.Sprint(i...)
	e.logger.Error(msg)
	panic(msg)
}

func (e *EchoLogger) Panicj(j log.JSON) {
	if b, err := json.Marshal(j); err == nil {
		msg := string(b)
		e.logger.Error(msg)
		panic(msg)
	}
}

func (e *EchoLogger) Panicf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	e.logger.Error(msg)
	panic(msg)
}

// EnableColor enables color output
func (e *EchoLogger) EnableColor() {
	// This is a no-op for our slog implementation
}

// DisableColor disables color output
func (e *EchoLogger) DisableColor() {
	// This is a no-op for our slog implementation
}
