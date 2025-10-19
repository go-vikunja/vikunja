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
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"
	"xorm.io/builder"
	"xorm.io/xorm"
)

// ReactionsService represents a service for managing task reactions.
type ReactionsService struct {
	DB       *xorm.Engine
	Registry *ServiceRegistry
}

// NewReactionsService creates a new ReactionsService.
// Deprecated: Use ServiceRegistry.Reactions() instead.
func NewReactionsService(db *xorm.Engine) *ReactionsService {
	registry := NewServiceRegistry(db)
	return registry.Reactions()
}

// AddReactionsToTasks adds reaction data to a map of tasks.
// This method implements proper service layer separation for task expansion.
func (rs *ReactionsService) AddReactionsToTasks(s *xorm.Session, taskIDs []int64, taskMap map[int64]*models.Task) error {
	if len(taskIDs) == 0 {
		return nil
	}

	// Use the same logic as models.getReactionsForEntityIDs but adapted for service layer
	where := builder.And(
		builder.Eq{"entity_kind": models.ReactionKindTask},
		builder.In("entity_id", taskIDs),
	)

	reactions := []*models.Reaction{}
	err := s.Where(where).Find(&reactions)
	if err != nil {
		return err
	}

	if len(reactions) == 0 {
		// Leave task.Reactions as nil (zero value) for tasks without reactions
		// This ensures JSON serialization shows "reactions": null instead of "reactions": {}
		return nil
	}

	// Get all users who made these reactions
	cond := builder.
		Select("user_id").
		From("reactions").
		Where(where)

	users, err := user.GetUsersByCond(s, builder.In("id", cond))
	if err != nil {
		return err
	}

	// Build reaction maps by task ID
	reactionsWithTasks := make(map[int64]models.ReactionMap)
	for _, reaction := range reactions {
		if _, taskExists := reactionsWithTasks[reaction.EntityID]; !taskExists {
			reactionsWithTasks[reaction.EntityID] = make(models.ReactionMap)
		}

		if _, has := reactionsWithTasks[reaction.EntityID][reaction.Value]; !has {
			reactionsWithTasks[reaction.EntityID][reaction.Value] = []*user.User{}
		}

		reactionsWithTasks[reaction.EntityID][reaction.Value] = append(reactionsWithTasks[reaction.EntityID][reaction.Value], users[reaction.UserID])
	}

	// Assign reactions to tasks
	// Only set Reactions for tasks that actually have reactions
	// Leave task.Reactions as nil (zero value) for tasks without reactions
	for taskID, task := range taskMap {
		if reactions, exists := reactionsWithTasks[taskID]; exists {
			task.Reactions = reactions
		}
		// Don't set task.Reactions for tasks without reactions (leave as nil)
	}

	return nil
}

// Create creates a new reaction for an entity (task or comment).
// This method is idempotent - creating a duplicate reaction will not result in an error.
func (rs *ReactionsService) Create(s *xorm.Session, reaction *models.Reaction, a web.Auth) error {
	// Set the user who created the reaction
	reaction.UserID = a.GetID()

	// Check if reaction already exists (idempotent behavior)
	exists, err := s.Where("user_id = ? AND entity_id = ? AND entity_kind = ? AND value = ?",
		reaction.UserID, reaction.EntityID, reaction.EntityKind, reaction.Value).
		Exist(&models.Reaction{})
	if err != nil {
		return err
	}

	if exists {
		return nil // Already exists, treat as success (idempotent)
	}

	// Reset ID to ensure new record
	reaction.ID = 0

	// Insert the reaction
	_, err = s.Insert(reaction)
	return err
}

// Delete removes a reaction from an entity.
// Only the user who created the reaction can delete it.
func (rs *ReactionsService) Delete(s *xorm.Session, entityID int64, userID int64, value string, entityKind models.ReactionKind) error {
	_, err := s.Where("user_id = ? AND entity_id = ? AND entity_kind = ? AND value = ?",
		userID, entityID, entityKind, value).
		Delete(&models.Reaction{})
	return err
}

// GetAll retrieves all reactions for a specific entity, grouped by reaction value.
// Returns a ReactionMap where keys are reaction values (emoji) and values are lists of users who reacted.
func (rs *ReactionsService) GetAll(s *xorm.Session, entityID int64, entityKind models.ReactionKind) (models.ReactionMap, error) {
	where := builder.And(
		builder.Eq{"entity_kind": entityKind},
		builder.Eq{"entity_id": entityID},
	)

	reactions := []*models.Reaction{}
	err := s.Where(where).Find(&reactions)
	if err != nil {
		return nil, err
	}

	if len(reactions) == 0 {
		return models.ReactionMap{}, nil
	}

	// Get all users who made these reactions
	cond := builder.
		Select("user_id").
		From("reactions").
		Where(where)

	users, err := user.GetUsersByCond(s, builder.In("id", cond))
	if err != nil {
		return nil, err
	}

	// Build reaction map
	reactionMap := make(models.ReactionMap)
	for _, reaction := range reactions {
		if _, has := reactionMap[reaction.Value]; !has {
			reactionMap[reaction.Value] = []*user.User{}
		}
		reactionMap[reaction.Value] = append(reactionMap[reaction.Value], users[reaction.UserID])
	}

	return reactionMap, nil
}

// CanRead checks if a user can read a reaction
func (rs *ReactionsService) CanRead(s *xorm.Session, entityID int64, entityKind models.ReactionKind, a web.Auth) (bool, int, error) {
	task, err := rs.getTaskForReaction(s, entityID, entityKind)
	if err != nil {
		return false, 0, err
	}
	// Delegate to task permissions
	return task.CanRead(s, a)
}

// CanCreate checks if a user can create a reaction on an entity
func (rs *ReactionsService) CanCreate(s *xorm.Session, entityID int64, entityKind models.ReactionKind, a web.Auth) (bool, error) {
	task, err := rs.getTaskForReaction(s, entityID, entityKind)
	if err != nil {
		return false, err
	}
	// Delegate to task permissions (need update permission to add reactions)
	return task.CanUpdate(s, a)
}

// CanDelete checks if a user can delete a reaction
func (rs *ReactionsService) CanDelete(s *xorm.Session, entityID int64, entityKind models.ReactionKind, a web.Auth) (bool, error) {
	task, err := rs.getTaskForReaction(s, entityID, entityKind)
	if err != nil {
		return false, err
	}
	// Delegate to task permissions (need update permission to delete reactions)
	return task.CanUpdate(s, a)
}

// getTaskForReaction retrieves the task associated with a reaction (either directly or via comment)
func (rs *ReactionsService) getTaskForReaction(s *xorm.Session, entityID int64, entityKind models.ReactionKind) (*models.Task, error) {
	task := &models.Task{ID: entityID}

	if entityKind == models.ReactionKindComment {
		// Get the comment to find the associated task
		// We need to query the comment directly since GetByID requires a user
		tc := &models.TaskComment{}
		exists, err := s.Where("id = ?", entityID).Get(tc)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, models.ErrTaskCommentDoesNotExist{ID: entityID}
		}
		task.ID = tc.TaskID
	}

	return task, nil
}
