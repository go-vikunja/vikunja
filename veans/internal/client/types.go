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

// Package client is a hand-rolled JSON client for the Vikunja REST API. It
// mirrors the wire types as plain Go structs so we don't pull XORM into the
// CLI binary.
package client

import "time"

// User mirrors the public fields of pkg/user.User on the wire.
type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name,omitempty"`
	Email    string `json:"email,omitempty"`
}

// BotUser is what `PUT /bots` returns.
type BotUser struct {
	ID       int64     `json:"id"`
	Username string    `json:"username"`
	Name     string    `json:"name,omitempty"`
	Status   int       `json:"status,omitempty"`
	Created  time.Time `json:"created,omitempty"`
}

// BotUserCreate is the request body for PUT /bots.
type BotUserCreate struct {
	Username string `json:"username"`
	Name     string `json:"name,omitempty"`
}

// Project mirrors pkg/models/project.Project.
type Project struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Identifier  string `json:"identifier,omitempty"`
	IsArchived  bool   `json:"is_archived,omitempty"`
}

// ProjectView is a saved view (Kanban/List/Gantt/Table) on a project.
// view_kind is serialized as a string on the wire ("list" / "gantt" /
// "table" / "kanban"), not an int — Vikunja's ProjectViewKind has a
// custom MarshalJSON.
type ProjectView struct {
	ID             int64  `json:"id"`
	Title          string `json:"title"`
	ProjectID      int64  `json:"project_id"`
	// view_kind / bucket_configuration_mode are serialized as strings on
	// the wire (custom MarshalJSON on the parent enums), not ints.
	ViewKind       string `json:"view_kind"`
	BucketConfMode string `json:"bucket_configuration_mode,omitempty"`
}

const (
	ViewKindList   = "list"
	ViewKindGantt  = "gantt"
	ViewKindTable  = "table"
	ViewKindKanban = "kanban"
)

// Bucket is a kanban bucket bound to a single project view.
type Bucket struct {
	ID            int64   `json:"id"`
	Title         string  `json:"title"`
	ProjectViewID int64   `json:"project_view_id"`
	Limit         int64   `json:"limit,omitempty"`
	Position      float64 `json:"position,omitempty"`
}

// Task mirrors the on-the-wire task representation. Many fields are omitted —
// veans only consumes what its commands print or filter on.
type Task struct {
	ID          int64       `json:"id"`
	Title       string      `json:"title"`
	Description string      `json:"description,omitempty"`
	Done        bool        `json:"done"`
	DoneAt      *time.Time  `json:"done_at,omitempty"`
	Priority    int64       `json:"priority,omitempty"`
	ProjectID   int64       `json:"project_id"`
	Index       int64       `json:"index,omitempty"`
	Identifier  string      `json:"identifier,omitempty"`
	Position    float64     `json:"position,omitempty"`
	Created     time.Time   `json:"created,omitempty"`
	Updated     time.Time   `json:"updated,omitempty"`
	// BucketID is only set by Vikunja when sending a task to a server-
	// side endpoint (e.g. the bucket-move POST); reads return it as 0.
	// The current bucket(s) — one per Kanban view — are exposed via
	// ?expand=buckets in the Buckets slice.
	BucketID    int64       `json:"bucket_id,omitempty"`
	Buckets     []*Bucket   `json:"buckets,omitempty"`
	Assignees   []*User     `json:"assignees,omitempty"`
	Labels      []*Label    `json:"labels,omitempty"`
	StartDate   *time.Time  `json:"start_date,omitempty"`
	DueDate     *time.Time  `json:"due_date,omitempty"`
	EndDate     *time.Time  `json:"end_date,omitempty"`
	PercentDone float64     `json:"percent_done,omitempty"`
	Reactions   interface{} `json:"reactions,omitempty"`
}

// TaskComment matches pkg/models/task_comments.TaskComment.
type TaskComment struct {
	ID      int64     `json:"id"`
	Comment string    `json:"comment"`
	Author  *User     `json:"author,omitempty"`
	Created time.Time `json:"created,omitempty"`
	Updated time.Time `json:"updated,omitempty"`
}

// Label is a global (per-user) label.
type Label struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	HexColor    string    `json:"hex_color,omitempty"`
	Created     time.Time `json:"created,omitempty"`
	Updated     time.Time `json:"updated,omitempty"`
}

// LabelTask is the body for `PUT /tasks/{id}/labels`.
type LabelTask struct {
	LabelID int64 `json:"label_id"`
}

// TaskRelation is the body for `PUT /tasks/{id}/relations` and the row
// returned. RelationKind is one of: subtask, parenttask, related, duplicates,
// duplicateof, blocking, blocked, precedes, follows, copiedfrom, copiedto.
type TaskRelation struct {
	TaskID       int64  `json:"task_id,omitempty"`
	OtherTaskID  int64  `json:"other_task_id"`
	RelationKind string `json:"relation_kind"`
}

// TaskAssignee is the body for `PUT /tasks/{id}/assignees`.
type TaskAssignee struct {
	UserID int64 `json:"user_id"`
}

// ProjectUser is the body and response for `PUT /projects/{id}/users`.
type ProjectUser struct {
	ID         int64  `json:"id,omitempty"`
	Username   string `json:"username"`
	Permission int    `json:"permission"`
}

// Permission constants for project sharing.
const (
	PermissionRead      = 0
	PermissionReadWrite = 1
	PermissionAdmin     = 2
)

// APIToken is the request and response shape for `PUT /tokens`. The plaintext
// `Token` field is only populated on creation. Vikunja requires ExpiresAt;
// callers that want a long-lived token use FarFuture (year 9999).
type APIToken struct {
	ID          int64               `json:"id,omitempty"`
	Title       string              `json:"title"`
	Token       string              `json:"token,omitempty"`
	Permissions map[string][]string `json:"permissions"`
	ExpiresAt   time.Time           `json:"expires_at"`
	OwnerID     int64               `json:"owner_id,omitempty"`
	Created     time.Time           `json:"created,omitempty"`
}

// FarFuture is what veans uses for "no expiry" since Vikunja's API token
// model marks expires_at as required. Year 9999 is well past any reasonable
// rotation horizon and is what the frontend uses for its "never" option.
var FarFuture = time.Date(9999, time.December, 31, 0, 0, 0, 0, time.UTC)

// Info is the parsed shape of `GET /info`.
type Info struct {
	Version             string `json:"version"`
	FrontendURL         string `json:"frontend_url"`
	MOTD                string `json:"motd,omitempty"`
	LinkSharingEnabled  bool   `json:"link_sharing_enabled"`
	RegistrationEnabled bool   `json:"registration_enabled"`
	Auth                struct {
		Local struct {
			Enabled bool `json:"enabled"`
		} `json:"local"`
		OpenIDConnect struct {
			Enabled   bool `json:"enabled"`
			Providers []struct {
				Key  string `json:"key"`
				Name string `json:"name"`
			} `json:"providers"`
		} `json:"openid_connect"`
	} `json:"auth"`
}

// LoginRequest is the body for `POST /login`.
type LoginRequest struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	TOTPPasscode string `json:"totp_passcode,omitempty"`
	LongToken    bool   `json:"long_token,omitempty"`
}

// LoginResponse is the JWT bundle.
type LoginResponse struct {
	Token string `json:"token"`
}

// OAuthTokenRequest is the JSON body for POST /api/v1/oauth/token. Vikunja's
// OAuth server explicitly rejects form-encoded requests; everything is JSON.
type OAuthTokenRequest struct {
	GrantType    string `json:"grant_type"`
	Code         string `json:"code,omitempty"`
	ClientID     string `json:"client_id,omitempty"`
	RedirectURI  string `json:"redirect_uri,omitempty"`
	CodeVerifier string `json:"code_verifier,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

// OAuthTokenResponse mirrors the standard RFC 6749 response.
type OAuthTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
}
