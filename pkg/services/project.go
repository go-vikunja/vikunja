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

package services

import (
	"math"

	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/utils"
	"xorm.io/xorm"
)

// Project is a service for projects.
type Project struct {
	DB *xorm.Engine
}

// Get gets a project by its ID.
func (p *Project) Get(s *xorm.Session, projectID int64, u *user.User) (*models.Project, error) {
	return nil, nil
}

// Create creates a new project.
func (p *Project) Create(s *xorm.Session, project *models.Project, u *user.User) (*models.Project, error) {
	if project.ParentProjectID != 0 {
		// parent := &models.Project{ID: project.ParentProjectID}
		// TODO: Move this to the service
		//can, err := parent.CanWrite(s, u)
		//if err != nil {
		//	return nil, err
		//}
		//if !can {
		//	return nil, errors.New("cannot write to parent project")
		//}
	}

	project.ID = 0
	project.OwnerID = u.ID
	project.Owner = u

	err := p.validate(s, project)
	if err != nil {
		return nil, err
	}

	project.HexColor = utils.NormalizeHex(project.HexColor)

	_, err = s.Insert(project)
	if err != nil {
		return nil, err
	}

	project.Position = calculateDefaultPosition(project.ID, project.Position)
	_, err = s.Where("id = ?", project.ID).Update(project)
	if err != nil {
		return nil, err
	}
	if project.IsFavorite {
		if err := addToFavorites(s, project.ID, u, models.FavoriteKindProject); err != nil {
			return nil, err
		}
	}

	err = CreateDefaultViewsForProject(s, project, u, true, true)
	if err != nil {
		return nil, err
	}

	err = events.Dispatch(&models.ProjectCreatedEvent{
		Project: project,
		Doer:    u,
	})
	if err != nil {
		return nil, err
	}

	fullProject, err := models.GetProjectSimpleByID(s, project.ID)
	if err != nil {
		return nil, err
	}

	return fullProject, err
}

func (p *Project) validate(s *xorm.Session, project *models.Project) (err error) {
	if project.ParentProjectID < 0 {
		return &ErrProjectCannotBelongToAPseudoParentProject{ProjectID: project.ID, ParentProjectID: project.ParentProjectID}
	}

	// Check if the parent project exists
	if project.ParentProjectID > 0 {
		if project.ParentProjectID == project.ID {
			return &ErrProjectCannotBeChildOfItself{
				ProjectID: project.ID,
			}
		}

		allProjects, err := models.GetAllParentProjects(s, project.ParentProjectID)
		if err != nil {
			return err
		}

		var parent *models.Project
		parent = allProjects[project.ParentProjectID]

		// Check if there's a cycle in the parent relation
		parentsVisited := make(map[int64]bool)
		parentsVisited[project.ID] = true
		for parent.ParentProjectID != 0 {

			parent = allProjects[parent.ParentProjectID]

			if parentsVisited[parent.ID] {
				return &ErrProjectCannotHaveACyclicRelationship{
					ProjectID: project.ID,
				}
			}

			parentsVisited[parent.ID] = true
		}
	}

	// Check if the identifier is unique and not empty
	if project.Identifier != "" {
		exists, err := s.
			Where("identifier = ?", project.Identifier).
			And("id != ?", project.ID).
			Exist(&models.Project{})
		if err != nil {
			return err
		}
		if exists {
			return ErrProjectIdentifierIsNotUnique{Identifier: project.Identifier}
		}
	}

	return nil
}

func calculateDefaultPosition(id int64, position float64) float64 {
	if position < 0.1 {
		return float64(id) * math.Pow(2, 32)
	}

	return position
}

func addToFavorites(s *xorm.Session, entityID int64, u *user.User, kind models.FavoriteKind) (err error) {
	fav := &models.Favorite{
		UserID:   u.ID,
		EntityID: entityID,
		Kind:     kind,
	}
	_, err = s.Insert(fav)
	return
}

func CreateDefaultViewsForProject(s *xorm.Session, project *models.Project, u *user.User, createBacklogBucket bool, createDefaultBuckets bool) (err error) {
	_, err = s.Insert([]*models.ProjectView{
		{
			ProjectID: project.ID,
			Title:     "List",
			ViewKind:  models.ProjectViewKindList,
			Position:  100,
		},
		{
			ProjectID: project.ID,
			Title:     "Gantt",
			ViewKind:  models.ProjectViewKindGantt,
			Position:  200,
		},
		{
			ProjectID: project.ID,
			Title:     "Table",
			ViewKind:  models.ProjectViewKindTable,
			Position:  300,
		},
	})
	if err != nil {
		return
	}

	kanbanView := &models.ProjectView{
		ProjectID:             project.ID,
		Title:                 "Kanban",
		ViewKind:              models.ProjectViewKindKanban,
		Position:              400,
		BucketConfigurationMode: models.BucketConfigurationModeManual,
	}
	_, err = s.Insert(kanbanView)
	if err != nil {
		return
	}

	if !createDefaultBuckets {
		return
	}

	buckets := []*models.Bucket{}
	if createBacklogBucket {
		buckets = append(buckets, &models.Bucket{
			Title:         "Backlog",
			Position:      100,
			ProjectViewID: kanbanView.ID,
		})
	}
	buckets = append(buckets, []*models.Bucket{
		{
			Title:         "To-Do",
			Position:      200,
			ProjectViewID: kanbanView.ID,
		},
		{
			Title:         "Doing",
			Position:      300,
			ProjectViewID: kanbanView.ID,
		},
		{
			Title:         "Done",
			Position:      400,
			ProjectViewID: kanbanView.ID,
		},
	}...)

	_, err = s.Insert(buckets)
	if err != nil {
		return
	}

	kanbanView.DefaultBucketID = buckets[0].ID
	kanbanView.DoneBucketID = buckets[len(buckets)-1].ID
	_, err = s.ID(kanbanView.ID).Cols("default_bucket_id", "done_bucket_id").Update(kanbanView)
	return
}
