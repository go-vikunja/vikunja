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
	"testing"

	"code.vikunja.io/api/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestGetMessageSetsMessageID(t *testing.T) {
	config.ServicePublicURL.Set("https://tasks.example.com/")
	config.MailerFromEmail.Set("test@example.com")

	opts := &Opts{
		To:          "recipient@example.com",
		Subject:     "Test",
		Message:     "Hello",
		ContentType: ContentTypePlain,
	}

	m := getMessage(opts)

	msgID := m.GetMessageID()
	assert.NotEmpty(t, msgID)
	assert.Contains(t, msgID, "@tasks.example.com>")
}
