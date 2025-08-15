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

package v2

import (
	"github.com/labstack/echo/v4"
)

// RegisterRoutes registers all routes for the v2 API.
func RegisterRoutes(a *echo.Group) {
	a.GET("/info", GetInfo)
	a.GET("/timezones", GetTimezones)
	a.GET("/webhooks/events", GetWebhookEvents)

	a.GET("/projects", GetProjects)
	a.POST("/projects", CreateProject)
	a.GET("/projects/:id", GetProject)
	a.PUT("/projects/:id", UpdateProject)
	a.DELETE("/projects/:id", DeleteProject)
	a.GET("/projects/:id/users", ListUsersForProject)

	a.GET("/projects/:id/tasks", GetTasks)
	a.POST("/projects/:id/tasks", CreateTask)
	a.GET("/tasks/:id", GetTask)
	a.PUT("/tasks/:id", UpdateTask)
	a.DELETE("/tasks/:id", DeleteTask)
	a.POST("/tasks/:id/attachments", UploadTaskAttachment)
	a.GET("/tasks/:id/attachments/:attachmentid", GetTaskAttachment)

	a.GET("/labels", GetLabels)
	a.POST("/labels", CreateLabel)
	a.GET("/labels/:id", GetLabel)
	a.PUT("/labels/:id", UpdateLabel)
	a.DELETE("/labels/:id", DeleteLabel)

	a.GET("/teams", GetTeams)
	a.POST("/teams", CreateTeam)
	a.GET("/teams/:id", GetTeam)
	a.PUT("/teams/:id", UpdateTeam)
	a.DELETE("/teams/:id", DeleteTeam)

	a.GET("/users", GetUsers)
	a.POST("/users", CreateUser)
	a.GET("/users/:id", GetUserByID)
	a.PUT("/users/:id", UpdateUser)
	a.DELETE("/users/:id", DeleteUser)

	a.POST("/users/confirm-email", ConfirmEmail)
	a.POST("/users/password-reset-token", RequestPasswordResetToken)
	a.POST("/users/password-reset", ResetPassword)

	a.GET("/user", GetCurrentUser)
	a.GET("/user/settings", GetUserSettings)
	a.PUT("/user/settings", UpdateUserSettings)
	a.POST("/user/email", UpdateUserEmail)
	a.PUT("/user/password", UpdateUserPassword)

	a.POST("/user/deletion-request", UserRequestDeletion)
	a.POST("/user/deletion-confirm", UserConfirmDeletion)
	a.POST("/user/deletion-cancel", UserCancelDeletion)
	a.GET("/user/export", GetUserExportStatus)
	a.POST("/user/export-request", RequestUserDataExport)
	a.POST("/user/export-download", DownloadUserDataExport)

	a.POST("/auth/login", Login)
	a.POST("/auth/logout", Logout)
	a.POST("/auth/token/renew", RenewToken)

	a.POST("/shares/:token/auth", AuthenticateLinkShare)

	a.GET("/user/caldav-tokens", GetCaldavTokens)
	a.POST("/user/caldav-tokens", GenerateCaldavToken)
	a.DELETE("/user/caldav-tokens/:id", DeleteCaldavToken)

	a.GET("/user/totp", GetTOTP)
	a.POST("/user/totp/enroll", EnrollTOTP)
	a.GET("/user/totp/qrcode", GetTOTPQrCode)
	a.POST("/user/totp/enable", EnableTOTP)
	a.DELETE("/user/totp", DisableTOTP)

	a.POST("/notifications/mark-all-as-read", MarkAllNotificationsAsRead)

	a.GET("/info", GetInfo)
	a.GET("/timezones", GetTimezones)

	p := a.Group("/projects/:id")
	p.GET("/webhooks", GetWebhooks)
	p.POST("/webhooks", CreateWebhook)
	p.GET("/webhooks/:id", GetWebhook)
	p.PUT("/webhooks/:id", UpdateWebhook)
	p.DELETE("/webhooks/:id", DeleteWebhook)
	a.GET("/webhooks/events", GetWebhookEvents)
}
