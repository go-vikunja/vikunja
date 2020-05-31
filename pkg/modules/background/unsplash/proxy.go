// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2020 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package unsplash

import (
	"code.vikunja.io/web/handler"
	"github.com/labstack/echo/v4"
	"net/http"
)

func unsplashImage(url string, c echo.Context) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	if resp.StatusCode > 399 {
		return echo.ErrNotFound
	}
	return c.Stream(http.StatusOK, "image/jpg", resp.Body)
}

// ProxyUnsplashImage proxies a thumbnail from unsplash for privacy reasons.
// @Summary Get an unsplash image
// @Description Get an unsplash image. **Returns json on error.**
// @tags list
// @Produce octet-stream
// @Param thumb path int true "Unsplash Image ID"
// @Security JWTKeyAuth
// @Success 200 {} string "The image"
// @Failure 404 {object} models.Message "The image does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /backgrounds/unsplash/image/{image} [get]
func ProxyUnsplashImage(c echo.Context) error {
	photo, err := getUnsplashPhotoInfoByID(c.Param("image"))
	if err != nil {
		return handler.HandleHTTPError(err, c)
	}
	pingbackByPhotoID(photo.ID)
	return unsplashImage(photo.Urls.Raw, c)
}

// ProxyUnsplashThumb proxies a thumbnail from unsplash for privacy reasons.
// @Summary Get an unsplash thumbnail image
// @Description Get an unsplash thumbnail image. The thumbnail is cropped to a max width of 200px. **Returns json on error.**
// @tags list
// @Produce octet-stream
// @Param thumb path int true "Unsplash Image ID"
// @Security JWTKeyAuth
// @Success 200 {} string "The thumbnail"
// @Failure 404 {object} models.Message "The image does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /backgrounds/unsplash/image/{image}/thumb [get]
func ProxyUnsplashThumb(c echo.Context) error {
	photo, err := getUnsplashPhotoInfoByID(c.Param("image"))
	if err != nil {
		return handler.HandleHTTPError(err, c)
	}
	pingbackByPhotoID(photo.ID)
	return unsplashImage("https://images.unsplash.com/photo-"+getImageID(photo.Urls.Raw)+"?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=200&fit=max&ixid=eyJhcHBfaWQiOjcyODAwfQ", c)
}
