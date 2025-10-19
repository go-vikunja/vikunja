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

// CRUD tests removed - covered by service layer tests in pkg/services/reactions_test.go
// Model methods are deprecated facades that delegate to ReactionsService
//
// Testing Strategy:
// - Business logic tests → pkg/services/reactions_test.go
// - Integration tests → Service tests with testutil.Init()
// - Route tests → pkg/routes/api/v1/reaction.go (if needed)
//
// Model tests should only cover:
// - TableName() function
// - Struct field validation (not database operations)
// - Pure data structure behavior
//
// Permission and helper function tests will be refactored in T-PERMISSIONS task
