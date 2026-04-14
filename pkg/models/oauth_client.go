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

	"xorm.io/xorm"
)

// OAuthClient represents a dynamically registered OAuth 2.0 client.
type OAuthClient struct {
	ClientID     string    `xorm:"varchar(50) not null unique pk" json:"client_id"`
	ClientName   string    `xorm:"varchar(255) not null" json:"client_name"`
	RedirectURIs string    `xorm:"text not null" json:"redirect_uris"`
	Created      time.Time `xorm:"created not null" json:"created"`
}

func (*OAuthClient) TableName() string {
	return "oauth_clients"
}

// CreateOAuthClient creates a new OAuth client in the database.
func CreateOAuthClient(s *xorm.Session, oauthClient *OAuthClient) error {
	_, err := s.Insert(oauthClient)
	return err
}

// GetOAuthClientByClientID retrieves an OAuth client by its client_id.
func GetOAuthClientByClientID(s *xorm.Session, clientID string) (*OAuthClient, error) {
	oauthClient := &OAuthClient{}
	has, err := s.Where("client_id = ?", clientID).Get(oauthClient)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, &ErrOAuthClientNotFound{}
	}
	return oauthClient, nil
}

// GetOAuthClientByRegistrationAccessToken retrieves an OAuth client by its registration access token.
func GetOAuthClientByRegistrationAccessToken(s *xorm.Session, token string) (*OAuthClient, error) {
	oauthClient := &OAuthClient{}
	has, err := s.Where("registration_access_token = ?", token).Get(oauthClient)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, &ErrOAuthClientNotFound{}
	}
	return oauthClient, nil
}

// UpdateOAuthClient updates an existing OAuth client.
func UpdateOAuthClient(s *xorm.Session, oauthClient *OAuthClient) error {
	_, err := s.ID(oauthClient.ClientID).Update(oauthClient)
	return err
}

// DeleteOAuthClient deletes an OAuth client by its ID.
func DeleteOAuthClient(s *xorm.Session, id int64) error {
	_, err := s.ID(id).Delete(&OAuthClient{})
	return err
}
