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

// FavoriteService is a service for managing favorites.
type FavoriteService struct {
	DB *xorm.Engine
}

// NewFavoriteService creates a new FavoriteService.
func NewFavoriteService(db *xorm.Engine) *FavoriteService {
	return &FavoriteService{DB: db}
}

// AddToFavorite creates a favorite entry for the given entity and auth context.
// The caller is responsible for providing an active session when part of a larger transaction.
func (fs *FavoriteService) AddToFavorite(s *xorm.Session, entityID int64, a web.Auth, kind models.FavoriteKind) error {
	if a == nil {
		return nil
	}

	u, err := user.GetFromAuth(a)
	if err != nil {
		// Only error GetFromAuth is if it's a link share and we want to ignore that
		return nil
	}

	_, err = s.Insert(&models.Favorite{
		EntityID: entityID,
		UserID:   u.ID,
		Kind:     kind,
	})
	return err
}

// GetForUserByType gets all favorites for a user by type.
func (fs *FavoriteService) GetForUserByType(s *xorm.Session, u *user.User, kind models.FavoriteKind) ([]*models.Favorite, error) {
	favorites := []*models.Favorite{}
	err := s.Where("user_id = ? AND kind = ?", u.ID, kind).Find(&favorites)
	return favorites, err
}

// RemoveFromFavorite removes an entity from favorites.
func (fs *FavoriteService) RemoveFromFavorite(s *xorm.Session, entityID int64, a web.Auth, kind models.FavoriteKind) error {
	if a == nil {
		return nil
	}

	u, err := user.GetFromAuth(a)
	if err != nil {
		// Only error GetFromAuth is if it's a link share and we want to ignore that
		return nil
	}

	_, err = s.
		Where("entity_id = ? AND user_id = ? AND kind = ?", entityID, u.ID, kind).
		Delete(&models.Favorite{})
	return err
}

// IsFavorite checks if an entity is marked as favorite for the given auth context.
func (fs *FavoriteService) IsFavorite(s *xorm.Session, entityID int64, a web.Auth, kind models.FavoriteKind) (bool, error) {
	if a == nil {
		return false, nil
	}

	u, err := user.GetFromAuth(a)
	if err != nil {
		// Only error GetFromAuth is if it's a link share and we want to ignore that
		return false, nil
	}

	return s.
		Where("entity_id = ? AND user_id = ? AND kind = ?", entityID, u.ID, kind).
		Exist(&models.Favorite{})
}

// GetFavoritesMap returns a map of entity IDs to favorite status for the given auth context.
// This is useful for bulk checking if multiple entities are favorites.
func (fs *FavoriteService) GetFavoritesMap(s *xorm.Session, entityIDs []int64, a web.Auth, kind models.FavoriteKind) (map[int64]bool, error) {
	favorites := make(map[int64]bool)

	if a == nil {
		return favorites, nil
	}

	u, err := user.GetFromAuth(a)
	if err != nil {
		// Only error GetFromAuth is if it's a link share and we want to ignore that
		return favorites, nil
	}

	if len(entityIDs) == 0 {
		return favorites, nil
	}

	favs := []*models.Favorite{}
	err = s.Where(builder.And(
		builder.Eq{"user_id": u.ID},
		builder.Eq{"kind": kind},
		builder.In("entity_id", entityIDs),
	)).Find(&favs)

	if err != nil {
		return nil, err
	}

	for _, fav := range favs {
		favorites[fav.EntityID] = true
	}

	return favorites, nil
}
