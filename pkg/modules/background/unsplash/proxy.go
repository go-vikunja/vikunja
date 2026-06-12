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

package unsplash

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"

	"code.vikunja.io/api/pkg/utils"
	"code.vikunja.io/api/pkg/web"

	"github.com/labstack/echo/v5"
)

// ErrUnsplashImageDoesNotExist is returned when Unsplash answers an image proxy fetch
// with a non-success status, mirroring v1's echo.ErrNotFound. It satisfies
// web.HTTPErrorProcessor so the v2 error bridge maps it to a 404.
type ErrUnsplashImageDoesNotExist struct{}

// IsErrUnsplashImageDoesNotExist checks if an error is ErrUnsplashImageDoesNotExist.
func IsErrUnsplashImageDoesNotExist(err error) bool {
	var target *ErrUnsplashImageDoesNotExist
	return errors.As(err, &target)
}

func (err *ErrUnsplashImageDoesNotExist) Error() string {
	return "Unsplash image does not exist"
}

// HTTPError holds the http error description.
func (err *ErrUnsplashImageDoesNotExist) HTTPError() web.HTTPError {
	return web.HTTPError{HTTPCode: http.StatusNotFound, Message: "Not Found"}
}

// fetchUnsplashImage fetches an image from Unsplash through the SSRF-safe client and
// returns its still-open response body for the caller to stream and close. The url is
// rebased onto the hardcoded images.unsplash.com host (stripping any client-supplied
// host) so the proxy can only ever reach Unsplash. It returns
// ErrUnsplashImageDoesNotExist on a non-success upstream status.
func fetchUnsplashImage(url string) (io.ReadCloser, error) {
	// Replacing and appending the url for security reasons
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://images.unsplash.com/"+strings.Replace(url, "https://images.unsplash.com/", "", 1), nil)
	if err != nil {
		return nil, err
	}
	resp, err := utils.NewSSRFSafeHTTPClient().Do(req) //nolint:gosec // SSRF protection is handled by the SSRF-safe client
	if err != nil {
		return nil, err
	}
	if resp.StatusCode > 399 {
		_ = resp.Body.Close()
		return nil, &ErrUnsplashImageDoesNotExist{}
	}
	return resp.Body, nil
}

// FetchUnsplashImageByID resolves an Unsplash image by id, fires the required pingback,
// and returns the full-resolution image body for the caller to stream and close.
func FetchUnsplashImageByID(imageID string) (io.ReadCloser, error) {
	photo, err := getUnsplashPhotoInfoByID(imageID)
	if err != nil {
		return nil, err
	}
	pingbackByPhotoID(photo.ID)
	return fetchUnsplashImage(photo.Urls.Raw)
}

// FetchUnsplashThumbByID resolves an Unsplash image by id, fires the required pingback,
// and returns a thumbnail (max width 200px) body for the caller to stream and close.
func FetchUnsplashThumbByID(imageID string) (io.ReadCloser, error) {
	photo, err := getUnsplashPhotoInfoByID(imageID)
	if err != nil {
		return nil, err
	}
	pingbackByPhotoID(photo.ID)
	return fetchUnsplashImage("https://images.unsplash.com/" + getImageID(photo.Urls.Raw) + "?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=200&fit=max&ixid=eyJhcHBfaWQiOjcyODAwfQ")
}

// streamUnsplashImage streams a fetched image body to the v1 echo response, mapping the
// not-found sentinel back to echo.ErrNotFound so v1's wire response is unchanged.
func streamUnsplashImage(body io.ReadCloser, err error, c *echo.Context) error {
	if err != nil {
		if IsErrUnsplashImageDoesNotExist(err) {
			return echo.ErrNotFound
		}
		return err
	}
	defer body.Close()
	return c.Stream(http.StatusOK, "image/jpg", body)
}

// ProxyUnsplashImage proxies an image from unsplash for privacy reasons.
// @Summary Get an unsplash image
// @Description Get an unsplash image. **Returns json on error.**
// @tags project
// @Produce octet-stream
// @Param image path int true "Unsplash Image ID"
// @Security JWTKeyAuth
// @Success 200 {file} blob "The image"
// @Failure 404 {object} models.Message "The image does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /backgrounds/unsplash/image/{image} [get]
func ProxyUnsplashImage(c *echo.Context) error {
	body, err := FetchUnsplashImageByID(c.Param("image"))
	return streamUnsplashImage(body, err, c)
}

// ProxyUnsplashThumb proxies a thumbnail from unsplash for privacy reasons.
// @Summary Get an unsplash thumbnail image
// @Description Get an unsplash thumbnail image. The thumbnail is cropped to a max width of 200px. **Returns json on error.**
// @tags project
// @Produce octet-stream
// @Param image path int true "Unsplash Image ID"
// @Security JWTKeyAuth
// @Success 200 {file} blob "The thumbnail"
// @Failure 404 {object} models.Message "The image does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /backgrounds/unsplash/image/{image}/thumb [get]
func ProxyUnsplashThumb(c *echo.Context) error {
	body, err := FetchUnsplashThumbByID(c.Param("image"))
	return streamUnsplashImage(body, err, c)
}
