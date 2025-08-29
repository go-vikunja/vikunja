package services

import (
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"xorm.io/xorm"
)

// FavoriteService is a service for managing favorites.
type FavoriteService struct{}

// NewFavoriteService returns a new FavoriteService.
func NewFavoriteService() *FavoriteService {
	return &FavoriteService{}
}

// GetForUserByType fetches all of a user's favorites for a specific entity type.
// This moves the logic from the old models.getFavorites function.
func (fs *FavoriteService) GetForUserByType(s *xorm.Session, u *user.User, entityType models.FavoriteKind) ([]*models.Favorite, error) {
	var favorites []*models.Favorite
	err := s.Where("user_id = ? AND kind = ?", u.ID, entityType).Find(&favorites)
	return favorites, err
}
