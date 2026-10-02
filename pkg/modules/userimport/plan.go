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
	"fmt"

	"code.vikunja.io/api/pkg/user"
)

// smallInstanceThreshold is the number of eligible active users below which the percentage limit is
// replaced by "at most one user per run": with 5 users, 20% would forbid every single leaver.
const smallInstanceThreshold = 10

// userChange is a user whose fields already carry the new values, together with the columns
// to write. Writes always use an explicit column list.
type userChange struct {
	User *user.User
	Cols []string
	// Reenabled is set when the change also moves a deactivated user back to active.
	Reenabled bool
}

// Plan is what a run would do. It is built without touching the database so the safety limit
// can be checked before anything is written.
type Plan struct {
	// Create are people in the file that no user maps to.
	Create []*Person
	// Update are listed users whose name, job title, department or import id changed, or who are
	// re-enabled (status is then one of the columns).
	Update []userChange
	// Disable are active users that are not in the file, or are listed with accountEnabled false.
	Disable []*user.User

	// Reenabled is the number of Update entries that also change the status back to active.
	Reenabled int
	// Unchanged are listed or already deactivated users that need no write.
	Unchanged int
	// Exempt are users the import never changes: admins, accounts that do not sign in through Entra
	// (local, LDAP, other OpenID providers), and users who are locked or awaiting email
	// confirmation. Bots are not counted, they are never loaded.
	Exempt int
	// EligibleActive is the number of active, non-exempt users. It is the base of the safety limit.
	EligibleActive int
}

// isExempt reports whether the import must leave a user alone. Bots never get here, buildPlan
// skips them first.
//
// Only accounts that sign in through the Entra tenant of the list (see userimport.tenantid), plus
// imported rows that have not logged in yet, are managed. Local, LDAP, other OpenID providers and
// other Entra tenants are not described by the list, so deactivating them for being absent from it
// would lock out people it knows nothing about.
func isExempt(u *user.User) bool {
	if u.IsAdmin {
		return true
	}
	if u.Issuer != user.IssuerImport && !user.IsImportedEntraIssuer(u.Issuer) {
		return true
	}
	return u.Status == user.StatusAccountLocked || u.Status == user.StatusEmailConfirmationRequired
}

// buildPlan compares the parsed file with the existing non-bot users.
//
// A user that already has an import id is mapped by that id only. Every other user is mapped by
// email (case-insensitive). The distinction matters because mail addresses get reused: when a
// leaver's address is given to a new hire, the leaver keeps their own id, so they stay a leaver
// and the new hire gets an account of their own, instead of inheriting the leaver's row.
//
// Mapped and enabled users get their name, job title and department refreshed, and are re-enabled
// if they were deactivated. Everyone else who is active is deactivated. The email is never changed.
//
// It modifies the passed users in place to carry the new values.
func buildPlan(people []*Person, users []*user.User) *Plan {
	plan := &Plan{}

	byMail := make(map[string]*Person, len(people))
	byID := make(map[string]*Person, len(people))
	for _, p := range people {
		byMail[normalizeMail(p.Mail)] = p
		byID[p.ID] = p
	}

	// Which user already carries an import id, to never give the same id to two rows.
	idOwner := make(map[string]int64)
	for _, u := range users {
		if u.ImportID == "" {
			continue
		}
		if _, taken := idOwner[u.ImportID]; !taken {
			idOwner[u.ImportID] = u.ID
		}
	}

	matched := make(map[string]bool, len(people))

	for _, u := range users {
		if u.BotOwnerID > 0 {
			continue
		}

		person := findPerson(u, byMail, byID)
		if person != nil {
			// Also for exempt users: an admin listed in the file must not get a second account.
			matched[person.ID] = true
		}

		if isExempt(u) {
			plan.Exempt++
			continue
		}
		if u.Status == user.StatusActive {
			plan.EligibleActive++
		}

		if person == nil || !person.Enabled {
			if u.Status == user.StatusActive {
				plan.Disable = append(plan.Disable, u)
			} else {
				plan.Unchanged++
			}
			continue
		}

		var cols []string

		// The file is the source of truth for the display name, but an empty name never blanks one.
		if person.DisplayName != "" && person.DisplayName != u.Name {
			u.Name = person.DisplayName
			cols = append(cols, "name")
		}
		// A blank job title or department in the file clears the value.
		if person.JobTitle != u.JobTitle {
			u.JobTitle = person.JobTitle
			cols = append(cols, "job_title")
		}
		if person.Department != u.Department {
			u.Department = person.Department
			cols = append(cols, "department")
		}
		if u.ImportID == "" {
			if _, taken := idOwner[person.ID]; !taken {
				u.ImportID = person.ID
				idOwner[person.ID] = u.ID
				cols = append(cols, "import_id")
			}
		}
		reenabled := false
		if u.Status == user.StatusDisabled {
			u.Status = user.StatusActive
			cols = append(cols, "status")
			reenabled = true
			plan.Reenabled++
		}

		if len(cols) == 0 {
			plan.Unchanged++
			continue
		}
		plan.Update = append(plan.Update, userChange{User: u, Cols: cols, Reenabled: reenabled})
	}

	for _, p := range people {
		if p.Enabled && !matched[p.ID] {
			plan.Create = append(plan.Create, p)
		}
	}

	return plan
}

// findPerson maps a user to a row of the file. A user with an import id is matched by that id only,
// everybody else by email.
func findPerson(u *user.User, byMail, byID map[string]*Person) *Person {
	if u.ImportID != "" {
		return byID[u.ImportID]
	}
	if email := normalizeMail(u.Email); email != "" {
		return byMail[email]
	}
	return nil
}

// checkLimit refuses a plan that would deactivate too many users. That protects against a
// truncated, wrong or half-written file locking everyone out.
func (p *Plan) checkLimit(maxPercent int) error {
	disable := len(p.Disable)
	if disable == 0 {
		return nil
	}

	if p.EligibleActive < smallInstanceThreshold {
		if disable > 1 {
			return fmt.Errorf("the run would deactivate %d of %d active users, only 1 is allowed per run below %d users",
				disable, p.EligibleActive, smallInstanceThreshold)
		}
		return nil
	}

	if disable*100 > p.EligibleActive*maxPercent {
		return fmt.Errorf("the run would deactivate %d of %d active users, more than the allowed %d%%",
			disable, p.EligibleActive, maxPercent)
	}
	return nil
}
