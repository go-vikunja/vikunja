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

	"code.vikunja.io/api/pkg/web"
	"xorm.io/builder"
	"xorm.io/xorm"

	"code.vikunja.io/api/pkg/user"
)

type ReactionKind int

const (
	ReactionKindTask = iota
	ReactionKindComment
)

type Reaction struct {
	// The unique numeric id of this reaction
	ID int64 `xorm:"autoincr not null unique pk" json:"-" param:"reaction"`

	// The user who reacted
	User   *user.User `xorm:"-" json:"user" valid:"-"`
	UserID int64      `xorm:"bigint not null INDEX" json:"-"`

	// The id of the entity you're reacting to
	EntityID int64 `xorm:"bigint not null INDEX" json:"-" param:"entityid"`
	// The entity kind which you're reacting to. Can be 0 for task, 1 for comment.
	EntityKind       ReactionKind `xorm:"bigint not null INDEX" json:"-"`
	EntityKindString string       `xorm:"-" json:"-" param:"entitykind"`

	// The actual reaction. This can be any valid utf character or text, up to a length of 20.
	Value string `xorm:"varchar(20) not null INDEX" json:"value" valid:"required"`

	// A timestamp when this reaction was created. You cannot change this value.
	Created time.Time `xorm:"created not null" json:"created"`

	web.CRUDable    `xorm:"-" json:"-"`
	web.Permissions `xorm:"-" json:"-"`
}

func (*Reaction) TableName() string {
	return "reactions"
}

type ReactionMap map[string][]*user.User

// ReadAll gets all reactions for an entity
// @Summary Get all reactions for an entity
// @Description Returns all reactions for an entity
// @tags task
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "Entity ID"
// @Param kind path int true "The kind of the entity. Can be either `tasks` or `comments` for task comments"
// @Success 200 {array} models.ReactionMap "The reactions"
// @Failure 403 {object} web.HTTPError "The user does not have access to the entity"
// @Failure 500 {object} models.Message "Internal error"
// @Router /{kind}/{id}/reactions [get]
func (r *Reaction) ReadAll(s *xorm.Session, a web.Auth, _ string, _ int, _ int) (result interface{}, resultCount int, numberOfTotalItems int64, err error) {

	can, _, err := r.CanRead(s, a)
	if err != nil {
		return nil, 0, 0, err
	}
	if !can {
		return nil, 0, 0, ErrGenericForbidden{}
	}

	reactions, err := getReactionsForEntityIDs(s, r.EntityKind, []int64{r.EntityID})
	if err != nil {
		return
	}

	return reactions[r.EntityID], len(reactions[r.EntityID]), int64(len(reactions[r.EntityID])), nil
}

func getReactionsForEntityIDs(s *xorm.Session, entityKind ReactionKind, entityIDs []int64) (reactionsWithTasks map[int64]ReactionMap, err error) {

	where := builder.And(
		builder.Eq{"entity_kind": entityKind},
		builder.In("entity_id", entityIDs),
	)

	reactions := []*Reaction{}
	err = s.Where(where).Find(&reactions)
	if err != nil {
		return
	}

	if len(reactions) == 0 {
		return
	}

	cond := builder.
		Select("user_id").
		From("reactions").
		Where(where)

	users, err := user.GetUsersByCond(s, builder.In("id", cond))
	if err != nil {
		return
	}

	reactionsWithTasks = make(map[int64]ReactionMap)
	for _, reaction := range reactions {
		if _, taskExists := reactionsWithTasks[reaction.EntityID]; !taskExists {
			reactionsWithTasks[reaction.EntityID] = make(ReactionMap)
		}

		if _, has := reactionsWithTasks[reaction.EntityID][reaction.Value]; !has {
			reactionsWithTasks[reaction.EntityID][reaction.Value] = []*user.User{}
		}

		reactionsWithTasks[reaction.EntityID][reaction.Value] = append(reactionsWithTasks[reaction.EntityID][reaction.Value], users[reaction.UserID])
	}

	return
}

// Delete removes the user's own reaction
// @Summary Removes the user's reaction
// @Description Removes the reaction of that user on that entity.
// @tags task
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "Entity ID"
// @Param kind path int true "The kind of the entity. Can be either `tasks` or `comments` for task comments"
// @Param project body models.Reaction true "The reaction you want to add to the entity."
// @Success 200 {object} models.Message "The reaction was successfully removed."
// @Failure 403 {object} web.HTTPError "The user does not have access to the entity"
// @Failure 500 {object} models.Message "Internal error"
// @Router /{kind}/{id}/reactions/delete [post]
func (r *Reaction) Delete(s *xorm.Session, a web.Auth) (err error) {
	r.UserID = a.GetID()

	_, err = s.Where("user_id = ? AND entity_id = ? AND entity_kind = ? AND value = ?", r.UserID, r.EntityID, r.EntityKind, r.Value).
		Delete(&Reaction{})
	return
}

// Create adds a new reaction to an entity
// @Summary Add a reaction to an entity
// @Description Add a reaction to an entity. Will do nothing if the reaction already exists.
// @tags task
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "Entity ID"
// @Param kind path int true "The kind of the entity. Can be either `tasks` or `comments` for task comments"
// @Param project body models.Reaction true "The reaction you want to add to the entity."
// @Success 200 {object} models.Reaction "The created reaction"
// @Failure 403 {object} web.HTTPError "The user does not have access to the entity"
// @Failure 500 {object} models.Message "Internal error"
// @Router /{kind}/{id}/reactions [put]
func (r *Reaction) Create(s *xorm.Session, a web.Auth) (err error) {
	r.UserID = a.GetID()

	exists, err := s.Where("user_id = ? AND entity_id = ? AND entity_kind = ? AND value = ?", r.UserID, r.EntityID, r.EntityKind, r.Value).
		Exist(&Reaction{})
	if err != nil {
		return err
	}

	if exists {
		return
	}

	r.ID = 0
	_, err = s.Insert(r)
	return
}
