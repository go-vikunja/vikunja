// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2020 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package upload

import (
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/background"
	"code.vikunja.io/web"
	"strconv"
)

// Provider represents an upload provider
type Provider struct {
}

// Search is only used to implement the interface
func (p *Provider) Search(search string, page int64) (result []*background.Image, err error) {
	return
}

// Set handles setting a background through a file upload
// @Summary Upload a list background
// @Description Upload a list background.
// @tags list
// @Accept mpfd
// @Produce json
// @Param id path int true "List ID"
// @Param background formData string true "The file as single file."
// @Security JWTKeyAuth
// @Success 200 {object} models.Message "The background was set successfully."
// @Failure 403 {object} models.Message "No access to the list."
// @Failure 403 {object} models.Message "File too large."
// @Failure 404 {object} models.Message "The list does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /lists/{id}/backgrounds/upload [put]
func (p *Provider) Set(image *background.Image, list *models.List, auth web.Auth) (err error) {
	// Remove the old background if one exists
	if list.BackgroundFileID != 0 {
		file := files.File{ID: list.BackgroundFileID}
		if err := file.Delete(); err != nil {
			return err
		}
	}

	file := &files.File{}
	file.ID, err = strconv.ParseInt(image.ID, 10, 64)
	if err != nil {
		return
	}

	list.BackgroundInformation = &models.ListBackgroundType{Type: models.ListBackgroundUpload}

	return models.SetListBackground(list.ID, file)
}
