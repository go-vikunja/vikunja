// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2021 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package log

import (
	"io"
	"os"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/config"
	"github.com/op/go-logging"
	"github.com/spf13/viper"
)

// ErrFmt holds the format for all the console logging
const ErrFmt = `${time_rfc3339_nano}: ${level} ` + "\t" + `▶ ${prefix} ${short_file}:${line}`

// WebFmt holds the format for all logging related to web requests
const WebFmt = `${time_rfc3339_nano}: WEB ` + "\t" + `▶ ${remote_ip} ${id} ${method} ${status} ${uri} ${latency_human} - ${user_agent}`

// Fmt is the general log format
const Fmt = `%{color}%{time:` + time.RFC3339Nano + `}: %{level}` + "\t" + `▶ %{shortpkg}/%{shortfunc} %{id:03x}%{color:reset} %{message}`

const logModule = `vikunja`

// loginstance is the instance of the logger which is used under the hood to log
var logInstance = logging.MustGetLogger(logModule)

// InitLogger initializes the global log handler
func InitLogger() {
	if !config.LogEnabled.GetBool() {
		// Disable all logging when loggin in general is disabled, overwriting everything a user might have set.
		config.LogStandard.Set("off")
		config.LogDatabase.Set("off")
		config.LogHTTP.Set("off")
		config.LogEcho.Set("off")
		return
	}

	// This show correct caller functions
	logInstance.ExtraCalldepth = 1

	if config.LogStandard.GetString() == "file" {
		err := os.Mkdir(config.LogPath.GetString(), 0744)
		if err != nil && !os.IsExist(err) {
			Fatalf("Could not create log folder: %s", err.Error())
		}
	}

	// We define our two backends
	if config.LogStandard.GetString() != "off" {
		stdWriter := GetLogWriter("standard")

		level, err := logging.LogLevel(strings.ToUpper(config.LogLevel.GetString()))
		if err != nil {
			Fatalf("Error setting database log level: %s", err.Error())
		}

		logBackend := logging.NewLogBackend(stdWriter, "", 0)
		backend := logging.NewBackendFormatter(logBackend, logging.MustStringFormatter(Fmt+"\n"))

		backendLeveled := logging.AddModuleLevel(backend)
		backendLeveled.SetLevel(level, logModule)

		logInstance.SetBackend(backendLeveled)
	}
}

// GetLogWriter returns the writer to where the normal log goes, depending on the config
func GetLogWriter(logfile string) (writer io.Writer) {
	writer = os.Stdout // Set the default case to prevent nil pointer panics
	switch viper.GetString("log." + logfile) {
	case "file":
		fullLogFilePath := config.LogPath.GetString() + "/" + logfile + ".log"
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
