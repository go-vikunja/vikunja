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
	"fmt"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"xorm.io/builder"
	"xorm.io/xorm"
)

// WikiPage represents a wiki page or folder within a project
type WikiPage struct {
	// The unique, numeric id of this wiki page.
	ID int64 `xorm:"bigint autoincr not null unique pk" json:"id" param:"page"`
	// The project this wiki page belongs to.
	ProjectID int64 `xorm:"bigint not null INDEX" json:"project_id" param:"project"`
	// The parent page ID. Null if this is a root page.
	ParentID *int64 `xorm:"bigint null INDEX" json:"parent_id"`
	// The title of the wiki page.
	Title string `xorm:"varchar(250) not null" json:"title" valid:"required,runelength(1|250)" minLength:"1" maxLength:"250"`
	// The markdown content of the wiki page.
	Content string `xorm:"longtext null" json:"content"`
	// The full path of the page for easier querying and display.
	Path string `xorm:"varchar(500) not null INDEX" json:"path"`
	// Whether this is a folder (true) or a page (false).
	IsFolder bool `xorm:"bool default false" json:"is_folder"`
	// The position of this page within its parent folder for ordering.
	Position float64 `xorm:"double not null" json:"position"`
	// The user who created this page.
	CreatedByID int64      `xorm:"bigint not null" json:"-"`
	CreatedBy   *user.User `xorm:"-" json:"created_by"`

	// A timestamp when this page was created. You cannot change this value.
	Created time.Time `xorm:"created not null" json:"created"`
	// A timestamp when this page was last updated. You cannot change this value.
	Updated time.Time `xorm:"updated not null" json:"updated"`

	web.CRUDable    `xorm:"-" json:"-"`
	web.Permissions `xorm:"-" json:"-"`
}

// TableName returns the table name for wiki pages
func (wp *WikiPage) TableName() string {
	return "wiki_pages"
}

// GetWikiPageByID returns a wiki page by its ID
func GetWikiPageByID(s *xorm.Session, id int64) (page *WikiPage, err error) {
	page = &WikiPage{}
	exists, err := s.Where("id = ?", id).Get(page)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrWikiPageDoesNotExist{ID: id}
	}

	// Load creator
	page.CreatedBy, err = user.GetUserByID(s, page.CreatedByID)
	if err != nil {
		return nil, err
	}

	return page, nil
}

// ReadOne implements the CRUD interface
func (wp *WikiPage) ReadOne(s *xorm.Session, _ web.Auth) (err error) {
	page, err := GetWikiPageByID(s, wp.ID)
	if err != nil {
		return err
	}
	*wp = *page
	return nil
}

// ReadAll returns all wiki pages for a project
// @Summary Get all wiki pages in a project
// @Description Returns all wiki pages for a project, including folders.
// @tags wiki
// @Accept json
// @Produce json
// @Param project path int true "Project ID"
// @Security JWTKeyAuth
// @Success 200 {array} models.WikiPage "All wiki pages in the project"
// @Failure 500 {object} models.Message "Internal error"
// @Router /projects/{project}/wiki [get]
func (wp *WikiPage) ReadAll(s *xorm.Session, auth web.Auth, search string, page int, perPage int) (result interface{}, resultCount int, numberOfTotalItems int64, err error) {
	// Build the query
	var cond builder.Cond = builder.Eq{"project_id": wp.ProjectID}

	if search != "" {
		cond = builder.And(cond, builder.Or(
			builder.Like{"title", "%" + search + "%"},
			builder.Like{"content", "%" + search + "%"},
		))
	}

	// Get all pages (including children, not just root level)
	pages := []*WikiPage{}
	err = s.Where(cond).
		OrderBy("is_folder DESC, position ASC, created ASC").
		Find(&pages)
	if err != nil {
		return nil, 0, 0, err
	}

	// Load creators
	for _, p := range pages {
		p.CreatedBy, err = user.GetUserByID(s, p.CreatedByID)
		if err != nil {
			return nil, 0, 0, err
		}
	}

	numberOfTotalItems = int64(len(pages))
	return pages, len(pages), numberOfTotalItems, nil
}

// Create implements the CRUD interface
func (wp *WikiPage) Create(s *xorm.Session, auth web.Auth) (err error) {
	// Get the doer
	doer, err := user.GetFromAuth(auth)
	if err != nil {
		return err
	}

	wp.CreatedByID = doer.ID
	wp.CreatedBy = doer

	// Validate parent exists if specified
	if wp.ParentID != nil && *wp.ParentID > 0 {
		parent, err := GetWikiPageByID(s, *wp.ParentID)
		if err != nil {
			return err
		}
		if !parent.IsFolder {
			return ErrWikiPageParentMustBeFolder{ParentID: *wp.ParentID}
		}
		if parent.ProjectID != wp.ProjectID {
			return ErrWikiPageParentProjectMismatch{ParentID: *wp.ParentID, ProjectID: wp.ProjectID}
		}
	}

	// Calculate position if not set
	if wp.Position == 0 {
		wp.Position = calculateDefaultPosition(wp.ID, wp.Position)
	}

	// Build path
	wp.Path = wp.buildPath(s)

	// Validate path uniqueness
	exists, err := s.Where("project_id = ? AND path = ? AND id != ?", wp.ProjectID, wp.Path, wp.ID).
		Exist(&WikiPage{})
	if err != nil {
		return err
	}
	if exists {
		return ErrWikiPagePathNotUnique{Path: wp.Path}
	}

	// Create the page
	_, err = s.Insert(wp)
	if err != nil {
		return err
	}

	// Emit event
	if err := events.Dispatch(&WikiPageCreatedEvent{
		WikiPage: wp,
		Doer:     doer,
	}); err != nil {
		return err
	}

	return nil
}

// Update implements the CRUD interface
func (wp *WikiPage) Update(s *xorm.Session, auth web.Auth) (err error) {
	// Get existing page
	existing, err := GetWikiPageByID(s, wp.ID)
	if err != nil {
		return err
	}

	// Helper to compare nullable parent IDs
	parentIDChanged := func() bool {
		if wp.ParentID == nil && existing.ParentID == nil {
			return false
		}
		if wp.ParentID == nil || existing.ParentID == nil {
			return true
		}
		return *wp.ParentID != *existing.ParentID
	}()

	// Validate parent if changed
	if parentIDChanged {
		if wp.ParentID != nil && *wp.ParentID > 0 {
			parent, err := GetWikiPageByID(s, *wp.ParentID)
			if err != nil {
				return err
			}
			if !parent.IsFolder {
				return ErrWikiPageParentMustBeFolder{ParentID: *wp.ParentID}
			}
			if parent.ProjectID != wp.ProjectID {
				return ErrWikiPageParentProjectMismatch{ParentID: *wp.ParentID, ProjectID: wp.ProjectID}
			}
			// Check for cycles
			if err := wp.checkForCycles(s); err != nil {
				return err
			}
		}
	}

	// Rebuild path if title or parent changed
	if wp.Title != existing.Title || parentIDChanged {
		wp.Path = wp.buildPath(s)

		// Validate path uniqueness
		exists, err := s.Where("project_id = ? AND path = ? AND id != ?", wp.ProjectID, wp.Path, wp.ID).
			Exist(&WikiPage{})
		if err != nil {
			return err
		}
		if exists {
			return ErrWikiPagePathNotUnique{Path: wp.Path}
		}
		
		// Update paths of all descendants if this is a folder
		if existing.IsFolder {
			if err := wp.updateDescendantPaths(s); err != nil {
				return err
			}
		}
	}

	// Update the page
	_, err = s.ID(wp.ID).Update(wp)
	if err != nil {
		return err
	}

	// Get doer for event
	doer, err := user.GetFromAuth(auth)
	if err != nil {
		return err
	}

	// Emit event
	if err := events.Dispatch(&WikiPageUpdatedEvent{
		WikiPage: wp,
		Doer:     doer,
	}); err != nil {
		return err
	}

	return nil
}

// Delete implements the CRUD interface
func (wp *WikiPage) Delete(s *xorm.Session, auth web.Auth) (err error) {
	// Get the page to delete
	page, err := GetWikiPageByID(s, wp.ID)
	if err != nil {
		return err
	}

	// If it's a folder, delete all children recursively
	if page.IsFolder {
		children := []*WikiPage{}
		err = s.Where("parent_id = ?", wp.ID).Find(&children)
		if err != nil {
			return err
		}

		for _, child := range children {
			if err := child.Delete(s, auth); err != nil {
				return err
			}
		}
	}

	// Delete the page itself
	_, err = s.ID(wp.ID).Delete(&WikiPage{})
	if err != nil {
		return err
	}

	// Get doer for event
	doer, err := user.GetFromAuth(auth)
	if err != nil {
		return err
	}

	// Emit event
	if err := events.Dispatch(&WikiPageDeletedEvent{
		WikiPage: page,
		Doer:     doer,
	}); err != nil {
		return err
	}

	return nil
}

// updateDescendantPaths recursively updates paths for all descendants
func (wp *WikiPage) updateDescendantPaths(s *xorm.Session) error {
	// Get all direct children
	var children []*WikiPage
	err := s.Where("parent_id = ?", wp.ID).Find(&children)
	if err != nil {
		return err
	}

	// Update each child's path
	for _, child := range children {
		child.Path = child.buildPath(s)
		
		// Check for uniqueness
		exists, err := s.Where("project_id = ? AND path = ? AND id != ?", child.ProjectID, child.Path, child.ID).
			Exist(&WikiPage{})
		if err != nil {
			return err
		}
		if exists {
			return ErrWikiPagePathNotUnique{Path: child.Path}
		}
		
		// Update in database
		_, err = s.ID(child.ID).Cols("path").Update(child)
		if err != nil {
			return err
		}
		
		// Recursively update its descendants if it's a folder
		if child.IsFolder {
			if err := child.updateDescendantPaths(s); err != nil {
				return err
			}
		}
	}

	return nil
}

// buildPath builds the full path for this page
func (wp *WikiPage) buildPath(s *xorm.Session) string {
	// Simple path sanitization - remove slashes and trim spaces
	cleanTitle := strings.ReplaceAll(strings.TrimSpace(wp.Title), "/", "-")
	
	if wp.ParentID == nil || *wp.ParentID == 0 {
		return fmt.Sprintf("/%s", cleanTitle)
	}

	parent, err := GetWikiPageByID(s, *wp.ParentID)
	if err != nil {
		return fmt.Sprintf("/%s", cleanTitle)
	}

	return fmt.Sprintf("%s/%s", parent.Path, cleanTitle)
}

// checkForCycles checks if moving this page would create a cycle
func (wp *WikiPage) checkForCycles(s *xorm.Session) error {
	if wp.ParentID == nil || *wp.ParentID == 0 {
		return nil
	}

	visited := make(map[int64]bool)
	visited[wp.ID] = true

	currentParentID := *wp.ParentID
	for currentParentID != 0 {
		if visited[currentParentID] {
			return ErrWikiPageCyclicRelationship{PageID: wp.ID}
		}
		visited[currentParentID] = true

		parent, err := GetWikiPageByID(s, currentParentID)
		if err != nil {
			return err
		}

		if parent.ParentID == nil {
			break
		}
		currentParentID = *parent.ParentID
	}

	return nil
}

// CanRead checks if a user can read this wiki page (delegates to project permissions)
func (wp *WikiPage) CanRead(s *xorm.Session, a web.Auth) (bool, int, error) {
	project := &Project{ID: wp.ProjectID}
	return project.CanRead(s, a)
}

// CanWrite checks if a user can write to this wiki page (delegates to project permissions)
func (wp *WikiPage) CanWrite(s *xorm.Session, a web.Auth) (bool, error) {
	project := &Project{ID: wp.ProjectID}
	return project.CanWrite(s, a)
}

// CanCreate checks if a user can create wiki pages in this project (delegates to project permissions)
func (wp *WikiPage) CanCreate(s *xorm.Session, a web.Auth) (bool, error) {
	project := &Project{ID: wp.ProjectID}
	return project.CanWrite(s, a)
}

// CanDelete checks if a user can delete this wiki page (delegates to project permissions)
func (wp *WikiPage) CanDelete(s *xorm.Session, a web.Auth) (bool, error) {
	project := &Project{ID: wp.ProjectID}
	return project.CanWrite(s, a)
}

// CanUpdate checks if a user can update this wiki page (delegates to project permissions)
func (wp *WikiPage) CanUpdate(s *xorm.Session, a web.Auth) (bool, error) {
	project := &Project{ID: wp.ProjectID}
	return project.CanWrite(s, a)
}
