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
	"net/http"
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

	a.POST("/test/:infi/:Käsebrot/blub/:gedöns", func(c echo.Context) error {

		type testStruct struct {
			Integ  int64  `param:"infi" form:"infi"`
			Cheese string `param:"Käsebrot"`
			Kram   string `param:"gedöns"`
			Other  string
			Whooo  int64
			Blub   float64
			Test   string `form:"test"`
		}

		t := testStruct{}

		if err := crud.ParamBinder(&t, c); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "No or invalid model provided.")
		}

		return c.JSON(http.StatusOK, &t)
	})

	// ===== Routes with Authetification =====
	// Authetification
	a.Use(middleware.JWT(models.Config.JWTLoginSecret))
	a.POST("/tokenTest", apiv1.CheckToken)

	listHandler := &crud.WebHandler{
		CObject: &models.List{},
	}
	a.GET("/lists", listHandler.ReadAllWeb)
	a.GET("/lists/:list", listHandler.ReadOneWeb)
	a.POST("/lists/:list", listHandler.UpdateWeb)
	a.DELETE("/lists/:list", listHandler.DeleteWeb)
	a.PUT("/namespaces/:namespace/lists", listHandler.CreateWeb)

	itemHandler := &crud.WebHandler{
		CObject: &models.ListItem{},
	}
	a.PUT("/lists/:list", itemHandler.CreateWeb)
	a.DELETE("/items/:listitem", itemHandler.DeleteWeb)
	a.POST("/items/:listitem", itemHandler.UpdateWeb)

	namespaceHandler := &crud.WebHandler{
		CObject: &models.Namespace{},
	}
	a.GET("/namespaces", namespaceHandler.ReadAllWeb)
	a.PUT("/namespaces", namespaceHandler.CreateWeb)
	a.GET("/namespaces/:namespace", namespaceHandler.ReadOneWeb)
	a.POST("/namespaces/:namespace", namespaceHandler.UpdateWeb)
	a.DELETE("/namespaces/:namespace", namespaceHandler.DeleteWeb)
	a.GET("/namespaces/:namespace/lists", apiv1.GetListsByNamespaceID)

	namespaceTeamHandler := &crud.WebHandler{
		CObject: &models.TeamNamespace{},
	}
	a.GET("/namespaces/:namespace/teams", namespaceTeamHandler.ReadAllWeb)
	a.PUT("/namespaces/:namespace/teams", namespaceTeamHandler.CreateWeb)
	a.DELETE("/namespaces/:namespace/teams/:team", namespaceTeamHandler.DeleteWeb)

	teamHandler := &crud.WebHandler{
		CObject: &models.Team{},
	}
	a.GET("/teams", teamHandler.ReadAllWeb)
	a.GET("/teams/:team", teamHandler.ReadOneWeb)
	a.PUT("/teams", teamHandler.CreateWeb)
	a.POST("/teams/:team", teamHandler.UpdateWeb)
	a.DELETE("/teams/:team", teamHandler.DeleteWeb)
}
