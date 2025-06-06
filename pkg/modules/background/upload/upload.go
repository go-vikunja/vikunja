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

package upload

import (
	"strconv"

	"xorm.io/xorm"

	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/background"
	"code.vikunja.io/api/pkg/web"
)

// Provider represents an upload provider
type Provider struct {
}

// Search is only used to implement the interface
func (p *Provider) Search(_ *xorm.Session, _ string, _ int64) (result []*background.Image, err error) {
	return
}

// Set handles setting a background through a file upload
// @Summary Upload a project background
// @Description Upload a project background.
// @tags project
// @Accept mpfd
// @Produce json
// @Param id path int true "Project ID"
// @Param background formData string true "The file as single file."
// @Security JWTKeyAuth
// @Success 200 {object} models.Message "The background was set successfully."
// @Failure 400 {object} models.Message "File is no image."
// @Failure 403 {object} models.Message "No access to the project."
// @Failure 403 {object} models.Message "File too large."
// @Failure 404 {object} models.Message "The project does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /projects/{id}/backgrounds/upload [put]
func (p *Provider) Set(s *xorm.Session, img *background.Image, project *models.Project, _ web.Auth) (err error) {
	// Remove the old background if one exists
	if project.BackgroundFileID != 0 {
		file := files.File{ID: project.BackgroundFileID}
		err := file.Delete(s)
		if err != nil && !files.IsErrFileDoesNotExist(err) {
			return err
		}
	}

	file := &files.File{}
	file.ID, err = strconv.ParseInt(img.ID, 10, 64)
	if err != nil {
		return
	}

	project.BackgroundInformation = &models.ProjectBackgroundType{Type: models.ProjectBackgroundUpload}

	return models.SetProjectBackground(s, project.ID, file, project.BackgroundBlurHash)
}
