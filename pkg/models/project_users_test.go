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

// CRUD tests for ProjectUser model have been removed as part of T-CLEANUP-6.
//
// The ProjectUser model's CRUD methods (Create, Update, Delete, ReadAll) are deprecated
// facades that delegate to pkg/services/project_users service layer with zero business logic.
//
// Business logic for project users is now tested comprehensively in:
// - pkg/services/project_users_test.go (service layer business logic tests)
// - pkg/routes/api/v1/project_users_test.go (route integration tests, if any)
//
// Testing deprecated delegation methods provides no value - they simply call the service
// layer, which is already tested. Model tests should focus on:
// - TableName() function (if it exists)
// - Struct field validation (not database operations)
// - Pure data structure behavior
//
// For the complete testing strategy for refactored components, see:
// /home/aron/projects/vikunja/REFACTORING_GUIDE.md - Section 5
//
// Removed tests (all CRUD operations):
// - TestProjectUser_Create (~113 lines, 7 subtests)
// - TestProjectUser_ReadAll (~123 lines, 6 subtests)
// - TestProjectUser_Update (~89 lines, 5 subtests)
// - TestProjectUser_Delete (~79 lines, 5 subtests)
//
// Total removed: ~404 lines of CRUD tests that duplicate service layer coverage.

// NOTE: Permission tests (CanCreate, CanUpdate, CanDelete) were tested within the CRUD tests.
// These permission checks will be moved to the service layer in task T-PERMISSIONS.
// Once T-PERMISSIONS is complete, permission logic will be tested in the service layer tests.
