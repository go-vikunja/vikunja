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
	"strconv"

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/utils"
	"code.vikunja.io/api/pkg/web"
	"xorm.io/xorm"
)

func init() {
	// Wire dependency inversion for backward compatibility
	models.SubscriptionCreateFunc = func(s *xorm.Session, sub *models.Subscription, auth web.Auth) error {
		service := NewSubscriptionService(s.Engine())
		return service.Create(s, sub, auth)
	}
	models.SubscriptionDeleteFunc = func(s *xorm.Session, entityType models.SubscriptionEntityType, entityID int64, auth web.Auth) error {
		service := NewSubscriptionService(s.Engine())
		return service.Delete(s, entityType, entityID, auth)
	}
	models.SubscriptionGetForUserFunc = func(s *xorm.Session, entityType models.SubscriptionEntityType, entityID int64, auth web.Auth) (*models.SubscriptionWithUser, error) {
		service := NewSubscriptionService(s.Engine())
		return service.GetForUser(s, entityType, entityID, auth)
	}
	models.SubscriptionGetForEntitiesFunc = func(s *xorm.Session, entityType models.SubscriptionEntityType, entityIDs []int64) (map[int64][]*models.SubscriptionWithUser, error) {
		service := NewSubscriptionService(s.Engine())
		return service.GetForEntities(s, entityType, entityIDs)
	}
	models.SubscriptionGetForEntitiesAndUserFunc = func(s *xorm.Session, entityType models.SubscriptionEntityType, entityIDs []int64, u *user.User) (map[int64][]*models.SubscriptionWithUser, error) {
		service := NewSubscriptionService(s.Engine())
		return service.GetForEntitiesAndUser(s, entityType, entityIDs, u)
	}
	models.SubscriptionGetForEntityFunc = func(s *xorm.Session, entityType models.SubscriptionEntityType, entityID int64) ([]*models.SubscriptionWithUser, error) {
		service := NewSubscriptionService(s.Engine())
		return service.GetForEntity(s, entityType, entityID)
	}
}

// SubscriptionService handles all business logic for subscription functionality
type SubscriptionService struct {
	DB *xorm.Engine
}

// NewSubscriptionService creates a new instance of SubscriptionService
func NewSubscriptionService(engine *xorm.Engine) *SubscriptionService {
	return &SubscriptionService{
		DB: engine,
	}
}

// Create subscribes the current user to an entity
func (ss *SubscriptionService) Create(s *xorm.Session, sub *models.Subscription, auth web.Auth) error {
	// Validate entity type
	if err := sub.EntityType.Validate(); err != nil {
		return err
	}

	// Check permissions
	canCreate, err := ss.canCreate(s, sub, auth)
	if err != nil {
		return err
	}
	if !canCreate {
		return models.ErrGenericForbidden{}
	}

	// Reset ID to ensure new record
	sub.ID = 0
	sub.UserID = auth.GetID()

	// Check if subscription already exists
	existingSub, err := ss.GetForUser(s, sub.EntityType, sub.EntityID, auth)
	if err != nil {
		return err
	}
	if existingSub != nil {
		return &models.ErrSubscriptionAlreadyExists{
			EntityID:   existingSub.EntityID,
			EntityType: existingSub.EntityType,
			UserID:     existingSub.UserID,
		}
	}

	// Insert into database
	_, err = s.Insert(sub)
	return err
}

// Delete unsubscribes the current user from an entity
func (ss *SubscriptionService) Delete(s *xorm.Session, entityType models.SubscriptionEntityType, entityID int64, auth web.Auth) error {
	// Validate entity type
	if err := entityType.Validate(); err != nil {
		return err
	}

	// Check permissions
	canDelete, err := ss.canDelete(s, entityType, entityID, auth)
	if err != nil {
		return err
	}
	if !canDelete {
		return models.ErrGenericForbidden{}
	}

	// Delete the subscription
	_, err = s.
		Where("entity_id = ? AND entity_type = ? AND user_id = ?", entityID, entityType, auth.GetID()).
		Delete(&models.Subscription{})
	return err
}

// GetForUser returns a subscription for a specific entity and user
func (ss *SubscriptionService) GetForUser(s *xorm.Session, entityType models.SubscriptionEntityType, entityID int64, auth web.Auth) (*models.SubscriptionWithUser, error) {
	u, is := auth.(*user.User)
	if !is || u == nil {
		return nil, nil
	}

	subs, err := ss.GetForEntitiesAndUser(s, entityType, []int64{entityID}, u)
	if err != nil || len(subs) == 0 || len(subs[entityID]) == 0 {
		return nil, err
	}

	return subs[entityID][0], nil
}

// GetForEntities returns subscriptions for multiple entities
func (ss *SubscriptionService) GetForEntities(s *xorm.Session, entityType models.SubscriptionEntityType, entityIDs []int64) (map[int64][]*models.SubscriptionWithUser, error) {
	return ss.getForEntitiesAndUser(s, entityType, entityIDs, nil, false)
}

// GetForEntitiesAndUser returns subscriptions for multiple entities filtered by user
func (ss *SubscriptionService) GetForEntitiesAndUser(s *xorm.Session, entityType models.SubscriptionEntityType, entityIDs []int64, u *user.User) (map[int64][]*models.SubscriptionWithUser, error) {
	return ss.getForEntitiesAndUser(s, entityType, entityIDs, u, true)
}

// GetForEntity returns subscriptions for a single entity
func (ss *SubscriptionService) GetForEntity(s *xorm.Session, entityType models.SubscriptionEntityType, entityID int64) ([]*models.SubscriptionWithUser, error) {
	subs, err := ss.GetForEntities(s, entityType, []int64{entityID})
	if err != nil || len(subs[entityID]) == 0 {
		return nil, err
	}

	return subs[entityID], nil
}

// getForEntitiesAndUser is the core method that returns subscriptions with support for inheritance
// It handles both project and task subscriptions, including parent project inheritance
func (ss *SubscriptionService) getForEntitiesAndUser(s *xorm.Session, entityType models.SubscriptionEntityType, entityIDs []int64, u *user.User, userOnly bool) (subscriptions map[int64][]*models.SubscriptionWithUser, err error) {
	// Validate entity type
	if err := entityType.Validate(); err != nil {
		return nil, err
	}

	if userOnly && u == nil {
		return nil, &models.ErrMustProvideUser{}
	}

	rawSubscriptions := []*models.SubscriptionResolved{}
	entityIDString := utils.JoinInt64Slice(entityIDs, ", ")

	var sUserCond string
	if userOnly {
		sUserCond = " AND s.user_id = " + strconv.FormatInt(u.ID, 10)
	}

	switch entityType {
	case models.SubscriptionEntityProject:
		err = s.SQL(`
WITH RECURSIVE project_hierarchy AS (
    -- Base case: Start with the specified projects
    SELECT
        id,
        parent_project_id,
        0 AS level,
        id AS original_project_id
    FROM projects
    WHERE id IN (`+entityIDString+`)

    UNION ALL

    -- Recursive case: Get parent projects
    SELECT
        p.id,
        p.parent_project_id,
        ph.level + 1,
        ph.original_project_id
    FROM projects p
             INNER JOIN project_hierarchy ph ON p.id = ph.parent_project_id
),

subscription_hierarchy AS (
    -- Check for project subscriptions (including parent projects)
    SELECT
        s.id,
        s.entity_type,
        s.entity_id,
        s.created,
        s.user_id,
        CASE
            WHEN s.entity_id = ph.original_project_id THEN 1  -- Direct project match
            ELSE ph.level + 1  -- Parent projects
            END AS priority,
        ph.original_project_id
    FROM subscriptions s
             INNER JOIN project_hierarchy ph ON s.entity_id = ph.id
    WHERE s.entity_type = ?`+sUserCond+`
)

SELECT
    p.id AS original_entity_id,
    sh.id AS subscription_id,
    sh.entity_type,
    sh.entity_id,
    sh.created,
    sh.user_id,
    CASE
        WHEN sh.priority = 1 THEN 'Direct Project'
        ELSE 'Parent Project'
        END 
	AS subscription_level,
    users.*
FROM projects p
         LEFT JOIN (
    SELECT *,
           ROW_NUMBER() OVER (PARTITION BY original_project_id, user_id ORDER BY priority) AS rn
    FROM subscription_hierarchy
) sh ON p.id = sh.original_project_id AND sh.rn = 1
    LEFT JOIN users ON sh.user_id = users.id
WHERE p.id IN (`+entityIDString+`)
ORDER BY p.id, sh.user_id`, models.SubscriptionEntityProject).
			Find(&rawSubscriptions)

	case models.SubscriptionEntityTask:
		err = s.SQL(`
WITH RECURSIVE project_hierarchy AS (
    -- Base case: Start with the projects associated with the tasks
    SELECT
        p.id,
        p.parent_project_id,
        0 AS level,
        t.id AS task_id
    FROM tasks t
             JOIN projects p ON t.project_id = p.id
    WHERE t.id IN (`+entityIDString+`)

    UNION ALL

    -- Recursive case: Get parent projects
    SELECT
        p.id,
        p.parent_project_id,
        ph.level + 1,
        ph.task_id
    FROM projects p
             INNER JOIN project_hierarchy ph ON p.id = ph.parent_project_id
),

subscription_hierarchy AS (
    -- Check for task subscriptions
    SELECT
        s.id,
        s.entity_type,
        s.entity_id,
        s.created,
        s.user_id,
        1 AS priority,
        t.id AS task_id
    FROM subscriptions s
             JOIN tasks t ON s.entity_id = t.id
    WHERE s.entity_type = ? AND t.id IN (`+entityIDString+`)`+sUserCond+`

    UNION ALL

    -- Check for project subscriptions (including parent projects)
    SELECT
        s.id,
        s.entity_type,
        s.entity_id,
        s.created,
        s.user_id,
        ph.level + 2 AS priority,
        ph.task_id
    FROM subscriptions s
             INNER JOIN project_hierarchy ph ON s.entity_id = ph.id
    WHERE s.entity_type = ?`+sUserCond+`
)

SELECT
    t.id AS original_entity_id,
    sh.id AS subscription_id,
    sh.entity_type,
    sh.entity_id,
    sh.created,
    sh.user_id,
    CASE
        WHEN sh.entity_type = ? THEN 'Task'
        WHEN sh.priority = ? THEN 'Direct Project'
        ELSE 'Parent Project'
    END
    AS subscription_level,
	users.*
FROM tasks t
    LEFT JOIN (
    SELECT *,
           ROW_NUMBER() OVER (PARTITION BY task_id, user_id ORDER BY priority) AS rn
    FROM subscription_hierarchy
) sh ON t.id = sh.task_id AND sh.rn = 1
    LEFT JOIN users ON sh.user_id = users.id
WHERE t.id IN (`+entityIDString+`)
ORDER BY t.id, sh.user_id`,
			models.SubscriptionEntityTask, models.SubscriptionEntityProject, models.SubscriptionEntityTask, models.SubscriptionEntityProject).
			Find(&rawSubscriptions)
	}

	if err != nil {
		return nil, err
	}

	// Process results into map structure
	subscriptions = make(map[int64][]*models.SubscriptionWithUser)
	for _, sub := range rawSubscriptions {
		if sub.EntityID == 0 {
			continue
		}

		_, has := subscriptions[sub.OriginalEntityID]
		if !has {
			subscriptions[sub.OriginalEntityID] = []*models.SubscriptionWithUser{}
		}

		sub.ID = sub.SubscriptionID
		if sub.User != nil {
			sub.User.ID = sub.UserID
		}

		subscriptions[sub.OriginalEntityID] = append(subscriptions[sub.OriginalEntityID], &sub.SubscriptionWithUser)
	}

	return subscriptions, nil
}

// canCreate checks if a user can subscribe to an entity
func (ss *SubscriptionService) canCreate(s *xorm.Session, sub *models.Subscription, auth web.Auth) (bool, error) {
	// Link shares cannot subscribe
	if _, is := auth.(*models.LinkSharing); is {
		return false, models.ErrGenericForbidden{}
	}

	switch sub.EntityType {
	case models.SubscriptionEntityProject:
		project := &models.Project{ID: sub.EntityID}
		can, _, err := project.CanRead(s, auth)
		return can, err
	case models.SubscriptionEntityTask:
		task := &models.Task{ID: sub.EntityID}
		can, _, err := task.CanRead(s, auth)
		return can, err
	default:
		return false, &models.ErrUnknownSubscriptionEntityType{EntityType: sub.EntityType}
	}
}

// canDelete checks if a user can delete a subscription
func (ss *SubscriptionService) canDelete(s *xorm.Session, entityType models.SubscriptionEntityType, entityID int64, auth web.Auth) (bool, error) {
	// Link shares cannot unsubscribe
	if _, is := auth.(*models.LinkSharing); is {
		return false, models.ErrGenericForbidden{}
	}

	// Check if subscription exists for this user
	realSub := &models.Subscription{}
	exists, err := s.
		Where("entity_id = ? AND entity_type = ? AND user_id = ?", entityID, entityType, auth.GetID()).
		Get(realSub)
	if err != nil {
		return false, err
	}

	return exists, nil
}
