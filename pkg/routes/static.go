package routes

import (
	"code.vikunja.io/api/frontend"
	"github.com/labstack/echo/v4/middleware"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

func staticWithConfig() echo.MiddlewareFunc {
	// Defaults
	if config.Root == "" {
		config.Root = "." // For security we want to restrict to CWD.
	}
	if config.Skipper == nil {
		config.Skipper = DefaultStaticConfig.Skipper
	}
	if config.Index == "" {
		config.Index = DefaultStaticConfig.Index
	}
	if config.Filesystem == nil {
		config.Filesystem = http.Dir(config.Root)
		config.Root = "."
	}

	// Index template
	t, tErr := template.New("index").Parse(html)
	if tErr != nil {
		panic(fmt.Errorf("echo: %w", tErr))
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			if config.Skipper(c) {
				return next(c)
			}

			p := c.Request().URL.Path
			if strings.HasSuffix(c.Path(), "*") { // When serving from a group, e.g. `/static*`.
				p = c.Param("*")
			}
			p, err = url.PathUnescape(p)
			if err != nil {
				return
			}
			name := path.Join(config.Root, path.Clean("/"+p)) // "/"+ for security

			if config.IgnoreBase {
				routePath := path.Base(strings.TrimRight(c.Path(), "/*"))
				baseURLPath := path.Base(p)
				if baseURLPath == routePath {
					i := strings.LastIndex(name, routePath)
					name = name[:i] + strings.Replace(name[i:], routePath, "", 1)
				}
			}

			file, err := config.Filesystem.Open(name)
			if err != nil {
				if !isIgnorableOpenFileError(err) {
					return err
				}

				// file with that path did not exist, so we continue down in middleware/handler chain, hoping that we end up in
				// handler that is meant to handle this request
				if err = next(c); err == nil {
					return err
				}

				var he *echo.HTTPError
				if !(errors.As(err, &he) && config.HTML5 && he.Code == http.StatusNotFound) {
					return err
				}

				file, err = config.Filesystem.Open(path.Join(config.Root, config.Index))
				if err != nil {
					return err
				}
			}

			defer file.Close()

			info, err := file.Stat()
			if err != nil {
				return err
			}

			if info.IsDir() {
				index, err := config.Filesystem.Open(path.Join(name, config.Index))
				if err != nil {
					if config.Browse {
						return listDir(t, name, file, c.Response())
					}

					return next(c)
				}

				defer index.Close()

				info, err = index.Stat()
				if err != nil {
					return err
				}

				return serveFile(c, index, info)
			}

			return serveFile(c, file, info)
		}
	}
}

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
