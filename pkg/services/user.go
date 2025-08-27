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
	"xorm.io/xorm"
)

// UserService is a service for users.
type UserService struct {
	DB *xorm.Engine
}

func init() {
	models.GetUsersOrLinkSharesFromIDsFunc = func(s *xorm.Session, ids []int64) (map[int64]*user.User, error) {
		userService := &UserService{DB: s.Engine()}
		return userService.GetUsersAndProxiesFromIDs(s, ids)
	}
	models.NewUserProxyFromLinkShareFunc = func(share *models.LinkSharing) *user.User {
		userService := &UserService{}
		return userService.NewUserProxyFromLinkShare(share)
	}
}

func (us *UserService) NewUserProxyFromLinkShare(share *models.LinkSharing) *user.User {
	suffix := "Link Share"
	if share.Name != "" {
		suffix = " (" + suffix + ")"
	}

	username := "link-share-" + strconv.FormatInt(share.ID, 10)

	return &user.User{
		ID:       share.ID * -1,
		Name:     share.Name + suffix,
		Username: username,
		Created:  share.Created,
		Updated:  share.Updated,
	}
}

// GetUsersAndProxiesFromIDs returns all users or pseudo link shares from a slice of ids. ids < 0 are considered to be a link share in that case.
func (us *UserService) GetUsersAndProxiesFromIDs(s *xorm.Session, ids []int64) (users map[int64]*user.User, err error) {
	if s == nil {
		s = us.DB.NewSession()
		defer s.Close()
	}

	users = make(map[int64]*user.User)
	var userIDs []int64
	var linkShareIDs []int64
	for _, id := range ids {
		if id < 0 {
			linkShareIDs = append(linkShareIDs, id*-1)
			continue
		}

		userIDs = append(userIDs, id)
	}

	if len(userIDs) > 0 {
		users, err = user.GetUsersByIDs(s, userIDs)
		if err != nil {
			return
		}
	}

	if len(linkShareIDs) == 0 {
		return
	}

	shares, err := models.GetLinkSharesByIDs(s, linkShareIDs)
	if err != nil {
		return nil, err
	}

	for _, share := range shares {
		users[share.ID*-1] = us.NewUserProxyFromLinkShare(share)
	}

	return
}
