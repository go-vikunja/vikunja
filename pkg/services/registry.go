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
	"sync"

	"xorm.io/xorm"
)

// ServiceRegistry provides centralized, thread-safe access to all service instances.
// This replaces the previous pattern of services creating other services in their constructors,
// which created circular dependencies and service duplication issues.
//
// Benefits:
// - Thread-safe lazy initialization using double-check locking pattern
// - Singleton pattern: each service created exactly once per registry
// - Breaks circular dependencies naturally
// - Explicit dependency graph: service → registry → other services
// - Better performance: no duplicate service creation
// - Consistent pattern used everywhere
// - Easy to mock registry in tests
//
// Usage:
//
//	registry := NewServiceRegistry(db)
//	projectService := registry.Project()
//	taskService := registry.Task()
type ServiceRegistry struct {
	db *xorm.Engine

	// Service instances (lazily initialized)
	apiToken   *APITokenService
	attachment *AttachmentService
	bulkTask   *BulkTaskService
	comment    *CommentService
	favorite   *FavoriteService
	kanban     *KanbanService
	label      *LabelService
	linkShare  *LinkShareService
	// notifications is NOT cached here - it requires per-request session, use NewNotificationsService(s) directly
	permissions  *PermissionService
	project      *ProjectService
	projectDup   *ProjectDuplicateService
	projectTeams *ProjectTeamService
	projectUsers *ProjectUsersService
	projectViews *ProjectViewService
	reactions    *ReactionsService
	savedFilter  *SavedFilterService
	subscription *SubscriptionService
	task         *TaskService
	team         *TeamService
	user         *UserService
	userExport   *UserExportService
	userMentions *UserMentionsService
	webhook      *WebhookService

	// Mutex for thread-safe initialization
	mu sync.RWMutex
}

// NewServiceRegistry creates a new service registry.
func NewServiceRegistry(db *xorm.Engine) *ServiceRegistry {
	return &ServiceRegistry{
		db: db,
	}
}

// APIToken returns the APITokenService instance (thread-safe lazy init).
func (r *ServiceRegistry) APIToken() *APITokenService {
	r.mu.RLock()
	if r.apiToken != nil {
		defer r.mu.RUnlock()
		return r.apiToken
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.apiToken == nil {
		r.apiToken = &APITokenService{
			DB: r.db,
		}
	}
	return r.apiToken
}

// Attachment returns the AttachmentService instance (thread-safe lazy init).
func (r *ServiceRegistry) Attachment() *AttachmentService {
	r.mu.RLock()
	if r.attachment != nil {
		defer r.mu.RUnlock()
		return r.attachment
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.attachment == nil {
		r.attachment = &AttachmentService{
			DB:       r.db,
			Registry: r,
		}
	}
	return r.attachment
}

// BulkTask returns the BulkTaskService instance (thread-safe lazy init).
func (r *ServiceRegistry) BulkTask() *BulkTaskService {
	r.mu.RLock()
	if r.bulkTask != nil {
		defer r.mu.RUnlock()
		return r.bulkTask
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.bulkTask == nil {
		r.bulkTask = &BulkTaskService{
			DB:       r.db,
			Registry: r,
		}
	}
	return r.bulkTask
}

// Comment returns the CommentService instance (thread-safe lazy init).
func (r *ServiceRegistry) Comment() *CommentService {
	r.mu.RLock()
	if r.comment != nil {
		defer r.mu.RUnlock()
		return r.comment
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.comment == nil {
		r.comment = &CommentService{
			DB:       r.db,
			Registry: r,
		}
	}
	return r.comment
}

// Favorite returns the FavoriteService instance (thread-safe lazy init).
func (r *ServiceRegistry) Favorite() *FavoriteService {
	r.mu.RLock()
	if r.favorite != nil {
		defer r.mu.RUnlock()
		return r.favorite
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.favorite == nil {
		r.favorite = &FavoriteService{
			DB: r.db,
		}
	}
	return r.favorite
}

// Kanban returns the KanbanService instance (thread-safe lazy init).
func (r *ServiceRegistry) Kanban() *KanbanService {
	r.mu.RLock()
	if r.kanban != nil {
		defer r.mu.RUnlock()
		return r.kanban
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.kanban == nil {
		r.kanban = &KanbanService{
			DB:       r.db,
			Registry: r,
		}
	}
	return r.kanban
}

// Label returns the LabelService instance (thread-safe lazy init).
func (r *ServiceRegistry) Label() *LabelService {
	r.mu.RLock()
	if r.label != nil {
		defer r.mu.RUnlock()
		return r.label
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.label == nil {
		r.label = &LabelService{
			DB:       r.db,
			Registry: r,
		}
	}
	return r.label
}

// LinkShare returns the LinkShareService instance (thread-safe lazy init).
func (r *ServiceRegistry) LinkShare() *LinkShareService {
	r.mu.RLock()
	if r.linkShare != nil {
		defer r.mu.RUnlock()
		return r.linkShare
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.linkShare == nil {
		r.linkShare = &LinkShareService{
			DB:       r.db,
			Registry: r,
		}
	}
	return r.linkShare
}

// Permissions returns the PermissionService instance (thread-safe lazy init).
func (r *ServiceRegistry) Permissions() *PermissionService {
	r.mu.RLock()
	if r.permissions != nil {
		defer r.mu.RUnlock()
		return r.permissions
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.permissions == nil {
		r.permissions = &PermissionService{
			DB:       r.db,
			Registry: r,
		}
	}
	return r.permissions
}

// Project returns the ProjectService instance (thread-safe lazy init).
func (r *ServiceRegistry) Project() *ProjectService {
	r.mu.RLock()
	if r.project != nil {
		defer r.mu.RUnlock()
		return r.project
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.project == nil {
		r.project = &ProjectService{
			DB:       r.db,
			Registry: r,
		}
	}
	return r.project
}

// ProjectDuplicate returns the ProjectDuplicateService instance (thread-safe lazy init).
func (r *ServiceRegistry) ProjectDuplicate() *ProjectDuplicateService {
	r.mu.RLock()
	if r.projectDup != nil {
		defer r.mu.RUnlock()
		return r.projectDup
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.projectDup == nil {
		r.projectDup = &ProjectDuplicateService{
			DB:       r.db,
			Registry: r,
		}
	}
	return r.projectDup
}

// ProjectTeams returns the ProjectTeamService instance (thread-safe lazy init).
func (r *ServiceRegistry) ProjectTeams() *ProjectTeamService {
	r.mu.RLock()
	if r.projectTeams != nil {
		defer r.mu.RUnlock()
		return r.projectTeams
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.projectTeams == nil {
		r.projectTeams = &ProjectTeamService{
			DB:       r.db,
			Registry: r,
		}
	}
	return r.projectTeams
}

// ProjectUsers returns the ProjectUsersService instance (thread-safe lazy init).
func (r *ServiceRegistry) ProjectUsers() *ProjectUsersService {
	r.mu.RLock()
	if r.projectUsers != nil {
		defer r.mu.RUnlock()
		return r.projectUsers
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.projectUsers == nil {
		r.projectUsers = &ProjectUsersService{
			DB:       r.db,
			Registry: r,
		}
	}
	return r.projectUsers
}

// ProjectViews returns the ProjectViewService instance (thread-safe lazy init).
func (r *ServiceRegistry) ProjectViews() *ProjectViewService {
	r.mu.RLock()
	if r.projectViews != nil {
		defer r.mu.RUnlock()
		return r.projectViews
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.projectViews == nil {
		r.projectViews = &ProjectViewService{
			DB:       r.db,
			Registry: r,
		}
	}
	return r.projectViews
}

// Reactions returns the ReactionsService instance (thread-safe lazy init).
func (r *ServiceRegistry) Reactions() *ReactionsService {
	r.mu.RLock()
	if r.reactions != nil {
		defer r.mu.RUnlock()
		return r.reactions
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.reactions == nil {
		r.reactions = &ReactionsService{
			DB:       r.db,
			Registry: r,
		}
	}
	return r.reactions
}

// SavedFilter returns the SavedFilterService instance (thread-safe lazy init).
func (r *ServiceRegistry) SavedFilter() *SavedFilterService {
	r.mu.RLock()
	if r.savedFilter != nil {
		defer r.mu.RUnlock()
		return r.savedFilter
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.savedFilter == nil {
		r.savedFilter = &SavedFilterService{
			DB: r.db,
		}
	}
	return r.savedFilter
}

// Subscription returns the SubscriptionService instance (thread-safe lazy init).
func (r *ServiceRegistry) Subscription() *SubscriptionService {
	r.mu.RLock()
	if r.subscription != nil {
		defer r.mu.RUnlock()
		return r.subscription
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.subscription == nil {
		r.subscription = &SubscriptionService{
			DB: r.db,
		}
	}
	return r.subscription
}

// Task returns the TaskService instance (thread-safe lazy init).
func (r *ServiceRegistry) Task() *TaskService {
	r.mu.RLock()
	if r.task != nil {
		defer r.mu.RUnlock()
		return r.task
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.task == nil {
		r.task = &TaskService{
			DB:       r.db,
			Registry: r,
		}
	}
	return r.task
}

// Team returns the TeamService instance (thread-safe lazy init).
func (r *ServiceRegistry) Team() *TeamService {
	r.mu.RLock()
	if r.team != nil {
		defer r.mu.RUnlock()
		return r.team
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.team == nil {
		r.team = &TeamService{
			DB: r.db,
		}
	}
	return r.team
}

// User returns the UserService instance (thread-safe lazy init).
func (r *ServiceRegistry) User() *UserService {
	r.mu.RLock()
	if r.user != nil {
		defer r.mu.RUnlock()
		return r.user
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.user == nil {
		r.user = &UserService{
			DB:       r.db,
			Registry: r,
		}
	}
	return r.user
}

// UserExport returns the UserExportService instance (thread-safe lazy init).
func (r *ServiceRegistry) UserExport() *UserExportService {
	r.mu.RLock()
	if r.userExport != nil {
		defer r.mu.RUnlock()
		return r.userExport
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.userExport == nil {
		r.userExport = &UserExportService{
			DB: r.db,
		}
	}
	return r.userExport
}

// UserMentions returns the UserMentionsService instance (thread-safe lazy init).
func (r *ServiceRegistry) UserMentions() *UserMentionsService {
	r.mu.RLock()
	if r.userMentions != nil {
		defer r.mu.RUnlock()
		return r.userMentions
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.userMentions == nil {
		r.userMentions = &UserMentionsService{}
	}
	return r.userMentions
}

// Webhook returns the WebhookService instance (thread-safe lazy init).
func (r *ServiceRegistry) Webhook() *WebhookService {
	r.mu.RLock()
	if r.webhook != nil {
		defer r.mu.RUnlock()
		return r.webhook
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.webhook == nil {
		r.webhook = &WebhookService{
			DB:       r.db,
			Registry: r,
		}
	}
	return r.webhook
}
