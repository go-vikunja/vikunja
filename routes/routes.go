package routes

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"git.kolaente.de/konrad/list/models"
	apiv1 "git.kolaente.de/konrad/list/routes/api/v1"
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

	// API Routes
	a := e.Group("/api/v1")

	// CORS_SHIT
	a.OPTIONS("/login", SetCORSHeader)
	a.OPTIONS("/users", SetCORSHeader)
	a.OPTIONS("/users/:id", SetCORSHeader)

	a.POST("/login", Login)

	// ===== Routes with Authetification =====
	// Authetification
	a.Use(middleware.JWT(models.Config.JWTLoginSecret))
	a.POST("/tokenTest", apiv1.CheckToken)
}
