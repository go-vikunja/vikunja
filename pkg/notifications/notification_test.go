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

package notifications

import (
	"testing"

	"code.vikunja.io/api/pkg/db"
	"github.com/stretchr/testify/assert"
	"xorm.io/xorm/schemas"
)

type testNotification struct {
	Test       string
	OtherValue int64
}

// ToMail returns the mail notification for testNotification
func (n *testNotification) ToMail() *Mail {
	return NewMail().
		Subject("Test Notification").
		Line(n.Test)
}

// ToDB returns the testNotification notification in a format which can be saved in the db
func (n *testNotification) ToDB() interface{} {
	data := make(map[string]interface{}, 2)
	data["test"] = n.Test
	data["other_value"] = n.OtherValue
	return data
}

// Name returns the name of the notification
func (n *testNotification) Name() string {
	return "test.notification"
}

type testNotifiable struct {
}

// RouteForMail routes a test notification for mail
func (t *testNotifiable) RouteForMail() (string, error) {
	return "some@email.com", nil
}

// RouteForDB routes a test notification for db
func (t *testNotifiable) RouteForDB() int64 {
	return 42
}

func TestNotify(t *testing.T) {
	tn := &testNotification{
		Test:       "somethingsomething",
		OtherValue: 42,
	}
	tnf := &testNotifiable{}

	err := Notify(tnf, tn)

	assert.NoError(t, err)
	vals := map[string]interface{}{
		"notifiable_id": 42,
		"notification":  "'{\"other_value\":42,\"test\":\"somethingsomething\"}'",
	}

	if db.Type() == schemas.POSTGRES {
		vals["notification::jsonb"] = vals["notification"].(string) + "::jsonb"
		delete(vals, "notification")
	}

	if db.Type() == schemas.SQLITE {
		vals["CAST(notification AS BLOB)"] = "CAST(" + vals["notification"].(string) + " AS BLOB)"
		delete(vals, "notification")
	}

	db.AssertExists(t, "notifications", vals, true)
}
