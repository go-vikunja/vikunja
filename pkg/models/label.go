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

package models

import (
	"code.vikunja.io/api/pkg/timeutil"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/web"
)

// Label represents a label
type Label struct {
	// The unique, numeric id of this label.
	ID int64 `xorm:"int(11) autoincr not null unique pk" json:"id" param:"label"`
	// The title of the lable. You'll see this one on tasks associated with it.
	Title string `xorm:"varchar(250) not null" json:"title" valid:"runelength(3|250)" minLength:"3" maxLength:"250"`
	// The label description.
	Description string `xorm:"longtext null" json:"description"`
	// The color this label has
	HexColor string `xorm:"varchar(6) null" json:"hex_color" valid:"runelength(0|6)" maxLength:"6"`

	CreatedByID int64 `xorm:"int(11) not null" json:"-"`
	// The user who created this label
	CreatedBy *user.User `xorm:"-" json:"created_by"`

	// A timestamp when this label was created. You cannot change this value.
	Created timeutil.TimeStamp `xorm:"created not null" json:"created"`
	// A timestamp when this label was last updated. You cannot change this value.
	Updated timeutil.TimeStamp `xorm:"updated not null" json:"updated"`

	web.CRUDable `xorm:"-" json:"-"`
	web.Rights   `xorm:"-" json:"-"`
}

// TableName makes a pretty table name
func (Label) TableName() string {
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
// @Success 200 {object} models.Label "The created label object."
// @Failure 400 {object} code.vikunja.io/web.HTTPError "Invalid label object provided."
// @Failure 500 {object} models.Message "Internal error"
// @Router /labels [put]
func (l *Label) Create(a web.Auth) (err error) {
	u, err := user.GetFromAuth(a)
	if err != nil {
		return
	}

	l.CreatedBy = u
	l.CreatedByID = u.ID

	_, err = x.Insert(l)
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
// @Failure 400 {object} code.vikunja.io/web.HTTPError "Invalid label object provided."
// @Failure 403 {object} code.vikunja.io/web.HTTPError "Not allowed to update the label."
// @Failure 404 {object} code.vikunja.io/web.HTTPError "Label not found."
// @Failure 500 {object} models.Message "Internal error"
// @Router /labels/{id} [put]
func (l *Label) Update() (err error) {
	_, err = x.
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

	err = l.ReadOne()
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
// @Failure 403 {object} code.vikunja.io/web.HTTPError "Not allowed to delete the label."
// @Failure 404 {object} code.vikunja.io/web.HTTPError "Label not found."
// @Failure 500 {object} models.Message "Internal error"
// @Router /labels/{id} [delete]
func (l *Label) Delete() (err error) {
	_, err = x.ID(l.ID).Delete(&Label{})
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
func (l *Label) ReadAll(a web.Auth, search string, page int, perPage int) (ls interface{}, resultCount int, numberOfEntries int64, err error) {
	if _, is := a.(*LinkSharing); is {
		return nil, 0, 0, ErrGenericForbidden{}
	}

	u := &user.User{ID: a.GetID()}

	// Get all tasks
	taskIDs, err := getUserTaskIDs(u)
	if err != nil {
		return nil, 0, 0, err
	}

	return getLabelsByTaskIDs(&LabelByTaskIDsOptions{
		Search:              search,
		User:                u,
		TaskIDs:             taskIDs,
		Page:                page,
		PerPage:             perPage,
		GetUnusedLabels:     true,
		GroupByLabelIDsOnly: true,
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
// @Failure 403 {object} code.vikunja.io/web.HTTPError "The user does not have access to the label"
// @Failure 404 {object} code.vikunja.io/web.HTTPError "Label not found"
// @Failure 500 {object} models.Message "Internal error"
// @Router /labels/{id} [get]
func (l *Label) ReadOne() (err error) {
	label, err := getLabelByIDSimple(l.ID)
	if err != nil {
		return err
	}
	*l = *label

	user, err := user.GetUserByID(l.CreatedByID)
	if err != nil {
		return err
	}

	l.CreatedBy = user
	return
}

func getLabelByIDSimple(labelID int64) (*Label, error) {
	label := Label{}
	exists, err := x.ID(labelID).Get(&label)
	if err != nil {
		return &label, err
	}

	if !exists {
		return &Label{}, ErrLabelDoesNotExist{labelID}
	}
	return &label, err
}

// Helper method to get all task ids a user has
func getUserTaskIDs(u *user.User) (taskIDs []int64, err error) {

	// Get all lists
	lists, _, _, err := getRawListsForUser(&listOptions{
		user: u,
		page: -1,
	})
	if err != nil {
		return nil, err
	}

	tasks, _, _, err := getRawTasksForLists(lists, &taskOptions{
		page:    -1,
		perPage: 0,
	})
	if err != nil {
		return nil, err
	}

	// make a slice of task ids
	for _, t := range tasks {
		taskIDs = append(taskIDs, t.ID)
	}

	return
}
