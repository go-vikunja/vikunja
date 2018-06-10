package routes

import (
	"github.com/labstack/echo"
	"net/http"
)

// SetCORSHeader sets relevant CORS headers for Cross-Site-Requests to the api
func SetCORSHeader(c echo.Context) error {
	res := c.Response()
	res.Header().Set("Access-Control-Allow-Origin", "*")
	res.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
	res.Header().Set("Access-Control-Allow-Headers", "authorization,content-type")
	res.Header().Set("Access-Control-Expose-Headers", "authorization,content-type")
	return c.String(http.StatusOK, "")
}
