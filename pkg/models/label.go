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

	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/utils"

	"code.vikunja.io/api/pkg/web"
	"xorm.io/xorm"
)

// Label represents a label
type Label struct {
	// The unique, numeric id of this label.
	ID int64 `xorm:"bigint autoincr not null unique pk" json:"id" param:"label"`
	// The title of the label. You'll see this one on tasks associated with it.
	Title string `xorm:"varchar(250) not null" json:"title" valid:"runelength(1|250)" minLength:"1" maxLength:"250"`
	// The label description.
	Description string `xorm:"longtext null" json:"description"`
	// The color this label has in hex format.
	HexColor string `xorm:"varchar(6) null" json:"hex_color" valid:"runelength(0|7)" maxLength:"7"`

	CreatedByID int64 `xorm:"bigint not null" json:"-"`
	// The user who created this label
	CreatedBy *user.User `xorm:"-" json:"created_by"`

	// A timestamp when this label was created. You cannot change this value.
	Created time.Time `xorm:"created not null" json:"created"`
	// A timestamp when this label was last updated. You cannot change this value.
	Updated time.Time `xorm:"updated not null" json:"updated"`

	web.CRUDable    `xorm:"-" json:"-"`
	web.Permissions `xorm:"-" json:"-"`
}

// TableName makes a pretty table name
func (*Label) TableName() string {
	return "labels"
}

// Create creates a new label
// @Summary Create a label
// @Description Creates a new label.
// @tags labels
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param label body models.Label true "The label object"
// @Success 201 {object} models.Label "The created label object."
// @Failure 400 {object} web.HTTPError "Invalid label object provided."
// @Failure 500 {object} models.Message "Internal error"
// @Router /labels [put]
func (l *Label) Create(s *xorm.Session, a web.Auth) (err error) {
	u, err := user.GetFromAuth(a)
	if err != nil {
		return
	}

	l.ID = 0
	l.HexColor = utils.NormalizeHex(l.HexColor)
	l.CreatedBy = u
	l.CreatedByID = u.ID

	_, err = s.Insert(l)
	return
}

// Update updates a label
// @Summary Update a label
// @Description Update an existing label. The user needs to be the creator of the label to be able to do this.
// @tags labels
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "Label ID"
// @Param label body models.Label true "The label object"
// @Success 200 {object} models.Label "The created label object."
// @Failure 400 {object} web.HTTPError "Invalid label object provided."
// @Failure 403 {object} web.HTTPError "Not allowed to update the label."
// @Failure 404 {object} web.HTTPError "Label not found."
// @Failure 500 {object} models.Message "Internal error"
// @Router /labels/{id} [put]
func (l *Label) Update(s *xorm.Session, a web.Auth) (err error) {

	l.HexColor = utils.NormalizeHex(l.HexColor)

	_, err = s.
		ID(l.ID).
		Cols(
			"title",
			"description",
			"hex_color",
		).
		Update(l)
	if err != nil {
		return
	}

	err = l.ReadOne(s, a)
	return
}

// Delete deletes a label
// @Summary Delete a label
// @Description Delete an existing label. The user needs to be the creator of the label to be able to do this.
// @tags labels
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "Label ID"
// @Success 200 {object} models.Label "The label was successfully deleted."
// @Failure 403 {object} web.HTTPError "Not allowed to delete the label."
// @Failure 404 {object} web.HTTPError "Label not found."
// @Failure 500 {object} models.Message "Internal error"
// @Router /labels/{id} [delete]
func (l *Label) Delete(s *xorm.Session, _ web.Auth) (err error) {
	_, err = s.ID(l.ID).Delete(&Label{})
	return err
}

// ReadAll gets all labels a user can use
// @Summary Get all labels a user has access to
// @Description Returns all labels which are either created by the user or associated with a task the user has at least read-access to.
// @tags labels
// @Accept json
// @Produce json
// @Param page query int false "The page number. Used for pagination. If not provided, the first page of results is returned."
// @Param per_page query int false "The maximum number of items per page. Note this parameter is limited by the configured maximum of items per page."
// @Param s query string false "Search labels by label text."
// @Security JWTKeyAuth
// @Success 200 {array} models.Label "The labels"
// @Failure 500 {object} models.Message "Internal error"
// @Router /labels [get]
func (l *Label) ReadAll(s *xorm.Session, a web.Auth, search string, page int, perPage int) (ls interface{}, resultCount int, numberOfEntries int64, err error) {
	return GetLabelsByTaskIDs(s, &LabelByTaskIDsOptions{
		Search:              []string{search},
		User:                a,
		Page:                page,
		PerPage:             perPage,
		GetUnusedLabels:     true,
		GroupByLabelIDsOnly: true,
		GetForUser:          true,
	})
}

// ReadOne gets one label
// @Summary Gets one label
// @Description Returns one label by its ID.
// @tags labels
// @Accept json
// @Produce json
// @Param id path int true "Label ID"
// @Security JWTKeyAuth
// @Success 200 {object} models.Label "The label"
// @Failure 403 {object} web.HTTPError "The user does not have access to the label"
// @Failure 404 {object} web.HTTPError "Label not found"
// @Failure 500 {object} models.Message "Internal error"
// @Router /labels/{id} [get]
func (l *Label) ReadOne(s *xorm.Session, _ web.Auth) (err error) {
	label, err := getLabelByIDSimple(s, l.ID)
	if err != nil {
		return
	}
	*l = *label

	u, err := user.GetUserByID(s, l.CreatedByID)
	if err != nil {
		return
	}

	l.CreatedBy = u
	return
}

func getLabelByIDSimple(s *xorm.Session, labelID int64) (*Label, error) {
	return GetLabelSimple(s, &Label{ID: labelID})
}

func GetLabelSimple(s *xorm.Session, l *Label) (*Label, error) {
	exists, err := s.Get(l)
	if err != nil {
		return l, err
	}
	if !exists {
		return &Label{}, ErrLabelDoesNotExist{l.ID}
	}
	return l, err
}
