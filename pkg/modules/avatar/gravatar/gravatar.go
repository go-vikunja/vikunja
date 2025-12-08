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

package gravatar

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/modules/keyvalue"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/utils"
	"golang.org/x/sync/singleflight"
)

type avatar struct {
	Content  []byte    `json:"content"`
	MimeType string    `json:"mime_type"`
	LoadedAt time.Time `json:"loaded_at"`
}

// Provider is the gravatar provider
type Provider struct {
}

var gravatarFetchGroup singleflight.Group

// FlushCache removes all gravatar cache entries for a user
func (g *Provider) FlushCache(u *user.User) error {
	return keyvalue.DelPrefix(keyPrefix + u.Username + "_")
}

const keyPrefix = "gravatar_avatar_"

// GetAvatar implements getting the avatar for the user
func (g *Provider) GetAvatar(user *user.User, size int64) ([]byte, string, error) {
	sizeString := strconv.FormatInt(size, 10)
	cacheKey := keyPrefix + user.Username + "_" + sizeString

	var av avatar
	exists, err := keyvalue.GetWithValue(cacheKey, &av)
	if err != nil {
		log.Errorf("Error retrieving gravatar from keyvalue store: %s", err)
	}

	if exists {
		elapsed := time.Since(av.LoadedAt)
		needsRefetch := elapsed > time.Duration(config.AvatarGravaterExpiration.GetInt64())*time.Second
		if needsRefetch {
			log.Debugf("Refetching avatar for user %d after %v", user.ID, elapsed)
		} else {
			log.Debugf("Serving avatar for user %d from cache", user.ID)
			return av.Content, av.MimeType, nil
		}
	}

	result, err, _ := gravatarFetchGroup.Do(cacheKey, func() (interface{}, error) {
		cached, cacheErr := keyvalue.GetWithValue(cacheKey, &av)
		if cacheErr != nil {
			log.Errorf("Error retrieving gravatar from keyvalue store: %s", cacheErr)
		}
		if cached && !g.avatarExpired(av) {
			log.Debugf("Serving avatar for user %d from cache", user.ID)
			return av, nil
		}

		log.Debugf("Gravatar for user %d with size %d not cached, requesting from gravatar...", user.ID, size)
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://www.gravatar.com/avatar/"+utils.Md5String(strings.ToLower(user.Email))+"?s="+sizeString+"&d=mp", nil)
		if err != nil {
			return nil, err
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		avatarContent, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		mimeType := "image/jpeg"
		if contentType := resp.Header.Get("Content-Type"); contentType != "" {
			mimeType = contentType
		}

		av = avatar{
			Content:  avatarContent,
			MimeType: mimeType,
			LoadedAt: time.Now(),
		}

		if err := keyvalue.Put(cacheKey, av); err != nil {
			log.Errorf("Error storing gravatar in keyvalue store: %s", err)
		}

		return av, nil
	})
	if err != nil {
		return nil, "", err
	}

	av = result.(avatar)
	return av.Content, av.MimeType, nil
}

func (g *Provider) avatarExpired(av avatar) bool {
	return time.Since(av.LoadedAt) > time.Duration(config.AvatarGravaterExpiration.GetInt64())*time.Second
}
