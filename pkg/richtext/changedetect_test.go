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

package richtext

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChanged(t *testing.T) {
	tests := []struct {
		name     string
		stored   string
		incoming string
		want     bool
	}{
		{
			name:     "markdown projection equals incoming",
			stored:   "<p>Hello <strong>world</strong></p>",
			incoming: "Hello **world**",
			want:     false,
		},
		{
			name:     "genuinely edited",
			stored:   "<p>Hello <strong>world</strong></p>",
			incoming: "Hello **mars**",
			want:     true,
		},
		{
			name:     "line ending only difference",
			stored:   "<p>line one</p><p>line two</p>",
			incoming: "line one\r\n\r\nline two",
			want:     false,
		},
		{
			name:     "trailing whitespace only difference",
			stored:   "<p>same</p>",
			incoming: "same\n\n   ",
			want:     false,
		},
		{
			name:     "equivalent markdown flavors compare equal",
			stored:   "<p><em>x</em></p>",
			incoming: "_x_",
			want:     false,
		},
		{
			name:     "empty stored vs empty incoming",
			stored:   "<p></p>",
			incoming: "",
			want:     false,
		},
		{
			name:     "empty stored vs new content",
			stored:   "",
			incoming: "now has text",
			want:     true,
		},
		{
			name:     "task list round trip unchanged",
			stored:   `<ul data-type="taskList"><li data-checked="true" data-type="taskItem"><label><input type="checkbox" checked="checked"><span></span></label><div><p>done</p></div></li></ul>`,
			incoming: "- [x] done",
			want:     false,
		},
		{
			name:     "mention round trip unchanged",
			stored:   `<p>cc <mention-user data-id="user1" data-label="User One">@User One</mention-user></p>`,
			incoming: "cc @user1",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, Changed(tt.stored, tt.incoming))
		})
	}
}

func TestHTMLIsEmpty(t *testing.T) {
	assert.True(t, HTMLIsEmpty(""))
	assert.True(t, HTMLIsEmpty("<p></p>"))
	assert.True(t, HTMLIsEmpty("   "))
	assert.True(t, HTMLIsEmpty("<p>   </p>"))
	assert.False(t, HTMLIsEmpty("<p>content</p>"))
}
