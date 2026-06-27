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
	"github.com/stretchr/testify/require"
)

func TestHTMLToMarkdown_TipTap(t *testing.T) {
	tests := []struct {
		name string
		html string
		want string
	}{
		{
			name: "mention uses data-id and drops label",
			html: `<p><mention-user data-id="actualuser" data-label="Different Label">@differentlabel</mention-user></p>`,
			want: "@actualuser",
		},
		{
			name: "empty mention keeps following space",
			html: `<p><mention-user data-id="frederick" data-label="Frederick"></mention-user> hello</p>`,
			want: "@frederick hello",
		},
		{
			name: "mention next to punctuation stays intact",
			html: `<p>cc <mention-user data-id="jane">@jane</mention-user>, please review</p>`,
			want: "cc @jane, please review",
		},
		{
			name: "multiple mentions in one block",
			html: `<p>ping <mention-user data-id="user1">@user1</mention-user> and <mention-user data-id="user2">@user2</mention-user></p>`,
			want: "ping @user1 and @user2",
		},
		{
			name: "mention without data-id keeps inner text",
			html: `<p><mention-user>@someuser</mention-user> hi</p>`,
			want: "@someuser hi",
		},
		{
			name: "tiptap task list checked and unchecked",
			html: `<ul data-type="taskList"><li data-checked="true" data-type="taskItem"><label><input type="checkbox" checked="checked"><span></span></label><div><p>done item</p></div></li><li data-checked="false" data-type="taskItem"><label><input type="checkbox"><span></span></label><div><p>todo item</p></div></li></ul>`,
			want: "- [x] done item\n- [ ] todo item",
		},
		{
			name: "task list bare data-checked form",
			html: `<ul data-type="taskList"><li data-type="taskItem" data-checked="false"><p>Item 1</p></li></ul>`,
			want: "- [ ] Item 1",
		},
		{
			name: "nested task list items",
			html: `<ul data-type="taskList"><li data-checked="false" data-type="taskItem"><div><p>parent</p></div><ul data-type="taskList"><li data-checked="true" data-type="taskItem"><div><p>child</p></div></li></ul></li></ul>`,
			want: "- [ ] parent\n  \n  - [x] child",
		},
		{
			name: "mention inside task list item",
			html: `<ul data-type="taskList"><li data-checked="false" data-type="taskItem"><div><p>ask <mention-user data-id="bob">@bob</mention-user></p></div></li></ul>`,
			want: "- [ ] ask @bob",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := HTMLToMarkdown(tt.html)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
