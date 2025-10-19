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

	"code.vikunja.io/api/pkg/web"
	"xorm.io/xorm"
)

// LabelServiceProvider is a function type that returns a label service instance
// This is used to avoid import cycles between models and services packages
type LabelServiceProvider func() interface {
	Create(s *xorm.Session, label *Label, u *user.User) error
	Update(s *xorm.Session, label *Label, u *user.User) error
	Delete(s *xorm.Session, label *Label, u *user.User) error
	GetAll(s *xorm.Session, u *user.User, search string, page int, perPage int) (interface{}, int, int64, error)
	GetByID(s *xorm.Session, labelID int64) (*Label, error)
}

// labelServiceProvider is the registered service provider function
var labelServiceProvider LabelServiceProvider

// RegisterLabelService registers a service provider for label operations
// This should be called during application initialization by the services package
func RegisterLabelService(provider LabelServiceProvider) {
	labelServiceProvider = provider
}

// getLabelService returns the registered label service instance
func getLabelService() interface {
	Create(s *xorm.Session, label *Label, u *user.User) error
	Update(s *xorm.Session, label *Label, u *user.User) error
	Delete(s *xorm.Session, label *Label, u *user.User) error
	GetAll(s *xorm.Session, u *user.User, search string, page int, perPage int) (interface{}, int, int64, error)
	GetByID(s *xorm.Session, labelID int64) (*Label, error)
} {
	if labelServiceProvider == nil {
		panic("LabelService not registered - did you forget to call services.InitializeDependencies()?")
	}
	return labelServiceProvider()
}

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
// @Deprecated: This method is deprecated and will be removed in a future release. Use LabelService.Create instead.
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
	return getLabelService().Create(s, l, u)
}

// Update updates a label
// @Deprecated: This method is deprecated and will be removed in a future release. Use LabelService.Update instead.
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
	u, err := user.GetFromAuth(a)
	if err != nil {
		return
	}
	return getLabelService().Update(s, l, u)
}

// Delete deletes a label
// @Deprecated: This method is deprecated and will be removed in a future release. Use LabelService.Delete instead.
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
func (l *Label) Delete(s *xorm.Session, a web.Auth) (err error) {
	u, err := user.GetFromAuth(a)
	if err != nil {
		return
	}
	return getLabelService().Delete(s, l, u)
}

// ReadAll gets all labels a user can use
// @Deprecated: This method is deprecated and will be removed in a future release. Use LabelService.GetAll instead.
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
	u, err := user.GetFromAuth(a)
	if err != nil {
		return nil, 0, 0, err
	}
	return getLabelService().GetAll(s, u, search, page, perPage)
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
	ls := getLabelService()
	label, err := ls.GetByID(s, l.ID)
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

// ===== Permission Methods =====
// These methods delegate to the service layer via function pointers

// CanRead checks if the user can read a label
func (l *Label) CanRead(s *xorm.Session, a web.Auth) (bool, int, error) {
	if CheckLabelReadFunc == nil {
		return false, 0, ErrPermissionDelegationNotInitialized{}
	}
	return CheckLabelReadFunc(s, l.ID, a)
}

// CanWrite checks if the user can write to a label
func (l *Label) CanWrite(s *xorm.Session, a web.Auth) (bool, error) {
	if CheckLabelWriteFunc == nil {
		return false, ErrPermissionDelegationNotInitialized{}
	}
	return CheckLabelWriteFunc(s, l.ID, a)
}

// CanUpdate checks if the user can update a label
func (l *Label) CanUpdate(s *xorm.Session, a web.Auth) (bool, error) {
	if CheckLabelUpdateFunc == nil {
		return false, ErrPermissionDelegationNotInitialized{}
	}
	return CheckLabelUpdateFunc(s, l.ID, a)
}

// CanDelete checks if the user can delete a label
func (l *Label) CanDelete(s *xorm.Session, a web.Auth) (bool, error) {
	if CheckLabelDeleteFunc == nil {
		return false, ErrPermissionDelegationNotInitialized{}
	}
	return CheckLabelDeleteFunc(s, l.ID, a)
}

// CanCreate checks if the user can create a label
func (l *Label) CanCreate(s *xorm.Session, a web.Auth) (bool, error) {
	if CheckLabelCreateFunc == nil {
		return false, ErrPermissionDelegationNotInitialized{}
	}
	return CheckLabelCreateFunc(s, l, a)
}
