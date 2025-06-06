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
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"
	"xorm.io/builder"
	"xorm.io/xorm"
)

// FavoriteKind represents the kind of entities that can be marked as favorite
type FavoriteKind int

const (
	FavoriteKindUnknown FavoriteKind = iota
	FavoriteKindTask
	FavoriteKindProject
)

// Favorite represents an entity which is a favorite to someone
type Favorite struct {
	EntityID int64        `xorm:"bigint not null pk"`
	UserID   int64        `xorm:"bigint not null pk"`
	Kind     FavoriteKind `xorm:"int not null pk"`
}

// TableName is the table name
func (t *Favorite) TableName() string {
	return "favorites"
}

func addToFavorites(s *xorm.Session, entityID int64, a web.Auth, kind FavoriteKind) error {
	u, err := user.GetFromAuth(a)
	if err != nil {
		// Only error GetFromAuth is if it's a link share and we want to ignore that
		return nil
	}

	fav := &Favorite{
		EntityID: entityID,
		UserID:   u.ID,
		Kind:     kind,
	}

	_, err = s.Insert(fav)
	return err
}

func removeFromFavorite(s *xorm.Session, entityID int64, a web.Auth, kind FavoriteKind) error {
	u, err := user.GetFromAuth(a)
	if err != nil {
		// Only error GetFromAuth is if it's a link share and we want to ignore that
		return nil
	}

	_, err = s.
		Where("entity_id = ? AND user_id = ? AND kind = ?", entityID, u.ID, kind).
		Delete(&Favorite{})
	return err
}

func isFavorite(s *xorm.Session, entityID int64, a web.Auth, kind FavoriteKind) (is bool, err error) {
	u, err := user.GetFromAuth(a)
	if err != nil {
		// Only error GetFromAuth is if it's a link share and we want to ignore that
		return false, nil
	}

	return s.
		Where("entity_id = ? AND user_id = ? AND kind = ?", entityID, u.ID, kind).
		Exist(&Favorite{})
}

func getFavorites(s *xorm.Session, entityIDs []int64, a web.Auth, kind FavoriteKind) (favorites map[int64]bool, err error) {
	favorites = make(map[int64]bool)
	u, err := user.GetFromAuth(a)
	if err != nil {
		// Only error GetFromAuth is if it's a link share and we want to ignore that
		return favorites, nil
	}

	favs := []*Favorite{}
	err = s.Where(builder.And(
		builder.Eq{"user_id": u.ID},
		builder.Eq{"kind": kind},
		builder.In("entity_id", entityIDs),
	)).
		Find(&favs)

	for _, fav := range favs {
		favorites[fav.EntityID] = true
	}
	return
}
