package routes

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

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"git.kolaente.de/konrad/list/models"
	apiv1 "git.kolaente.de/konrad/list/routes/api/v1"
	_ "git.kolaente.de/konrad/list/routes/api/v1/swagger" // for docs generation
	"git.kolaente.de/konrad/list/routes/crud"
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

	// TODO: Use proper cors middleware by echo

	// Middleware for cors
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			res := c.Response()
			res.Header().Set("Access-Control-Allow-Origin", "*")
			res.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
			res.Header().Set("Access-Control-Allow-Headers", "authorization,content-type")
			res.Header().Set("Access-Control-Expose-Headers", "authorization,content-type")
			return next(c)
		}
	})

	// Swagger UI
	e.Static("/swagger", "public/swagger")

	// API Routes
	a := e.Group("/api/v1")

	// CORS_SHIT
	a.OPTIONS("/login", SetCORSHeader)
	a.OPTIONS("/register", SetCORSHeader)
	a.OPTIONS("/users", SetCORSHeader)
	a.OPTIONS("/users/:id", SetCORSHeader)
	a.OPTIONS("/lists", SetCORSHeader)
	a.OPTIONS("/lists/:id", SetCORSHeader)

	a.POST("/login", apiv1.Login)
	a.POST("/register", apiv1.RegisterUser)

	// ===== Routes with Authetification =====
	// Authetification
	a.Use(middleware.JWT(models.Config.JWTLoginSecret))
	a.POST("/tokenTest", apiv1.CheckToken)

	listHandler := &crud.WebHandler{
		CObject: &models.List{},
	}
	a.GET("/lists", listHandler.ReadAllWeb)
	a.GET("/lists/:id", listHandler.ReadOneWeb)
	a.POST("/lists/:id", listHandler.UpdateWeb)
	a.DELETE("/lists/:id", listHandler.DeleteWeb)
	a.PUT("/namespaces/:id/lists", listHandler.CreateWeb)

	itemHandler := &crud.WebHandler{
		CObject: &models.ListItem{},
	}
	a.PUT("/lists/:id", itemHandler.CreateWeb)
	a.DELETE("/items/:id", itemHandler.DeleteWeb)
	a.POST("/items/:id", itemHandler.UpdateWeb)

	namespaceHandler := &crud.WebHandler{
		CObject: &models.Namespace{},
	}
	a.GET("/namespaces", namespaceHandler.ReadAllWeb)
	a.PUT("/namespaces", namespaceHandler.CreateWeb)
	a.GET("/namespaces/:id", namespaceHandler.ReadOneWeb)
	a.POST("/namespaces/:id", namespaceHandler.UpdateWeb)
	a.DELETE("/namespaces/:id", namespaceHandler.DeleteWeb)
	a.GET("/namespaces/:id/lists", apiv1.GetListsByNamespaceID)

	teamHandler := &crud.WebHandler{
		CObject: &models.Team{},
	}
	a.GET("/teams", teamHandler.ReadAllWeb)
	a.GET("/teams/:id", teamHandler.ReadOneWeb)
	a.PUT("/teams", teamHandler.CreateWeb)
	a.POST("/teams/:id", teamHandler.UpdateWeb)
	a.DELETE("/teams/:id", teamHandler.DeleteWeb)
}
