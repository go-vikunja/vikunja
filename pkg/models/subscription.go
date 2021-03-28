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

	"xorm.io/builder"

	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/web"
	"xorm.io/xorm"
)

// SubscriptionEntityType represents all entities which can be subscribed to
type SubscriptionEntityType int

const (
	SubscriptionEntityUnknown = iota
	SubscriptionEntityNamespace
	SubscriptionEntityList
	SubscriptionEntityTask
)

const (
	entityNamespace = `namespace`
	entityList      = `list`
	entityTask      = `task`
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
	case entityNamespace:
		return SubscriptionEntityNamespace
	case entityList:
		return SubscriptionEntityList
	case entityTask:
		return SubscriptionEntityTask
	}

	return SubscriptionEntityUnknown
}

// String returns a human-readable string of an entity
func (et SubscriptionEntityType) String() string {
	switch et {
	case SubscriptionEntityNamespace:
		return entityNamespace
	case SubscriptionEntityList:
		return entityList
	case SubscriptionEntityTask:
		return entityTask
	}

	return ""
}

func (et SubscriptionEntityType) validate() error {
	if et == SubscriptionEntityNamespace ||
		et == SubscriptionEntityList ||
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
// @Param entity path string true "The entity the user subscribes to. Can be either `namespace`, `list` or `task`."
// @Param entityID path string true "The numeric id of the entity to subscribe to."
// @Success 200 {object} models.Subscription "The subscription"
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
// @Param entity path string true "The entity the user subscribed to. Can be either `namespace`, `list` or `task`."
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

func getSubscriberCondForEntity(entityType SubscriptionEntityType, entityID int64) (cond builder.Cond) {
	if entityType == SubscriptionEntityNamespace {
		cond = builder.And(
			builder.Eq{"entity_id": entityID},
			builder.Eq{"entity_type": SubscriptionEntityNamespace},
		)
	}

	if entityType == SubscriptionEntityList {
		cond = builder.Or(
			builder.And(
				builder.Eq{"entity_id": entityID},
				builder.Eq{"entity_type": SubscriptionEntityList},
			),
			builder.And(
				builder.Eq{"entity_id": builder.
					Select("namespace_id").
					From("lists").
					Where(builder.Eq{"id": entityID}),
				},
				builder.Eq{"entity_type": SubscriptionEntityNamespace},
			),
		)
	}

	if entityType == SubscriptionEntityTask {
		cond = builder.Or(
			builder.And(
				builder.Eq{"entity_id": entityID},
				builder.Eq{"entity_type": SubscriptionEntityTask},
			),
			builder.And(
				builder.Eq{"entity_id": builder.
					Select("namespace_id").
					From("lists").
					Join("INNER", "tasks", "lists.id = tasks.list_id").
					Where(builder.Eq{"tasks.id": entityID}),
				},
				builder.Eq{"entity_type": SubscriptionEntityNamespace},
			),
			builder.And(
				builder.Eq{"entity_id": builder.
					Select("list_id").
					From("tasks").
					Where(builder.Eq{"id": entityID}),
				},
				builder.Eq{"entity_type": SubscriptionEntityList},
			),
		)
	}

	return
}

// GetSubscription returns a matching subscription for an entity and user.
// It will return the next parent of a subscription. That means for tasks, it will first look for a subscription for
// that task, if there is none it will look for a subscription on the list the task belongs to and if that also
// doesn't exist it will check for a subscription for the namespace the list is belonging to.
func GetSubscription(s *xorm.Session, entityType SubscriptionEntityType, entityID int64, a web.Auth) (subscription *Subscription, err error) {
	u, is := a.(*user.User)
	if !is {
		return
	}

	if err := entityType.validate(); err != nil {
		return nil, err
	}

	subscription = &Subscription{}
	cond := getSubscriberCondForEntity(entityType, entityID)
	exists, err := s.
		Where("user_id = ?", u.ID).
		And(cond).
		Get(subscription)
	if !exists {
		return nil, err
	}

	subscription.Entity = subscription.EntityType.String()

	return subscription, err
}

func getSubscribersForEntity(s *xorm.Session, entityType SubscriptionEntityType, entityID int64) (subscriptions []*Subscription, err error) {
	if err := entityType.validate(); err != nil {
		return nil, err
	}

	cond := getSubscriberCondForEntity(entityType, entityID)
	err = s.
		Where(cond).
		Find(&subscriptions)
	if err != nil {
		return
	}

	userIDs := []int64{}
	for _, subscription := range subscriptions {
		userIDs = append(userIDs, subscription.UserID)
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
