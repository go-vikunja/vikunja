package routes

import (
	"code.vikunja.io/api/frontend"
	"github.com/labstack/echo/v4/middleware"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

func setupStaticFrontendFilesHandler(e *echo.Echo) {
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level:     6,
		MinLength: 256,
		Skipper: func(c echo.Context) bool {
			return strings.HasPrefix(c.Path(), "/api/")
		},
	}))

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if strings.HasPrefix(c.Path(), "/api/") {
				return next(c)
			}

			c.Response().Header().Set("Server", "Vikunja")
			c.Response().Header().Set("Vary", "Accept-Encoding")

			// TODO how to get last modified and etag header?
			// Cache-Control: https://www.rfc-editor.org/rfc/rfc9111#section-5.2
			/*

			   nginx returns these headers:

			   --content-encoding: gzip
			   --content-type: text/html; charset=utf-8
			   --date: Thu, 08 Feb 2024 15:53:23 GMT
			   etag: W/"65c39587-bf7"
			   --last-modified: Wed, 07 Feb 2024 14:36:55 GMT
			   --server: nginx
			   --vary: Accept-Encoding
			   cache-control: public, max-age=0, s-maxage=0, must-revalidate

			*/

			return next(c)
		}
	})

	e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Filesystem: http.FS(frontend.Files),
		HTML5:      true,
		Root:       "dist/",
	}))
}
