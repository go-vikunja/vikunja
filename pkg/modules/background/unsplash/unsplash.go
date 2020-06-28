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
	"bytes"
	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/background"
	"code.vikunja.io/web"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const unsplashAPIURL = `https://api.unsplash.com/`

// Provider represents an unsplash image provider
type Provider struct {
}

// SearchResult is a search result from unsplash's api
type SearchResult struct {
	Total      int      `json:"total"`
	TotalPages int      `json:"total_pages"`
	Results    []*Photo `json:"results"`
}

// Photo represents an unpslash photo as returned by their api
type Photo struct {
	ID          string `json:"id"`
	CreatedAt   string `json:"created_at"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	Color       string `json:"color"`
	Description string `json:"description"`
	User        struct {
		Username string `json:"username"`
		Name     string `json:"name"`
	} `json:"user"`
	Urls struct {
		Raw     string `json:"raw"`
		Full    string `json:"full"`
		Regular string `json:"regular"`
		Small   string `json:"small"`
		Thumb   string `json:"thumb"`
	} `json:"urls"`
	Links struct {
		Self             string `json:"self"`
		HTML             string `json:"html"`
		Download         string `json:"download"`
		DownloadLocation string `json:"download_location"`
	} `json:"links"`
}

// Very simple caching method - pretty much only used to retain information when saving an image
// FIXME: Should use a proper cache
var photos map[string]*Photo

// We're caching the initial collection to save a few api requests as this is retrieved every time a
// user opens the settings page.
type initialCollection struct {
	lastCached time.Time
	// images contains a slice of images by page they belong to
	// this allows us to cache individual pages.
	images map[int64][]*background.Image
}

var emptySearchResult *initialCollection

func init() {
	photos = make(map[string]*Photo)
}

func doGet(url string, result ...interface{}) (err error) {
	req, err := http.NewRequest("GET", unsplashAPIURL+url, nil)
	if err != nil {
		return
	}

	req.Header.Add("Authorization", "Client-ID "+config.BackgroundsUnsplashAccessToken.GetString())
	hc := http.Client{}
	resp, err := hc.Do(req)
	if err != nil {
		return
	}

	if len(result) > 0 {
		return json.NewDecoder(resp.Body).Decode(result[0])
	}

	return
}

func getImageID(fullURL string) string {
	// Unsplash image urls have the form
	// https://images.unsplash.com/photo-1590622878565-c662a7fd1394?ixlib=rb-1.2.1&q=80&fm=jpg&crop=entropy&cs=tinysrgb&w=200&fit=max&ixid=eyJhcHBfaWQiOjcyODAwfQ
	// We only need the "photo-*" part of it.
	return strings.Replace(strings.Split(fullURL, "?")[0], "https://images.unsplash.com/", "", 1)
}

// Gets an unsplash photo either from cache or directly from the unsplash api
func getUnsplashPhotoInfoByID(photoID string) (photo *Photo, err error) {
	var exists bool
	photo, exists = photos[photoID]
	if !exists {
		log.Debugf("Image information for unsplash photo %s not cached, requesting from unsplash...", photoID)
		photo = &Photo{}
		err = doGet("photos/"+photoID, photo)
		if err != nil {
			return
		}
	}
	return
}

// Search is the implementation to search on unsplash
// @Summary Search for a background from unsplash
// @Description Search for a list background from unsplash
// @tags list
// @Produce json
// @Security JWTKeyAuth
// @Param s query string false "Search backgrounds from unsplash with this search term."
// @Param p query int false "The page number. Used for pagination. If not provided, the first page of results is returned."
// @Success 200 {array} background.Image "An array with photos"
// @Failure 500 {object} models.Message "Internal error"
// @Router /backgrounds/unsplash/search [get]
func (p *Provider) Search(search string, page int64) (result []*background.Image, err error) {

	// If we don't have a search query, return results from the unsplash featured collection
	if search == "" {

		var existsForPage bool

		if emptySearchResult != nil &&
			time.Since(emptySearchResult.lastCached) < time.Hour {
			_, existsForPage = emptySearchResult.images[page]
		}

		if existsForPage {
			log.Debugf("Serving initial unsplash collection for page %d from cache, last updated at %v", page, emptySearchResult.lastCached)
			return emptySearchResult.images[page], nil
		}

		log.Debugf("Retrieving initial unsplash collection for page %d from unsplash api", page)

		collectionResult := []*Photo{}
		err = doGet("collections/317099/photos?page="+strconv.FormatInt(page, 10)+"&per_page=25&order_by=latest", &collectionResult)
		if err != nil {
			return
		}

		result = []*background.Image{}
		for _, p := range collectionResult {
			result = append(result, &background.Image{
				ID:  p.ID,
				URL: getImageID(p.Urls.Raw),
				Info: &models.UnsplashPhoto{
					UnsplashID: p.ID,
					Author:     p.User.Username,
					AuthorName: p.User.Name,
				},
			})
			photos[p.ID] = p
		}

		// Put the collection in cache
		if emptySearchResult == nil {
			emptySearchResult = &initialCollection{
				images: make(map[int64][]*background.Image),
			}
		}

		emptySearchResult.lastCached = time.Now()
		emptySearchResult.images[page] = result

		return
	}

	searchResult := &SearchResult{}
	err = doGet("search/photos?query="+url.QueryEscape(search)+"&page="+strconv.FormatInt(page, 10)+"&per_page=25", &searchResult)
	if err != nil {
		return
	}

	result = []*background.Image{}
	for _, p := range searchResult.Results {
		result = append(result, &background.Image{
			ID:  p.ID,
			URL: getImageID(p.Urls.Raw),
			Info: &models.UnsplashPhoto{
				UnsplashID: p.ID,
				Author:     p.User.Username,
				AuthorName: p.User.Name,
			},
		})
		photos[p.ID] = p
	}

	return
}

// Set sets an unsplash photo as list background
// @Summary Set an unsplash photo as list background
// @Description Sets a photo from unsplash as list background.
// @tags list
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "List ID"
// @Param list body background.Image true "The image you want to set as background"
// @Success 200 {object} models.List "The background has been successfully set."
// @Failure 400 {object} web.HTTPError "Invalid image object provided."
// @Failure 403 {object} web.HTTPError "The user does not have access to the list"
// @Failure 500 {object} models.Message "Internal error"
// @Router /lists/{id}/backgrounds/unsplash [post]
func (p *Provider) Set(image *background.Image, list *models.List, auth web.Auth) (err error) {

	// Find the photo
	photo, err := getUnsplashPhotoInfoByID(image.ID)
	if err != nil {
		return
	}

	// Download the photo from unsplash
	// The parameters crop the image to a max width of 2560 and a max height of 2048 to save bandwidth and storage.
	resp, err := http.Get(photo.Urls.Raw + "&w=2560&h=2048&q=90")
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode > 399 {
		b := bytes.Buffer{}
		_, _ = b.ReadFrom(resp.Body)
		log.Errorf("Error getting unsplash photo %s: Request failed with status %d, message was %s", photo.ID, resp.StatusCode, b.String())
		return
	}

	log.Debugf("Downloaded unsplash photo %s", image.ID)

	// Ping the unsplash download endpoint (again, unsplash api guidelines)
	err = doGet(strings.Replace(photo.Links.DownloadLocation, unsplashAPIURL, "", 1))
	if err != nil {
		return
	}
	log.Debugf("Pinged unsplash download endpoint for photo %s", image.ID)

	// Save it as a file in vikunja
	file, err := files.Create(resp.Body, "", 0, auth)
	if err != nil {
		return
	}

	// Remove the old background if one exists
	if list.BackgroundFileID != 0 {
		file := files.File{ID: list.BackgroundFileID}
		if err := file.Delete(); err != nil {
			return err
		}

		if err := models.RemoveUnsplashPhoto(list.BackgroundFileID); err != nil {
			return err
		}
	}

	// Save the relation that we got it from unsplash
	unsplashPhoto := &models.UnsplashPhoto{
		FileID:     file.ID,
		UnsplashID: image.ID,
		Author:     photo.User.Username,
		AuthorName: photo.User.Name,
	}
	err = unsplashPhoto.Save()
	if err != nil {
		return
	}
	log.Debugf("Saved unsplash photo %s as file %d with new entry %d", image.ID, file.ID, unsplashPhoto.ID)

	// Set the file in the list
	list.BackgroundFileID = file.ID
	list.BackgroundInformation = unsplashPhoto

	// Set it as the list background
	return models.SetListBackground(list.ID, file)
}

// Pingback pings the unsplash api if an unsplash photo has been accessed.
func Pingback(f *files.File) {
	// Check if the file is actually downloaded from unsplash
	unsplashPhoto, err := models.GetUnsplashPhotoByFileID(f.ID)
	if err != nil {
		if files.IsErrFileIsNotUnsplashFile(err) {
			return
		}
		log.Errorf("Unsplash Pingback: %s", err.Error())
	}

	// Do the ping
	pingbackByPhotoID(unsplashPhoto.UnsplashID)
}

func pingbackByPhotoID(photoID string) {
	if _, err := http.Get("https://views.unsplash.com/v?app_id=" + config.BackgroundsUnsplashApplicationID.GetString() + "&photo_id=" + photoID); err != nil {
		log.Errorf("Unsplash Pingback Failed: %s", err.Error())
	}
	log.Debugf("Pinged unsplash for photo %s", photoID)

}
