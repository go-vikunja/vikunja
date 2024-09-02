// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
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

	"xorm.io/builder"

	"code.vikunja.io/api/pkg/user"
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

const (
	entityProject = `project`
	entityTask    = `task`
)

// Subscription represents a subscription for an entity
type Subscription struct {
	// The numeric ID of the subscription
	ID int64 `xorm:"autoincr not null unique pk" json:"id"`

	EntityType SubscriptionEntityType `xorm:"index not null" json:"-"`
	Entity     string                 `xorm:"-" json:"entity" param:"entity"`
	// The id of the entity to subscribe to.
	EntityID int64 `xorm:"bigint index not null" json:"entity_id" param:"entityID"`

	// The user who made this subscription
	User   *user.User `xorm:"-" json:"user"`
	UserID int64      `xorm:"bigint index not null" json:"-"`

	// A timestamp when this subscription was created. You cannot change this value.
	Created time.Time `xorm:"created not null" json:"created"`

	web.CRUDable `xorm:"-" json:"-"`
	web.Rights   `xorm:"-" json:"-"`
}

// TableName gives us a better tabel name for the subscriptions table
func (sb *Subscription) TableName() string {
	return "subscriptions"
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

// String returns a human-readable string of an entity
func (et SubscriptionEntityType) String() string {
	switch et {
	case SubscriptionEntityProject:
		return entityProject
	case SubscriptionEntityTask:
		return entityTask
	}

	return ""
}

func (et SubscriptionEntityType) validate() error {
	if et == SubscriptionEntityProject ||
		et == SubscriptionEntityTask {
		return nil
	}

	return &ErrUnknownSubscriptionEntityType{EntityType: et}
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
	// Rights method alread does the validation of the entity type so we don't need to do that here

	sb.UserID = auth.GetID()

	sub, err := GetSubscription(s, sb.EntityType, sb.EntityID, auth)
	if err != nil {
		return err
	}
	if sub != nil {
		return &ErrSubscriptionAlreadyExists{
			EntityID:   sb.EntityID,
			EntityType: sb.EntityType,
			UserID:     sb.UserID,
		}
	}

	_, err = s.Insert(sb)
	if err != nil {
		return
	}

	sb.User, err = user.GetFromAuth(auth)
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

func getSubscriberCondForEntities(entityType SubscriptionEntityType, entityIDs []int64) (cond builder.Cond) {
	if entityType == SubscriptionEntityProject {
		return builder.And(
			builder.In("entity_id", entityIDs),
			builder.Eq{"entity_type": SubscriptionEntityProject},
		)
	}

	if entityType == SubscriptionEntityTask {
		return builder.Or(
			builder.And(
				builder.In("entity_id", entityIDs),
				builder.Eq{"entity_type": SubscriptionEntityTask},
			),
			builder.And(
				builder.Eq{"entity_id": builder.
					Select("project_id").
					From("tasks").
					Where(builder.In("id", entityIDs)),
				// TODO parent project
				},
				builder.Eq{"entity_type": SubscriptionEntityProject},
			),
		)
	}

	return
}

// GetSubscription returns a matching subscription for an entity and user.
// It will return the next parent of a subscription. That means for tasks, it will first look for a subscription for
// that task, if there is none it will look for a subscription on the project the task belongs to.
func GetSubscription(s *xorm.Session, entityType SubscriptionEntityType, entityID int64, a web.Auth) (subscription *Subscription, err error) {
	subs, err := GetSubscriptions(s, entityType, []int64{entityID}, a)
	if err != nil || len(subs) == 0 {
		return nil, err
	}
	if sub, exists := subs[entityID]; exists && len(sub) > 0 {
		return sub[0], nil // Take exact match first, if available
	}
	for _, sub := range subs {
		if len(sub) > 0 {
			return sub[0], nil // For parents, take next available
		}
	}
	return nil, nil
}

// GetSubscriptions returns a map of subscriptions to a set of given entity IDs
func GetSubscriptions(s *xorm.Session, entityType SubscriptionEntityType, entityIDs []int64, a web.Auth) (projectsToSubscriptions map[int64][]*Subscription, err error) {
	u, is := a.(*user.User)
	if u != nil && !is {
		return
	}
	if err := entityType.validate(); err != nil {
		return nil, err
	}

	switch entityType {
	case SubscriptionEntityProject:
		projects, err := GetProjectsByIDs(s, entityIDs)
		if err != nil {
			return nil, err
		}
		return GetSubscriptionsForProjects(s, projects, u)
	case SubscriptionEntityTask:
		subs, err := getSubscriptionsForTasks(s, entityIDs, u)
		if err != nil {
			return nil, err
		}

		projects, err := GetProjectsSimplByTaskIDs(s, entityIDs)
		if err != nil {
			return nil, err
		}

		tasks, err := GetTasksSimpleByIDs(s, entityIDs)
		if err != nil {
			return nil, err
		}

		projectSubscriptions, err := GetSubscriptionsForProjects(s, projects, u)
		if err != nil {
			return nil, err
		}

		for _, task := range tasks {
			// If a task is already subscribed through the parent project,
			// remove the task subscription since that's a duplicate.
			// But if the user is not subscribed to the task but a parent project is, add that to the subscriptions
			psub, hasProjectSub := projectSubscriptions[task.ProjectID]
			_, hasTaskSub := subs[task.ID]
			if hasProjectSub && hasTaskSub {
				delete(subs, task.ID)
			}

			if !hasTaskSub && !hasProjectSub {
				subs[task.ID] = psub
			}
		}

		return subs, nil
	}

	return
}

func GetSubscriptionsForProjects(s *xorm.Session, projects []*Project, a web.Auth) (projectsToSubscriptions map[int64][]*Subscription, err error) {
	u, is := a.(*user.User)
	if u != nil && !is {
		return
	}

	var ps = make(map[int64]*Project)
	origProjectIDs := make([]int64, 0, len(projects))
	allProjectIDs := make([]int64, 0, len(projects))

	for _, p := range projects {
		ps[p.ID] = p
		origProjectIDs = append(origProjectIDs, p.ID)
		allProjectIDs = append(allProjectIDs, p.ID)
	}

	// We can't just use the projects we have, we need to fetch the parents
	// because they may not be loaded in the same object

	for _, p := range projects {
		if p.ParentProjectID == 0 {
			continue
		}

		if _, has := ps[p.ParentProjectID]; has {
			continue
		}

		parents, err := GetAllParentProjects(s, p.ID)
		if err != nil {
			return nil, err
		}

		// Walk the tree up until we reach the top
		var parent = parents[p.ParentProjectID] // parent now has a pointer…
		ps[p.ID].ParentProject = parents[p.ParentProjectID]
		for parent != nil {
			allProjectIDs = append(allProjectIDs, parent.ID)
			parent = parents[parent.ParentProjectID] // … which means we can update it here and then update the pointer in the map
		}
	}

	var subscriptions []*Subscription
	if u != nil {
		err = s.
			Where("user_id = ?", u.ID).
			And(getSubscriberCondForEntities(SubscriptionEntityProject, allProjectIDs)).
			Find(&subscriptions)
	} else {
		err = s.
			And(getSubscriberCondForEntities(SubscriptionEntityProject, allProjectIDs)).
			Find(&subscriptions)
	}
	if err != nil {
		return nil, err
	}

	projectsToSubscriptions = make(map[int64][]*Subscription)
	for _, sub := range subscriptions {
		sub.Entity = sub.EntityType.String()
		projectsToSubscriptions[sub.EntityID] = append(projectsToSubscriptions[sub.EntityID], sub)
	}

	// Rearrange so that subscriptions trickle down

	for _, eID := range origProjectIDs {
		// If the current project does not have a subscription, climb up the tree until a project has one,
		// then use that subscription for all child projects
		_, has := projectsToSubscriptions[eID]
		_, hasProject := ps[eID]
		if !has && hasProject {
			_, exists := ps[eID]
			if !exists {
				continue
			}
			var parent = ps[eID].ParentProject
			for parent != nil {
				sub, has := projectsToSubscriptions[parent.ID]
				projectsToSubscriptions[eID] = sub
				parent = parent.ParentProject
				if has { // reached the top of the tree
					break
				}
			}
		}
	}

	return projectsToSubscriptions, nil
}

func getSubscriptionsForTasks(s *xorm.Session, taskIDs []int64, u *user.User) (projectsToSubscriptions map[int64][]*Subscription, err error) {
	var subscriptions []*Subscription
	if u != nil {
		err = s.
			Where("user_id = ?", u.ID).
			And(getSubscriberCondForEntities(SubscriptionEntityTask, taskIDs)).
			Find(&subscriptions)
	} else {
		err = s.
			And(getSubscriberCondForEntities(SubscriptionEntityTask, taskIDs)).
			Find(&subscriptions)
	}
	if err != nil {
		return nil, err
	}

	projectsToSubscriptions = make(map[int64][]*Subscription)
	for _, sub := range subscriptions {
		sub.Entity = sub.EntityType.String()
		projectsToSubscriptions[sub.EntityID] = append(projectsToSubscriptions[sub.EntityID], sub)
	}

	return
}

func getSubscribersForEntity(s *xorm.Session, entityType SubscriptionEntityType, entityID int64) (subscriptions []*Subscription, err error) {
	if err := entityType.validate(); err != nil {
		return nil, err
	}

	subs, err := GetSubscriptions(s, entityType, []int64{entityID}, nil)
	if err != nil {
		return
	}

	userIDs := []int64{}
	subscriptions = make([]*Subscription, 0, len(subs))
	for _, subss := range subs {
		for _, subscription := range subss {
			userIDs = append(userIDs, subscription.UserID)
			subscriptions = append(subscriptions, subscription)
		}
	}

	users, err := user.GetUsersByIDs(s, userIDs)
	if err != nil {
		return
	}

	for _, subscription := range subscriptions {
		subscription.User = users[subscription.UserID]
	}
	return
}
