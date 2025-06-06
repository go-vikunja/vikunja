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
	"io"
	"os"
	"strings"
	"time"

	"github.com/op/go-logging"
)

// ErrFmt holds the format for all the console logging
const ErrFmt = `${time_rfc3339}: ${level} ` + "\t" + `▶ ${prefix} ${short_file}:${line}`

// WebFmt holds the format for all logging related to web requests
const WebFmt = `${time_rfc3339}: WEB ` + "\t" + `▶ ${remote_ip} ${id} ${method} ${status} ${uri} ${latency_human} - ${user_agent}`

// Fmt is the general log format
const Fmt = `%{color}%{time:` + time.RFC3339 + `}: %{level}` + "\t" + `▶ %{id:03x}%{color:reset} %{message}`

const logModule = `vikunja`

// loginstance is the instance of the logger which is used under the hood to log
var logInstance = logging.MustGetLogger(logModule)

// logpath is the path in which log files will be written.
// This value is a mere fallback for other modules that could but shouldn't be used before calling ConfigureLogger
var logPath = "."

// InitLogger initializes the global log handler
func InitLogger() {
	// This show correct caller functions
	logInstance.ExtraCalldepth = 1

	// Init with stdout and INFO as default format and level
	logBackend := logging.NewLogBackend(os.Stdout, "", 0)
	backend := logging.NewBackendFormatter(logBackend, logging.MustStringFormatter(Fmt+"\n"))

	backendLeveled := logging.AddModuleLevel(backend)
	backendLeveled.SetLevel(logging.INFO, logModule)

	logInstance.SetBackend(backendLeveled)
}

// ConfigureLogger configures the global log handler
func ConfigureLogger(configLogEnabled bool, configLogStandard string, configLogPath string, configLogLevel string) {
	lvl := strings.ToUpper(configLogLevel)
	level, err := logging.LogLevel(lvl)
	if err != nil {
		Fatalf("Error setting standard log level %s: %s", lvl, err.Error())
	}

	logPath = configLogPath

	// The backend is the part which actually handles logging the log entries somewhere.
	var backend logging.Backend
	backend = &NoopBackend{}
	if configLogEnabled && configLogStandard != "off" {
		logBackend := logging.NewLogBackend(GetLogWriter(configLogStandard, "standard"), "", 0)
		backend = logging.NewBackendFormatter(logBackend, logging.MustStringFormatter(Fmt+"\n"))
	}

	backendLeveled := logging.AddModuleLevel(backend)
	backendLeveled.SetLevel(level, logModule)

	logInstance.SetBackend(backendLeveled)
}

// GetLogWriter returns the writer to where the normal log goes, depending on the config
func GetLogWriter(logfmt string, logfile string) (writer io.Writer) {
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
func GetLogger() *logging.Logger {
	return logInstance
}

// The following functions are to be used as an "eye-candy", so one can just write log.Error() instead of log.Log.Error()

// Debug is for debug messages
func Debug(args ...interface{}) {
	logInstance.Debug(args...)
}

// Debugf is for debug messages
func Debugf(format string, args ...interface{}) {
	logInstance.Debugf(format, args...)
}

// Info is for info messages
func Info(args ...interface{}) {
	logInstance.Info(args...)
}

// Infof is for info messages
func Infof(format string, args ...interface{}) {
	logInstance.Infof(format, args...)
}

// Error is for error messages
func Error(args ...interface{}) {
	logInstance.Error(args...)
}

// Errorf is for error messages
func Errorf(format string, args ...interface{}) {
	logInstance.Errorf(format, args...)
}

// Warning is for warning messages
func Warning(args ...interface{}) {
	logInstance.Warning(args...)
}

// Warningf is for warning messages
func Warningf(format string, args ...interface{}) {
	logInstance.Warningf(format, args...)
}

// Critical is for critical messages
func Critical(args ...interface{}) {
	logInstance.Critical(args...)
}

// Criticalf is for critical messages
func Criticalf(format string, args ...interface{}) {
	logInstance.Criticalf(format, args...)
}

// Fatal is for fatal messages
func Fatal(args ...interface{}) {
	logInstance.Fatal(args...)
}

// Fatalf is for fatal messages
func Fatalf(format string, args ...interface{}) {
	logInstance.Fatalf(format, args...)
}
