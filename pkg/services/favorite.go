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

// GetForUserByType gets all favorites for a user by type.
func (fs *FavoriteService) GetForUserByType(s *xorm.Session, u *user.User, kind models.FavoriteKind) ([]*models.Favorite, error) {
	favorites := []*models.Favorite{}
	err := s.Where("user_id = ? AND kind = ?", u.ID, kind).Find(&favorites)
	return favorites, err
}

// RemoveFromFavorite removes an entity from favorites.
func (fs *FavoriteService) RemoveFromFavorite(s *xorm.Session, entityID int64, a web.Auth, kind models.FavoriteKind) error {
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
