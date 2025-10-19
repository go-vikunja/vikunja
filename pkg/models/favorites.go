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
	"code.vikunja.io/api/pkg/web"
	"xorm.io/xorm"
)

// FavoriteServiceProvider is a function type that returns a favorite service instance
// This is used to avoid import cycles between models and services packages
type FavoriteServiceProvider func() interface {
	AddToFavorite(s *xorm.Session, entityID int64, a web.Auth, kind FavoriteKind) error
	RemoveFromFavorite(s *xorm.Session, entityID int64, a web.Auth, kind FavoriteKind) error
	IsFavorite(s *xorm.Session, entityID int64, a web.Auth, kind FavoriteKind) (bool, error)
	GetFavoritesMap(s *xorm.Session, entityIDs []int64, a web.Auth, kind FavoriteKind) (map[int64]bool, error)
}

// favoriteServiceProvider is the registered service provider function
var favoriteServiceProvider FavoriteServiceProvider

// RegisterFavoriteService registers a service provider for favorite operations
// This should be called during application initialization by the services package
func RegisterFavoriteService(provider FavoriteServiceProvider) {
	favoriteServiceProvider = provider
}

// getFavoriteService returns the registered favorite service instance
func getFavoriteService() interface {
	AddToFavorite(s *xorm.Session, entityID int64, a web.Auth, kind FavoriteKind) error
	RemoveFromFavorite(s *xorm.Session, entityID int64, a web.Auth, kind FavoriteKind) error
	IsFavorite(s *xorm.Session, entityID int64, a web.Auth, kind FavoriteKind) (bool, error)
	GetFavoritesMap(s *xorm.Session, entityIDs []int64, a web.Auth, kind FavoriteKind) (map[int64]bool, error)
} {
	if favoriteServiceProvider == nil {
		panic("FavoriteService not registered - did you forget to call services.InitializeDependencies()?")
	}
	return favoriteServiceProvider()
}

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

// @Deprecated: This function is deprecated and will be removed in a future release. Use FavoriteService.AddToFavorite instead.
func AddToFavorites(s *xorm.Session, entityID int64, a web.Auth, kind FavoriteKind) error {
	return getFavoriteService().AddToFavorite(s, entityID, a, kind)
}

// @Deprecated: This function is deprecated and will be removed in a future release. Use FavoriteService.RemoveFromFavorite instead.
func RemoveFromFavorite(s *xorm.Session, entityID int64, a web.Auth, kind FavoriteKind) error {
	return getFavoriteService().RemoveFromFavorite(s, entityID, a, kind)
}

// @Deprecated: This function is deprecated and will be removed in a future release. Use FavoriteService.IsFavorite instead.
func IsFavorite(s *xorm.Session, entityID int64, a web.Auth, kind FavoriteKind) (is bool, err error) {
	return getFavoriteService().IsFavorite(s, entityID, a, kind)
}

// @Deprecated: This function is deprecated and will be removed in a future release. Use FavoriteService.GetFavoritesMap instead.
func getFavorites(s *xorm.Session, entityIDs []int64, a web.Auth, kind FavoriteKind) (favorites map[int64]bool, err error) {
	return getFavoriteService().GetFavoritesMap(s, entityIDs, a, kind)
}
