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
	"code.vikunja.io/api/pkg/entitlement"
	"code.vikunja.io/api/pkg/files"

	"xorm.io/xorm"
)

// countProjectsOwnedBy counts every project owned by userID, archived and
// child projects included. Saved filters are not projects and do not count.
func countProjectsOwnedBy(s *xorm.Session, userID int64) (int64, error) {
	return s.Where("owner_id = ?", userID).Count(&Project{})
}

// storageUsedBy sums the attachment files of tasks in projects owned by userID
// (soft-deleted tasks included, their files still exist) plus the background
// files of those projects. Avatars and exports are not charged.
func storageUsedBy(s *xorm.Session, userID int64) (int64, error) {
	attachments, err := s.Table("files").
		Join("INNER", "task_attachments", "task_attachments.file_id = files.id").
		Join("INNER", "tasks", "tasks.id = task_attachments.task_id").
		Join("INNER", "projects", "projects.id = tasks.project_id").
		Where("projects.owner_id = ?", userID).
		SumInt(&files.File{}, "files.size")
	if err != nil {
		return 0, err
	}
	backgrounds, err := s.Table("files").
		Join("INNER", "projects", "projects.background_file_id = files.id").
		Where("projects.owner_id = ?", userID).
		SumInt(&files.File{}, "files.size")
	if err != nil {
		return 0, err
	}
	return attachments + backgrounds, nil
}

// checkProjectLimit blocks a new project once the owner is at max_projects.
func checkProjectLimit(s *xorm.Session, ownerID int64) error {
	limit, limited, err := entitlement.Limit(s, ownerID, entitlement.FeatureMaxProjects)
	if err != nil || !limited {
		return err
	}
	current, err := countProjectsOwnedBy(s, ownerID)
	if err != nil {
		return err
	}
	if current >= limit {
		return entitlement.ErrLimitReached{Feature: entitlement.FeatureMaxProjects, Limit: limit, Current: current}
	}
	return nil
}

// CheckStorageLimit blocks storing incoming bytes once that would push the
// owner past max_storage_bytes. Charged to the project owner, not the uploader.
func CheckStorageLimit(s *xorm.Session, ownerID int64, incoming int64) error {
	limit, limited, err := entitlement.Limit(s, ownerID, entitlement.FeatureMaxStorageBytes)
	if err != nil || !limited {
		return err
	}
	current, err := storageUsedBy(s, ownerID)
	if err != nil {
		return err
	}
	if current+incoming > limit {
		return entitlement.ErrLimitReached{Feature: entitlement.FeatureMaxStorageBytes, Limit: limit, Current: current}
	}
	return nil
}

var limitUsage = map[entitlement.Feature]func(s *xorm.Session, userID int64) (int64, error){
	entitlement.FeatureMaxProjects:     countProjectsOwnedBy,
	entitlement.FeatureMaxStorageBytes: storageUsedBy,
}

// EntitlementUsage measures the given limit features for the /user response.
func EntitlementUsage(s *xorm.Session, userID int64, limits []entitlement.Feature) (map[entitlement.Feature]int64, error) {
	usage := make(map[entitlement.Feature]int64, len(limits))
	for _, f := range limits {
		current, ok := limitUsage[f]
		if !ok {
			continue
		}
		used, err := current(s, userID)
		if err != nil {
			return nil, err
		}
		usage[f] = used
	}
	return usage, nil
}
