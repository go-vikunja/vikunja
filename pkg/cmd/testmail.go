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

package cmd

import (
	"strings"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/initialize"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/mail"
	"code.vikunja.io/api/pkg/notifications"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(testmailCmd)
}

var testmailCmd = &cobra.Command{
	Use:   "testmail [email]",
	Short: "Send a test mail using the configured smtp connection",
	Args:  cobra.ExactArgs(1),
	PreRun: func(_ *cobra.Command, _ []string) {
		initialize.LightInit()

		// Start the mail daemon
		mail.StartMailDaemon()
	},
	Run: func(_ *cobra.Command, args []string) {
		log.Info("Sending testmail...")
		message := notifications.NewMail().
			From("Vikunja <"+config.MailerFromEmail.GetString()+">").
			To(args[0]).
			Subject("Test from Vikunja").
			Line("This is a test mail!").
			Line("If you received this, Vikunja is correctly set up to send emails.").
			Action("Go to your instance", config.ServicePublicURL.GetString())

		opts, err := notifications.RenderMail(message, "en")
		if err != nil {
			log.Errorf("Error rendering test mail: %s", err.Error())
			return
		}
		if err := mail.SendTestMail(opts); err != nil &&
			!strings.HasPrefix(err.Error(), "failed to close connction: not connected to SMTP server") {
			log.Errorf("Error sending test mail: %s", err.Error())
			return
		}
		log.Info("Testmail successfully sent.")
	},
}
