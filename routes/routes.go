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
	a.OPTIONS("/register", SetCORSHeader)
	a.OPTIONS("/users", SetCORSHeader)
	a.OPTIONS("/users/:id", SetCORSHeader)
	a.OPTIONS("/lists", SetCORSHeader)
	a.OPTIONS("/lists/:id", SetCORSHeader)

	a.POST("/login", apiv1.Login)
	a.POST("/register", apiv1.UserAddOrUpdate)

	// ===== Routes with Authetification =====
	// Authetification
	a.Use(middleware.JWT(models.Config.JWTLoginSecret))
	a.POST("/tokenTest", apiv1.CheckToken)

	a.PUT("/lists", apiv1.AddOrUpdateList)
	a.GET("/lists", apiv1.GetListsByUser)
	a.GET("/lists/:id", apiv1.GetListByID)
	a.POST("/lists/:id", apiv1.AddOrUpdateList)
	a.PUT("/lists/:id", apiv1.AddOrUpdateListItem)
	a.DELETE("/lists/:id", apiv1.DeleteListByID)

	a.DELETE("/item/:id", apiv1.DeleteListItemByIDtemByID)
}
