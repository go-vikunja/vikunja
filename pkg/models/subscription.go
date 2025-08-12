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
	"encoding/json"
	"strconv"
	"time"

	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/utils"
	"code.vikunja.io/api/pkg/web"

	"xorm.io/xorm"
)

// SubscriptionEntityType represents all entities which can be subscribed to
type SubscriptionEntityType int

const (
	SubscriptionEntityUnknown   = iota
	SubscriptionEntityNamespace // Kept even though not used anymore since we don't want to manually change all ids
	SubscriptionEntityProject
	SubscriptionEntityTask
)

func (st *SubscriptionEntityType) UnmarshalJSON(bytes []byte) error {
	var value string
	err := json.Unmarshal(bytes, &value)
	if err != nil {
		return err
	}

	switch value {
	case "project":
		*st = SubscriptionEntityProject
	case "task":
		*st = SubscriptionEntityTask
	default:
		return &ErrUnknownSubscriptionEntityType{EntityType: *st}
	}

	return nil
}

func (st SubscriptionEntityType) MarshalJSON() ([]byte, error) {
	switch st {
	case SubscriptionEntityProject:
		return []byte(`"project"`), nil
	case SubscriptionEntityTask:
		return []byte(`"task"`), nil
	}

	return []byte(`nil`), nil
}

func getEntityTypeFromString(entityType string) SubscriptionEntityType {
	switch entityType {
	case entityProject:
		return SubscriptionEntityProject
	case entityTask:
		return SubscriptionEntityTask
	}

	return SubscriptionEntityUnknown
}

func (st SubscriptionEntityType) validate() error {
	if st == SubscriptionEntityProject ||
		st == SubscriptionEntityTask {
		return nil
	}

	return &ErrUnknownSubscriptionEntityType{EntityType: st}
}

const (
	entityProject = `project`
	entityTask    = `task`
)

// Subscription represents a subscription for an entity
type Subscription struct {
	// The numeric ID of the subscription
	ID int64 `xorm:"autoincr not null unique pk" json:"id"`

	EntityType SubscriptionEntityType `xorm:"index not null" json:"entity"`
	Entity     string                 `xorm:"-" json:"-" param:"entity"`
	// The id of the entity to subscribe to.
	EntityID int64 `xorm:"bigint index not null" json:"entity_id" param:"entityID"`

	// The user who made this subscription
	UserID int64 `xorm:"bigint index not null" json:"-"`

	// A timestamp when this subscription was created. You cannot change this value.
	Created time.Time `xorm:"created not null" json:"created"`

	web.CRUDable    `xorm:"-" json:"-"`
	web.Permissions `xorm:"-" json:"-"`
}

type SubscriptionWithUser struct {
	Subscription `xorm:"extends"`
	User         *user.User `xorm:"extends" json:"user"`
}

type subscriptionResolved struct {
	OriginalEntityID     int64
	SubscriptionID       int64
	SubscriptionWithUser `xorm:"extends"`
}

// TableName gives us a better table name for the subscriptions table
func (sb *Subscription) TableName() string {
	return "subscriptions"
}

// Create subscribes the current user to an entity
// @Summary Subscribes the current user to an entity.
// @Description Subscribes the current user to an entity.
// @tags subscriptions
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param entity path string true "The entity the user subscribes to. Can be either `project` or `task`."
// @Param entityID path string true "The numeric id of the entity to subscribe to."
// @Success 201 {object} models.Subscription "The subscription"
// @Failure 403 {object} web.HTTPError "The user does not have access to subscribe to this entity."
// @Failure 412 {object} web.HTTPError "The subscription already exists."
// @Failure 412 {object} web.HTTPError "The subscription entity is invalid."
// @Failure 500 {object} models.Message "Internal error"
// @Router /subscriptions/{entity}/{entityID} [put]
func (sb *Subscription) Create(s *xorm.Session, auth web.Auth) (err error) {
	// Permissions method already does the validation of the entity type, so we don't need to do that here

	sb.ID = 0
	sb.UserID = auth.GetID()

	sub, err := GetSubscriptionForUser(s, sb.EntityType, sb.EntityID, auth)
	if err != nil {
		return err
	}
	if sub != nil {
		return &ErrSubscriptionAlreadyExists{
			EntityID:   sub.EntityID,
			EntityType: sub.EntityType,
			UserID:     sub.UserID,
		}
	}

	_, err = s.Insert(sb)
	return
}

// Delete unsubscribes the current user to an entity
// @Summary Unsubscribe the current user from an entity.
// @Description Unsubscribes the current user to an entity.
// @tags subscriptions
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param entity path string true "The entity the user subscribed to. Can be either `project` or `task`."
// @Param entityID path string true "The numeric id of the subscribed entity to."
// @Success 200 {object} models.Subscription "The subscription"
// @Failure 403 {object} web.HTTPError "The user does not have access to subscribe to this entity."
// @Failure 404 {object} web.HTTPError "The subscription does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /subscriptions/{entity}/{entityID} [delete]
func (sb *Subscription) Delete(s *xorm.Session, auth web.Auth) (err error) {
	sb.UserID = auth.GetID()

	_, err = s.
		Where("entity_id = ? AND entity_type = ? AND user_id = ?", sb.EntityID, sb.EntityType, sb.UserID).
		Delete(&Subscription{})
	return
}

func GetSubscriptionForUser(s *xorm.Session, entityType SubscriptionEntityType, entityID int64, a web.Auth) (subscription *SubscriptionWithUser, err error) {
	u, is := a.(*user.User)
	if !is || u == nil {
		return
	}

	subs, err := GetSubscriptionsForEntitiesAndUser(s, entityType, []int64{entityID}, u)
	if err != nil || len(subs) == 0 || len(subs[entityID]) == 0 {
		return nil, err
	}

	return subs[entityID][0], nil
}

// GetSubscriptionsForEntities returns a list of subscriptions to for an entity ID
func GetSubscriptionsForEntities(s *xorm.Session, entityType SubscriptionEntityType, entityIDs []int64) (subscriptions map[int64][]*SubscriptionWithUser, err error) {
	return getSubscriptionsForEntitiesAndUser(s, entityType, entityIDs, nil, false)
}

func GetSubscriptionsForEntitiesAndUser(s *xorm.Session, entityType SubscriptionEntityType, entityIDs []int64, u *user.User) (subscriptions map[int64][]*SubscriptionWithUser, err error) {
	return getSubscriptionsForEntitiesAndUser(s, entityType, entityIDs, u, true)
}

func GetSubscriptionsForEntity(s *xorm.Session, entityType SubscriptionEntityType, entityID int64) (subscriptions []*SubscriptionWithUser, err error) {
	subs, err := GetSubscriptionsForEntities(s, entityType, []int64{entityID})
	if err != nil || len(subs[entityID]) == 0 {
		return
	}

	return subs[entityID], nil
}

// This function returns a matching subscription for an entity and user.
// It will return the next parent of a subscription. That means for tasks, it will first look for a subscription for
// that task, if there is none it will look for a subscription on the project the task belongs to.
// It will return a map where the key is the entity id and the value is a slice with all subscriptions for that entity.
func getSubscriptionsForEntitiesAndUser(s *xorm.Session, entityType SubscriptionEntityType, entityIDs []int64, u *user.User, userOnly bool) (subscriptions map[int64][]*SubscriptionWithUser, err error) {
	if err := entityType.validate(); err != nil {
		return nil, err
	}

	rawSubscriptions := []*subscriptionResolved{}
	entityIDString := utils.JoinInt64Slice(entityIDs, ", ")

	var sUserCond string
	if userOnly {
		if u == nil {
			return nil, &ErrMustProvideUser{}
		}
		sUserCond = " AND s.user_id = " + strconv.FormatInt(u.ID, 10)
	}

	switch entityType {
	case SubscriptionEntityProject:
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
ORDER BY p.id, sh.user_id`, SubscriptionEntityProject).
			Find(&rawSubscriptions)
	case SubscriptionEntityTask:
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
			SubscriptionEntityTask, SubscriptionEntityProject, SubscriptionEntityTask, SubscriptionEntityProject).
			Find(&rawSubscriptions)
	}
	if err != nil {
		return nil, err
	}

	subscriptions = make(map[int64][]*SubscriptionWithUser)
	for _, sub := range rawSubscriptions {

		if sub.EntityID == 0 {
			continue
		}

		_, has := subscriptions[sub.OriginalEntityID]
		if !has {
			subscriptions[sub.OriginalEntityID] = []*SubscriptionWithUser{}
		}

		sub.ID = sub.SubscriptionID
		if sub.User != nil {
			sub.User.ID = sub.UserID
		}

		subscriptions[sub.OriginalEntityID] = append(subscriptions[sub.OriginalEntityID], &sub.SubscriptionWithUser)
	}

	return subscriptions, nil
}
