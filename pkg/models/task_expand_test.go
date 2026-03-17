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

package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseExpandParameters(t *testing.T) {
	tests := []struct {
		name        string
		csvExpand   string
		arrayExpand []TaskCollectionExpandable
		want        []TaskCollectionExpandable
		wantErr     bool
	}{
		{
			name:        "empty parameters",
			csvExpand:   "",
			arrayExpand: []TaskCollectionExpandable{},
			want:        []TaskCollectionExpandable{},
			wantErr:     false,
		},
		{
			name:        "CSV single value",
			csvExpand:   "comments",
			arrayExpand: []TaskCollectionExpandable{},
			want:        []TaskCollectionExpandable{TaskCollectionExpandComments},
			wantErr:     false,
		},
		{
			name:        "CSV multiple values",
			csvExpand:   "comments,reactions",
			arrayExpand: []TaskCollectionExpandable{},
			want:        []TaskCollectionExpandable{TaskCollectionExpandComments, TaskCollectionExpandReactions},
			wantErr:     false,
		},
		{
			name:        "CSV with spaces",
			csvExpand:   "comments, reactions, buckets",
			arrayExpand: []TaskCollectionExpandable{},
			want:        []TaskCollectionExpandable{TaskCollectionExpandComments, TaskCollectionExpandReactions, TaskCollectionExpandBuckets},
			wantErr:     false,
		},
		{
			name:        "array single value",
			csvExpand:   "",
			arrayExpand: []TaskCollectionExpandable{TaskCollectionExpandSubtasks},
			want:        []TaskCollectionExpandable{TaskCollectionExpandSubtasks},
			wantErr:     false,
		},
		{
			name:        "array multiple values",
			csvExpand:   "",
			arrayExpand: []TaskCollectionExpandable{TaskCollectionExpandSubtasks, TaskCollectionExpandCommentCount},
			want:        []TaskCollectionExpandable{TaskCollectionExpandSubtasks, TaskCollectionExpandCommentCount},
			wantErr:     false,
		},
		{
			name:        "mixed CSV and array",
			csvExpand:   "comments,reactions",
			arrayExpand: []TaskCollectionExpandable{TaskCollectionExpandSubtasks, TaskCollectionExpandBuckets},
			want:        []TaskCollectionExpandable{TaskCollectionExpandSubtasks, TaskCollectionExpandBuckets, TaskCollectionExpandComments, TaskCollectionExpandReactions},
			wantErr:     false,
		},
		{
			name:        "deduplication",
			csvExpand:   "comments,reactions",
			arrayExpand: []TaskCollectionExpandable{TaskCollectionExpandComments, TaskCollectionExpandBuckets},
			want:        []TaskCollectionExpandable{TaskCollectionExpandComments, TaskCollectionExpandBuckets, TaskCollectionExpandReactions},
			wantErr:     false,
		},
		{
			name:        "invalid CSV value",
			csvExpand:   "comments,invalid",
			arrayExpand: []TaskCollectionExpandable{},
			want:        nil,
			wantErr:     true,
		},
		{
			name:        "invalid array value",
			csvExpand:   "",
			arrayExpand: []TaskCollectionExpandable{"invalid"},
			want:        nil,
			wantErr:     true,
		},
		{
			name:        "CSV with empty values",
			csvExpand:   "comments,,reactions",
			arrayExpand: []TaskCollectionExpandable{},
			want:        []TaskCollectionExpandable{TaskCollectionExpandComments, TaskCollectionExpandReactions},
			wantErr:     false,
		},
		{
			name:        "all valid expand values",
			csvExpand:   "subtasks,buckets,reactions,comments,comment_count,is_unread",
			arrayExpand: []TaskCollectionExpandable{},
			want: []TaskCollectionExpandable{
				TaskCollectionExpandSubtasks,
				TaskCollectionExpandBuckets,
				TaskCollectionExpandReactions,
				TaskCollectionExpandComments,
				TaskCollectionExpandCommentCount,
				TaskCollectionExpandIsUnread,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseExpandParameters(tt.csvExpand, tt.arrayExpand)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.ElementsMatch(t, tt.want, got)
			}
		})
	}
}