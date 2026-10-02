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
	"net/mail"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/i18n"
	"code.vikunja.io/api/pkg/modules/avatar"
	"code.vikunja.io/api/pkg/user"

	"xorm.io/builder"
	"xorm.io/xorm"
)

// The people area (/api/v2/manage) is gated by the instance-admin flag only, on purpose and by the
// operator's decision, independent of pkg/license. The functions here carry no license check
// themselves, like the other *AsAdmin model functions.

// Where a user comes from, as shown in the people area.
const (
	UserSourceLocal  = "local"
	UserSourceEntra  = "entra"
	UserSourceImport = "import"
	UserSourceLDAP   = "ldap"
	UserSourceOther  = "other"
)

// UserSource classifies how a user signs in. "import" is a person the scheduled user list import
// created who has not logged in yet.
func UserSource(u *user.User) string {
	switch {
	case u.Issuer == "" || u.Issuer == user.IssuerLocal:
		return UserSourceLocal
	case u.Issuer == user.IssuerImport:
		return UserSourceImport
	case u.Issuer == user.IssuerLDAP:
		return UserSourceLDAP
	case user.IsEntraIssuer(u.Issuer):
		return UserSourceEntra
	}
	return UserSourceOther
}

// ManagedUser is the view of a user in the people area.
type ManagedUser struct {
	ID                 int64       `json:"id" readOnly:"true" doc:"The numeric id of the user."`
	Username           string      `json:"username" readOnly:"true" doc:"The username. It cannot be changed after the user was created."`
	Name               string      `json:"name" doc:"The display name."`
	Email              string      `json:"email" doc:"The email address."`
	Language           string      `json:"language" doc:"The language of the user."`
	JobTitle           string      `json:"job_title" readOnly:"true" doc:"The job title, maintained by the scheduled user list import."`
	Department         string      `json:"department" readOnly:"true" doc:"The department, maintained by the scheduled user list import."`
	Status             user.Status `json:"status" readOnly:"true" doc:"Account status (0=active, 1=email-confirmation required, 2=disabled, 3=locked)."`
	IsAdmin            bool        `json:"is_admin" readOnly:"true" doc:"Whether the user is an instance admin."`
	Source             string      `json:"source" readOnly:"true" enum:"local,entra,import,ldap,other" doc:"How the user signs in: local, entra, import (created by the user list import, never logged in yet), ldap or other."`
	Issuer             string      `json:"issuer" readOnly:"true" doc:"The authentication issuer."`
	ImportID           string      `json:"import_id,omitempty" readOnly:"true" doc:"The Microsoft Entra object id from the user list import."`
	MustChangePassword bool        `json:"must_change_password" readOnly:"true" doc:"Whether the user has to choose a new password at the next login."`
	ProfileEditable    bool        `json:"profile_editable" readOnly:"true" doc:"False for users whose profile is managed by a third-party provider or the user list import; their name, email, job title and department cannot be changed here."`
	Created            time.Time   `json:"created" readOnly:"true" doc:"When the user was created."`
}

// NewManagedUser builds the people area view of a user.
func NewManagedUser(u *user.User) *ManagedUser {
	return &ManagedUser{
		ID:                 u.ID,
		Username:           u.Username,
		Name:               u.Name,
		Email:              u.Email,
		Language:           u.Language,
		JobTitle:           u.JobTitle,
		Department:         u.Department,
		Status:             u.Status,
		IsAdmin:            u.IsAdmin,
		Source:             UserSource(u),
		Issuer:             u.Issuer,
		ImportID:           u.ImportID,
		MustChangePassword: u.MustChangePassword,
		ProfileEditable:    u.IsLocalUser(),
		Created:            u.Created,
	}
}

// GetManagedUser returns one user, whatever their status, with their email address. The normal
// user getters blank the email and refuse disabled and locked accounts.
func GetManagedUser(s *xorm.Session, id int64) (*user.User, error) {
	return loadAdminTargetUser(s, id)
}

// ManageUserFilter narrows the list of the people area.
type ManageUserFilter struct {
	Search string
	// Status is -1 for any.
	Status int
	// Source is empty for any, otherwise one of the UserSource constants.
	Source string
}

func sourceCond(source string) builder.Cond {
	localIssuer := builder.Or(builder.Eq{"issuer": user.IssuerLocal}, builder.Eq{"issuer": ""}, builder.IsNull{"issuer"})
	switch source {
	case UserSourceLocal:
		return localIssuer
	case UserSourceImport:
		return builder.Eq{"issuer": user.IssuerImport}
	case UserSourceLDAP:
		return builder.Eq{"issuer": user.IssuerLDAP}
	case UserSourceEntra:
		return builder.Like{"issuer", "https://login.microsoftonline.com/%"}
	case UserSourceOther:
		return builder.And(
			builder.Neq{"issuer": user.IssuerImport},
			builder.Neq{"issuer": user.IssuerLDAP},
			builder.Expr("issuer NOT LIKE ?", "https://login.microsoftonline.com/%"),
			builder.Neq{"issuer": user.IssuerLocal},
			builder.Neq{"issuer": ""},
			builder.NotNull{"issuer"},
		)
	}
	return nil
}

// ListManagedUsers lists the non-bot users, filtered and paginated. Emails are part of the result,
// so the read is audited like the licensed admin list.
func ListManagedUsers(s *xorm.Session, doer *user.User, f ManageUserFilter, page, perPage int) ([]*user.User, int64, error) {
	events.DispatchOnCommit(s, &AdminUsersListedEvent{Doer: doer})

	cond := builder.Or(builder.IsNull{"bot_owner_id"}, builder.Eq{"bot_owner_id": 0})
	if f.Search != "" {
		q := "%" + f.Search + "%"
		cond = builder.And(cond, builder.Or(
			builder.Like{"username", q},
			builder.Like{"name", q},
			builder.Like{"email", q},
		))
	}
	if f.Status >= 0 {
		cond = builder.And(cond, builder.Eq{"status": f.Status})
	}
	if c := sourceCond(f.Source); c != nil {
		cond = builder.And(cond, c)
	}

	var users []*user.User
	total, err := s.
		Where(cond).
		OrderBy("id ASC").
		Limit(perPage, (page-1)*perPage).
		FindAndCount(&users)
	if err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

// ManageProfileUpdate holds the fields an admin may change on a local user. Username is not part
// of it: stored rich text references users by username, so a rename would break their mentions.
type ManageProfileUpdate struct {
	Name     string
	Email    string
	Language string
}

// UpdateUserProfileAsAdmin changes the name, email and language of a local user and nothing else.
// It writes an explicit column list: user.UpdateUser writes dozens of columns, so reusing it with a
// partial body would reset them. Users managed by a third-party provider are refused. It does not
// commit; the caller owns the transaction.
func UpdateUserProfileAsAdmin(s *xorm.Session, doer *user.User, id int64, update ManageProfileUpdate) (*user.User, error) {
	target, err := loadAdminTargetUser(s, id)
	if err != nil {
		return nil, err
	}
	if target.IsBot() {
		return nil, ErrInvalidData{Message: "bot accounts cannot be edited here"}
	}
	if !target.IsLocalUser() {
		return nil, &user.ErrProfileManagedExternally{UserID: target.ID}
	}

	email := strings.TrimSpace(update.Email)
	if email == "" {
		return nil, ErrInvalidData{Message: "the email address must not be empty"}
	}
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email {
		return nil, ErrInvalidData{Message: "the email address is not valid"}
	}
	if update.Language != "" && !i18n.HasLanguage(update.Language) {
		return nil, ErrInvalidData{Message: "the language does not exist"}
	}

	var fields, cols []string
	flushAvatar := false

	if update.Name != target.Name {
		flushAvatar = target.AvatarProvider == "initials"
		target.Name = update.Name
		cols = append(cols, "name")
		fields = append(fields, "name")
	}
	if !strings.EqualFold(email, target.Email) {
		err = user.CheckEmailNotTaken(s, email)
		if err != nil {
			return nil, err
		}
		target.Email = email
		cols = append(cols, "email")
		fields = append(fields, "email")
	}
	// An address an admin set takes effect at once, so a change the user left unconfirmed is obsolete.
	if target.PendingEmail != "" {
		target.PendingEmail = ""
		cols = append(cols, "pending_email")
	}
	if update.Language != "" && update.Language != target.Language {
		target.Language = update.Language
		cols = append(cols, "language")
		fields = append(fields, "language")
	}

	if len(cols) == 0 {
		return target, nil
	}

	_, err = s.ID(target.ID).Cols(cols...).Update(target)
	if err != nil {
		return nil, err
	}
	if flushAvatar {
		avatar.FlushAllCaches(target)
	}
	if len(fields) > 0 {
		events.DispatchOnCommit(s, &ManageUserProfileUpdatedEvent{User: target, Doer: doer, Fields: fields})
	}
	return target, nil
}

// TransferableProject is a project a user owns, as shown before transferring it.
type TransferableProject struct {
	ID              int64  `json:"id" doc:"The project id."`
	Title           string `json:"title" doc:"The project title."`
	IsArchived      bool   `json:"is_archived" doc:"Whether the project is archived."`
	ParentProjectID int64  `json:"parent_project_id" doc:"The parent project id, 0 for a top-level project."`
}

// ListOwnedProjects returns every project the user owns, archived included.
func ListOwnedProjects(s *xorm.Session, userID int64) ([]*TransferableProject, error) {
	_, err := loadAdminTargetUser(s, userID)
	if err != nil {
		return nil, err
	}

	var projects []*Project
	err = s.Where("owner_id = ?", userID).OrderBy("id ASC").Find(&projects)
	if err != nil {
		return nil, err
	}

	out := make([]*TransferableProject, 0, len(projects))
	for _, p := range projects {
		out = append(out, &TransferableProject{
			ID:              p.ID,
			Title:           p.Title,
			IsArchived:      p.IsArchived,
			ParentProjectID: p.parentID(),
		})
	}
	return out, nil
}

// TransferProjectsAsAdmin hands the projects of one user to another, for example before a leaver
// is deleted: deleting a user removes every project they own. projectIDs selects projects, nil
// means every project the source owns. Nothing is moved unless every selected project can be. It
// does not commit; the caller owns the transaction. It returns the number of projects moved.
func TransferProjectsAsAdmin(s *xorm.Session, doer *user.User, fromUserID, toUserID int64, projectIDs []int64) (int, error) {
	if fromUserID == toUserID {
		return 0, ErrInvalidData{Message: "the new owner must be a different user"}
	}

	_, err := loadAdminTargetUser(s, fromUserID)
	if err != nil {
		return 0, err
	}
	newOwner, err := loadAdminTargetUser(s, toUserID)
	if err != nil {
		return 0, err
	}
	if newOwner.IsBot() {
		return 0, ErrInvalidData{Message: "a bot cannot own projects"}
	}
	if newOwner.Status != user.StatusActive {
		return 0, ErrInvalidData{Message: "the new owner must be an active user"}
	}

	var owned []*Project
	err = s.Where("owner_id = ?", fromUserID).OrderBy("id ASC").Find(&owned)
	if err != nil {
		return 0, err
	}

	selected := owned
	if projectIDs != nil {
		byID := make(map[int64]*Project, len(owned))
		for _, p := range owned {
			byID[p.ID] = p
		}
		selected = make([]*Project, 0, len(projectIDs))
		seen := make(map[int64]bool, len(projectIDs))
		for _, id := range projectIDs {
			if seen[id] {
				continue
			}
			seen[id] = true
			p, ok := byID[id]
			if !ok {
				return 0, ErrInvalidData{Message: "a selected project is not owned by the source user"}
			}
			selected = append(selected, p)
		}
	}

	for _, p := range selected {
		_, err = ReassignProjectOwner(s, doer, p.ID, toUserID)
		if err != nil {
			return 0, err
		}
	}
	return len(selected), nil
}
