// Package v1 List API.
//
// This documentation describes the List API.
//
//     Schemes: http, https
//     BasePath: /api/v1
//     Version: 0.1
//     License: GPLv3
//
//     Consumes:
//     - application/json
//
//     Produces:
//     - application/json
//
//     Security:
//     - AuthorizationHeaderToken :
//
//     SecurityDefinitions:
//     AuthorizationHeaderToken:
//          type: apiKey
//          name: Authorization
//          in: header
//
// swagger:meta

package routes

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"code.vikunja.io/api/models"
	apiv1 "code.vikunja.io/api/routes/api/v1"
	_ "code.vikunja.io/api/routes/api/v1/swagger" // for docs generation
	"code.vikunja.io/api/routes/crud"
	"github.com/spf13/viper"
)

// NewEcho registers a new Echo instance
func NewEcho() *echo.Echo {
	e := echo.New()

	// Logger
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${time_rfc3339_nano}: ${remote_ip} ${method} ${status} ${uri} ${latency_human} - ${user_agent}\n",
	}))

	return e
}

// RegisterRoutes registers all routes for the application
func RegisterRoutes(e *echo.Echo) {

	// CORS_SHIT
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
	}))

	// API Routes
	a := e.Group("/api/v1")

	// Swagger UI
	a.Static("/swagger", "public/swagger")

	a.POST("/login", apiv1.Login)
	a.POST("/register", apiv1.RegisterUser)
	a.POST("/user/password/token", apiv1.UserRequestResetPasswordToken)
	a.POST("/user/password/reset", apiv1.UserResetPassword)

	// ===== Routes with Authetification =====
	// Authetification
	a.Use(middleware.JWT([]byte(viper.GetString("service.JWTSecret"))))
	a.POST("/tokenTest", apiv1.CheckToken)

	// User stuff
	a.GET("/user", apiv1.UserShow)
	a.POST("/user/password", apiv1.UserChangePassword)
	a.GET("/users", apiv1.UserList)

	listHandler := &crud.WebHandler{
		EmptyStruct: func() crud.CObject {
			return &models.List{}
		},
	}
	a.GET("/lists", listHandler.ReadAllWeb)
	a.GET("/lists/:list", listHandler.ReadOneWeb)
	a.POST("/lists/:list", listHandler.UpdateWeb)
	a.DELETE("/lists/:list", listHandler.DeleteWeb)
	a.PUT("/namespaces/:namespace/lists", listHandler.CreateWeb)

	taskHandler := &crud.WebHandler{
		EmptyStruct: func() crud.CObject {
			return &models.ListTask{}
		},
	}
	a.PUT("/lists/:list", taskHandler.CreateWeb)
	a.DELETE("/tasks/:listtask", taskHandler.DeleteWeb)
	a.POST("/tasks/:listtask", taskHandler.UpdateWeb)

	listTeamHandler := &crud.WebHandler{
		EmptyStruct: func() crud.CObject {
			return &models.TeamList{}
		},
	}
	a.GET("/lists/:list/teams", listTeamHandler.ReadAllWeb)
	a.PUT("/lists/:list/teams", listTeamHandler.CreateWeb)
	a.DELETE("/lists/:list/teams/:team", listTeamHandler.DeleteWeb)
	a.POST("/lists/:list/teams/:team", listTeamHandler.UpdateWeb)

	listUserHandler := &crud.WebHandler{
		EmptyStruct: func() crud.CObject {
			return &models.ListUser{}
		},
	}
	a.GET("/lists/:list/users", listUserHandler.ReadAllWeb)
	a.PUT("/lists/:list/users", listUserHandler.CreateWeb)
	a.DELETE("/lists/:list/users/:user", listUserHandler.DeleteWeb)
	a.POST("/lists/:list/users/:user", listUserHandler.UpdateWeb)

	namespaceHandler := &crud.WebHandler{
		EmptyStruct: func() crud.CObject {
			return &models.Namespace{}
		},
	}
	a.GET("/namespaces", namespaceHandler.ReadAllWeb)
	a.PUT("/namespaces", namespaceHandler.CreateWeb)
	a.GET("/namespaces/:namespace", namespaceHandler.ReadOneWeb)
	a.POST("/namespaces/:namespace", namespaceHandler.UpdateWeb)
	a.DELETE("/namespaces/:namespace", namespaceHandler.DeleteWeb)
	a.GET("/namespaces/:namespace/lists", apiv1.GetListsByNamespaceID)

	namespaceTeamHandler := &crud.WebHandler{
		EmptyStruct: func() crud.CObject {
			return &models.TeamNamespace{}
		},
	}
	a.GET("/namespaces/:namespace/teams", namespaceTeamHandler.ReadAllWeb)
	a.PUT("/namespaces/:namespace/teams", namespaceTeamHandler.CreateWeb)
	a.DELETE("/namespaces/:namespace/teams/:team", namespaceTeamHandler.DeleteWeb)
	a.POST("/namespaces/:namespace/teams/:team", namespaceTeamHandler.UpdateWeb)

	namespaceUserHandler := &crud.WebHandler{
		EmptyStruct: func() crud.CObject {
			return &models.NamespaceUser{}
		},
	}
	a.GET("/namespaces/:namespace/users", namespaceUserHandler.ReadAllWeb)
	a.PUT("/namespaces/:namespace/users", namespaceUserHandler.CreateWeb)
	a.DELETE("/namespaces/:namespace/users/:user", namespaceUserHandler.DeleteWeb)
	a.POST("/namespaces/:namespace/users/:user", namespaceUserHandler.UpdateWeb)

	teamHandler := &crud.WebHandler{
		EmptyStruct: func() crud.CObject {
			return &models.Team{}
		},
	}
	a.GET("/teams", teamHandler.ReadAllWeb)
	a.GET("/teams/:team", teamHandler.ReadOneWeb)
	a.PUT("/teams", teamHandler.CreateWeb)
	a.POST("/teams/:team", teamHandler.UpdateWeb)
	a.DELETE("/teams/:team", teamHandler.DeleteWeb)

	teamMemberHandler := &crud.WebHandler{
		EmptyStruct: func() crud.CObject {
			return &models.TeamMember{}
		},
	}
	a.PUT("/teams/:team/members", teamMemberHandler.CreateWeb)
	a.DELETE("/teams/:team/members/:user", teamMemberHandler.DeleteWeb)
}
