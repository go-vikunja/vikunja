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
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"xorm.io/builder"
	"xorm.io/xorm"
)

// subtreeIDs returns the project and its descendants; trashed narrows them to live (false) or trashed (true) ones.
func subtreeIDs(s *xorm.Session, projectID int64, trashed bool) (ids []int64, err error) {
	deletedCond := "projects.deleted_at IS NULL"
	if trashed {
		deletedCond = "projects.deleted_at IS NOT NULL"
	}
	ids = []int64{}
	err = s.Table("project_ancestors").
		Join("INNER", "projects", "projects.id = project_ancestors.project_id AND "+deletedCond).
		Where("project_ancestors.ancestor_id = ?", projectID).
		Cols("project_ancestors.project_id").
		Find(&ids)
	return
}

// trashProject moves the project and all its live descendants to the trash in one go.
func trashProject(s *xorm.Session, project *Project, a web.Auth) error {
	ids, err := subtreeIDs(s, project.ID, false)
	if err != nil {
		return err
	}

	doer := doerFromAuth(s, a)
	_, err = s.In("id", ids).Cols("deleted_by_id").NoAutoTime().Update(&Project{DeletedByID: doer.ID})
	if err != nil {
		return err
	}
	// The deleted tag turns this into one update, so every project gets the same deleted_at.
	_, err = s.In("id", ids).Delete(&Project{})
	if err != nil {
		return err
	}

	_, err = s.In("default_project_id", ids).
		Cols("default_project_id").
		Update(&user.User{DefaultProjectID: 0})
	if err != nil {
		return err
	}

	events.DispatchOnCommit(s, &ProjectDeletedEvent{
		Project: project,
		Doer:    doer,
	})
	return nil
}

// hardDeleteProject permanently removes a project, its sub-projects and everything in them.
// Time entries survive without a container, kept by their authors.
func hardDeleteProject(s *xorm.Session, p *Project) (err error) {
	// Lock first: the task loop below deletes task_positions rows before the views are dropped further down.
	_, err = lockProjectViewsForPositionUpdate(s, p.ID)
	if err != nil {
		return err
	}

	// Including trashed tasks: there is nothing to restore them into once the project is gone.
	tasks := []*Task{}
	err = s.Unscoped().Where("project_id = ?", p.ID).Find(&tasks)
	if err != nil {
		return
	}
	for _, task := range tasks {
		// Moves the task's time entries onto the project, detached below
		err = hardDeleteTask(s, task)
		if err != nil {
			return err
		}
	}

	_, err = s.Table("time_entries").
		Where("project_id = ?", p.ID).
		Update(map[string]any{"project_id": 0, "task_id": 0})
	if err != nil {
		return
	}

	err = p.DeleteBackgroundFileIfExists(s)
	if err != nil {
		return
	}

	_, err = s.Where("default_project_id = ?", p.ID).
		Cols("default_project_id").
		Update(&user.User{DefaultProjectID: 0})
	if err != nil {
		return
	}

	views, err := getViewsForProject(s, p.ID)
	if err != nil {
		return
	}
	viewIDs := []int64{}
	for _, v := range views {
		viewIDs = append(viewIDs, v.ID)
	}

	_, err = s.In("project_view_id", viewIDs).Delete(&Bucket{})
	if err != nil {
		return
	}

	_, err = s.In("id", viewIDs).Delete(&ProjectView{})
	if err != nil {
		return
	}

	// Every user's favorite, not just the doer's
	_, err = s.Where("entity_id = ? AND kind = ?", p.ID, FavoriteKindProject).Delete(&Favorite{})
	if err != nil {
		return
	}

	_, err = s.Where("project_id = ?", p.ID).Delete(&LinkSharing{})
	if err != nil {
		return
	}

	_, err = s.Where("project_id = ?", p.ID).Delete(&ProjectUser{})
	if err != nil {
		return
	}

	_, err = s.Where("project_id = ?", p.ID).Delete(&TeamProject{})
	if err != nil {
		return
	}

	err = deleteProjectAncestors(s, p.ID)
	if err != nil {
		return
	}

	_, err = s.Where("project_id = ?", p.ID).Delete(&TaskIndexAlias{})
	if err != nil {
		return
	}

	_, err = s.ID(p.ID).Delete(&ProjectTaskCounter{})
	if err != nil {
		return
	}

	_, err = s.ID(p.ID).Unscoped().Delete(&Project{})
	if err != nil {
		return
	}

	children := []*Project{}
	err = s.Unscoped().Where("parent_project_id = ?", p.ID).Find(&children)
	if err != nil {
		return
	}
	for _, child := range children {
		err = hardDeleteProject(s, child)
		if err != nil {
			return
		}
	}

	return
}

func getTrashedProject(s *xorm.Session, projectID int64) (*Project, error) {
	project := &Project{}
	has, err := s.Unscoped().
		Where(builder.And(builder.Eq{"id": projectID}, builder.NotNull{"deleted_at"})).
		Get(project)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, ErrProjectDoesNotExist{ID: projectID}
	}
	return project, nil
}

// Trashed projects can't be reached through the regular permission checks, which skip them.
func canAdminTrashedProject(s *xorm.Session, a web.Auth, project *Project) (bool, error) {
	if _, isShare := a.(*LinkSharing); isShare {
		return false, nil
	}
	if isInstanceAdmin(s, a) || project.OwnerID == a.GetID() {
		return true, nil
	}
	access, err := getProjectAccessIncludingTrash(s, a.GetID())
	if err != nil {
		return false, err
	}
	permission, has := access.permission(project.ID)
	return has && permission == PermissionAdmin, nil
}

// Readers who could reach the project get ErrProjectIsInTrash; everyone else keeps notFound.
func trashedProjectReadError(s *xorm.Session, a web.Auth, projectID int64, notFound error) error {
	if _, isShare := a.(*LinkSharing); isShare {
		return notFound
	}
	trashed, err := getTrashedProject(s, projectID)
	if IsErrProjectDoesNotExist(err) {
		return notFound
	}
	if err != nil {
		return err
	}

	access, err := getProjectAccessIncludingTrash(s, a.GetID())
	if err != nil {
		return err
	}
	if _, has := access.permission(projectID); !has && trashed.OwnerID != a.GetID() {
		return notFound
	}

	inTrash := ErrProjectIsInTrash{ID: trashed.ID, DeletedAt: trashed.DeletedAt}
	inTrash.CanRestore, err = canAdminTrashedProject(s, a, trashed)
	if err != nil {
		return err
	}
	if trashed.DeletedByID != 0 {
		deleters, err := getUsersOrLinkSharesFromIDs(s, []int64{trashed.DeletedByID})
		if err != nil {
			return err
		}
		inTrash.DeletedBy = deleters[trashed.DeletedByID]
	}
	return inTrash
}

// canDoTrashedProject loads the trashed project into p and checks admin access to it.
func (p *Project) canDoTrashedProject(s *xorm.Session, a web.Auth) (bool, error) {
	trashed, err := getTrashedProject(s, p.ID)
	if err != nil {
		return false, err
	}
	*p = *trashed
	return canAdminTrashedProject(s, a, p)
}

// CanRestore checks whether the auth may restore the trashed project.
func (p *Project) CanRestore(s *xorm.Session, a web.Auth) (bool, error) {
	return p.canDoTrashedProject(s, a)
}

// Restore brings a trashed project back with all its trashed sub-projects,
// whenever they were deleted. Trashed tasks inside stay in the trash.
func (p *Project) Restore(s *xorm.Session, a web.Auth) error {
	if parentID := p.parentID(); parentID > 0 {
		parent := &Project{}
		has, err := s.Unscoped().ID(parentID).Get(parent)
		if err != nil {
			return err
		}
		switch {
		case !has:
			// Unreachable while purging takes all sub-projects along, but a dangling parent must not block the restore.
			p.ParentProjectID = noParentProjectID()
			if _, err := s.ID(p.ID).Unscoped().Cols("parent_project_id").Update(p); err != nil {
				return err
			}
			if err := moveProjectAncestors(s, p.ID, 0); err != nil {
				return err
			}
		case !parent.DeletedAt.IsZero():
			return ErrParentProjectIsInTrash{ProjectID: p.ID, ParentProjectID: parentID}
		case parent.IsArchived:
			return ErrParentProjectIsArchived{ProjectID: p.ID, ParentProjectID: parentID}
		}
	}

	ids, err := subtreeIDs(s, p.ID, true)
	if err != nil {
		return err
	}
	_, err = s.Table("projects").
		In("id", ids).
		Update(map[string]any{"deleted_at": nil, "deleted_by_id": nil})
	if err != nil {
		return err
	}
	p.DeletedAt = time.Time{}
	p.DeletedByID = 0

	_, err = s.In("id", ids).Cols("updated").Update(&Project{})
	if err != nil {
		return err
	}
	// CalDAV clients dropped the collections and only pick the tasks up again through this
	_, err = s.In("project_id", ids).Cols("updated").Update(&Task{})
	if err != nil {
		return err
	}

	events.DispatchOnCommit(s, &ProjectRestoredEvent{
		Project: p,
		Doer:    doerFromAuth(s, a),
	})
	return nil
}

// CanDeletePermanently checks whether the auth may purge the trashed project.
func (p *Project) CanDeletePermanently(s *xorm.Session, a web.Auth) (bool, error) {
	return p.canDoTrashedProject(s, a)
}

// DeletePermanently purges a trashed project with all its sub-projects right away.
func (p *Project) DeletePermanently(s *xorm.Session) error {
	return hardDeleteProject(s, p)
}

// Topmost trashed projects the user is admin of; trashed sub-projects are folded into their trashed ancestor.
func trashedProjectsCond(s *xorm.Session, a web.Auth) (builder.Cond, error) {
	if _, isShare := a.(*LinkSharing); isShare {
		return nil, ErrGenericForbidden{}
	}
	access, err := getProjectAccessIncludingTrash(s, a.GetID())
	if err != nil {
		return nil, err
	}
	adminIDs := []int64{}
	for _, id := range access.sortedIDs {
		if access.permissions[id] == PermissionAdmin {
			adminIDs = append(adminIDs, id)
		}
	}

	return builder.And(
		builder.NotNull{"projects.deleted_at"},
		builder.In("projects.id", adminIDs),
		builder.Or(
			builder.IsNull{"projects.parent_project_id"},
			builder.NotIn("projects.parent_project_id",
				builder.Select("id").From("projects").Where(builder.NotNull{"deleted_at"}),
			),
		),
	), nil
}

// GetTrashedProjects returns the project trash of the auth, newest first.
func GetTrashedProjects(s *xorm.Session, a web.Auth, page, perPage int) (projects []*Project, total int64, err error) {
	cond, err := trashedProjectsCond(s, a)
	if err != nil {
		return nil, 0, err
	}

	total, err = s.Unscoped().Where(cond).Count(&Project{})
	if err != nil {
		return nil, 0, err
	}

	projects = []*Project{}
	err = s.Unscoped().
		Where(cond).
		OrderBy("projects.deleted_at DESC, projects.id DESC").
		Limit(getLimitFromPageIndex(page, perPage)).
		Find(&projects)
	if err != nil || len(projects) == 0 {
		return projects, total, err
	}

	projectIDs := make([]int64, 0, len(projects))
	userIDs := []int64{}
	for _, project := range projects {
		projectIDs = append(projectIDs, project.ID)
		userIDs = append(userIDs, project.OwnerID)
		if project.DeletedByID != 0 {
			userIDs = append(userIDs, project.DeletedByID)
		}
	}

	type subProjectCount struct {
		AncestorID int64 `xorm:"ancestor_id"`
		Count      int64 `xorm:"count"`
	}
	counts := []*subProjectCount{}
	err = s.Table("project_ancestors").
		Select("project_ancestors.ancestor_id, COUNT(*) AS count").
		Join("INNER", "projects", "projects.id = project_ancestors.project_id AND projects.deleted_at IS NOT NULL").
		Where(builder.And(builder.In("project_ancestors.ancestor_id", projectIDs), builder.Gt{"project_ancestors.depth": 0})).
		GroupBy("project_ancestors.ancestor_id").
		Find(&counts)
	if err != nil {
		return nil, 0, err
	}
	countByProject := make(map[int64]int64, len(counts))
	for _, c := range counts {
		countByProject[c.AncestorID] = c.Count
	}

	users, err := getUsersOrLinkSharesFromIDs(s, userIDs)
	if err != nil {
		return nil, 0, err
	}
	for _, project := range projects {
		project.Owner = users[project.OwnerID]
		project.DeletedBy = users[project.DeletedByID]
		project.TrashedSubProjects = countByProject[project.ID]
	}

	return projects, total, nil
}

// EmptyProjectTrash purges every project in the auth's trash, including ones others deleted.
func EmptyProjectTrash(s *xorm.Session, a web.Auth) (purged int, err error) {
	cond, err := trashedProjectsCond(s, a)
	if err != nil {
		return 0, err
	}

	projects := []*Project{}
	if err = s.Unscoped().Where(cond).Find(&projects); err != nil {
		return 0, err
	}
	for _, project := range projects {
		if err = hardDeleteProject(s, project); err != nil {
			return 0, err
		}
	}
	return len(projects), nil
}

// RegisterProjectCleanupCron registers the cron job that permanently removes
// projects which were trashed more than TrashRetention ago.
func RegisterProjectCleanupCron() {
	err := cron.Schedule("30 * * * *", func() {
		deleteExpiredProjects(time.Now())
	})
	if err != nil {
		log.Errorf("Could not register project cleanup cron: %s", err.Error())
	}
}

func deleteExpiredProjects(now time.Time) {
	s := db.NewSession()
	projectIDs := []int64{}
	err := s.Table("projects").
		Where(builder.And(
			builder.NotNull{"deleted_at"},
			builder.Lt{"deleted_at": now.Add(-TrashRetention)},
		)).
		Cols("id").
		Find(&projectIDs)
	s.Close()
	if err != nil {
		log.Errorf("Could not get projects scheduled for permanent deletion: %s", err)
		return
	}

	for _, projectID := range projectIDs {
		func() {
			ps := db.NewSession()
			defer ps.Close()

			// An earlier iteration may have purged it with its parent
			project, err := getTrashedProject(ps, projectID)
			if IsErrProjectDoesNotExist(err) {
				return
			}
			if err == nil {
				err = hardDeleteProject(ps, project)
			}
			if err != nil {
				_ = ps.Rollback()
				log.Errorf("Could not permanently delete project %d: %s", projectID, err)
				return
			}
			if err = ps.Commit(); err != nil {
				log.Errorf("Could not commit transaction: %s", err)
			}
		}()
	}
}

// ProjectCount sums up what moving a project to the trash affects.
type ProjectCount struct {
	SubProjects      int64 `json:"sub_projects" doc:"Live sub-projects at any depth, which move to the trash with the project."`
	Tasks            int64 `json:"tasks" doc:"Tasks in the project and its sub-projects, not counting tasks already in the trash."`
	TimeEntries      int64 `json:"time_entries" doc:"Time entries on the project, its sub-projects or their tasks."`
	TimeEntryAuthors int64 `json:"time_entry_authors" doc:"How many people logged those time entries."`
}

// GetProjectCount counts what the project and its live sub-projects contain. Admins only.
func GetProjectCount(s *xorm.Session, a web.Auth, projectID int64) (*ProjectCount, error) {
	project := &Project{ID: projectID}
	can, err := project.IsAdmin(s, a)
	if err != nil {
		return nil, err
	}
	if !can {
		return nil, ErrGenericForbidden{}
	}

	ids, err := subtreeIDs(s, projectID, false)
	if err != nil {
		return nil, err
	}

	count := &ProjectCount{SubProjects: int64(len(ids)) - 1}
	count.Tasks, err = s.In("project_id", ids).Count(&Task{})
	if err != nil {
		return nil, err
	}

	entriesCond := builder.Or(
		builder.In("project_id", ids),
		builder.In("task_id", builder.Select("id").From("tasks").Where(builder.In("project_id", ids))),
	)
	count.TimeEntries, err = s.Table("time_entries").Where(entriesCond).Count()
	if err != nil {
		return nil, err
	}
	authors := []int64{}
	err = s.Table("time_entries").Where(entriesCond).Distinct("user_id").Find(&authors)
	if err != nil {
		return nil, err
	}
	count.TimeEntryAuthors = int64(len(authors))

	return count, nil
}
