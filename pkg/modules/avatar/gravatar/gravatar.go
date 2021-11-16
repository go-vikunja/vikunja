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

package gravatar

import (
	"context"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/utils"
)

type avatar struct {
	content  []byte
	loadedAt time.Time
}

// Provider is the gravatar provider
type Provider struct {
}

// avatars is a global map which contains cached avatars of the users
var avatars map[string]*avatar

func init() {
	avatars = make(map[string]*avatar)
}

// GetAvatar implements getting the avatar for the user
func (g *Provider) GetAvatar(user *user.User, size int64) ([]byte, string, error) {
	sizeString := strconv.FormatInt(size, 10)
	cacheKey := user.Username + "_" + sizeString
	a, exists := avatars[cacheKey]
	var needsRefetch bool
	if exists {
		// elaped is alway < 0 so the next check would always succeed.
		// To have it make sense, we flip that.
		elapsed := time.Until(a.loadedAt) * -1
		needsRefetch = elapsed > time.Duration(config.AvatarGravaterExpiration.GetInt64())*time.Second
		if needsRefetch {
			log.Debugf("Refetching avatar for user %d after %v", user.ID, elapsed)
		} else {
			log.Debugf("Serving avatar for user %d from cache", user.ID)
		}
	}
	if !exists || needsRefetch {
		log.Debugf("Gravatar for user %d with size %d not cached, requesting from gravatar...", user.ID, size)
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://www.gravatar.com/avatar/"+utils.Md5String(strings.ToLower(user.Email))+"?s="+sizeString+"&d=mp", nil)
		if err != nil {
			return nil, "", err
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, "", err
		}
		defer resp.Body.Close()
		avatarContent, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, "", err
		}
		avatars[cacheKey] = &avatar{
			content:  avatarContent,
			loadedAt: time.Now(),
		}
	}
	return avatars[cacheKey].content, "image/jpg", nil
}
