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

package v2

import "time"

// TaskAttachment represents a task attachment.
type TaskAttachment struct {
	ID      int64     `json:"id"`
	TaskID  int64     `json:"task_id"`
	FileID  int64     `json:"file_id"`
	Created time.Time `json:"created"`
	Links   *TaskAttachmentLinks `json:"_links"`
}

// TaskAttachmentLinks represents the links for a task attachment.
type TaskAttachmentLinks struct {
	Self *Link `json:"self"`
	Task *Link `json:"task"`
	File *Link `json:"file"`
}
