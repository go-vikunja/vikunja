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

package userimport

import (
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const entraIssuer = "https://login.microsoftonline.com/tenant/v2.0"

func entraUser(id int64, email string) *user.User {
	return &user.User{
		ID:       id,
		Username: email,
		Email:    email,
		Issuer:   entraIssuer,
		Subject:  "sub-" + email,
		Status:   user.StatusActive,
	}
}

func listed(id, mail, name string) *Person {
	return &Person{ID: id, Mail: mail, DisplayName: name, Enabled: true}
}

func disabledIDs(p *Plan) []int64 {
	ids := []int64{}
	for _, u := range p.Disable {
		ids = append(ids, u.ID)
	}
	return ids
}

func TestBuildPlan(t *testing.T) {
	t.Run("(1) mapped by mail, case-insensitive: name, job title and department are updated, email is not", func(t *testing.T) {
		u := entraUser(1, "jane@corp.com")
		u.Name = "Old Name"
		p := &Person{ID: "aaa-1", Mail: "JANE@Corp.com", DisplayName: "Jane Doe", JobTitle: "Engineer", Department: "IT", Enabled: true}

		plan := buildPlan([]*Person{p}, []*user.User{u})

		require.Len(t, plan.Update, 1)
		assert.ElementsMatch(t, []string{"name", "job_title", "department", "import_id"}, plan.Update[0].Cols)
		assert.NotContains(t, plan.Update[0].Cols, "email")
		assert.Equal(t, "Jane Doe", u.Name)
		assert.Equal(t, "Engineer", u.JobTitle)
		assert.Equal(t, "IT", u.Department)
		assert.Equal(t, "aaa-1", u.ImportID)
		assert.Equal(t, "jane@corp.com", u.Email)
		assert.Empty(t, plan.Create)
		assert.Empty(t, plan.Disable)
	})
	t.Run("only changed columns are written", func(t *testing.T) {
		u := entraUser(1, "jane@corp.com")
		u.Name, u.JobTitle, u.Department, u.ImportID = "Jane", "Engineer", "IT", "aaa-1"
		p := &Person{ID: "aaa-1", Mail: "jane@corp.com", DisplayName: "Jane", JobTitle: "Manager", Department: "IT", Enabled: true}

		plan := buildPlan([]*Person{p}, []*user.User{u})

		require.Len(t, plan.Update, 1)
		assert.Equal(t, []string{"job_title"}, plan.Update[0].Cols)
	})
	t.Run("nothing changed is not a write", func(t *testing.T) {
		u := entraUser(1, "jane@corp.com")
		u.Name, u.ImportID = "Jane", "aaa-1"
		plan := buildPlan([]*Person{listed("aaa-1", "jane@corp.com", "Jane")}, []*user.User{u})
		assert.Empty(t, plan.Update)
		assert.Equal(t, 1, plan.Unchanged)
	})
	t.Run("a blank job title or department clears the value, a blank name does not", func(t *testing.T) {
		u := entraUser(1, "jane@corp.com")
		u.Name, u.JobTitle, u.Department, u.ImportID = "Jane", "Engineer", "IT", "aaa-1"

		plan := buildPlan([]*Person{{ID: "aaa-1", Mail: "jane@corp.com", DisplayName: "", Enabled: true}}, []*user.User{u})

		require.Len(t, plan.Update, 1)
		assert.ElementsMatch(t, []string{"job_title", "department"}, plan.Update[0].Cols)
		assert.Equal(t, "Jane", u.Name)
		assert.Empty(t, u.JobTitle)
		assert.Empty(t, u.Department)
	})
	t.Run("mapped by import id when the mail differs, so the person is not deactivated", func(t *testing.T) {
		u := entraUser(1, "token-mail@corp.com")
		u.ImportID = "aaa-1"
		plan := buildPlan([]*Person{listed("aaa-1", "file-mail@corp.com", "Jane")}, []*user.User{u})

		assert.Empty(t, plan.Disable)
		assert.Empty(t, plan.Create)
		require.Len(t, plan.Update, 1)
		assert.Equal(t, "token-mail@corp.com", u.Email, "the email is never changed")
	})
	t.Run("a user with an import id is matched by that id only, not by mail", func(t *testing.T) {
		u := entraUser(1, "jane@corp.com")
		u.ImportID = "aaa-2"
		plan := buildPlan([]*Person{listed("aaa-1", "jane@corp.com", "Same Mail"), listed("aaa-2", "other@corp.com", "By Id")}, []*user.User{u})

		assert.Equal(t, "By Id", u.Name)
		require.Len(t, plan.Create, 1)
		assert.Equal(t, "aaa-1", plan.Create[0].ID, "the person who only shares the mail gets an account of their own")
	})
	t.Run("a reused mail address does not hand the leaver's account to the new hire", func(t *testing.T) {
		leaver := entraUser(1, "shared@corp.com")
		leaver.Name = "Leaver"
		leaver.ImportID = "aaa-old"
		leaver.Status = user.StatusDisabled
		newHire := listed("aaa-new", "shared@corp.com", "New Hire")

		plan := buildPlan([]*Person{newHire}, []*user.User{leaver})

		assert.Equal(t, "Leaver", leaver.Name, "the leaver's row is left alone")
		assert.Equal(t, user.StatusDisabled, leaver.Status, "and not re-enabled")
		assert.Empty(t, plan.Update)
		require.Len(t, plan.Create, 1)
		assert.Equal(t, "aaa-new", plan.Create[0].ID)
	})
	t.Run("an active leaver whose mail was reused is deactivated", func(t *testing.T) {
		leaver := entraUser(1, "shared@corp.com")
		leaver.ImportID = "aaa-old"

		plan := buildPlan([]*Person{listed("aaa-new", "shared@corp.com", "New Hire")}, []*user.User{leaver})

		assert.Equal(t, []int64{1}, disabledIDs(plan))
		assert.Len(t, plan.Create, 1)
	})
	t.Run("an imported row that has not logged in yet is managed like any other user", func(t *testing.T) {
		waiting := entraUser(1, "waiting@corp.com")
		waiting.Issuer = user.IssuerImport
		waiting.ImportID = "aaa-1"

		gone := buildPlan([]*Person{listed("aaa-9", "other@corp.com", "Other")}, []*user.User{waiting})
		assert.Equal(t, []int64{1}, disabledIDs(gone))

		waiting.Status = user.StatusActive
		listedPlan := buildPlan([]*Person{listed("aaa-1", "waiting@corp.com", "Waiting")}, []*user.User{waiting})
		assert.Empty(t, listedPlan.Disable)
	})
	t.Run("(2) an active user who is not in the file is deactivated", func(t *testing.T) {
		gone := entraUser(1, "gone@corp.com")
		stays := entraUser(2, "stays@corp.com")
		plan := buildPlan([]*Person{listed("aaa-2", "stays@corp.com", "Stays")}, []*user.User{gone, stays})
		assert.Equal(t, []int64{1}, disabledIDs(plan))
	})
	t.Run("(2) also an Entra user without an import id who signed in before the import existed", func(t *testing.T) {
		u := entraUser(1, "early@corp.com")
		assert.Empty(t, u.ImportID)
		plan := buildPlan([]*Person{listed("aaa-9", "someone-else@corp.com", "Else")}, []*user.User{u})
		assert.Equal(t, []int64{1}, disabledIDs(plan))
	})
	t.Run("an already deactivated user who is not in the file is left alone", func(t *testing.T) {
		u := entraUser(1, "gone@corp.com")
		u.Status = user.StatusDisabled
		plan := buildPlan([]*Person{listed("aaa-2", "stays@corp.com", "Stays")}, []*user.User{u})
		assert.Empty(t, plan.Disable)
		assert.Empty(t, plan.Update)
		assert.Equal(t, 1, plan.Unchanged)
	})
	t.Run("accountEnabled false deactivates and does not create", func(t *testing.T) {
		existing := entraUser(1, "off@corp.com")
		off := &Person{ID: "aaa-1", Mail: "off@corp.com", Enabled: false}
		newOff := &Person{ID: "aaa-2", Mail: "newoff@corp.com", Enabled: false}

		plan := buildPlan([]*Person{off, newOff}, []*user.User{existing})

		assert.Equal(t, []int64{1}, disabledIDs(plan))
		assert.Empty(t, plan.Create)
		assert.Empty(t, plan.Update)
	})
	t.Run("a deactivated user who is back in the file is re-enabled", func(t *testing.T) {
		u := entraUser(1, "back@corp.com")
		u.Status = user.StatusDisabled
		plan := buildPlan([]*Person{listed("aaa-1", "back@corp.com", "Back")}, []*user.User{u})

		require.Len(t, plan.Update, 1)
		assert.Contains(t, plan.Update[0].Cols, "status")
		assert.True(t, plan.Update[0].Reenabled)
		assert.Equal(t, user.StatusActive, u.Status)
		assert.Equal(t, 1, plan.Reenabled)
	})
	t.Run("a deactivated user listed with accountEnabled false stays deactivated", func(t *testing.T) {
		u := entraUser(1, "off@corp.com")
		u.Status = user.StatusDisabled
		plan := buildPlan([]*Person{{ID: "aaa-1", Mail: "off@corp.com", Enabled: false}}, []*user.User{u})
		assert.Empty(t, plan.Update)
		assert.Empty(t, plan.Disable)
		assert.Equal(t, user.StatusDisabled, u.Status)
	})
	t.Run("a person nobody maps to is created", func(t *testing.T) {
		plan := buildPlan([]*Person{listed("aaa-1", "new@corp.com", "New")}, []*user.User{entraUser(1, "other@corp.com")})
		require.Len(t, plan.Create, 1)
		assert.Equal(t, "aaa-1", plan.Create[0].ID)
	})

	t.Run("with a tenant id set, users of another Entra tenant are not managed", func(t *testing.T) {
		config.UserImportTenantID.Set("tenant")
		t.Cleanup(func() { config.UserImportTenantID.Set("") })

		own := entraUser(1, "own@corp.com")
		partner := entraUser(2, "partner@other.com")
		partner.Issuer = "https://login.microsoftonline.com/other-tenant/v2.0"

		plan := buildPlan([]*Person{listed("aaa-9", "someone@corp.com", "Someone")}, []*user.User{own, partner})

		assert.Equal(t, []int64{1}, disabledIDs(plan), "only the user of the configured tenant is deactivated")
		assert.Equal(t, 1, plan.Exempt)
	})
	t.Run("without a tenant id every Entra tenant is managed", func(t *testing.T) {
		config.UserImportTenantID.Set("")

		a := entraUser(1, "a@corp.com")
		b := entraUser(2, "b@other.com")
		b.Issuer = "https://login.microsoftonline.com/other-tenant/v2.0"

		plan := buildPlan([]*Person{listed("aaa-9", "someone@corp.com", "Someone")}, []*user.User{a, b})

		assert.ElementsMatch(t, []int64{1, 2}, disabledIDs(plan))
	})
	t.Run("exempt users are never changed, even when absent from the file", func(t *testing.T) {
		admin := entraUser(1, "admin@corp.com")
		admin.IsAdmin = true
		bot := entraUser(2, "bot@corp.com")
		bot.BotOwnerID = 1
		local := entraUser(3, "local@corp.com")
		local.Issuer = user.IssuerLocal
		ldap := entraUser(4, "ldap@corp.com")
		ldap.Issuer = user.IssuerLDAP
		noIssuer := entraUser(5, "old@corp.com")
		noIssuer.Issuer = ""
		locked := entraUser(6, "locked@corp.com")
		locked.Status = user.StatusAccountLocked
		unconfirmed := entraUser(7, "unconfirmed@corp.com")
		unconfirmed.Status = user.StatusEmailConfirmationRequired
		otherOIDC := entraUser(8, "google@corp.com")
		otherOIDC.Issuer = "https://accounts.google.com"

		plan := buildPlan(
			[]*Person{listed("aaa-1", "someone@corp.com", "Someone")},
			[]*user.User{admin, bot, local, ldap, noIssuer, locked, unconfirmed, otherOIDC},
		)

		assert.Empty(t, plan.Disable, "including accounts of another OpenID provider, the list says nothing about them")
		assert.Empty(t, plan.Update)
		assert.Equal(t, 7, plan.Exempt, "the bot is skipped before it is even counted")
		assert.Equal(t, 0, plan.EligibleActive)
		assert.Equal(t, user.StatusAccountLocked, locked.Status)
		assert.Equal(t, user.StatusEmailConfirmationRequired, unconfirmed.Status)
	})
	t.Run("exempt users listed in the file are not updated and do not get a duplicate account", func(t *testing.T) {
		admin := entraUser(1, "admin@corp.com")
		admin.IsAdmin = true
		admin.Name = "Admin"

		plan := buildPlan([]*Person{{ID: "aaa-1", Mail: "admin@corp.com", DisplayName: "Renamed", JobTitle: "CTO", Enabled: true}}, []*user.User{admin})

		assert.Empty(t, plan.Create)
		assert.Empty(t, plan.Update)
		assert.Equal(t, "Admin", admin.Name)
		assert.Empty(t, admin.JobTitle)
	})

	t.Run("two users sharing one mail are both updated and nobody is created", func(t *testing.T) {
		a := entraUser(1, "shared@corp.com")
		b := entraUser(2, "shared@corp.com")
		plan := buildPlan([]*Person{{ID: "aaa-1", Mail: "shared@corp.com", DisplayName: "Shared", Enabled: true}}, []*user.User{a, b})

		assert.Empty(t, plan.Create)
		assert.Len(t, plan.Update, 2)
		assert.Equal(t, "aaa-1", a.ImportID)
		assert.Empty(t, b.ImportID, "an import id is never given to two rows")
	})
	t.Run("an import id already carried by another row is not stamped again", func(t *testing.T) {
		claimed := entraUser(1, "claimed@corp.com")
		claimed.ImportID = "aaa-1"
		byMail := entraUser(2, "jane@corp.com")

		buildPlan([]*Person{listed("aaa-1", "jane@corp.com", "Jane")}, []*user.User{claimed, byMail})

		assert.Empty(t, byMail.ImportID)
	})
	t.Run("counts the active eligible users, the base of the safety limit", func(t *testing.T) {
		active := entraUser(1, "a@corp.com")
		deactivated := entraUser(2, "b@corp.com")
		deactivated.Status = user.StatusDisabled
		admin := entraUser(3, "c@corp.com")
		admin.IsAdmin = true

		plan := buildPlan([]*Person{listed("aaa-1", "a@corp.com", "A")}, []*user.User{active, deactivated, admin})

		assert.Equal(t, 1, plan.EligibleActive)
	})
}

func TestCheckLimit(t *testing.T) {
	plan := func(eligible, disable int) *Plan {
		p := &Plan{EligibleActive: eligible}
		for i := 0; i < disable; i++ {
			p.Disable = append(p.Disable, &user.User{ID: int64(i + 1)})
		}
		return p
	}

	t.Run("nothing to deactivate always passes, including with no users at all", func(t *testing.T) {
		require.NoError(t, plan(0, 0).checkLimit(20))
		require.NoError(t, plan(100, 0).checkLimit(0))
	})
	t.Run("at the limit passes, above it aborts", func(t *testing.T) {
		require.NoError(t, plan(100, 20).checkLimit(20))
		require.Error(t, plan(100, 21).checkLimit(20))
	})
	t.Run("everybody at once is refused", func(t *testing.T) {
		require.Error(t, plan(50, 50).checkLimit(20))
	})
	t.Run("a limit of zero forbids any deactivation", func(t *testing.T) {
		require.Error(t, plan(100, 1).checkLimit(0))
	})
	t.Run("a small instance may lose one user per run, not two", func(t *testing.T) {
		require.NoError(t, plan(5, 1).checkLimit(20))
		require.Error(t, plan(5, 2).checkLimit(20))
		require.NoError(t, plan(9, 1).checkLimit(20))
	})
	t.Run("the percentage rule applies from the threshold on", func(t *testing.T) {
		require.NoError(t, plan(smallInstanceThreshold, 2).checkLimit(20))
		require.Error(t, plan(smallInstanceThreshold, 3).checkLimit(20))
	})
}

func TestImportUsername(t *testing.T) {
	tests := []struct {
		name string
		p    *Person
		want string
	}{
		{"local part of the userPrincipalName", &Person{UserPrincipalName: "jane.doe@corp.com", Mail: "j@corp.com"}, "jane.doe"},
		{"falls back to the mail", &Person{Mail: "jane.doe@corp.com"}, "jane.doe"},
		{"reserved bot- prefix falls through to the mail", &Person{UserPrincipalName: "bot-x@corp.com", Mail: "jane@corp.com"}, "jane"},
		{"reserved everywhere means random", &Person{UserPrincipalName: "bot-x@corp.com", Mail: "bot-y@corp.com"}, ""},
		{"link share pattern is reserved", &Person{UserPrincipalName: "link-share-12@corp.com", Mail: "link-share-13@corp.com"}, ""},
		{"guest marker is dropped", &Person{UserPrincipalName: "jane_ext.com#EXT#@corp.onmicrosoft.com", Mail: "jane@ext.com"}, "jane_ext.com"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, importUsername(tt.p))
		})
	}
}
