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

package mail

import (
	"crypto/tls"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"
	"gopkg.in/gomail.v2"
)

// Queue is the mail queue
var Queue chan *gomail.Message

func getDialer() *gomail.Dialer {
	d := gomail.NewDialer(config.MailerHost.GetString(), config.MailerPort.GetInt(), config.MailerUsername.GetString(), config.MailerPassword.GetString())
	// #nosec
	d.TLSConfig = &tls.Config{
		InsecureSkipVerify: config.MailerSkipTLSVerify.GetBool(),
		ServerName:         config.MailerHost.GetString(),
	}
	d.SSL = config.MailerForceSSL.GetBool()
	return d
}

// StartMailDaemon starts the mail daemon
func StartMailDaemon() {
	Queue = make(chan *gomail.Message, config.MailerQueuelength.GetInt())

	if !config.MailerEnabled.GetBool() {
		return
	}

	if config.MailerHost.GetString() == "" {
		log.Warning("Mailer seems to be not configured! Please see the config docs for more details.")
		return
	}

	go func() {
		d := getDialer()

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
						log.Error("Error during connect to smtp server: %s", err)
						break
					}
					open = true
				}
				if err := gomail.Send(s, m); err != nil {
					log.Error("Error when sending mail: %s", err)
					break
				}
				// Close the connection to the SMTP server if no email was sent in
				// the last 30 seconds.
			case <-time.After(config.MailerQueueTimeout.GetDuration() * time.Second):
				if open {
					open = false
					if err := s.Close(); err != nil {
						log.Error("Error closing the mail server connection: %s\n", err)
						break
					}
					log.Infof("Closed connection to mailserver")
				}
			}
		}
	}()
}

// StopMailDaemon closes the mail queue channel, aka stops the daemon
func StopMailDaemon() {
	close(Queue)
}
