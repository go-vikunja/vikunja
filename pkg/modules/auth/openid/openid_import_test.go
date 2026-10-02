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

package openid

import (
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/builder"
)

const testEntraIssuer = "https://login.microsoftonline.com/tenant/v2.0"

// insertImportedUser inserts a row like the scheduled user import creates it.
func insertImportedUser(t *testing.T, email, name, oid string, status user.Status) *user.User {
	t.Helper()
	return insertTestUser(t, &user.User{
		Username: "imported-" + oid,
		Email:    email,
		Name:     name,
		Issuer:   user.IssuerImport,
		Subject:  oid,
		ImportID: oid,
		Status:   status,
	})
}

func insertTestUser(t *testing.T, u *user.User) *user.User {
	t.Helper()
	s := db.NewSession()
	defer s.Close()
	_, err := s.Insert(u)
	require.NoError(t, err)
	require.NoError(t, s.Commit())
	return u
}

func login(t *testing.T, cl *claims, issuer, subject string) (*user.User, error) {
	t.Helper()
	s := db.NewSession()
	defer s.Close()

	u, err := getOrCreateUser(s, cl, &Provider{}, &oidc.IDToken{Issuer: issuer, Subject: subject})
	if err != nil {
		_ = s.Rollback()
		return nil, err
	}
	require.NoError(t, s.Commit())
	return u, nil
}

func countByEmail(t *testing.T, email string) int64 {
	t.Helper()
	s := db.NewSession()
	defer s.Close()
	n, err := s.Where(builder.Eq{"email": email}).Count(&user.User{})
	require.NoError(t, err)
	return n
}

func TestGetOrCreateUserImported(t *testing.T) {
	t.Run("an Entra login claims the imported row by oid and the file keeps owning the name", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		imported := insertImportedUser(t, "jane@corp.com", "Jane From File", "oid-1", user.StatusActive)

		u, err := login(t, &claims{Email: "jane.token@corp.com", Name: "Jane From Token", PreferredUsername: "jane@corp.com", OID: "oid-1"}, testEntraIssuer, "sub-abc")
		require.NoError(t, err)

		assert.Equal(t, imported.ID, u.ID, "no second account is created")
		db.AssertExists(t, "users", map[string]interface{}{
			"id":        imported.ID,
			"issuer":    testEntraIssuer,
			"subject":   "sub-abc",
			"import_id": "oid-1",
			"name":      "Jane From File",
			"email":     "jane.token@corp.com",
		}, false)
	})
	t.Run("the next login finds the claimed user by issuer and subject", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		imported := insertImportedUser(t, "jane@corp.com", "Jane From File", "oid-1", user.StatusActive)
		cl := &claims{Email: "jane@corp.com", Name: "Jane From Token", OID: "oid-1"}

		_, err := login(t, cl, testEntraIssuer, "sub-abc")
		require.NoError(t, err)
		again, err := login(t, cl, testEntraIssuer, "sub-abc")
		require.NoError(t, err)

		assert.Equal(t, imported.ID, again.ID)
		assert.Equal(t, int64(1), countByEmail(t, "jane@corp.com"))
	})
	t.Run("a token with an oid that matches nothing does not claim a row that has the same email", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		imported := insertImportedUser(t, "jane@corp.com", "Jane From File", "oid-1", user.StatusActive)

		u, err := login(t, &claims{Email: "jane@corp.com", Name: "Impostor", PreferredUsername: "impostor", OID: "oid-someone-else"}, testEntraIssuer, "sub-xyz")
		require.NoError(t, err)

		assert.NotEqual(t, imported.ID, u.ID)
		db.AssertExists(t, "users", map[string]interface{}{"id": imported.ID, "issuer": user.IssuerImport, "subject": "oid-1"}, false)
	})
	t.Run("without an oid the email and preferred username never claim a row", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		imported := insertImportedUser(t, "jane@corp.com", "Jane From File", "oid-1", user.StatusActive)

		// Same mail as the imported person, but no oid to prove it is them.
		u, err := login(t, &claims{Email: "JANE@Corp.com", Name: "Not Jane", PreferredUsername: "jane@corp.com"}, testEntraIssuer, "sub-abc")
		require.NoError(t, err)

		assert.NotEqual(t, imported.ID, u.ID)
		db.AssertExists(t, "users", map[string]interface{}{"id": imported.ID, "issuer": user.IssuerImport, "subject": "oid-1"}, false)
	})
	t.Run("empty claims never match an imported row", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		badRow := insertImportedUser(t, "", "No Mail", "oid-1", user.StatusActive)

		_, err := login(t, &claims{Email: "", PreferredUsername: ""}, testEntraIssuer, "sub-abc")

		require.Error(t, err, "a login without an email is still refused")
		db.AssertExists(t, "users", map[string]interface{}{"id": badRow.ID, "issuer": user.IssuerImport}, false)
	})
	t.Run("only Entra tokens can claim an imported row", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		imported := insertImportedUser(t, "jane@corp.com", "Jane From File", "oid-1", user.StatusActive)

		u, err := login(t, &claims{Email: "jane@corp.com", PreferredUsername: "jane", OID: "oid-1"}, "https://some.other.issuer", "sub-abc")
		require.NoError(t, err)

		assert.NotEqual(t, imported.ID, u.ID)
		assert.Empty(t, u.ImportID, "the oid claim is not trusted for other issuers")
		db.AssertExists(t, "users", map[string]interface{}{"id": imported.ID, "issuer": user.IssuerImport}, false)
	})
	t.Run("an already claimed row is never taken by a second login", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		imported := insertImportedUser(t, "jane@corp.com", "Jane From File", "oid-1", user.StatusActive)
		first, err := login(t, &claims{Email: "jane@corp.com", PreferredUsername: "jane"}, testEntraIssuer, "sub-first")
		require.NoError(t, err)
		require.Equal(t, imported.ID, first.ID)

		second, err := login(t, &claims{Email: "jane@corp.com", PreferredUsername: "jane-second", OID: "oid-1"}, testEntraIssuer, "sub-second")
		require.NoError(t, err)

		assert.NotEqual(t, first.ID, second.ID)
		assert.Empty(t, second.ImportID, "the import id is never given to two rows")
		db.AssertExists(t, "users", map[string]interface{}{"id": first.ID, "subject": "sub-first"}, false)
	})
	t.Run("a disabled imported user is matched, so no duplicate is created, and reported as disabled", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		imported := insertImportedUser(t, "jane@corp.com", "Jane From File", "oid-1", user.StatusDisabled)

		u, err := login(t, &claims{Email: "jane@corp.com", OID: "oid-1"}, testEntraIssuer, "sub-abc")
		require.NoError(t, err)

		assert.Equal(t, imported.ID, u.ID)
		assert.Equal(t, user.StatusDisabled, u.Status, "AuthenticateCallback turns this into the account disabled error")
		assert.Equal(t, int64(1), countByEmail(t, "jane@corp.com"))
	})
	t.Run("a reserved preferred username gets a generated username instead of failing the login", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		u, err := login(t, &claims{Email: "helper@corp.com", PreferredUsername: "bot-helper@corp.com", OID: "oid-helper"}, testEntraIssuer, "sub-helper")
		require.NoError(t, err)

		assert.NotEmpty(t, u.Username)
		assert.NotEqual(t, "bot-helper", u.Username)
	})
	t.Run("a username held by a disabled user is not reused, the login gets a generated one", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		// user17 is a disabled fixture user.
		u, err := login(t, &claims{Email: "newcomer@corp.com", PreferredUsername: "user17@corp.com", OID: "oid-newcomer"}, testEntraIssuer, "sub-newcomer")
		require.NoError(t, err)

		assert.NotEmpty(t, u.Username)
		assert.NotEqual(t, "user17", u.Username)
	})
	t.Run("a new Entra user gets the oid as import id", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		u, err := login(t, &claims{Email: "new@corp.com", Name: "New", PreferredUsername: "new@corp.com", OID: "oid-new"}, testEntraIssuer, "sub-new")
		require.NoError(t, err)

		db.AssertExists(t, "users", map[string]interface{}{"id": u.ID, "import_id": "oid-new", "username": "new"}, false)
	})
	t.Run("an existing Entra user without an import id is stamped, the import owns the name from then on", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		existing := insertTestUser(t, &user.User{
			Username: "early", Email: "early@corp.com", Name: "Early Name",
			Issuer: testEntraIssuer, Subject: "sub-early", Status: user.StatusActive,
		})

		// The login that stamps the id still refreshes the name once.
		_, err := login(t, &claims{Email: "early@corp.com", Name: "Token Name", OID: "oid-early"}, testEntraIssuer, "sub-early")
		require.NoError(t, err)
		db.AssertExists(t, "users", map[string]interface{}{"id": existing.ID, "import_id": "oid-early", "name": "Token Name"}, false)

		// Afterwards the token no longer overwrites the name, the scheduled import does.
		_, err = login(t, &claims{Email: "early@corp.com", Name: "Changed In Token", OID: "oid-early"}, testEntraIssuer, "sub-early")
		require.NoError(t, err)
		db.AssertExists(t, "users", map[string]interface{}{"id": existing.ID, "import_id": "oid-early", "name": "Token Name"}, false)
	})
	t.Run("an admin keeps taking the name from the token even with an import id", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		admin := insertTestUser(t, &user.User{
			Username: "boss", Email: "boss@corp.com", Name: "Old Boss Name", IsAdmin: true,
			Issuer: testEntraIssuer, Subject: "sub-boss", ImportID: "oid-boss", Status: user.StatusActive,
		})

		_, err := login(t, &claims{Email: "boss@corp.com", Name: "New Boss Name", OID: "oid-boss"}, testEntraIssuer, "sub-boss")
		require.NoError(t, err)

		db.AssertExists(t, "users", map[string]interface{}{"id": admin.ID, "name": "New Boss Name"}, false)
	})
	t.Run("with a tenant id set, only that tenant can claim a row or be stamped", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		config.UserImportTenantID.Set("tenant")
		t.Cleanup(func() { config.UserImportTenantID.Set("") })
		imported := insertImportedUser(t, "jane@corp.com", "Jane From File", "oid-1", user.StatusActive)

		// Same oid, but issued by another tenant: an object id is only unique per tenant.
		foreign, err := login(t, &claims{Email: "jane@partner.com", Name: "Foreign", PreferredUsername: "jane@partner.com", OID: "oid-1"},
			"https://login.microsoftonline.com/other-tenant/v2.0", "sub-foreign")
		require.NoError(t, err)
		assert.NotEqual(t, imported.ID, foreign.ID)
		assert.Empty(t, foreign.ImportID, "the oid of a foreign tenant is not trusted")
		db.AssertExists(t, "users", map[string]interface{}{"id": imported.ID, "issuer": user.IssuerImport}, false)

		// The configured tenant still works.
		own, err := login(t, &claims{Email: "jane@corp.com", OID: "oid-1"}, testEntraIssuer, "sub-own")
		require.NoError(t, err)
		assert.Equal(t, imported.ID, own.ID)
	})
	t.Run("the stamp is skipped when another row already carries the id", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		insertImportedUser(t, "someone@corp.com", "Someone", "oid-shared", user.StatusActive)
		existing := insertTestUser(t, &user.User{
			Username: "early", Email: "early@corp.com", Name: "Early Name",
			Issuer: testEntraIssuer, Subject: "sub-early", Status: user.StatusActive,
		})

		_, err := login(t, &claims{Email: "early@corp.com", Name: "Token Name", OID: "oid-shared"}, testEntraIssuer, "sub-early")
		require.NoError(t, err)

		u := &user.User{}
		s := db.NewSession()
		defer s.Close()
		found, err := s.ID(existing.ID).Get(u)
		require.NoError(t, err)
		require.True(t, found)
		assert.Empty(t, u.ImportID)
		assert.Equal(t, "Token Name", u.Name, "without an import id the token still provides the name")
	})
	t.Run("other providers are unaffected: the token name still wins and no import id is stored", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		existing := insertTestUser(t, &user.User{
			Username: "other", Email: "other@example.com", Name: "Old Name",
			Issuer: "https://some.other.issuer", Subject: "sub-other", Status: user.StatusActive,
		})

		_, err := login(t, &claims{Email: "other@example.com", Name: "Token Name", OID: "ignored"}, "https://some.other.issuer", "sub-other")
		require.NoError(t, err)

		db.AssertExists(t, "users", map[string]interface{}{"id": existing.ID, "name": "Token Name"}, false)
		u := &user.User{}
		s := db.NewSession()
		defer s.Close()
		found, err := s.ID(existing.ID).Get(u)
		require.NoError(t, err)
		require.True(t, found)
		assert.Empty(t, u.ImportID)
	})
}
