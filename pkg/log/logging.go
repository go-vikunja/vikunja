//  Vikunja is a todo-list application to facilitate your life.
//  Copyright 2018 Vikunja and contributors. All rights reserved.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU General Public License as published by
//  the Free Software Foundation, either version 3 of the License, or
//  (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU General Public License for more details.
//
//  You should have received a copy of the GNU General Public License
//  along with this program.  If not, see <https://www.gnu.org/licenses/>.

package log

import (
	"code.vikunja.io/api/pkg/config"
	"github.com/op/go-logging"
	"github.com/spf13/viper"
	"io"
	"log"
	"os"
	"time"
)

// ErrFmt holds the format for all the console logging
const ErrFmt = `${time_rfc3339_nano}: ${level} ` + "\t" + `▶ ${prefix} ${short_file}:${line}`

// WebFmt holds the format for all logging related to web requests
const WebFmt = `${time_rfc3339_nano}: WEB ` + "\t" + `▶ ${remote_ip} ${id} ${method} ${status} ${uri} ${latency_human} - ${user_agent}`

// Fmt is the general log format
const Fmt = `%{color}%{time:` + time.RFC3339Nano + `}: %{level}` + "\t" + `▶ %{shortpkg}/%{shortfunc} %{id:03x}%{color:reset} %{message}`

// Log is the handler for the logger
var Log = logging.MustGetLogger("vikunja")

// InitLogger initializes the global log handler
func InitLogger() {
	if !config.LogEnabled.GetBool() {
		// Disable all logging when loggin in general is disabled, overwriting everything a user might have set.
		config.LogErrors.Set("off")
		config.LogStandard.Set("off")
		config.LogDatabase.Set("off")
		config.LogHTTP.Set("off")
		config.LogEcho.Set("off")
		return
	}

	if config.LogErrors.GetString() == "file" || config.LogStandard.GetString() == "file" {
		err := os.Mkdir(config.LogPath.GetString(), 0744)
		if err != nil && !os.IsExist(err) {
			log.Fatal("Could not create log folder: ", err.Error())
		}
	}

	var logBackends []logging.Backend

	// We define our two backends
	if config.LogStandard.GetString() != "off" {
		stdWriter := GetLogWriter("standard")
		stdBackend := logging.NewLogBackend(stdWriter, "", 0)

		// Set the standard backend
		logBackends = append(logBackends, logging.NewBackendFormatter(stdBackend, logging.MustStringFormatter(Fmt+"\n")))
	}

	if config.LogErrors.GetString() != "off" {
		errWriter := GetLogWriter("error")
		errBackend := logging.NewLogBackend(errWriter, "", 0)

		// Only warnings and more severe messages should go to the error backend
		errBackendLeveled := logging.AddModuleLevel(errBackend)
		errBackendLeveled.SetLevel(logging.WARNING, "")
		logBackends = append(logBackends, errBackendLeveled)
	}

	// Set our backends
	logging.SetBackend(logBackends...)
}

// GetLogWriter returns the writer to where the normal log goes, depending on the config
func GetLogWriter(logfile string) (writer io.Writer) {
	writer = os.Stderr // Set the default case to prevent nil pointer panics
	switch viper.GetString("log." + logfile) {
	case "file":
		f, err := os.OpenFile(config.LogPath.GetString()+"/"+logfile+".log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
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
