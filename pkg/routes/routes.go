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

// @title Vikunja API
// @description This is the documentation for the [Vikunja](https://vikunja.io) API. Vikunja is a cross-platform To-do-application with a lot of features, such as sharing projects with users or teams. <!-- ReDoc-Inject: <security-definitions> -->

// @description # Pagination
// @description Every endpoint capable of pagination will return two headers:
// @description * `x-pagination-total-pages`: The total number of available pages for this request
// @description * `x-pagination-result-count`: The number of items returned for this request.
// @description # Permissions
// @description All endpoints which return a single item (project, task, etc.) - no array - will also return a `x-max-permission` header with the max permission the user has on this item as an int where `0` is `Read Only`, `1` is `Read & Write` and `2` is `Admin`.
// @description This can be used to show or hide ui elements based on the permissions the user has.
// @description # Errors
// @description All errors have an error code and a human-readable error message in addition to the http status code. You should always check for the status code in the response, not only the http status code.
// @description Due to limitations in the swagger library we're using for this document, only one error per http status code is documented here. Make sure to check the [error docs](https://vikunja.io/docs/errors/) in Vikunja's documentation for a full list of available error codes.
// @description # Authorization
// @description **JWT-Auth:** Main authorization method, used for most of the requests. Needs `Authorization: Bearer <jwt-token>`-header to authenticate successfully.
// @description
// @description **API Token:** You can create scoped API tokens for your user and use the token to make authenticated requests in the context of that user. The token must be provided via an `Authorization: Bearer <token>` header, similar to jwt auth. See the documentation for the `api` group to manage token creation and revocation.
// @description
// @description **BasicAuth:** Only used when requesting tasks via CalDAV.
// @description <!-- ReDoc-Inject: <security-definitions> -->
// @BasePath /api/v1

// @license.url https://code.vikunja.io/api/src/branch/main/LICENSE
// @license.name AGPL-3.0-or-later

// @contact.url https://vikunja.io/contact/
// @contact.name General Vikunja contact
// @contact.email hello@vikunja.io

// @securityDefinitions.basic BasicAuth

// @securityDefinitions.apikey JWTKeyAuth
// @in header
// @name Authorization

package routes

import (
	"errors"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth/openid"
	"code.vikunja.io/api/pkg/modules/background"
	backgroundHandler "code.vikunja.io/api/pkg/modules/background/handler"
	"code.vikunja.io/api/pkg/modules/background/unsplash"
	"code.vikunja.io/api/pkg/modules/background/upload"
	"code.vikunja.io/api/pkg/modules/migration"
	migrationHandler "code.vikunja.io/api/pkg/modules/migration/handler"
	microsofttodo "code.vikunja.io/api/pkg/modules/migration/microsoft-todo"
	"code.vikunja.io/api/pkg/modules/migration/ticktick"
	"code.vikunja.io/api/pkg/modules/migration/todoist"
	"code.vikunja.io/api/pkg/modules/migration/trello"
	vikunja_file "code.vikunja.io/api/pkg/modules/migration/vikunja-file"
	"code.vikunja.io/api/pkg/plugins"
	apiv1 "code.vikunja.io/api/pkg/routes/api/v1"
	"code.vikunja.io/api/pkg/routes/caldav"
	"code.vikunja.io/api/pkg/version"
	"code.vikunja.io/api/pkg/web/handler"

	"github.com/getsentry/sentry-go"
	sentryecho "github.com/getsentry/sentry-go/echo"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/ulule/limiter/v3"
)

// slogHTTPMiddleware creates a custom HTTP logging middleware using slog
func slogHTTPMiddleware(logger *slog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return echo.HandlerFunc(func(c echo.Context) error {
			start := time.Now()

			err := next(c)
			if err != nil {
				c.Error(err)
			}

			req := c.Request()
			res := c.Response()

			logger.InfoContext(c.Request().Context(),
				req.Method+" "+req.RequestURI,
				"status", res.Status,
				"remote_ip", c.RealIP(),
				"latency", time.Since(start),
				"user_agent", req.UserAgent(),
			)

			return err
		})
	}
}

// NewEcho registers a new Echo instance
func NewEcho() *echo.Echo {
	e := echo.New()

	e.HideBanner = true

	e.Logger = log.NewEchoLogger(config.LogEnabled.GetBool(), config.LogHTTP.GetString(), config.LogFormat.GetString())

	// Logger
	if config.LogEnabled.GetBool() && config.LogHTTP.GetString() != "off" {
		httpLogger := log.NewHTTPLogger(config.LogEnabled.GetBool(), config.LogHTTP.GetString(), config.LogFormat.GetString())
		e.Use(slogHTTPMiddleware(httpLogger))
	}

	// panic recover
	e.Use(middleware.Recover())

	setupSentry(e)

	// Validation
	e.Validator = &CustomValidator{}

	return e
}

func setupSentry(e *echo.Echo) {
	if !config.SentryEnabled.GetBool() {
		return
	}

	if err := sentry.Init(sentry.ClientOptions{
		Dsn:              config.SentryDsn.GetString(),
		AttachStacktrace: true,
		Release:          version.Version,
	}); err != nil {
		log.Criticalf("Sentry init failed: %s", err)
	}
	defer sentry.Flush(5 * time.Second)

	e.Use(sentryecho.New(sentryecho.Options{
		Repanic: true,
	}))

	e.HTTPErrorHandler = func(err error, c echo.Context) {
		// Only capture errors not already handled by echo
		var herr *echo.HTTPError
		if errors.As(err, &herr) && herr.Code > 499 {
			var errToReport = err
			if herr.Internal == nil {
				errToReport = herr.Internal
			}

			hub := sentryecho.GetHubFromContext(c)
			if hub != nil {
				hub.WithScope(func(scope *sentry.Scope) {
					scope.SetExtra("url", c.Request().URL)
					hub.CaptureException(errToReport)
				})
			} else {
				sentry.CaptureException(errToReport)
				log.Debugf("Could not add context for sending error '%s' to sentry", err.Error())
			}
			log.Debugf("Error '%s' sent to sentry", err.Error())
		}
		e.DefaultHTTPErrorHandler(err, c)
	}
}

// RegisterRoutes registers all routes for the application
func RegisterRoutes(e *echo.Echo) {

	if config.ServiceEnableCaldav.GetBool() {
		// Caldav routes
		wkg := e.Group("/.well-known")
		wkg.Use(middleware.BasicAuth(caldav.BasicAuth))
		wkg.Any("/caldav", caldav.PrincipalHandler)
		wkg.Any("/caldav/", caldav.PrincipalHandler)
		c := e.Group("/dav")
		registerCalDavRoutes(c)
	}

	// healthcheck
	e.GET("/health", HealthcheckHandler)

	setupStaticFrontendFilesHandler(e)

	// CORS
	if config.CorsEnable.GetBool() {
		log.Debugf("CORS enabled with origins: %s", strings.Join(config.CorsOrigins.GetStringSlice(), ", "))
		e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins: config.CorsOrigins.GetStringSlice(),
			MaxAge:       config.CorsMaxAge.GetInt(),
			Skipper: func(context echo.Context) bool {
				// Since it is not possible to register this middleware just for the api group,
				// we just disable it when for caldav requests.
				// Caldav requires OPTIONS requests to be answered in a specific manner,
				// not doing this would break the caldav implementation
				return strings.HasPrefix(context.Path(), "/dav")
			},
		}))
	}

	// API Routes
	a := e.Group("/api/v1")
	e.OnAddRouteHandler = func(_ string, route echo.Route, _ echo.HandlerFunc, middlewares []echo.MiddlewareFunc) {
		models.CollectRoutesForAPITokenUsage(route, middlewares)
	}
	registerAPIRoutes(a)
}

func registerAPIRoutes(a *echo.Group) {

	// This is the group with no auth
	// It is its own group to be able to rate limit this based on different heuristics
	n := a.Group("")
	setupRateLimit(n, "ip")

	// Echo does not unescape url path params by default. To make sure values bound as :param in urls are passed
	// properly to handlers, we use this middleware to unescape them.
	// See https://kolaente.dev/vikunja/vikunja/issues/1224
	// See https://github.com/labstack/echo/issues/766
	a.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			params := make([]string, 0, len(c.ParamValues()))
			for _, param := range c.ParamValues() {
				p, err := url.PathUnescape(param)
				if err != nil {
					return err
				}
				params = append(params, p)
			}
			c.SetParamValues(params...)
			return next(c)
		}
	})

	// Docs
	n.GET("/docs.json", apiv1.DocsJSON)
	n.GET("/docs", apiv1.RedocUI)

	// Prometheus endpoint
	setupMetrics(n)

	// Separate route for unauthenticated routes to enable rate limits for it
	ur := a.Group("")
	rate := limiter.Rate{
		Period: 60 * time.Second,
		Limit:  config.RateLimitNoAuthRoutesLimit.GetInt64(),
	}
	rateLimiter := createRateLimiter(rate)
	ur.Use(RateLimit(rateLimiter, "ip"))

	if config.AuthLocalEnabled.GetBool() {
		ur.POST("/register", apiv1.RegisterUser)
		ur.POST("/user/password/token", apiv1.UserRequestResetPasswordToken)
		ur.POST("/user/password/reset", apiv1.UserResetPassword)
		ur.POST("/user/confirm", apiv1.UserConfirmEmail)
	}

	if config.AuthLocalEnabled.GetBool() || config.AuthLdapEnabled.GetBool() {
		ur.POST("/login", apiv1.Login)
	}

	if config.AuthOpenIDEnabled.GetBool() {
		ur.POST("/auth/openid/:provider/callback", openid.HandleCallback)
	}

	// Testing
	if config.ServiceTestingtoken.GetString() != "" {
		n.PATCH("/test/:table", apiv1.HandleTesting)
	}

	// Info endpoint
	n.GET("/info", apiv1.Info)

	// Link share auth
	if config.ServiceEnableLinkSharing.GetBool() {
		ur.POST("/shares/:share/auth", apiv1.AuthenticateLinkShare)
	}

	// ===== Routes with Authentication =====
	a.Use(SetupTokenMiddleware())

	// Rate limit
	setupRateLimit(a, config.RateLimitKind.GetString())

	// Middleware to collect metrics
	setupMetricsMiddleware(a)

	a.GET("/token/test", apiv1.TestToken)
	a.POST("/token/test", apiv1.CheckToken)
	a.GET("/routes", models.GetAvailableAPIRoutesForToken)

	// Avatar endpoint
	a.GET("/avatar/:username", apiv1.GetAvatar)

	// User stuff
	u := a.Group("/user")

	u.GET("", apiv1.UserShow)
	u.POST("/password", apiv1.UserChangePassword)
	u.GET("s", apiv1.UserList)
	u.POST("/token", apiv1.RenewToken)
	u.POST("/settings/email", apiv1.UpdateUserEmail)
	u.GET("/settings/avatar", apiv1.GetUserAvatarProvider)
	u.POST("/settings/avatar", apiv1.ChangeUserAvatarProvider)
	u.PUT("/settings/avatar/upload", apiv1.UploadAvatar)
	u.POST("/settings/general", apiv1.UpdateGeneralUserSettings)
	u.POST("/export/request", apiv1.RequestUserDataExport)
	u.POST("/export/download", apiv1.DownloadUserDataExport)
	u.GET("/export", apiv1.GetUserExportStatus)
	u.GET("/timezones", apiv1.GetAvailableTimezones)
	u.PUT("/settings/token/caldav", apiv1.GenerateCaldavToken)
	u.GET("/settings/token/caldav", apiv1.GetCaldavTokens)
	u.DELETE("/settings/token/caldav/:id", apiv1.DeleteCaldavToken)

	if config.ServiceEnableTotp.GetBool() {
		u.GET("/settings/totp", apiv1.UserTOTP)
		u.POST("/settings/totp/enroll", apiv1.UserTOTPEnroll)
		u.POST("/settings/totp/enable", apiv1.UserTOTPEnable)
		u.POST("/settings/totp/disable", apiv1.UserTOTPDisable)
		u.GET("/settings/totp/qrcode", apiv1.UserTOTPQrCode)
	}

	// User deletion
	if config.ServiceEnableUserDeletion.GetBool() {
		u.POST("/deletion/request", apiv1.UserRequestDeletion)
		u.POST("/deletion/confirm", apiv1.UserConfirmDeletion)
		u.POST("/deletion/cancel", apiv1.UserCancelDeletion)
	}

	projectHandler := &handler.WebHandler{
		EmptyStruct: func() handler.CObject {
			return &models.Project{}
		},
	}
	a.GET("/projects", projectHandler.ReadAllWeb)
	a.GET("/projects/:project", projectHandler.ReadOneWeb)
	a.POST("/projects/:project", projectHandler.UpdateWeb)
	a.DELETE("/projects/:project", projectHandler.DeleteWeb)
	a.PUT("/projects", projectHandler.CreateWeb)
	a.GET("/projects/:project/projectusers", apiv1.ListUsersForProject)

	if config.ServiceEnableLinkSharing.GetBool() {
		projectSharingHandler := &handler.WebHandler{
			EmptyStruct: func() handler.CObject {
				return &models.LinkSharing{}
			},
		}
		a.PUT("/projects/:project/shares", projectSharingHandler.CreateWeb)
		a.GET("/projects/:project/shares", projectSharingHandler.ReadAllWeb)
		a.GET("/projects/:project/shares/:share", projectSharingHandler.ReadOneWeb)
		a.DELETE("/projects/:project/shares/:share", projectSharingHandler.DeleteWeb)
	}

	taskCollectionHandler := &handler.WebHandler{
		EmptyStruct: func() handler.CObject {
			return &models.TaskCollection{}
		},
	}
	a.GET("/projects/:project/views/:view/tasks", taskCollectionHandler.ReadAllWeb)
	a.GET("/projects/:project/tasks", taskCollectionHandler.ReadAllWeb)

	kanbanBucketHandler := &handler.WebHandler{
		EmptyStruct: func() handler.CObject {
			return &models.Bucket{}
		},
	}
	a.GET("/projects/:project/views/:view/buckets", kanbanBucketHandler.ReadAllWeb)
	a.PUT("/projects/:project/views/:view/buckets", kanbanBucketHandler.CreateWeb)
	a.POST("/projects/:project/views/:view/buckets/:bucket", kanbanBucketHandler.UpdateWeb)
	a.DELETE("/projects/:project/views/:view/buckets/:bucket", kanbanBucketHandler.DeleteWeb)

	projectDuplicateHandler := &handler.WebHandler{
		EmptyStruct: func() handler.CObject {
			return &models.ProjectDuplicate{}
		},
	}
	a.PUT("/projects/:projectid/duplicate", projectDuplicateHandler.CreateWeb)

	taskHandler := &handler.WebHandler{
		EmptyStruct: func() handler.CObject {
			return &models.Task{}
		},
	}
	a.PUT("/projects/:project/tasks", taskHandler.CreateWeb)
	a.GET("/tasks/:projecttask", taskHandler.ReadOneWeb)
	a.GET("/tasks/all", taskCollectionHandler.ReadAllWeb)
	a.DELETE("/tasks/:projecttask", taskHandler.DeleteWeb)
	a.POST("/tasks/:projecttask", taskHandler.UpdateWeb)

	taskPositionHandler := &handler.WebHandler{
		EmptyStruct: func() handler.CObject {
			return &models.TaskPosition{}
		},
	}
	a.POST("/tasks/:task/position", taskPositionHandler.UpdateWeb)

	bulkTaskHandler := &handler.WebHandler{
		EmptyStruct: func() handler.CObject {
			return &models.BulkTask{}
		},
	}
	a.POST("/tasks/bulk", bulkTaskHandler.UpdateWeb)

	assigneeTaskHandler := &handler.WebHandler{
		EmptyStruct: func() handler.CObject {
			return &models.TaskAssginee{}
		},
	}
	a.PUT("/tasks/:projecttask/assignees", assigneeTaskHandler.CreateWeb)
	a.DELETE("/tasks/:projecttask/assignees/:user", assigneeTaskHandler.DeleteWeb)
	a.GET("/tasks/:projecttask/assignees", assigneeTaskHandler.ReadAllWeb)

	bulkAssigneeHandler := &handler.WebHandler{
		EmptyStruct: func() handler.CObject {
			return &models.BulkAssignees{}
		},
	}
	a.POST("/tasks/:projecttask/assignees/bulk", bulkAssigneeHandler.CreateWeb)

	labelTaskHandler := &handler.WebHandler{
		EmptyStruct: func() handler.CObject {
			return &models.LabelTask{}
		},
	}
	a.PUT("/tasks/:projecttask/labels", labelTaskHandler.CreateWeb)
	a.DELETE("/tasks/:projecttask/labels/:label", labelTaskHandler.DeleteWeb)
	a.GET("/tasks/:projecttask/labels", labelTaskHandler.ReadAllWeb)

	bulkLabelTaskHandler := &handler.WebHandler{
		EmptyStruct: func() handler.CObject {
			return &models.LabelTaskBulk{}
		},
	}
	a.POST("/tasks/:projecttask/labels/bulk", bulkLabelTaskHandler.CreateWeb)

	taskRelationHandler := &handler.WebHandler{
		EmptyStruct: func() handler.CObject {
			return &models.TaskRelation{}
		},
	}
	a.PUT("/tasks/:task/relations", taskRelationHandler.CreateWeb)
	a.DELETE("/tasks/:task/relations/:relationKind/:otherTask", taskRelationHandler.DeleteWeb)

	if config.ServiceEnableTaskAttachments.GetBool() {
		taskAttachmentHandler := &handler.WebHandler{
			EmptyStruct: func() handler.CObject {
				return &models.TaskAttachment{}
			},
		}
		a.GET("/tasks/:task/attachments", taskAttachmentHandler.ReadAllWeb)
		a.DELETE("/tasks/:task/attachments/:attachment", taskAttachmentHandler.DeleteWeb)
		a.PUT("/tasks/:task/attachments", apiv1.UploadTaskAttachment)
		a.GET("/tasks/:task/attachments/:attachment", apiv1.GetTaskAttachment)
	}

	if config.ServiceEnableTaskComments.GetBool() {
		taskCommentHandler := &handler.WebHandler{
			EmptyStruct: func() handler.CObject {
				return &models.TaskComment{}
			},
		}
		a.GET("/tasks/:task/comments", taskCommentHandler.ReadAllWeb)
		a.PUT("/tasks/:task/comments", taskCommentHandler.CreateWeb)
		a.DELETE("/tasks/:task/comments/:commentid", taskCommentHandler.DeleteWeb)
		a.POST("/tasks/:task/comments/:commentid", taskCommentHandler.UpdateWeb)
		a.GET("/tasks/:task/comments/:commentid", taskCommentHandler.ReadOneWeb)
	}

	labelHandler := &handler.WebHandler{
		EmptyStruct: func() handler.CObject {
			return &models.Label{}
		},
	}
	a.GET("/labels", labelHandler.ReadAllWeb)
	a.GET("/labels/:label", labelHandler.ReadOneWeb)
	a.PUT("/labels", labelHandler.CreateWeb)
	a.DELETE("/labels/:label", labelHandler.DeleteWeb)
	a.POST("/labels/:label", labelHandler.UpdateWeb)

	projectTeamHandler := &handler.WebHandler{
		EmptyStruct: func() handler.CObject {
			return &models.TeamProject{}
		},
	}
	a.GET("/projects/:project/teams", projectTeamHandler.ReadAllWeb)
	a.PUT("/projects/:project/teams", projectTeamHandler.CreateWeb)
	a.DELETE("/projects/:project/teams/:team", projectTeamHandler.DeleteWeb)
	a.POST("/projects/:project/teams/:team", projectTeamHandler.UpdateWeb)

	projectUserHandler := &handler.WebHandler{
		EmptyStruct: func() handler.CObject {
			return &models.ProjectUser{}
		},
	}
	a.GET("/projects/:project/users", projectUserHandler.ReadAllWeb)
	a.PUT("/projects/:project/users", projectUserHandler.CreateWeb)
	a.DELETE("/projects/:project/users/:user", projectUserHandler.DeleteWeb)
	a.POST("/projects/:project/users/:user", projectUserHandler.UpdateWeb)

	savedFiltersHandler := &handler.WebHandler{
		EmptyStruct: func() handler.CObject {
			return &models.SavedFilter{}
		},
	}
	a.GET("/filters/:filter", savedFiltersHandler.ReadOneWeb)
	a.PUT("/filters", savedFiltersHandler.CreateWeb)
	a.DELETE("/filters/:filter", savedFiltersHandler.DeleteWeb)
	a.POST("/filters/:filter", savedFiltersHandler.UpdateWeb)

	teamHandler := &handler.WebHandler{
		EmptyStruct: func() handler.CObject {
			return &models.Team{}
		},
	}
	a.GET("/teams", teamHandler.ReadAllWeb)
	a.GET("/teams/:team", teamHandler.ReadOneWeb)
	a.PUT("/teams", teamHandler.CreateWeb)
	a.POST("/teams/:team", teamHandler.UpdateWeb)
	a.DELETE("/teams/:team", teamHandler.DeleteWeb)

	teamMemberHandler := &handler.WebHandler{
		EmptyStruct: func() handler.CObject {
			return &models.TeamMember{}
		},
	}
	a.PUT("/teams/:team/members", teamMemberHandler.CreateWeb)
	a.DELETE("/teams/:team/members/:user", teamMemberHandler.DeleteWeb)
	a.POST("/teams/:team/members/:user/admin", teamMemberHandler.UpdateWeb)

	// Subscriptions
	subscriptionHandler := &handler.WebHandler{
		EmptyStruct: func() handler.CObject {
			return &models.Subscription{}
		},
	}
	a.PUT("/subscriptions/:entity/:entityID", subscriptionHandler.CreateWeb)
	a.DELETE("/subscriptions/:entity/:entityID", subscriptionHandler.DeleteWeb)

	// Notifications
	notificationHandler := &handler.WebHandler{
		EmptyStruct: func() handler.CObject {
			return &models.DatabaseNotifications{}
		},
	}
	a.GET("/notifications", notificationHandler.ReadAllWeb)
	a.POST("/notifications/:notificationid", notificationHandler.UpdateWeb)
	a.POST("/notifications", apiv1.MarkAllNotificationsAsRead)

	// Migrations
	m := a.Group("/migration")
	registerMigrations(m)

	// Project Backgrounds
	if config.BackgroundsEnabled.GetBool() {
		a.GET("/projects/:project/background", backgroundHandler.GetProjectBackground)
		a.DELETE("/projects/:project/background", backgroundHandler.RemoveProjectBackground)
		if config.BackgroundsUploadEnabled.GetBool() {
			uploadBackgroundProvider := &backgroundHandler.BackgroundProvider{
				Provider: func() background.Provider {
					return &upload.Provider{}
				},
			}
			a.PUT("/projects/:project/backgrounds/upload", uploadBackgroundProvider.UploadBackground)
		}
		if config.BackgroundsUnsplashEnabled.GetBool() {
			unsplashBackgroundProvider := &backgroundHandler.BackgroundProvider{
				Provider: func() background.Provider {
					return &unsplash.Provider{}
				},
			}
			a.GET("/backgrounds/unsplash/search", unsplashBackgroundProvider.SearchBackgrounds)
			a.POST("/projects/:project/backgrounds/unsplash", unsplashBackgroundProvider.SetBackground)
			a.GET("/backgrounds/unsplash/images/:image/thumb", unsplash.ProxyUnsplashThumb)
			a.GET("/backgrounds/unsplash/images/:image", unsplash.ProxyUnsplashImage)
		}
	}

	// API Tokens
	apiTokenProvider := &handler.WebHandler{
		EmptyStruct: func() handler.CObject {
			return &models.APIToken{}
		},
	}
	a.GET("/tokens", apiTokenProvider.ReadAllWeb)
	a.PUT("/tokens", apiTokenProvider.CreateWeb)
	a.DELETE("/tokens/:token", apiTokenProvider.DeleteWeb)

	// Webhooks
	if config.WebhooksEnabled.GetBool() {
		webhookProvider := &handler.WebHandler{
			EmptyStruct: func() handler.CObject {
				return &models.Webhook{}
			},
		}
		a.GET("/projects/:project/webhooks", webhookProvider.ReadAllWeb)
		a.PUT("/projects/:project/webhooks", webhookProvider.CreateWeb)
		a.DELETE("/projects/:project/webhooks/:webhook", webhookProvider.DeleteWeb)
		a.POST("/projects/:project/webhooks/:webhook", webhookProvider.UpdateWeb)
		a.GET("/webhooks/events", apiv1.GetAvailableWebhookEvents)
	}

	// Reactions
	reactionProvider := &handler.WebHandler{
		EmptyStruct: func() handler.CObject {
			return &models.Reaction{}
		},
	}
	a.GET("/:entitykind/:entityid/reactions", reactionProvider.ReadAllWeb)
	a.POST("/:entitykind/:entityid/reactions/delete", reactionProvider.DeleteWeb)
	a.PUT("/:entitykind/:entityid/reactions", reactionProvider.CreateWeb)

	// Project views
	projectViewProvider := &handler.WebHandler{
		EmptyStruct: func() handler.CObject {
			return &models.ProjectView{}
		},
	}
	a.GET("/projects/:project/views", projectViewProvider.ReadAllWeb)
	a.GET("/projects/:project/views/:view", projectViewProvider.ReadOneWeb)
	a.PUT("/projects/:project/views", projectViewProvider.CreateWeb)
	a.DELETE("/projects/:project/views/:view", projectViewProvider.DeleteWeb)
	a.POST("/projects/:project/views/:view", projectViewProvider.UpdateWeb)

	// Kanban Task Bucket Relation
	taskBucketProvider := &handler.WebHandler{
		EmptyStruct: func() handler.CObject {
			return &models.TaskBucket{}
		},
	}
	a.POST("/projects/:project/views/:view/buckets/:bucket/tasks", taskBucketProvider.UpdateWeb)

	// Plugin routes
	if config.PluginsEnabled.GetBool() {
		// Authenticated plugin routes
		authenticatedPluginGroup := a.Group("/plugins")

		// Unauthenticated plugin routes (with basic IP rate limiting)
		unauthenticatedPluginGroup := n.Group("/plugins")

		plugins.RegisterPluginRoutes(authenticatedPluginGroup, unauthenticatedPluginGroup)
	}
}

func registerMigrations(m *echo.Group) {
	// Todoist
	if config.MigrationTodoistEnable.GetBool() {
		todoistMigrationHandler := &migrationHandler.MigrationWeb{
			MigrationStruct: func() migration.Migrator {
				return &todoist.Migration{}
			},
		}
		todoistMigrationHandler.RegisterMigrator(m)
	}

	// Trello
	if config.MigrationTrelloEnable.GetBool() {
		trelloMigrationHandler := &migrationHandler.MigrationWeb{
			MigrationStruct: func() migration.Migrator {
				return &trello.Migration{}
			},
		}
		trelloMigrationHandler.RegisterMigrator(m)
	}

	// Microsoft Todo
	if config.MigrationMicrosoftTodoEnable.GetBool() {
		microsoftTodoMigrationHandler := &migrationHandler.MigrationWeb{
			MigrationStruct: func() migration.Migrator {
				return &microsofttodo.Migration{}
			},
		}
		microsoftTodoMigrationHandler.RegisterMigrator(m)
	}

	// Vikunja File Migrator
	vikunjaFileMigrationHandler := &migrationHandler.FileMigratorWeb{
		MigrationStruct: func() migration.FileMigrator {
			return &vikunja_file.FileMigrator{}
		},
	}
	vikunjaFileMigrationHandler.RegisterRoutes(m)

	// TickTick File Migrator
	tickTickFileMigrator := migrationHandler.FileMigratorWeb{
		MigrationStruct: func() migration.FileMigrator {
			return &ticktick.Migrator{}
		},
	}
	tickTickFileMigrator.RegisterRoutes(m)
}

func registerCalDavRoutes(c *echo.Group) {

	// Basic auth middleware
	c.Use(middleware.BasicAuth(caldav.BasicAuth))

	// THIS is the entry point for caldav clients, otherwise projects will show up double
	c.Any("", caldav.EntryHandler)
	c.Any("/", caldav.EntryHandler)
	c.Any("/principals/*/", caldav.PrincipalHandler)
	c.Any("/projects", caldav.ProjectHandler)
	c.Any("/projects/", caldav.ProjectHandler)
	c.Any("/projects/:project", caldav.ProjectHandler)
	c.Any("/projects/:project/", caldav.ProjectHandler)
	c.Any("/projects/:project/:task", caldav.TaskHandler) // Mostly used for editing
}
