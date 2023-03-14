// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2021 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
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

		err = s.Begin()
		if err != nil {
			log.Errorf("Could not start transaction: %s", err)
			return
		}

		err = DeleteUser(s, u)
		if err != nil {
			_ = s.Rollback()
			log.Errorf("Could not delete u %d: %s", u.ID, err)
			return
		}

		log.Debugf("Deleted user %d", u.ID)

		err = s.Commit()
		if err != nil {
			log.Errorf("Could not commit transaction: %s", err)
			return
		}
	}
}

func getNamespacesToDelete(s *xorm.Session, u *user.User) (namespacesToDelete []*Namespace, err error) {
	namespacesToDelete = []*Namespace{}
	nm := &Namespace{IsArchived: true}
	res, _, _, err := nm.ReadAll(s, u, "", 1, -1)
	if err != nil {
		return nil, err
	}

	if res == nil {
		return nil, nil
	}

	namespaces := res.([]*NamespaceWithProjects)
	for _, n := range namespaces {
		if n.ID < 0 {
			continue
		}

		hadUsers, err := ensureNamespaceAdminUser(s, &n.Namespace)
		if err != nil {
			return nil, err
		}
		if hadUsers {
			continue
		}
		hadTeams, err := ensureNamespaceAdminTeam(s, &n.Namespace)
		if err != nil {
			return nil, err
		}
		if hadTeams {
			continue
		}

		namespacesToDelete = append(namespacesToDelete, &n.Namespace)
	}

	return
}

func getProjectsToDelete(s *xorm.Session, u *user.User) (projectsToDelete []*Project, err error) {
	projectsToDelete = []*Project{}
	lm := &Project{IsArchived: true}
	res, _, _, err := lm.ReadAll(s, u, "", 0, -1)
	if err != nil {
		return nil, err
	}

	if res == nil {
		return nil, nil
	}

	projects := res.([]*Project)
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

// DeleteUser completely removes a user and all their associated projects, namespaces and tasks.
// This action is irrevocable.
// Public to allow deletion from the CLI.
func DeleteUser(s *xorm.Session, u *user.User) (err error) {
	namespacesToDelete, err := getNamespacesToDelete(s, u)
	if err != nil {
		return err
	}

	projectsToDelete, err := getProjectsToDelete(s, u)
	if err != nil {
		return err
	}

	// Delete everything not shared with anybody else
	for _, n := range namespacesToDelete {
		err = deleteNamespace(s, n, u, false)
		if err != nil {
			return err
		}
	}

	for _, l := range projectsToDelete {
		err = l.Delete(s, u)
		if err != nil {
			return err
		}
	}

	_, err = s.Where("id = ?", u.ID).Delete(&user.User{})
	if err != nil {
		return err
	}

	return notifications.Notify(u, &user.AccountDeletedNotification{
		User: u,
	})
}

func ensureNamespaceAdminUser(s *xorm.Session, n *Namespace) (hadUsers bool, err error) {
	namespaceUsers := []*NamespaceUser{}
	err = s.Where("namespace_id = ?", n.ID).Find(&namespaceUsers)
	if err != nil {
		return
	}

	if len(namespaceUsers) == 0 {
		return false, nil
	}

	for _, lu := range namespaceUsers {
		if lu.Right == RightAdmin {
			// Project already has more than one admin, no need to do anything
			return true, nil
		}
	}

	firstUser := namespaceUsers[0]
	firstUser.Right = RightAdmin
	_, err = s.Where("id = ?", firstUser.ID).
		Cols("right").
		Update(firstUser)
	return true, err
}

func ensureNamespaceAdminTeam(s *xorm.Session, n *Namespace) (hadTeams bool, err error) {
	namespaceTeams := []*TeamNamespace{}
	err = s.Where("namespace_id = ?", n.ID).Find(&namespaceTeams)
	if err != nil {
		return
	}

	if len(namespaceTeams) == 0 {
		return false, nil
	}

	for _, lu := range namespaceTeams {
		if lu.Right == RightAdmin {
			// Project already has more than one admin, no need to do anything
			return true, nil
		}
	}

	firstTeam := namespaceTeams[0]
	firstTeam.Right = RightAdmin
	_, err = s.Where("id = ?", firstTeam.ID).
		Cols("right").
		Update(firstTeam)
	return true, err
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
		if lu.Right == RightAdmin {
			// Project already has more than one admin, no need to do anything
			return true, nil
		}
	}

	firstUser := projectUsers[0]
	firstUser.Right = RightAdmin
	_, err = s.Where("id = ?", firstUser.ID).
		Cols("right").
		Update(firstUser)
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
		if lu.Right == RightAdmin {
			// Project already has more than one admin, no need to do anything
			return true, nil
		}
	}

	firstTeam := projectTeams[0]
	firstTeam.Right = RightAdmin
	_, err = s.Where("id = ?", firstTeam.ID).
		Cols("right").
		Update(firstTeam)
	return true, err
}
