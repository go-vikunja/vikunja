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
	"os"
	"strings"
)

// logInstance is the instance of the logger which is used under the hood to log
var logInstance *slog.Logger

// logpath is the path in which log files will be written.
// This value is a mere fallback for other modules that could but shouldn't be used before calling ConfigureLogger
var logPath = "."

// InitLogger initializes the global log handler
func InitLogger() {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	logInstance = slog.New(handler)
}

func makeLogHandler(enabled bool, output string, level string, format string) (slog.Handler, io.Writer) {
	var slogLevel slog.Level
	switch strings.ToUpper(level) {
	case "CRITICAL", "ERROR":
		slogLevel = slog.LevelError
	case "WARNING":
		slogLevel = slog.LevelWarn
	case "NOTICE", "INFO":
		slogLevel = slog.LevelInfo
	case "DEBUG":
		slogLevel = slog.LevelDebug
	default:
		slogLevel = slog.LevelInfo
	}

	format = strings.ToLower(format)
	if format == "" {
		format = "text"
	}
	if format != "text" && format != "structured" {
		Fatalf("invalid log format %s", format)
	}

	writer := io.Discard
	if enabled && output != "off" {
		writer = getLogWriter(output, "standard")
	}

	return createHandler(writer, slogLevel, format), writer
}

// createHandler creates a consistent slog handler for all loggers
func createHandler(writer io.Writer, level slog.Level, format string) slog.Handler {
	handlerOpts := &slog.HandlerOptions{Level: level}
	if strings.ToLower(format) == "structured" {
		return slog.NewJSONHandler(writer, handlerOpts)
	}

	return slog.NewTextHandler(writer, handlerOpts)
}

// NewHTTPLogger creates and initializes a new HTTP logger
func NewHTTPLogger(enabled bool, output string, format string) *slog.Logger {
	handler, _ := makeLogHandler(enabled, output, "DEBUG", format)

	return slog.New(handler).With("component", "http")
}

// ConfigureStandardLogger configures the global log handler
func ConfigureStandardLogger(enabled bool, output string, path string, level string, format string) {
	handler, _ := makeLogHandler(enabled, output, level, format)
	logInstance = slog.New(handler)
	logPath = path
}

// wrapLogger is used for libraries requiring a Debugf method.
type wrapLogger struct{}

func (wrapLogger) Debugf(format string, args ...interface{}) {
	logInstance.Debug(fmt.Sprintf(format, args...))
}

// GetLogWriter returns the writer to where the normal log goes, depending on the config
func getLogWriter(logfmt string, logfile string) (writer io.Writer) {
	writer = os.Stdout // Set the default case to prevent nil pointer panics
	switch logfmt {
	case "file":
		if err := os.MkdirAll(logPath, 0744); err != nil {
			Fatalf("Could not create log path: %s", err.Error())
		}
		fullLogFilePath := logPath + "/" + logfile + ".log"
		f, err := os.OpenFile(fullLogFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			Fatalf("Could not create logfile %s: %s", fullLogFilePath, err.Error())
		}
		writer = f
	case "stderr":
		writer = os.Stderr
	case "stdout":
	default:
		writer = os.Stdout
	}
	return
}

// GetLogger returns the logging instance. DO NOT USE THIS TO LOG STUFF.
// GetLogger returns a logger which can be used by external libraries expecting a Debugf method.
// It only implements Debugf and forwards to the global logger.
func GetLogger() interface{ Debugf(string, ...interface{}) } {
	return wrapLogger{}
}

// The following functions are to be used as an "eye-candy", so one can just write log.Error() instead of log.Log.Error()

// Debug is for debug messages
func Debug(args ...interface{}) {
	logInstance.Debug(fmt.Sprint(args...))
}

// Debugf is for debug messages
func Debugf(format string, args ...interface{}) {
	logInstance.Debug(fmt.Sprintf(format, args...))
}

// Info is for info messages
func Info(args ...interface{}) {
	logInstance.Info(fmt.Sprint(args...))
}

// Infof is for info messages
func Infof(format string, args ...interface{}) {
	logInstance.Info(fmt.Sprintf(format, args...))
}

// Error is for error messages
func Error(args ...interface{}) {
	logInstance.Error(fmt.Sprint(args...))
}

// Errorf is for error messages
func Errorf(format string, args ...interface{}) {
	logInstance.Error(fmt.Sprintf(format, args...))
}

// Warning is for warning messages
func Warning(args ...interface{}) {
	logInstance.Warn(fmt.Sprint(args...))
}

// Warningf is for warning messages
func Warningf(format string, args ...interface{}) {
	logInstance.Warn(fmt.Sprintf(format, args...))
}

// Critical is for critical messages
func Critical(args ...interface{}) {
	logInstance.Error(fmt.Sprint(args...))
}

// Criticalf is for critical messages
func Criticalf(format string, args ...interface{}) {
	logInstance.Error(fmt.Sprintf(format, args...))
}

// Fatal is for fatal messages
func Fatal(args ...interface{}) {
	logInstance.Error(fmt.Sprint(args...))
	os.Exit(1)
}

// Fatalf is for fatal messages
func Fatalf(format string, args ...interface{}) {
	logInstance.Error(fmt.Sprintf(format, args...))
	os.Exit(1)
}
