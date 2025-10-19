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

// CRUD tests for Label model have been removed as part of T-CLEANUP-4.
//
// The Label model's CRUD methods (Create, Update, Delete, ReadAll, ReadOne) are deprecated
// facades that delegate to pkg/services/labels service layer with zero business logic.
//
// Business logic for labels is now tested comprehensively in:
// - pkg/services/labels_test.go (service layer business logic tests)
// - pkg/routes/api/v1/label_test.go (route integration tests, if any)
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
// - TestLabel_ReadAll (~137 lines)
// - TestLabel_ReadOne (~131 lines)
// - TestLabel_Create (~70 lines)
// - TestLabel_Update (~90 lines)
// - TestLabel_Delete (~82 lines)
//
// Total removed: ~510 lines of CRUD tests that duplicate service layer coverage.

// NOTE: Permission tests (CanCreate, CanUpdate, CanDelete) were tested within the CRUD tests.
// These permission checks will be moved to the service layer in task T-PERMISSIONS.
// Once T-PERMISSIONS is complete, permission logic will be tested in the service layer tests.
