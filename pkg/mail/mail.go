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

package mail

import (
	"code.vikunja.io/api/pkg/log"
	"crypto/tls"
	"github.com/spf13/viper"
	"gopkg.in/gomail.v2"
	"time"
)

// Queue is the mail queue
var Queue chan *gomail.Message

// StartMailDaemon starts the mail daemon
func StartMailDaemon() {
	Queue = make(chan *gomail.Message, viper.GetInt("mailer.queuelength"))

	if !viper.GetBool("mailer.enabled") {
		return
	}

	if viper.GetString("mailer.host") == "" {
		log.Log.Warning("Mailer seems to be not configured! Please see the config docs for more details.")
		return
	}

	go func() {
		d := gomail.NewDialer(viper.GetString("mailer.host"), viper.GetInt("mailer.port"), viper.GetString("mailer.username"), viper.GetString("mailer.password"))
		d.TLSConfig = &tls.Config{InsecureSkipVerify: viper.GetBool("mailer.skiptlsverify")}

		var s gomail.SendCloser
		var err error
		open := false
		for {
			select {
			case m, ok := <-Queue:
				if !ok {
					return
				}
				if !open {
					if s, err = d.Dial(); err != nil {
						log.Log.Error("Error during connect to smtp server: %s", err)
					}
					open = true
				}
				if err := gomail.Send(s, m); err != nil {
					log.Log.Error("Error when sending mail: %s", err)
				}
				// Close the connection to the SMTP server if no email was sent in
				// the last 30 seconds.
			case <-time.After(viper.GetDuration("mailer.queuetimeout") * time.Second):
				if open {
					if err := s.Close(); err != nil {
						log.Log.Error("Error closing the mail server connection: %s\n", err)
					}
					log.Log.Infof("Closed connection to mailserver")
					open = false
				}
			}
		}
	}()
}

// StopMailDaemon closes the mail queue channel, aka stops the daemon
func StopMailDaemon() {
	close(Queue)
}
