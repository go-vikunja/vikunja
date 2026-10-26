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
	"context"
	"slices"
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

// Includes the project itself.
func subtreeIDs(s *xorm.Session, projectID int64, cond builder.Cond) (ids []int64, err error) {
	ids = []int64{}
	err = s.Table("project_ancestors").
		Join("INNER", "projects", "projects.id = project_ancestors.project_id").
		Where(builder.And(
			builder.Eq{"project_ancestors.ancestor_id": projectID},
			cond,
		)).
		Cols("project_ancestors.project_id").
		Find(&ids)
	return
}

func trashProject(s *xorm.Session, project *Project, a web.Auth) error {
	ids, err := subtreeIDs(s, project.ID, builder.IsNull{"projects.deleted_at"})
	if err != nil {
		return err
	}

	// Owners may trash their own default project, nobody else's.
	blockedID := int64(0)
	has, err := s.Table("users").
		Join("INNER", "projects", "projects.id = users.default_project_id AND projects.owner_id = users.id").
		Where(builder.And(
			builder.In("projects.id", ids),
			builder.Neq{"projects.owner_id": a.GetID()},
		)).
		Cols("projects.id").
		Get(&blockedID)
	if err != nil {
		return err
	}
	if has {
		return &ErrCannotDeleteDefaultProject{ProjectID: blockedID}
	}

	projects := []*Project{}
	if err = s.In("id", ids).Find(&projects); err != nil {
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

	for _, p := range projects {
		events.DispatchOnCommit(s, &ProjectDeletedEvent{
			Project: p,
			Doer:    doer,
		})
	}
	return nil
}

// Time entries survive without a container, kept by their authors. A nil doer is the system.
func hardDeleteProject(s *xorm.Session, p *Project, doer *user.User) error {
	ancestorIDs := []int64{}
	err := s.Table(&ProjectAncestor{}).
		Where(builder.And(
			builder.Eq{"project_id": p.ID},
			builder.Gt{"depth": 0},
		)).
		Cols("ancestor_id").
		Find(&ancestorIDs)
	if err != nil {
		return err
	}
	return hardDeleteProjectTree(s, p, doer, ancestorIDs)
}

// Children get their ancestors passed down: deleting p's closure rows also drops their links to p.
func hardDeleteProjectTree(s *xorm.Session, p *Project, doer *user.User, ancestorIDs []int64) (err error) {
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
		Update(map[string]any{
			"project_id": 0,
			"task_id":    0,
		})
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
	events.DispatchOnCommit(s, &ProjectPurgedEvent{
		Project:     p,
		Doer:        doer,
		AncestorIDs: ancestorIDs,
	})

	children := []*Project{}
	err = s.Unscoped().Where("parent_project_id = ?", p.ID).Find(&children)
	if err != nil {
		return
	}
	childAncestorIDs := append(slices.Clone(ancestorIDs), p.ID)
	for _, child := range children {
		err = hardDeleteProjectTree(s, child, doer, childAncestorIDs)
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

// Unreachable trashed projects are ErrProjectDoesNotExist so their existence doesn't leak.
// Like regular reads, link shares never reach a trashed project.
func getReadableTrashedProject(s *xorm.Session, a web.Auth, projectID int64) (*Project, Permission, error) {
	if _, isShare := a.(*LinkSharing); isShare {
		return nil, 0, ErrProjectDoesNotExist{ID: projectID}
	}
	project, err := getTrashedProject(s, projectID)
	if err != nil {
		return nil, 0, err
	}
	if isInstanceAdmin(s, a) || project.OwnerID == a.GetID() {
		return project, PermissionAdmin, nil
	}
	access, err := getProjectAccessIncludingTrash(s, a.GetID())
	if err != nil {
		return nil, 0, err
	}
	permission, reachable := access.permission(project.ID)
	if !reachable {
		return nil, 0, ErrProjectDoesNotExist{ID: projectID}
	}
	return project, permission, nil
}

// Readers who could reach the project get ErrProjectIsInTrash; everyone else keeps notFound.
func trashedProjectReadError(s *xorm.Session, a web.Auth, projectID int64, notFound error) error {
	trashed, permission, err := getReadableTrashedProject(s, a, projectID)
	if IsErrProjectDoesNotExist(err) {
		return notFound
	}
	if err != nil {
		return err
	}

	inTrash := ErrProjectIsInTrash{
		ID:        trashed.ID,
		DeletedAt: trashed.DeletedAt,
	}
	inTrash.CanRestore, err = canRestoreTrashedProject(s, trashed, permission)
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
	inTrash.TrashedRootID, inTrash.CanRestoreRoot, err = readableTrashedRoot(s, a, trashed)
	if err != nil {
		return err
	}
	return inTrash
}

func canRestoreTrashedProject(s *xorm.Session, p *Project, permission Permission) (bool, error) {
	if permission != PermissionAdmin {
		return false, nil
	}
	blocker, err := restoreBlocker(s, p)
	return blocker == nil, err
}

// The topmost trashed ancestor of p if the caller can reach it (else 0), and whether the caller can restore it.
func readableTrashedRoot(s *xorm.Session, a web.Auth, p *Project) (rootID int64, canRestore bool, err error) {
	trashedAncestors := []*ProjectAncestor{}
	err = s.Table("project_ancestors").
		Join("INNER", "projects", "projects.id = project_ancestors.ancestor_id").
		Where(builder.And(
			builder.Eq{"project_ancestors.project_id": p.ID},
			builder.Gt{"project_ancestors.depth": 0},
			builder.NotNull{"projects.deleted_at"},
		)).
		Cols("project_ancestors.ancestor_id", "project_ancestors.depth").
		OrderBy("project_ancestors.depth").
		Find(&trashedAncestors)
	if err != nil {
		return 0, false, err
	}
	// Stop at the first live ancestor: restoring the root below it brings p back.
	for i, ancestor := range trashedAncestors {
		if ancestor.Depth != int64(i+1) {
			break
		}
		rootID = ancestor.AncestorID
	}
	if rootID == 0 {
		return 0, false, nil
	}
	root, permission, err := getReadableTrashedProject(s, a, rootID)
	if IsErrProjectDoesNotExist(err) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	canRestore, err = canRestoreTrashedProject(s, root, permission)
	return rootID, canRestore, err
}

// Loads the trashed project into p.
func (p *Project) canDoTrashedProject(s *xorm.Session, a web.Auth) (bool, error) {
	trashed, permission, err := getReadableTrashedProject(s, a, p.ID)
	if err != nil {
		return false, err
	}
	*p = *trashed
	return permission == PermissionAdmin, nil
}

// CanRestore requires admin access to the trashed project.
func (p *Project) CanRestore(s *xorm.Session, a web.Auth) (bool, error) {
	return p.canDoTrashedProject(s, a)
}

func restoreBlocker(s *xorm.Session, p *Project) (blocker error, err error) {
	parentID := p.parentID()
	if parentID == 0 {
		return nil, nil
	}
	parent := &Project{}
	has, err := s.Unscoped().ID(parentID).Get(parent)
	if err != nil {
		return nil, err
	}
	// Purging takes all sub-projects along and the orphan repair re-parents the rest.
	if !has {
		return ErrParentProjectDoesNotExist{
			ProjectID:       p.ID,
			ParentProjectID: parentID,
		}, nil
	}
	switch {
	case !parent.DeletedAt.IsZero():
		return ErrParentProjectIsInTrash{
			ProjectID:       p.ID,
			ParentProjectID: parentID,
		}, nil
	case parent.IsArchived:
		return ErrParentProjectIsArchived{
			ProjectID:       p.ID,
			ParentProjectID: parentID,
		}, nil
	}
	return nil, nil
}

// Restore brings back all trashed sub-projects, whenever they were trashed; trashed tasks stay.
func (p *Project) Restore(s *xorm.Session, a web.Auth) error {
	blocker, err := restoreBlocker(s, p)
	if err != nil {
		return err
	}
	if blocker != nil {
		return blocker
	}

	ids, err := subtreeIDs(s, p.ID, builder.NotNull{"projects.deleted_at"})
	if err != nil {
		return err
	}
	_, err = s.Table("projects").
		In("id", ids).
		Update(map[string]any{
			"deleted_at":    nil,
			"deleted_by_id": nil,
		})
	if err != nil {
		return err
	}

	_, err = s.In("id", ids).Cols("updated").Update(&Project{})
	if err != nil {
		return err
	}
	// CalDAV clients dropped the collections and only pick the tasks up again through this
	_, err = s.In("project_id", ids).Cols("updated").Update(&Task{})
	if err != nil {
		return err
	}

	restored := []*Project{}
	if err = s.In("id", ids).Find(&restored); err != nil {
		return err
	}
	doer := doerFromAuth(s, a)
	for _, project := range restored {
		events.DispatchOnCommit(s, &ProjectRestoredEvent{
			Project: project,
			Doer:    doer,
		})
	}
	return nil
}

// CanDeletePermanently requires admin access to the trashed project.
func (p *Project) CanDeletePermanently(s *xorm.Session, a web.Auth) (bool, error) {
	return p.canDoTrashedProject(s, a)
}

// DeletePermanently purges the project with all its sub-projects.
func (p *Project) DeletePermanently(s *xorm.Session, a web.Auth) error {
	return hardDeleteProject(s, p, doerFromAuth(s, a))
}

// TrashedSubProject is a trashed descendant shown under its topmost trashed ancestor in the trash.
type TrashedSubProject struct {
	ID              int64  `xorm:"id" json:"id" doc:"The id of the sub-project."`
	Title           string `xorm:"title" json:"title" doc:"The title of the sub-project."`
	ParentProjectID int64  `xorm:"parent_project_id" json:"parent_project_id" doc:"The id of its parent, either the listed project or another of its sub-projects."`
	Depth           int64  `xorm:"depth" json:"depth" doc:"How many levels below the listed project it sits, starting at 1."`
}

// TrashedProject is a topmost trashed project in the trash listing.
type TrashedProject struct {
	Project
	TrashedSubProjects []*TrashedSubProject `json:"trashed_sub_projects,omitempty" readOnly:"true" doc:"The trashed sub-projects at any depth that restoring this project brings back, shallowest first."`
}

type trashedSubProjectRow struct {
	AncestorID        int64 `xorm:"ancestor_id"`
	TrashedSubProject `xorm:"extends"`
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
	// Only trashed ids: the access map holds every live project too, which would bloat the IN list.
	trashedAdminIDs := []int64{}
	for chunk := range slices.Chunk(adminIDs, maxInListSize) {
		batch := []int64{}
		err = s.Table("projects").
			Where(builder.And(
				builder.In("id", chunk),
				builder.NotNull{"deleted_at"},
			)).
			Cols("id").
			Find(&batch)
		if err != nil {
			return nil, err
		}
		trashedAdminIDs = append(trashedAdminIDs, batch...)
	}

	return builder.And(
		builder.NotNull{"projects.deleted_at"},
		builder.In("projects.id", trashedAdminIDs),
		builder.Or(
			builder.IsNull{"projects.parent_project_id"},
			builder.NotIn("projects.parent_project_id",
				builder.Select("id").From("projects").Where(builder.NotNull{"deleted_at"}),
			),
		),
	), nil
}

// GetTrashedProjects returns the project trash of the auth, newest first.
func GetTrashedProjects(s *xorm.Session, a web.Auth, page, perPage int) (trashed []*TrashedProject, total int64, err error) {
	cond, err := trashedProjectsCond(s, a)
	if err != nil {
		return nil, 0, err
	}

	total, err = s.Unscoped().Where(cond).Count(&Project{})
	if err != nil {
		return nil, 0, err
	}

	projects := []*Project{}
	err = s.Unscoped().
		Where(cond).
		OrderBy("projects.deleted_at DESC, projects.id DESC").
		Limit(getLimitFromPageIndex(page, perPage)).
		Find(&projects)
	if err != nil {
		return nil, 0, err
	}
	trashed = make([]*TrashedProject, 0, len(projects))
	if len(projects) == 0 {
		return trashed, total, nil
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

	subProjects := []*trashedSubProjectRow{}
	err = s.Table("project_ancestors").
		Select("project_ancestors.ancestor_id, project_ancestors.depth, projects.id, projects.title, "+
			"COALESCE(projects.parent_project_id, 0) AS parent_project_id").
		Join("INNER", "projects", "projects.id = project_ancestors.project_id").
		Where(builder.And(
			builder.In("project_ancestors.ancestor_id", projectIDs),
			builder.Gt{"project_ancestors.depth": 0},
			builder.NotNull{"projects.deleted_at"},
		)).
		OrderBy("project_ancestors.depth, projects.title, projects.id").
		Find(&subProjects)
	if err != nil {
		return nil, 0, err
	}
	subProjectsByRoot := make(map[int64][]*TrashedSubProject, len(projects))
	for _, sub := range subProjects {
		subProjectsByRoot[sub.AncestorID] = append(subProjectsByRoot[sub.AncestorID], &sub.TrashedSubProject)
	}

	users, err := getUsersOrLinkSharesFromIDs(s, userIDs)
	if err != nil {
		return nil, 0, err
	}
	for _, project := range projects {
		project.Owner = users[project.OwnerID]
		project.DeletedBy = users[project.DeletedByID]
		trashed = append(trashed, &TrashedProject{
			Project:            *project,
			TrashedSubProjects: subProjectsByRoot[project.ID],
		})
	}

	return trashed, total, nil
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
	doer := doerFromAuth(s, a)
	for _, project := range projects {
		if err = hardDeleteProject(s, project, doer); err != nil {
			return 0, err
		}
	}
	return len(projects), nil
}

// RegisterProjectCleanupCron purges projects trashed more than TrashRetention ago.
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
			defer events.CleanupPending(ps)

			// An earlier iteration may have purged it with its parent
			project, err := getTrashedProject(ps, projectID)
			if IsErrProjectDoesNotExist(err) {
				return
			}
			if err == nil {
				err = hardDeleteProject(ps, project, nil)
			}
			if err != nil {
				_ = ps.Rollback()
				log.Errorf("Could not permanently delete project %d: %s", projectID, err)
				return
			}
			if err = ps.Commit(); err != nil {
				log.Errorf("Could not commit transaction: %s", err)
				return
			}
			events.DispatchPending(context.Background(), ps)
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
	// An admin share would see totals of sub-projects it can't reach.
	if _, isShare := a.(*LinkSharing); isShare {
		return nil, ErrGenericForbidden{}
	}

	project := &Project{ID: projectID}
	can, err := project.IsAdmin(s, a)
	if IsErrProjectDoesNotExist(err) {
		return nil, trashedProjectReadError(s, a, projectID, err)
	}
	if err != nil {
		return nil, err
	}
	if !can {
		return nil, ErrGenericForbidden{}
	}

	ids, err := subtreeIDs(s, projectID, builder.IsNull{"projects.deleted_at"})
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
		builder.In("task_id", builder.Select("id").From("tasks").Where(builder.And(
			builder.In("project_id", ids),
			builder.IsNull{"deleted_at"},
		))),
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
