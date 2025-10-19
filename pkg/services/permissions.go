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

package services

import (
	"xorm.io/xorm"
)

// PermissionService provides centralized permission checking.
// This service is used by other services to check permissions consistently.
//
// Design Philosophy:
// - All permission logic resides in the service layer
// - Models are pure data structures with zero business logic
// - Cross-entity permission checks handled via lazy-loaded service dependencies
// - Avoids circular dependencies through on-demand service initialization
//
// Migration Strategy:
// This service will gradually absorb all Can* methods currently in model *_permissions.go files.
// Each permission method migration maintains exact behavioral compatibility via baseline tests.
type PermissionService struct {
	DB       *xorm.Engine
	Registry *ServiceRegistry
}

// NewPermissionService creates a new PermissionService.
// Deprecated: Use ServiceRegistry.Permissions() instead.
func NewPermissionService(db *xorm.Engine) *PermissionService {
	registry := NewServiceRegistry(db)
	return registry.Permissions()
}

// InitPermissionService sets up the permission service infrastructure.
// This function is called during application initialization to prepare
// the permission delegation system.
//
// The permission delegation infrastructure is established in T-PERM-003.
// As permission methods are migrated from models to services (T-PERM-006 onwards),
// this function will be populated with delegation function assignments that allow
// models to call service-layer permission checks without import cycles.
//
// Delegation Pattern:
// - Function variables defined in pkg/models/permissions_delegation.go
// - Service layer sets these variables during initialization
// - Model permission methods call delegated functions
// - Service layer implements actual permission logic
//
// Migration Progress:
// - T-PERM-003: ✅ Delegation infrastructure created
// - T-PERM-006: Project permissions (not yet migrated)
// - T-PERM-007: Task permissions (not yet migrated)
// - T-PERM-008: Label & Kanban permissions (not yet migrated)
// - T-PERM-009: Link Share & Subscription permissions (not yet migrated)
// - T-PERM-010: Task Relations permissions (not yet migrated)
// - T-PERM-011: Project Relations permissions (not yet migrated)
// - T-PERM-012: Misc permissions (not yet migrated)
func InitPermissionService() {
	// NOTE: Delegation function assignments will be added here as permission
	// methods are migrated in tasks T-PERM-006 through T-PERM-012.
	//
	// Example pattern (to be implemented in future tasks):
	// models.CheckProjectReadFunc = func(s *xorm.Session, projectID int64, a web.Auth) (bool, int, error) {
	//     ps := NewPermissionService(s.Engine())
	//     return ps.CheckProjectRead(s, projectID, a)
	// }

	// T-PERM-006: Project permission delegation ✅
	InitProjectService()

	// T-PERM-007: Task permission delegation ✅
	InitTaskService()

	// T-PERM-008: Label & Kanban permission delegation
	InitLabelService()
	// InitKanbanService() already called in InitializeDependencies

	// T-PERM-009: Link Share & Subscription permission delegation
	InitLinkShareService()
	InitSubscriptionService()

	// T-PERM-011: Project Relations permission delegation ✅
	InitProjectTeamService()
	InitProjectUserService()
	InitProjectViewService()

	// T-PERM-012: Misc permission delegation
	InitWebhookService()
	InitSavedFilterService()
	// InitReactionsService() already called in InitializeDependencies
	InitProjectDuplicateService()
	// InitTeamService() already called in InitializeDependencies
}

// NOTE: Permission checking methods will be added in subsequent migration tasks.
//
// Migration Plan:
// - T-PERM-006: Project permissions (CanRead, CanWrite, CanUpdate, CanDelete, CanCreate)
// - T-PERM-007: Task permissions (CanRead, CanWrite, CanUpdate, CanDelete, CanCreate)
// - T-PERM-008: Label & Kanban permissions
// - T-PERM-009: Link Share & Subscription permissions
// - T-PERM-010: Task Relations permissions
// - T-PERM-011: Project Relations permissions
// - T-PERM-012: Misc permissions (API tokens, teams, webhooks, etc.)
//
// Each migration task will:
// 1. Add permission methods to this service
// 2. Update baseline tests to verify exact behavioral match
// 3. Update models to delegate to service methods
// 4. Remove original model permission code after verification
