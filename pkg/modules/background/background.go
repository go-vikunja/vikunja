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

package background

import (
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/web"
	"xorm.io/xorm"
)

// Image represents an image which can be used as a project background
type Image struct {
	ID       string `json:"id"`
	URL      string `json:"url"`
	Thumb    string `json:"thumb,omitempty"`
	BlurHash string `json:"blur_hash"`
	// This can be used to supply extra information from an image provider to clients
	Info interface{} `json:"info,omitempty"`
}

const MaxBackgroundImageHeight = 3840

// Provider represents something that is able to get a project of images and set one of them as background
type Provider interface {
	// Search is used to either return a pre-defined project of Image or let the user search for an image
	Search(s *xorm.Session, search string, page int64) (result []*Image, err error)
	// Set sets an image which was most likely previously obtained by Search as project background
	Set(s *xorm.Session, image *Image, project *models.Project, auth web.Auth) (err error)
}
