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

package notifications

import (
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/i18n"
)

type disabledMailNotifiable struct{}

func (d *disabledMailNotifiable) RouteForMail() (string, error) { return "test@example.com", nil }
func (d *disabledMailNotifiable) RouteForDB() int64             { return 1 }
func (d *disabledMailNotifiable) ShouldNotify() (bool, error)   { return true, nil }
func (d *disabledMailNotifiable) Lang() string                  { return "en" }

type disabledMailNotification struct{}

func (n *disabledMailNotification) ToMail(string) *Mail {
	return NewMail().Subject("Test").Line("Test")
}
func (n *disabledMailNotification) ToDB() any    { return nil }
func (n *disabledMailNotification) Name() string { return "disabled.mail.notification" }

func TestNotifyDoesNotBlockWhenMailerDisabled(t *testing.T) {
	config.InitDefaultConfig()
	config.MailerEnabled.Set(false)
	i18n.Init()

	done := make(chan struct{})
	go func() {
		_ = Notify(&disabledMailNotifiable{}, &disabledMailNotification{})
		close(done)
	}()

	select {
	case <-done:
		// Success - Notify returned without blocking
	case <-time.After(1 * time.Second):
		t.Fatal("Notify blocked when mailer was disabled")
	}
}
