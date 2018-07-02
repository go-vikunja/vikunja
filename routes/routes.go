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

	"git.kolaente.de/konrad/list/models"
	apiv1 "git.kolaente.de/konrad/list/routes/api/v1"
	_ "git.kolaente.de/konrad/list/routes/api/v1/swagger" // for docs generation
)

// NewEcho registers a new Echo instance
func NewEcho() *echo.Echo {
	e := echo.New()

	// Logger
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${time_rfc3339}: ${remote_ip} ${method} ${status} ${uri} ${latency_human} - ${user_agent}\n",
	}))

	return e
}

// RegisterRoutes registers all routes for the application
func RegisterRoutes(e *echo.Echo) {

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

	a.GET("/lists", apiv1.GetListsByUser)
	a.PUT("/lists", apiv1.AddList)
	a.GET("/lists/:id", apiv1.GetListByID)
	a.POST("/lists/:id", apiv1.UpdateList)
	a.PUT("/lists/:id", apiv1.AddListItem)
	a.DELETE("/lists/:id", apiv1.DeleteListByID)

	a.DELETE("/item/:id", apiv1.DeleteListItemByIDtemByID)
	a.POST("/item/:id", apiv1.UpdateListItem)

	a.GET("/namespaces", apiv1.GetAllNamespacesByCurrentUser)
	a.PUT("/namespaces", apiv1.AddNamespace)
	a.GET("/namespaces/:id", apiv1.ShowNamespace)
	//a.GET("/namespaces/:id/lists") // Gets all lists for that namespace
	a.POST("/namespaces/:id", apiv1.UpdateNamespace)
	//a.PUT("/namespaces/:id") // Creates a new list in that namespace
	// a.DELETE("/namespaces/:id") // Deletes a namespace with all lists
}
