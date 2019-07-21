//  Vikunja is a todo-list application to facilitate your life.
//  Copyright 2018 Vikunja and contributors. All rights reserved.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU General Public License as published by
//  the Free Software Foundation, either version 3 of the License, or
//  (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU General Public License for more details.
//
//  You should have received a copy of the GNU General Public License
//  along with this program.  If not, see <https://www.gnu.org/licenses/>.

// @title Vikunja API
// @description This is the documentation for the [Vikunja](http://vikunja.io) API. Vikunja is a cross-plattform Todo-application with a lot of features, such as sharing lists with users or teams. <!-- ReDoc-Inject: <security-definitions> -->
// @description # Authorization
// @description **JWT-Auth:** Main authorization method, used for most of the requests. Needs ` + "`" + `Authorization: Bearer <jwt-token>` + "`" + `-header to authenticate successfully.
// @description
// @description **BasicAuth:** Only used when requesting tasks via caldav.
// @description <!-- ReDoc-Inject: <security-definitions> -->
// @BasePath /api/v1

// @license.url http://code.vikunja.io/api/src/branch/master/LICENSE
// @license.name GPLv3

// @contact.url http://vikunja.io/en/contact/
// @contact.name General Vikunja contact
// @contact.email hello@vikunja.io

// @securityDefinitions.basic BasicAuth

// @securityDefinitions.apikey JWTKeyAuth
// @in header
// @name Authorization

package routes

import (
	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	apiv1 "code.vikunja.io/api/pkg/routes/api/v1"
	"code.vikunja.io/api/pkg/routes/caldav"
	_ "code.vikunja.io/api/pkg/swagger" // To generate swagger docs
	"code.vikunja.io/web"
	"code.vikunja.io/web/handler"
	"github.com/asaskevich/govalidator"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	elog "github.com/labstack/gommon/log"
	"strings"
)

// CustomValidator is a dummy struct to use govalidator with echo
type CustomValidator struct{}

// Validate validates stuff
func (cv *CustomValidator) Validate(i interface{}) error {
	if _, err := govalidator.ValidateStruct(i); err != nil {

		var errs []string
		for field, e := range govalidator.ErrorsByField(err) {
			errs = append(errs, field+": "+e)
		}

		httperr := models.ValidationHTTPError{
			web.HTTPError{
				Code:    models.ErrCodeInvalidData,
				Message: "Invalid Data",
			},
			errs,
		}

		return httperr
	}
	return nil
}

// NewEcho registers a new Echo instance
func NewEcho() *echo.Echo {
	e := echo.New()

	e.HideBanner = true

	if l, ok := e.Logger.(*elog.Logger); ok {
		if config.LogEcho.GetString() == "off" {
			l.SetLevel(elog.OFF)
		}
		l.EnableColor()
		l.SetHeader(log.ErrFmt)
		l.SetOutput(log.GetLogWriter("echo"))
	}

	// Logger
	if config.LogHTTP.GetString() != "off" {
		e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
			Format: log.WebFmt + "\n",
			Output: log.GetLogWriter("http"),
		}))
	}

	// Validation
	e.Validator = &CustomValidator{}

	// Handler config
	handler.SetAuthProvider(&web.Auths{
		AuthObject: func(c echo.Context) (web.Auth, error) {
			return models.GetCurrentUser(c)
		},
	})
	handler.SetLoggingProvider(log.GetLogger())

	return e
}

// RegisterRoutes registers all routes for the application
func RegisterRoutes(e *echo.Echo) {

	if config.ServiceEnableCaldav.GetBool() {
		// Caldav routes
		wkg := e.Group("/.well-known")
		wkg.Use(middleware.BasicAuth(caldavBasicAuth))
		wkg.Any("/caldav", caldav.PrincipalHandler)
		wkg.Any("/caldav/", caldav.PrincipalHandler)
		c := e.Group("/dav")
		registerCalDavRoutes(c)
	}

	// CORS_SHIT
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		Skipper: func(context echo.Context) bool {
			// Since it is not possible to register this middleware just for the api group,
			// we just disable it when for caldav requests.
			// Caldav requires OPTIONS requests to be answered in a specific manner,
			// not doing this would break the caldav implementation
			return strings.HasPrefix(context.Path(), "/dav")
		},
	}))

	// API Routes
	a := e.Group("/api/v1")
	registerAPIRoutes(a)
}

func registerAPIRoutes(a *echo.Group) {

	// Docs
	a.GET("/docs.json", apiv1.DocsJSON)
	a.GET("/docs", apiv1.RedocUI)

	// Prometheus endpoint
	setupMetrics(a)

	// User stuff
	a.POST("/login", apiv1.Login)
	a.POST("/register", apiv1.RegisterUser)
	a.POST("/user/password/token", apiv1.UserRequestResetPasswordToken)
	a.POST("/user/password/reset", apiv1.UserResetPassword)
	a.POST("/user/confirm", apiv1.UserConfirmEmail)

	// Info endpoint
	a.GET("/info", apiv1.Info)

	// ===== Routes with Authetication =====
	// Authetification
	a.Use(middleware.JWT([]byte(config.ServiceJWTSecret.GetString())))

	// Rate limit
	setupRateLimit(a)

	// Middleware to collect metrics
	setupMetricsMiddleware(a)

	a.POST("/tokenTest", apiv1.CheckToken)

	// User stuff
	a.GET("/user", apiv1.UserShow)
	a.POST("/user/password", apiv1.UserChangePassword)
	a.GET("/users", apiv1.UserList)

	listHandler := &handler.WebHandler{
		EmptyStruct: func() handler.CObject {
			return &models.List{}
		},
	}
	a.GET("/lists", listHandler.ReadAllWeb)
	a.GET("/lists/:list", listHandler.ReadOneWeb)
	a.POST("/lists/:list", listHandler.UpdateWeb)
	a.DELETE("/lists/:list", listHandler.DeleteWeb)
	a.PUT("/namespaces/:namespace/lists", listHandler.CreateWeb)
	a.GET("/lists/:list/listusers", apiv1.ListUsersForList)

	taskHandler := &handler.WebHandler{
		EmptyStruct: func() handler.CObject {
			return &models.ListTask{}
		},
	}
	a.PUT("/lists/:list", taskHandler.CreateWeb)
	a.GET("/tasks/all", taskHandler.ReadAllWeb)
	a.DELETE("/tasks/:listtask", taskHandler.DeleteWeb)
	a.POST("/tasks/:listtask", taskHandler.UpdateWeb)

	bulkTaskHandler := &handler.WebHandler{
		EmptyStruct: func() handler.CObject {
			return &models.BulkTask{}
		},
	}
	a.POST("/tasks/bulk", bulkTaskHandler.UpdateWeb)

	assigneeTaskHandler := &handler.WebHandler{
		EmptyStruct: func() handler.CObject {
			return &models.ListTaskAssginee{}
		},
	}
	a.PUT("/tasks/:listtask/assignees", assigneeTaskHandler.CreateWeb)
	a.DELETE("/tasks/:listtask/assignees/:user", assigneeTaskHandler.DeleteWeb)
	a.GET("/tasks/:listtask/assignees", assigneeTaskHandler.ReadAllWeb)

	bulkAssigneeHandler := &handler.WebHandler{
		EmptyStruct: func() handler.CObject {
			return &models.BulkAssignees{}
		},
	}
	a.POST("/tasks/:listtask/assignees/bulk", bulkAssigneeHandler.CreateWeb)

	labelTaskHandler := &handler.WebHandler{
		EmptyStruct: func() handler.CObject {
			return &models.LabelTask{}
		},
	}
	a.PUT("/tasks/:listtask/labels", labelTaskHandler.CreateWeb)
	a.DELETE("/tasks/:listtask/labels/:label", labelTaskHandler.DeleteWeb)
	a.GET("/tasks/:listtask/labels", labelTaskHandler.ReadAllWeb)

	bulkLabelTaskHandler := &handler.WebHandler{
		EmptyStruct: func() handler.CObject {
			return &models.LabelTaskBulk{}
		},
	}
	a.POST("/tasks/:listtask/labels/bulk", bulkLabelTaskHandler.CreateWeb)

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

	listTeamHandler := &handler.WebHandler{
		EmptyStruct: func() handler.CObject {
			return &models.TeamList{}
		},
	}
	a.GET("/lists/:list/teams", listTeamHandler.ReadAllWeb)
	a.PUT("/lists/:list/teams", listTeamHandler.CreateWeb)
	a.DELETE("/lists/:list/teams/:team", listTeamHandler.DeleteWeb)
	a.POST("/lists/:list/teams/:team", listTeamHandler.UpdateWeb)

	listUserHandler := &handler.WebHandler{
		EmptyStruct: func() handler.CObject {
			return &models.ListUser{}
		},
	}
	a.GET("/lists/:list/users", listUserHandler.ReadAllWeb)
	a.PUT("/lists/:list/users", listUserHandler.CreateWeb)
	a.DELETE("/lists/:list/users/:user", listUserHandler.DeleteWeb)
	a.POST("/lists/:list/users/:user", listUserHandler.UpdateWeb)

	namespaceHandler := &handler.WebHandler{
		EmptyStruct: func() handler.CObject {
			return &models.Namespace{}
		},
	}
	a.GET("/namespaces", namespaceHandler.ReadAllWeb)
	a.PUT("/namespaces", namespaceHandler.CreateWeb)
	a.GET("/namespaces/:namespace", namespaceHandler.ReadOneWeb)
	a.POST("/namespaces/:namespace", namespaceHandler.UpdateWeb)
	a.DELETE("/namespaces/:namespace", namespaceHandler.DeleteWeb)
	a.GET("/namespaces/:namespace/lists", apiv1.GetListsByNamespaceID)

	namespaceTeamHandler := &handler.WebHandler{
		EmptyStruct: func() handler.CObject {
			return &models.TeamNamespace{}
		},
	}
	a.GET("/namespaces/:namespace/teams", namespaceTeamHandler.ReadAllWeb)
	a.PUT("/namespaces/:namespace/teams", namespaceTeamHandler.CreateWeb)
	a.DELETE("/namespaces/:namespace/teams/:team", namespaceTeamHandler.DeleteWeb)
	a.POST("/namespaces/:namespace/teams/:team", namespaceTeamHandler.UpdateWeb)

	namespaceUserHandler := &handler.WebHandler{
		EmptyStruct: func() handler.CObject {
			return &models.NamespaceUser{}
		},
	}
	a.GET("/namespaces/:namespace/users", namespaceUserHandler.ReadAllWeb)
	a.PUT("/namespaces/:namespace/users", namespaceUserHandler.CreateWeb)
	a.DELETE("/namespaces/:namespace/users/:user", namespaceUserHandler.DeleteWeb)
	a.POST("/namespaces/:namespace/users/:user", namespaceUserHandler.UpdateWeb)

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
}

func registerCalDavRoutes(c *echo.Group) {

	// Basic auth middleware
	c.Use(middleware.BasicAuth(caldavBasicAuth))

	// THIS is the entry point for caldav clients, otherwise lists will show up double
	c.Any("", caldav.EntryHandler)
	c.Any("/", caldav.EntryHandler)
	c.Any("/principals/*/", caldav.PrincipalHandler)
	c.Any("/lists", caldav.ListHandler)
	c.Any("/lists/", caldav.ListHandler)
	c.Any("/lists/:list", caldav.ListHandler)
	c.Any("/lists/:list/", caldav.ListHandler)
	c.Any("/lists/:list/:task", caldav.TaskHandler) // Mostly used for editing
}

func caldavBasicAuth(username, password string, c echo.Context) (bool, error) {
	creds := &models.UserLogin{
		Username: username,
		Password: password,
	}
	u, err := models.CheckUserCredentials(creds)
	if err != nil {
		log.Errorf("Error during basic auth for caldav: %v", err)
		return false, nil
	}
	// Save the user in echo context for later use
	c.Set("userBasicAuth", u)
	return true, nil
}
