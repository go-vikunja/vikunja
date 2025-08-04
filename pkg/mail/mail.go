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

package mail

import (
	"context"
	"crypto/tls"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"

	"github.com/wneessen/go-mail"
)

// Queue is the mail queue
var Queue chan *mail.Msg

func getClient() (*mail.Client, error) {

	var authType mail.SMTPAuthType
	switch config.MailerAuthType.GetString() {
	case "plain":
		authType = mail.SMTPAuthPlain
	case "login":
		authType = mail.SMTPAuthLogin
	case "cram-md5":
		authType = mail.SMTPAuthCramMD5
	}

	tlsPolicy := mail.TLSOpportunistic
	if config.MailerForceSSL.GetBool() {
		tlsPolicy = mail.TLSMandatory
	}

	opts := []mail.Option{
		mail.WithTLSPortPolicy(tlsPolicy),
		mail.WithTLSConfig(&tls.Config{
			//#nosec G402
			InsecureSkipVerify: config.MailerSkipTLSVerify.GetBool(),
			ServerName:         config.MailerHost.GetString(),
		}),
		mail.WithPort(config.MailerPort.GetInt()),
		mail.WithTimeout((config.MailerQueueTimeout.GetDuration() + 3) * time.Second), // 3s more for us to close before mail server timeout
		mail.WithLogger(log.NewMailLogger(config.LogEnabled.GetBool(), config.LogMail.GetString(), config.LogMailLevel.GetString(), config.LogFormat.GetString())),
		mail.WithDebugLog(),
	}

	if config.MailerForceSSL.GetBool() {
		opts = append(opts, mail.WithSSLPort(true))
	}

	if config.MailerUsername.GetString() != "" && config.MailerPassword.GetString() != "" {
		opts = append(opts, mail.WithSMTPAuth(authType))
	}

	if config.MailerUsername.GetString() != "" {
		opts = append(opts, mail.WithUsername(config.MailerUsername.GetString()))
	}

	if config.MailerPassword.GetString() != "" {
		opts = append(opts, mail.WithPassword(config.MailerPassword.GetString()))
	}

	return mail.NewClient(
		config.MailerHost.GetString(),
		opts...,
	)
}

// StartMailDaemon starts the mail daemon
func StartMailDaemon() {
	Queue = make(chan *mail.Msg, config.MailerQueuelength.GetInt())

	if !config.MailerEnabled.GetBool() {
		return
	}

	if config.MailerHost.GetString() == "" {
		log.Warning("Mailer seems to be not configured! Please see the config docs for more details.")
		return
	}

	c, err := getClient()
	if err != nil {
		log.Errorf("Could not create mail client: %v", err)
		return
	}
	go func() {
		var err error
		open := false
		for {
			select {
			case m, ok := <-Queue:
				if !ok {
					return
				}
				if !open {
					err = c.DialWithContext(context.Background())
					if err != nil {
						log.Errorf("Error during connect to smtp server: %s", err)
						break
					}
					open = true
				}
				err = c.Send(m)
				if err != nil {
					log.Errorf("Error when sending mail: %s", err)
					break
				}
				// Close the connection to the SMTP server if no email was sent in
				// the last 30 seconds.
			case <-time.After(config.MailerQueueTimeout.GetDuration() * time.Second):
				if open {
					open = false
					err = c.Close()
					if err != nil {
						log.Errorf("Error closing the mail server connection: %s\n", err)
						break
					}
					log.Info("Closed connection to mail server")
				}
			}
		}
	}()
}
