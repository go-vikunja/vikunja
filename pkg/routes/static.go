// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package routes

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"text/template"

	"code.vikunja.io/api/frontend"
	"code.vikunja.io/api/pkg/config"

	etaggenerator "github.com/hhsnopek/etag"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const (
	indexFile               = `index.html`
	rootPath                = `dist/`
	cacheControlMax         = `max-age=315360000, public, max-age=31536000, s-maxage=31536000, immutable`
	cacheControlNone        = `public, max-age=0, s-maxage=0, must-revalidate`
	configScriptTagTemplate = `
<script>
	window.SENTRY_ENABLED = {{ .SENTRY_ENABLED }}
	window.SENTRY_DSN = '{{ .SENTRY_DSN }}'
	window.ALLOW_ICON_CHANGES = {{ .ALLOW_ICON_CHANGES }}
	window.CUSTOM_LOGO_URL = '{{ .CUSTOM_LOGO_URL }}'
</script>`
)

// Because the files are embedded into the final binary, we can be absolutely sure the etag will never change
// and we can cache its generation pretty heavily.
var etagCache map[string]string
var etagLock sync.Mutex
var scriptConfigString string
var scriptConfigStringLock sync.Mutex

func init() {
	etagCache = make(map[string]string)
	etagLock = sync.Mutex{}
	scriptConfigStringLock = sync.Mutex{}
}

func serveIndexFile(c echo.Context, assetFs http.FileSystem) (err error) {
	index, err := assetFs.Open(path.Join(rootPath, indexFile))
	if err != nil {
		return err
	}
	defer index.Close()

	if scriptConfigString == "" {

		scriptConfigStringLock.Lock()
		defer scriptConfigStringLock.Unlock()

		// replace config variables
		tmpl, err := template.New("config").Parse(configScriptTagTemplate)
		if err != nil {
			return err
		}
		var tplOutput bytes.Buffer
		data := make(map[string]string)

		data["SENTRY_ENABLED"] = "false"
		if config.SentryFrontendEnabled.GetBool() {
			data["SENTRY_ENABLED"] = "true"
		}
		data["SENTRY_DSN"] = config.SentryFrontendDsn.GetString()
		data["ALLOW_ICON_CHANGES"] = "false"
		if config.ServiceAllowIconChanges.GetBool() {
			data["ALLOW_ICON_CHANGES"] = "true"
		}
		data["CUSTOM_LOGO_URL"] = config.ServiceCustomLogoURL.GetString()

		err = tmpl.Execute(&tplOutput, data)
		if err != nil {
			return err
		}
		scriptConfig := tplOutput.String()

		buf := bytes.Buffer{}
		_, err = buf.ReadFrom(index)
		if err != nil {
			return err
		}

		scriptConfigString = strings.ReplaceAll(buf.String(), `<div id="app"></div>`, `<div id="app"></div>`+scriptConfig)

		publicURL := config.ServicePublicURL.GetString()
		if publicURL == "" {
			publicURL = "/"
		}

		scriptConfigString = strings.ReplaceAll(scriptConfigString, "'http://localhost:3456/api/v1'", "'"+publicURL+"api/v1'")
	}

	reader := strings.NewReader(scriptConfigString)

	info, err := index.Stat()
	if err != nil {
		return err
	}

	//etag, err := generateEtag(index, info.Name())
	//if err != nil {
	//	return err
	//}
	return serveFile(c, reader, info, "")
}

// Copied from echo's middleware.StaticWithConfig simplified and adjusted for caching
func static() echo.MiddlewareFunc {
	assetFs := http.FS(frontend.Files)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			p := c.Request().URL.Path
			if strings.HasPrefix(p, "/api/") {
				return next(c)
			}
			if strings.HasSuffix(c.Path(), "*") { // When serving from a group, e.g. `/static*`.
				p = c.Param("*")
			}
			p, err = url.PathUnescape(p)
			if err != nil {
				return
			}
			name := path.Join(rootPath, path.Clean("/"+p)) // "/"+ for security

			file, err := assetFs.Open(name)
			if err != nil {
				if !os.IsNotExist(err) {
					return err
				}

				// file with that path did not exist, so we continue down in middleware/handler chain, hoping that we end up in
				// handler that is meant to handle this request
				if err = next(c); err == nil {
					return err
				}

				var he *echo.HTTPError
				if !errors.As(err, &he) || he.Code != http.StatusNotFound {
					return err
				}

				// Handle all other requests with the index file
				return serveIndexFile(c, assetFs)
			}

			defer file.Close()

			info, err := file.Stat()
			if err != nil {
				return err
			}

			if info.IsDir() {
				return serveIndexFile(c, assetFs)
			}

			etag, err := generateEtag(file, name)
			if err != nil {
				return err
			}

			return serveFile(c, file, info, etag)
		}
	}
}

func generateEtag(file http.File, name string) (etag string, err error) {
	etagLock.Lock()
	defer etagLock.Unlock()
	etag, has := etagCache[name]
	if !has {
		buf := bytes.Buffer{}
		_, err = buf.ReadFrom(file)
		if err != nil {
			return "", err
		}
		etag = etaggenerator.Generate(buf.Bytes(), true)
		etagCache[name] = etag
	}

	return etag, nil
}

// copied from http.serveContent
func getMimeType(name string, file io.ReadSeeker) (mineType string, err error) {
	mineType = mime.TypeByExtension(filepath.Ext(name))
	if mineType == "" {
		// read a chunk to decide between utf-8 text and binary
		var buf [512]byte
		n, _ := io.ReadFull(file, buf[:])
		mineType = http.DetectContentType(buf[:n])
		_, err := file.Seek(0, io.SeekStart) // rewind to output whole file
		if err != nil {
			return "", fmt.Errorf("seeker can't seek")
		}
	}

	return mineType, nil
}

func getCacheControlHeader(info os.FileInfo, file io.ReadSeeker) (header string, err error) {
	// Don't cache service worker and related files
	if info.Name() == "robots.txt" ||
		info.Name() == "sw.js" ||
		info.Name() == "manifest.webmanifest" {
		return cacheControlNone, nil
	}

	if strings.HasPrefix(info.Name(), "workbox-") {
		return cacheControlMax, nil
	}

	contentType, err := getMimeType(info.Name(), file)
	if err != nil {
		return "", err
	}

	// Cache everything looking like an asset
	if strings.HasPrefix(contentType, "image/") ||
		strings.HasPrefix(contentType, "font/") ||
		strings.HasPrefix(contentType, "~images/") ||
		strings.HasPrefix(contentType, "~font/") ||
		contentType == "text/css" ||
		contentType == "application/javascript" ||
		contentType == "text/javascript" ||
		contentType == "application/vnd.ms-fontobject" ||
		contentType == "application/x-font-ttf" ||
		contentType == "font/opentype" ||
		contentType == "font/woff2" ||
		contentType == "image/svg+xml" ||
		contentType == "image/x-icon" ||
		contentType == "audio/wav" {
		return cacheControlMax, nil
	}

	return cacheControlNone, nil
}

func serveFile(c echo.Context, file io.ReadSeeker, info os.FileInfo, etag string) error {

	c.Response().Header().Set("Server", "Vikunja")
	c.Response().Header().Set("Vary", "Accept-Encoding")
	if etag != "" {
		c.Response().Header().Set("Etag", etag)
	}

	cacheControl, err := getCacheControlHeader(info, file)
	if err != nil {
		return err
	}
	c.Response().Header().Set("Cache-Control", cacheControl)

	http.ServeContent(c.Response(), c.Request(), info.Name(), info.ModTime(), file)
	return nil
}

func setupStaticFrontendFilesHandler(e *echo.Echo) {
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level:     6,
		MinLength: 256,
		Skipper: func(c echo.Context) bool {
			return strings.HasPrefix(c.Path(), "/api/")
		},
	}))

	e.Use(static())
}
