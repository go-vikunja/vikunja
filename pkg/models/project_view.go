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
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"github.com/danielgtaylor/huma/v2"
	"xorm.io/builder"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

type ProjectViewKind int

func (p *ProjectViewKind) MarshalJSON() ([]byte, error) {
	switch *p {
	case ProjectViewKindList:
		return []byte(`"list"`), nil
	case ProjectViewKindGantt:
		return []byte(`"gantt"`), nil
	case ProjectViewKindTable:
		return []byte(`"table"`), nil
	case ProjectViewKindKanban:
		return []byte(`"kanban"`), nil
	}

	return []byte(`null`), nil
}

func (p *ProjectViewKind) UnmarshalJSON(bytes []byte) error {
	var value string
	err := json.Unmarshal(bytes, &value)
	if err != nil {
		return err
	}

	switch value {
	case "list":
		*p = ProjectViewKindList
	case "gantt":
		*p = ProjectViewKindGantt
	case "table":
		*p = ProjectViewKindTable
	case "kanban":
		*p = ProjectViewKindKanban
	default:
		return fmt.Errorf("unknown project view kind: %s", value)
	}

	return nil
}

// Schema lets Huma (/api/v2) reflect this type as a string enum. The custom
// Marshal/UnmarshalJSON above serialize it as a string, but the underlying Go
// type is an int — without this, Huma would generate an integer schema and
// reject the string form clients actually send.
func (*ProjectViewKind) Schema(_ huma.Registry) *huma.Schema {
	return &huma.Schema{
		Type: "string",
		Enum: []any{"list", "gantt", "table", "kanban"},
	}
}

// NOTE: When adding or changing enum values for ProjectViewKind,
// make sure to update the corresponding `enums` tag in the ProjectView struct
// to keep the OpenAPI documentation in sync.

const (
	ProjectViewKindList ProjectViewKind = iota
	ProjectViewKindGantt
	ProjectViewKindTable
	ProjectViewKindKanban
)

type BucketConfigurationModeKind int

// NOTE: When adding or changing enum values for BucketConfigurationModeKind,
// make sure to update the corresponding `enums` tag in the ProjectView struct
// to keep the OpenAPI documentation in sync.

const (
	BucketConfigurationModeNone BucketConfigurationModeKind = iota
	BucketConfigurationModeManual
	BucketConfigurationModeFilter
)

func (p *BucketConfigurationModeKind) MarshalJSON() ([]byte, error) {
	switch *p {
	case BucketConfigurationModeNone:
		return []byte(`"none"`), nil
	case BucketConfigurationModeManual:
		return []byte(`"manual"`), nil
	case BucketConfigurationModeFilter:
		return []byte(`"filter"`), nil
	}

	return []byte(`null`), nil
}

func (p *BucketConfigurationModeKind) UnmarshalJSON(bytes []byte) error {
	var value string
	err := json.Unmarshal(bytes, &value)
	if err != nil {
		return err
	}

	switch value {
	case "none":
		*p = BucketConfigurationModeNone
	case "manual":
		*p = BucketConfigurationModeManual
	case "filter":
		*p = BucketConfigurationModeFilter
	default:
		return fmt.Errorf("unknown bucket configuration mode kind: %s", value)
	}

	return nil
}

// Schema lets Huma (/api/v2) reflect this type as a string enum; see the note
// on ProjectViewKind.Schema for why this is needed.
func (*BucketConfigurationModeKind) Schema(_ huma.Registry) *huma.Schema {
	return &huma.Schema{
		Type: "string",
		Enum: []any{"none", "manual", "filter"},
	}
}

type ProjectViewBucketConfiguration struct {
	Title  string          `json:"title" doc:"The title of the bucket this configuration creates."`
	Filter *TaskCollection `json:"filter" doc:"The filter query that decides which tasks land in this bucket. See https://vikunja.io/docs/filters."`
}

type ProjectView struct {
	// The unique numeric id of this view
	ID int64 `xorm:"autoincr not null unique pk" json:"id" param:"view" readOnly:"true" doc:"The unique, numeric id of this view. Set by the server."`
	// The title of this view
	Title string `xorm:"varchar(255) not null" json:"title" valid:"required,runelength(1|250)" minLength:"1" maxLength:"250" doc:"The title of this view."`
	// The project this view belongs to
	ProjectID int64 `xorm:"not null index" json:"project_id" param:"project" readOnly:"true" doc:"The project this view belongs to. Taken from the URL path; ignored on write."`
	// The kind of this view. Can be `list`, `gantt`, `table` or `kanban`.
	ViewKind ProjectViewKind `xorm:"not null" json:"view_kind" swaggertype:"string" enums:"list,gantt,table,kanban" doc:"The kind of this view. One of list, gantt, table or kanban."`

	// The filter query to match tasks by. Check out https://vikunja.io/docs/filters for a full explanation.
	Filter *TaskCollection `xorm:"json null default null" query:"filter" json:"filter" doc:"The filter query used to match tasks shown in this view. See https://vikunja.io/docs/filters."`
	// The position of this view in the list. The list of all views will be sorted by this parameter.
	Position float64 `xorm:"double null" json:"position" doc:"The position of this view in the project's list of views. Views are sorted ascending by this value."`

	// The bucket configuration mode. Can be `none`, `manual` or `filter`. `manual` allows to move tasks between buckets as you normally would. `filter` creates buckets based on a filter for each bucket.
	BucketConfigurationMode BucketConfigurationModeKind `xorm:"default 0" json:"bucket_configuration_mode" swaggertype:"string" enums:"none,manual,filter" doc:"The bucket configuration mode. One of none, manual or filter. manual lets you move tasks between buckets; filter creates a bucket per filter."`
	// When the bucket configuration mode is not `manual`, this field holds the options of that configuration.
	BucketConfiguration []*ProjectViewBucketConfiguration `xorm:"json" json:"bucket_configuration" doc:"When the bucket configuration mode is filter, holds the title and filter of each bucket."`
	// The ID of the bucket where new tasks without a bucket are added to. By default, this is the leftmost bucket in a view.
	DefaultBucketID int64 `xorm:"bigint INDEX null" json:"default_bucket_id" doc:"The id of the bucket new tasks without a bucket are added to. Defaults to the leftmost bucket."`
	// If tasks are moved to the done bucket, they are marked as done. If they are marked as done individually, they are moved into the done bucket.
	DoneBucketID int64 `xorm:"bigint INDEX null" json:"done_bucket_id" doc:"The id of the done bucket. Tasks moved here are marked done, and tasks marked done are moved here."`

	// A timestamp when this view was updated. You cannot change this value.
	Updated time.Time `xorm:"updated not null" json:"updated" readOnly:"true" doc:"A timestamp when this view was last updated. You cannot change this value."`
	// A timestamp when this reaction was created. You cannot change this value.
	Created time.Time `xorm:"created not null" json:"created" readOnly:"true" doc:"A timestamp when this view was created. You cannot change this value."`

	web.CRUDable    `xorm:"-" json:"-"`
	web.Permissions `xorm:"-" json:"-"`
}

func (pv *ProjectView) TableName() string {
	return "project_views"
}

func getViewsForProject(s *xorm.Session, projectID int64) (views []*ProjectView, err error) {
	views = []*ProjectView{}
	err = s.
		Where("project_id = ?", projectID).
		OrderBy("position asc").
		Find(&views)
	return
}

// ReadAll gets all project views
// @Summary Get all project views for a project
// @Description Returns all project views for a sepcific project
// @tags project
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param project path int true "Project ID"
// @Success 200 {array} models.ProjectView "The project views"
// @Failure 500 {object} models.Message "Internal error"
// @Router /projects/{project}/views [get]
func (pv *ProjectView) ReadAll(s *xorm.Session, a web.Auth, _ string, _ int, _ int) (result interface{}, resultCount int, numberOfTotalItems int64, err error) {

	pp := &Project{ID: pv.ProjectID}
	can, _, err := pp.CanRead(s, a)
	if err != nil {
		return nil, 0, 0, err
	}
	if !can {
		return nil, 0, 0, ErrGenericForbidden{}
	}

	projectViews, err := getViewsForProject(s, pv.ProjectID)
	if err != nil {
		return nil, 0, 0, err
	}

	totalCount, err := s.
		Where("project_id = ?", pv.ProjectID).
		Count(&ProjectView{})
	if err != nil {
		return
	}

	return projectViews, len(projectViews), totalCount, nil
}

// ReadOne implements the CRUD method to get one project view
// @Summary Get one project view
// @Description Returns a project view by its ID.
// @tags project
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param project path int true "Project ID"
// @Param id path int true "Project View ID"
// @Success 200 {object} models.ProjectView "The project view"
// @Failure 403 {object} web.HTTPError "The user does not have access to this project view"
// @Failure 500 {object} models.Message "Internal error"
// @Router /projects/{project}/views/{id} [get]
func (pv *ProjectView) ReadOne(s *xorm.Session, _ web.Auth) (err error) {
	view, err := GetProjectViewByIDAndProject(s, pv.ID, pv.ProjectID)
	if err != nil {
		return err
	}

	*pv = *view
	return
}

// Delete removes the project view
// @Summary Delete a project view
// @Description Deletes a project view.
// @tags project
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param project path int true "Project ID"
// @Param id path int true "Project View ID"
// @Success 200 {object} models.Message "The project view was successfully deleted."
// @Failure 403 {object} web.HTTPError "The user does not have access to the project view"
// @Failure 500 {object} models.Message "Internal error"
// @Router /projects/{project}/views/{id} [delete]
func (pv *ProjectView) Delete(s *xorm.Session, _ web.Auth) (err error) {
	// Resolve the view under the path project first: the buckets/positions below are deleted by
	// view id alone, so without this guard a delete scoped to the wrong parent project would still
	// wipe another project's buckets and positions while matching zero project_views rows.
	if _, err = GetProjectViewByIDAndProject(s, pv.ID, pv.ProjectID); err != nil {
		return err
	}

	_, err = s.
		Where("id = ? AND project_id = ?", pv.ID, pv.ProjectID).
		Delete(&ProjectView{})
	if err != nil {
		return
	}

	_, err = s.Where("project_view_id = ?", pv.ID).Delete(&TaskBucket{})
	if err != nil {
		return
	}

	_, err = s.Where("project_view_id = ?", pv.ID).Delete(&TaskPosition{})
	return
}

// Create adds a new project view
// @Summary Create a project view
// @Description Create a project view in a specific project.
// @tags project
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param project path int true "Project ID"
// @Param view body models.ProjectView true "The project view you want to create."
// @Success 201 {object} models.ProjectView "The created project view"
// @Failure 403 {object} web.HTTPError "The user does not have access to create a project view"
// @Failure 500 {object} models.Message "Internal error"
// @Router /projects/{project}/views [put]
func (pv *ProjectView) Create(s *xorm.Session, a web.Auth) (err error) {
	return createProjectView(s, pv, a, true, true)
}

func validateProjectViewFilters(p *ProjectView) (err error) {
	if p.Filter != nil && p.Filter.Filter != "" {
		_, err = getTaskFiltersFromFilterString(p.Filter.Filter, p.Filter.FilterTimezone)
		if err != nil {
			return
		}
	}

	if p.BucketConfigurationMode == BucketConfigurationModeFilter {
		for _, configuration := range p.BucketConfiguration {
			if configuration == nil {
				continue
			}
			if configuration.Filter != nil && configuration.Filter.Filter != "" {
				_, err = getTaskFiltersFromFilterString(configuration.Filter.Filter, configuration.Filter.FilterTimezone)
				if err != nil {
					return
				}
			}
		}
	}

	return
}

func normalizeBucketConfigurationMode(p *ProjectView) {
	if p.ViewKind != ProjectViewKindKanban {
		p.BucketConfigurationMode = BucketConfigurationModeNone
		return
	}

	if p.BucketConfigurationMode == BucketConfigurationModeNone {
		p.BucketConfigurationMode = BucketConfigurationModeManual
	}
}

// bucket_configuration_mode is optional in v1, so an omitted mode must keep the
// stored one instead of converting a filter view to manual and mass-seeding it.
func normalizeBucketConfigurationModeOnUpdate(p, oldView *ProjectView) {
	if p.ViewKind == ProjectViewKindKanban &&
		p.BucketConfigurationMode == BucketConfigurationModeNone &&
		oldView.ViewKind == ProjectViewKindKanban &&
		oldView.BucketConfigurationMode != BucketConfigurationModeNone {
		p.BucketConfigurationMode = oldView.BucketConfigurationMode
		// Inheriting the mode without the configuration would leave a filter view
		// with zero buckets.
		if p.BucketConfigurationMode == BucketConfigurationModeFilter && len(p.BucketConfiguration) == 0 {
			p.BucketConfiguration = oldView.BucketConfiguration
		}
		return
	}

	normalizeBucketConfigurationMode(p)
}

func resolveBucketIDs(s *xorm.Session, p, oldView *ProjectView) (err error) {
	p.DefaultBucketID, err = resolveBucketID(s, p.ID, p.DefaultBucketID, oldView.DefaultBucketID)
	if err != nil {
		return err
	}

	p.DoneBucketID, err = resolveBucketID(s, p.ID, p.DoneBucketID, oldView.DoneBucketID)
	return err
}

func resolveBucketID(s *xorm.Session, viewID, incomingID, storedID int64) (int64, error) {
	if incomingID == 0 {
		return 0, nil
	}

	bucketViewID, exists, err := bucketProjectViewID(s, incomingID)
	if err != nil {
		return 0, err
	}
	if exists && bucketViewID == viewID {
		return incomingID, nil
	}

	// Rejecting an id the frontend merely echoed back - of a bucket that is gone,
	// or of the stale stored value - would lock the view until a reload.
	if !exists || incomingID == storedID {
		return 0, nil
	}

	return 0, ErrBucketDoesNotBelongToProjectView{
		BucketID:      incomingID,
		ProjectViewID: viewID,
	}
}

func bucketProjectViewID(s *xorm.Session, bucketID int64) (viewID int64, exists bool, err error) {
	exists, err = s.
		Table(&Bucket{}).
		Where(builder.Eq{"id": bucketID}).
		Cols("project_view_id").
		Get(&viewID)
	return
}

func createProjectView(s *xorm.Session, p *ProjectView, a web.Auth, createBacklogBucket bool, addExistingTasksToView bool) (err error) {
	normalizeBucketConfigurationMode(p)

	err = validateProjectViewFilters(p)
	if err != nil {
		return
	}

	p.ID = 0
	// No bucket can belong to a view which does not exist yet.
	p.DefaultBucketID = 0
	p.DoneBucketID = 0
	_, err = s.Insert(p)
	if err != nil {
		return
	}

	p.Position = calculateDefaultPosition(p.ID, p.Position)
	_, err = s.Where("id = ?", p.ID).Update(p)
	if err != nil {
		return
	}

	if p.ViewKind == ProjectViewKindKanban && createBacklogBucket && p.BucketConfigurationMode == BucketConfigurationModeManual {
		err = createDefaultKanbanBuckets(s, a, p, addExistingTasksToView)
		if err != nil {
			return
		}
	}

	if addExistingTasksToView {
		return RecalculateTaskPositions(s, p, a)
	}

	return
}

func createDefaultKanbanBuckets(s *xorm.Session, a web.Auth, p *ProjectView, addExistingTasksToView bool) (err error) {
	backlog := &Bucket{
		ProjectViewID: p.ID,
		Title:         "To-Do",
		Position:      100,
	}
	err = backlog.Create(s, a)
	if err != nil {
		return
	}

	doing := &Bucket{
		ProjectViewID: p.ID,
		Title:         "Doing",
		Position:      200,
	}
	err = doing.Create(s, a)
	if err != nil {
		return
	}

	done := &Bucket{
		ProjectViewID: p.ID,
		Title:         "Done",
		Position:      300,
	}
	err = done.Create(s, a)
	if err != nil {
		return
	}

	p.DefaultBucketID = backlog.ID
	p.DoneBucketID = done.ID
	_, err = s.ID(p.ID).Cols("default_bucket_id", "done_bucket_id").Update(p)
	if err != nil {
		return
	}

	if addExistingTasksToView {
		_, err = addTasksToView(s, p, backlog.ID)
		if err != nil {
			return
		}
	}

	return
}

func addTasksToView(s *xorm.Session, pv *ProjectView, bucketID int64) (added int, err error) {
	taskIDs, err := tasksWithoutBucketInView(s, pv)
	if err != nil {
		return 0, err
	}

	if err = insertTaskBuckets(s, pv.ID, bucketID, taskIDs); err != nil {
		return 0, err
	}

	return len(taskIDs), nil
}

// insertTaskBuckets places tasks in a bucket, keeping rows whose
// (task_id, project_view_id) pair already exists. Callers determine the missing
// tasks with a separate select, so a concurrent transaction can place one of
// them in the meantime; the existing row is then the newer state and a plain
// insert would fail on the unique index.
func insertTaskBuckets(s *xorm.Session, viewID, bucketID int64, taskIDs []int64) (err error) {
	// Keep statements well below the parameter limits of all supported databases.
	const batchSize = 100

	var onConflict string
	if db.Type() == schemas.MYSQL {
		// The self-assignment is a no-op update which only ignores duplicate
		// keys, unlike INSERT IGNORE which would swallow unrelated errors too.
		onConflict = "ON DUPLICATE KEY UPDATE bucket_id = bucket_id"
	} else {
		onConflict = "ON CONFLICT (task_id, project_view_id) DO NOTHING"
	}

	// Raw SQL bypasses xorm's bean-based table-name handling, so qualify the
	// table ourselves to honor a configured postgres schema (database.schema).
	table := s.Engine().TableName(&TaskBucket{}, true)

	for start := 0; start < len(taskIDs); start += batchSize {
		batch := taskIDs[start:min(start+batchSize, len(taskIDs))]

		placeholders := make([]string, 0, len(batch))
		args := make([]interface{}, 1, len(batch)*3+1)
		for _, taskID := range batch {
			placeholders = append(placeholders, "(?, ?, ?)")
			args = append(args, taskID, bucketID, viewID)
		}

		args[0] = "INSERT INTO " + table + " (task_id, bucket_id, project_view_id) VALUES " + strings.Join(placeholders, ", ") + " " + onConflict

		_, err = s.Exec(args...)
		if err != nil {
			return err
		}
	}

	return nil
}

func tasksWithoutBucketInView(s *xorm.Session, pv *ProjectView) (taskIDs []int64, err error) {
	if filterID := GetSavedFilterIDFromProjectID(pv.ProjectID); filterID > 0 {
		sf, err := GetSavedFilterSimpleByID(s, filterID)
		if err != nil {
			return nil, err
		}

		return filteredTasksWithoutBucketInView(s, pv.ID, sf)
	}

	return taskIDsWithoutBucketInView(s, pv.ID, builder.Eq{"tasks.project_id": pv.ProjectID}, false)
}

// Saved filter views have no project of their own, their task set only exists
// once the filter is evaluated.
func filteredTasksWithoutBucketInView(s *xorm.Session, viewID int64, sf *SavedFilter) (taskIDs []int64, err error) {
	parsedFilters, err := getTaskFiltersFromFilterString(sf.Filters.Filter, sf.Filters.FilterTimezone)
	if err != nil {
		return nil, err
	}

	// convertFiltersToDBFilterCond rewrites the field names in place, so this has
	// to be answered before it runs.
	joinTaskBuckets := hasBucketIDInParsedFilter(parsedFilters)

	filterCond, err := convertFiltersToDBFilterCond(parsedFilters, sf.Filters.FilterIncludeNulls)
	if err != nil {
		return nil, err
	}

	// The filter alone matches tasks across the whole instance, restrict it to
	// what its owner can actually see.
	projects, err := getRelevantProjectsFromCollection(s, &user.User{ID: sf.OwnerID}, &TaskCollection{})
	if err != nil {
		return nil, err
	}
	projectIDs := make([]int64, 0, len(projects))
	for _, p := range projects {
		projectIDs = append(projectIDs, p.ID)
	}

	return taskIDsWithoutBucketInView(s, viewID, builder.And(
		builder.In("tasks.project_id", projectIDs),
		filterCond,
	), joinTaskBuckets)
}

// Selecting ids instead of beans means xorm adds no soft-delete condition of its own.
func taskIDsWithoutBucketInView(s *xorm.Session, viewID int64, cond builder.Cond, joinTaskBuckets bool) (taskIDs []int64, err error) {
	taskIDs = []int64{}
	query := s.Table("tasks")
	// convertFiltersToDBFilterCond resolves bucket_id against task_buckets, which
	// is only in scope through this join.
	if joinTaskBuckets {
		query = query.Join("LEFT", "task_buckets", "task_buckets.task_id = tasks.id AND task_buckets.project_view_id = ?", viewID)
	}

	err = query.
		Where(cond).
		And(taskNotDeletedCond("tasks")).
		And(builder.NotIn("tasks.id", builder.
			Select("task_id").
			From("task_buckets").
			Where(builder.Eq{"project_view_id": viewID}))).
		Cols("tasks.id").
		Find(&taskIDs)
	return
}

// Update is the handler to update a project view
// @Summary Updates a project view
// @Description Updates a project view.
// @tags project
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param project path int true "Project ID"
// @Param id path int true "Project View ID"
// @Param view body models.ProjectView true "The project view with updated values you want to change."
// @Success 200 {object} models.ProjectView "The updated project view."
// @Failure 400 {object} web.HTTPError "Invalid project view object provided."
// @Failure 500 {object} models.Message "Internal error"
// @Router /projects/{project}/views/{id} [post]
func (pv *ProjectView) Update(s *xorm.Session, a web.Auth) (err error) {
	oldView, err := GetProjectViewByIDAndProject(s, pv.ID, pv.ProjectID)
	if err != nil {
		return
	}

	normalizeBucketConfigurationModeOnUpdate(pv, oldView)

	err = validateProjectViewFilters(pv)
	if err != nil {
		return
	}

	err = resolveBucketIDs(s, pv, oldView)
	if err != nil {
		return
	}

	cols := []string{
		"title",
		"view_kind",
		"filter",
		"position",
		"bucket_configuration_mode",
		"bucket_configuration",
	}
	// A view without buckets has no business writing bucket ids - keeping the
	// stored ones is what lets healBucketIDs restore them on the way back.
	if pv.ViewKind == ProjectViewKindKanban {
		cols = append(cols, "default_bucket_id", "done_bucket_id")
	}

	_, err = s.
		ID(pv.ID).
		Cols(cols...).
		Update(pv)
	if err != nil {
		return
	}

	return syncManualKanbanBuckets(s, a, pv, oldView)
}

func syncManualKanbanBuckets(s *xorm.Session, a web.Auth, pv, oldView *ProjectView) (err error) {
	// Pseudo views like favorites only exist in memory, seeding would write
	// buckets on a view id shared by every user.
	becameManualKanban := pv.ID > 0 &&
		pv.ViewKind == ProjectViewKindKanban &&
		pv.BucketConfigurationMode == BucketConfigurationModeManual &&
		(oldView.ViewKind != ProjectViewKindKanban || oldView.BucketConfigurationMode != BucketConfigurationModeManual)
	if !becameManualKanban {
		return
	}

	// Buckets survive switching the view kind away from kanban and back, only
	// seed defaults when there are none to avoid duplicates.
	hasBuckets, err := s.Where(builder.Eq{"project_view_id": pv.ID}).Exist(&Bucket{})
	if err != nil {
		return
	}

	if hasBuckets {
		err = healBucketIDs(s, pv, oldView)
	} else {
		err = createDefaultKanbanBuckets(s, a, pv, false)
	}
	if err != nil {
		return
	}

	// Tasks created while the view was not a kanban view are not in any bucket yet
	added, err := addTasksToView(s, pv, pv.DefaultBucketID)
	if err != nil || added == 0 {
		return err
	}

	return RecalculateTaskPositions(s, pv, a)
}

// Clients may omit both ids, restore them from the previous state unless the
// bucket is gone by now.
func healBucketIDs(s *xorm.Session, pv, oldView *ProjectView) (err error) {
	if pv.DefaultBucketID == 0 {
		pv.DefaultBucketID, err = existingBucketID(s, pv.ID, oldView.DefaultBucketID)
		if err != nil {
			return
		}
	}
	if pv.DefaultBucketID == 0 {
		pv.DefaultBucketID, err = getDefaultBucketID(s, pv)
		if err != nil {
			return
		}
	}
	if pv.DoneBucketID == 0 {
		pv.DoneBucketID, err = existingBucketID(s, pv.ID, oldView.DoneBucketID)
		if err != nil {
			return
		}
	}

	_, err = s.ID(pv.ID).Cols("default_bucket_id", "done_bucket_id").Update(pv)
	return
}

func existingBucketID(s *xorm.Session, viewID, bucketID int64) (int64, error) {
	if bucketID == 0 {
		return 0, nil
	}

	bucketViewID, exists, err := bucketProjectViewID(s, bucketID)
	if err != nil || !exists || bucketViewID != viewID {
		return 0, err
	}

	return bucketID, nil
}

func GetProjectViewByIDAndProject(s *xorm.Session, viewID, projectID int64) (view *ProjectView, err error) {
	if projectID == FavoritesPseudoProjectID && viewID < 0 {
		for _, v := range FavoritesPseudoProject.Views {
			if v.ID == viewID {
				return v, nil
			}
		}

		return nil, &ErrProjectViewDoesNotExist{
			ProjectViewID: viewID,
		}
	}

	view = &ProjectView{}
	exists, err := s.
		Where("id = ? AND project_id = ?", viewID, projectID).
		NoAutoCondition().
		Get(view)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, &ErrProjectViewDoesNotExist{
			ProjectViewID: viewID,
		}
	}

	return
}

func GetProjectViewByID(s *xorm.Session, id int64) (view *ProjectView, err error) {
	view = &ProjectView{}
	exists, err := s.
		Where("id = ?", id).
		NoAutoCondition().
		Get(view)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, &ErrProjectViewDoesNotExist{
			ProjectViewID: id,
		}
	}

	return
}

func CreateDefaultViewsForProject(s *xorm.Session, project *Project, a web.Auth, createBacklogBucket bool, createDefaultListFilter bool) (err error) {
	list := &ProjectView{
		ProjectID: project.ID,
		Title:     "List",
		ViewKind:  ProjectViewKindList,
		Position:  100,
	}
	if createDefaultListFilter {
		list.Filter = &TaskCollection{
			Filter: "done = false",
		}
	}
	err = createProjectView(s, list, a, createBacklogBucket, true)
	if err != nil {
		return
	}

	gantt := &ProjectView{
		ProjectID: project.ID,
		Title:     "Gantt",
		ViewKind:  ProjectViewKindGantt,
		Position:  200,
	}
	err = createProjectView(s, gantt, a, createBacklogBucket, true)
	if err != nil {
		return
	}

	table := &ProjectView{
		ProjectID: project.ID,
		Title:     "Table",
		ViewKind:  ProjectViewKindTable,
		Position:  300,
	}
	err = createProjectView(s, table, a, createBacklogBucket, true)
	if err != nil {
		return
	}

	kanban := &ProjectView{
		ProjectID:               project.ID,
		Title:                   "Kanban",
		ViewKind:                ProjectViewKindKanban,
		Position:                400,
		BucketConfigurationMode: BucketConfigurationModeManual,
	}
	err = createProjectView(s, kanban, a, createBacklogBucket, true)
	if err != nil {
		return
	}

	project.Views = []*ProjectView{
		list,
		gantt,
		table,
		kanban,
	}

	return
}
