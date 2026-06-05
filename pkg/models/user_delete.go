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
	"time"

	"code.vikunja.io/api/pkg/cron"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/notifications"
	"code.vikunja.io/api/pkg/user"

	"xorm.io/builder"
	"xorm.io/xorm"
)

// User deletion must happen here in this packaage because we want to delete everything associated to this user.
// Because most of these things are managed in the models package, using them has to happen here.

// RegisterUserDeletionCron registers the cron job that actually removes users who are scheduled to delete.
func RegisterUserDeletionCron() {
	err := cron.Schedule("0 * * * *", deleteUsers)
	if err != nil {
		log.Errorf("Could not register deletion cron: %s", err.Error())
	}
}

func deleteUsers() {
	s := db.NewSession()
	users := []*user.User{}
	err := s.Where(builder.Lt{"deletion_scheduled_at": time.Now()}).
		Find(&users)
	s.Close()
	if err != nil {
		log.Errorf("Could not get users scheduled for deletion: %s", err)
		return
	}

	if len(users) == 0 {
		return
	}

	log.Debugf("Found %d users scheduled for deletion", len(users))

	now := time.Now()

	for _, u := range users {
		if !u.DeletionScheduledAt.Before(now) {
			log.Debugf("User %d is not yet scheduled for deletion. Scheduled at %s, now is %s", u.ID, u.DeletionScheduledAt, now)
			continue
		}

		func() {
			us := db.NewSession()
			defer us.Close()

			err = DeleteUser(us, u)
			if err != nil {
				_ = us.Rollback()
				log.Errorf("Could not delete u %d: %s", u.ID, err)
				return
			}

			log.Debugf("Deleted user %d", u.ID)

			err = us.Commit()
			if err != nil {
				log.Errorf("Could not commit transaction: %s", err)
			}
		}()
	}
}

func getProjectsToDelete(s *xorm.Session, u *user.User) (projectsToDelete []*Project, err error) {
	projectsToDelete = []*Project{}
	projects, _, err := getAllProjectsForUser(s, u.ID, &projectOptions{
		page:        0,
		perPage:     -1,
		getArchived: true,
	})
	if err != nil {
		return nil, err
	}

	for _, l := range projects {
		if l.ID < 0 {
			continue
		}

		hadUsers, err := ensureProjectAdminUser(s, l)
		if err != nil {
			return nil, err
		}
		if hadUsers {
			continue
		}
		hadTeams, err := ensureProjectAdminTeam(s, l)
		if err != nil {
			return nil, err
		}

		if hadTeams {
			continue
		}

		projectsToDelete = append(projectsToDelete, l)
	}

	return
}

// DeleteUser completely removes a user and all their associated projects and tasks.
// This action is irrevocable.
// Public to allow deletion from the CLI.
func DeleteUser(s *xorm.Session, u *user.User) (err error) {
	// Delete any bot users owned by this user first (cascades their data too).
	var ownedBots []*user.User
	if err = s.Where("bot_owner_id = ?", u.ID).Find(&ownedBots); err != nil {
		return err
	}
	for _, bot := range ownedBots {
		if err = DeleteUser(s, bot); err != nil {
			return err
		}
	}

	projectsToDelete, err := getProjectsToDelete(s, u)
	if err != nil {
		return err
	}

	for _, p := range projectsToDelete {
		if p.ParentProjectID != 0 {
			// Child projects are deleted by p.Delete
			continue
		}
		err = p.Delete(s, u)
		// If the user is the owner of the default project it will be deleted, if they are not the owner
		// we can ignore the error as the project was shared in that case.
		if err != nil && !IsErrCannotDeleteDefaultProject(err) {
			return err
		}
	}

	// Delete all related entities
	relatedEntities := []struct {
		column string
		model  any
	}{
		{"user_id", &TaskAssginee{}},
		{"user_id", &Subscription{}},
		{"user_id", &TeamMember{}},
		{"owner_id", &SavedFilter{}},
		{"user_id", &Reaction{}},
		{"user_id", &Favorite{}},
		{"owner_id", &APIToken{}},
	}

	for _, entity := range relatedEntities {
		_, err = s.Where(entity.column+" = ?", u.ID).Delete(entity.model)
		if err != nil {
			return err
		}
	}

	// Notify before deleting the user row, because ShouldNotify will try to
	// look up the user and fail if the row is already gone.
	err = notifications.Notify(u, &user.AccountDeletedNotification{
		User: u,
	}, s)
	if err != nil {
		return err
	}

	_, err = s.Where("id = ?", u.ID).Delete(&user.User{})
	return err
}

func ensureProjectAdminUser(s *xorm.Session, l *Project) (hadUsers bool, err error) {
	projectUsers := []*ProjectUser{}
	err = s.Where("project_id = ?", l.ID).Find(&projectUsers)
	if err != nil {
		return
	}

	if len(projectUsers) == 0 {
		return false, nil
	}

	for _, lu := range projectUsers {
		if lu.Permission == PermissionAdmin {
			// Project already has more than one admin, no need to do anything
			return true, nil
		}
	}

	for _, lu := range projectUsers {
		if lu.Permission == PermissionWrite {
			lu.Permission = PermissionAdmin
			_, err = s.Where("id = ?", lu.ID).
				Cols("permission").
				Update(lu)
			return true, err
		}
	}

	firstUser := projectUsers[0]
	firstUser.Permission = PermissionAdmin
	_, err = s.Where("id = ?", firstUser.ID).
		Cols("permission").
		Update(firstUser)
	if err != nil {
		return true, err
	}

	_, err = s.Where("id = ?", l.ID).
		Cols("owner_id").
		Update(&Project{OwnerID: firstUser.UserID})
	if err != nil {
		return true, err
	}

	return true, err
}

func ensureProjectAdminTeam(s *xorm.Session, l *Project) (hadTeams bool, err error) {
	projectTeams := []*TeamProject{}
	err = s.Where("project_id = ?", l.ID).Find(&projectTeams)
	if err != nil {
		return
	}

	if len(projectTeams) == 0 {
		return false, nil
	}

	for _, lu := range projectTeams {
		if lu.Permission == PermissionAdmin {
			// Project already has more than one admin, no need to do anything
			return true, nil
		}
	}

	for _, lu := range projectTeams {
		if lu.Permission == PermissionWrite {
			lu.Permission = PermissionAdmin
			_, err = s.Where("id = ?", lu.ID).
				Cols("permission").
				Update(lu)
			return true, err
		}
	}

	firstTeam := projectTeams[0]
	firstTeam.Permission = PermissionAdmin
	_, err = s.Where("id = ?", firstTeam.ID).
		Cols("permission").
		Update(firstTeam)
	return true, err
}
