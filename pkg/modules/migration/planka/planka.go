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

package planka

import (
	"bytes"
	"context"

	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/modules/migration"
	"code.vikunja.io/api/pkg/user"
)

// Migrator imports projects, boards, cards and everything attached to them from a Planka v2 instance.
// It authenticates with either an API key / access token or username + password.
type Migrator struct {
	URL      string `json:"url"`
	Token    string `json:"token"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// Name is the name of the planka migration.
func (m *Migrator) Name() string {
	return "planka"
}

// AuthURL is empty: Planka has no OAuth flow, credentials are passed with the migrate request.
func (m *Migrator) AuthURL() string {
	return ""
}

// validate checks the request payload without talking to Planka.
func (m *Migrator) validate() error {
	if m.URL == "" || (m.Token == "" && (m.Username == "" || m.Password == "")) {
		return &ErrInvalidConfig{}
	}
	return nil
}

// CheckCredentials logs in to Planka and out again. Used to fail fast before the async migration is queued.
func (m *Migrator) CheckCredentials() error {
	ctx, cancel := context.WithTimeout(context.Background(), credentialCheckTimeout)
	defer cancel()

	c, err := m.connect(ctx)
	if err != nil {
		return err
	}
	c.logout(ctx)
	return nil
}

func (m *Migrator) connect(ctx context.Context) (*client, error) {
	if err := m.validate(); err != nil {
		return nil, err
	}
	c, err := newClient(m.URL)
	if err != nil {
		return nil, err
	}
	if err := c.login(ctx, m.Token, m.Username, m.Password); err != nil {
		return nil, err
	}
	return c, nil
}

// Migrate gets all projects, boards and cards from planka for a user and puts them into vikunja.
func (m *Migrator) Migrate(u *user.User) error {
	log.Debugf("[Planka Migration] Starting migration for user %d", u.ID)

	// the async migration is not bound by a request deadline, the client's own timeout applies
	c, err := m.connect(context.Background())
	if err != nil {
		return err
	}
	defer c.logout(context.Background())

	data, err := fetchAll(c)
	if err != nil {
		return err
	}

	log.Debugf("[Planka Migration] Fetched all planka data for user %d, converting", u.ID)

	hierarchy, err := convertPlankaToVikunja(data, func(a *plankaAttachment) (*bytes.Buffer, error) {
		return c.download(a.ID, a.Name)
	})
	if err != nil {
		return err
	}

	log.Debugf("[Planka Migration] Inserting %d projects for user %d", len(hierarchy), u.ID)

	if err := migration.InsertFromStructure(hierarchy, u); err != nil {
		return err
	}

	log.Debugf("[Planka Migration] Done migrating planka data for user %d", u.ID)
	return nil
}
