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
	DB *xorm.Engine
}

// NewReactionsService creates a new ReactionsService.
func NewReactionsService(db *xorm.Engine) *ReactionsService {
	return &ReactionsService{
		DB: db,
	}
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

// CreateReaction creates a new reaction for a task.
func (rs *ReactionsService) CreateReaction(s *xorm.Session, reaction *models.Reaction, a web.Auth) error {
	// Set the user who created the reaction
	reaction.UserID = a.GetID()
	reaction.EntityKind = models.ReactionKindTask

	// Insert the reaction
	_, err := s.Insert(reaction)
	return err
}

// DeleteReaction removes a reaction from a task.
func (rs *ReactionsService) DeleteReaction(s *xorm.Session, entityID int64, userID int64, value string) error {
	_, err := s.Where("entity_id = ? AND user_id = ? AND value = ? AND entity_kind = ?",
		entityID, userID, value, models.ReactionKindTask).Delete(&models.Reaction{})
	return err
}

// GetReactionsByTask retrieves all reactions for a specific task.
func (rs *ReactionsService) GetReactionsByTask(s *xorm.Session, taskID int64) ([]*models.Reaction, error) {
	reactions := []*models.Reaction{}
	err := s.Where("entity_id = ? AND entity_kind = ?", taskID, models.ReactionKindTask).Find(&reactions)
	return reactions, err
}
