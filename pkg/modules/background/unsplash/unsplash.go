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
	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/background"
	"code.vikunja.io/web"
	"encoding/json"
	"net/http"
	"strconv"
)

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
		Self     string `json:"self"`
		HTML     string `json:"html"`
		Download string `json:"download"`
	} `json:"links"`
}

// Very simple caching method - pretty much only used to retain information when saving an image
// FIXME: Should use a proper cache
var photos map[string]*Photo

func init() {
	photos = make(map[string]*Photo)
}

func doGet(url string, result interface{}) (err error) {
	req, err := http.NewRequest("GET", "https://api.unsplash.com/"+url, nil)
	if err != nil {
		return
	}

	req.Header.Add("Authorization", "Client-ID "+config.BackgroundsUnsplashAccessToken.GetString())
	hc := http.Client{}
	resp, err := hc.Do(req)
	if err != nil {
		return
	}

	err = json.NewDecoder(resp.Body).Decode(result)
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
		collectionResult := []*Photo{}
		err = doGet("collections/317099/photos?page="+strconv.FormatInt(page, 10)+"&per_page=25&order_by=latest", &collectionResult)
		if err != nil {
			return
		}

		result = []*background.Image{}
		for _, p := range collectionResult {
			result = append(result, &background.Image{
				ID:    p.ID,
				URL:   p.Urls.Raw,
				Thumb: p.Urls.Thumb,
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

	searchResult := &SearchResult{}
	err = doGet("search/photos?query="+search+"&page="+strconv.FormatInt(page, 10)+"&per_page=25", &searchResult)
	if err != nil {
		return
	}

	result = []*background.Image{}
	for _, p := range searchResult.Results {
		result = append(result, &background.Image{
			ID:    p.ID,
			URL:   p.Urls.Raw,
			Thumb: p.Urls.Thumb,
		})
		photos[p.ID] = p
	}

	return
}

// Set sets an unsplash photo as list background
// @Summary Set an unsplash photo as list background
// @Description Sets a photo from unsplash as list background.1
// @tags list
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "List ID"
// @Param list body background.Image true "The image you want to set as background"
// @Success 200 {object} models.List "The background has been successfully set."
// @Failure 400 {object} code.vikunja.io/web.HTTPError "Invalid image object provided."
// @Failure 403 {object} code.vikunja.io/web.HTTPError "The user does not have access to the list"
// @Failure 500 {object} models.Message "Internal error"
// @Router /lists/{id}/backgrounds/unsplash [post]
func (p *Provider) Set(image *background.Image, list *models.List, auth web.Auth) (err error) {

	// Find the photo
	var photo *Photo
	var exists bool
	photo, exists = photos[image.ID]
	if !exists {
		log.Debugf("Image information for unsplash photo %s not cached, requesting from unsplash...", image.ID)
		photo = &Photo{}
		err = doGet("photos/"+image.ID, photo)
		if err != nil {
			return
		}
	}

	// Download the photo from unsplash
	// The parameters crop the image to a max width of 2560 and a max height of 2048 to save bandwidth and storage.
	resp, err := http.Get(photo.Urls.Raw + "&w=2560&h=2048&q=90")
	if err != nil {
		return
	}
	defer resp.Body.Close()

	log.Debugf("Downloaded Unsplash Photo %s", image.ID)

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
		log.Errorf("Unsplash Pingback: %s", err.Error())
	}

	// Do the ping
	if _, err := http.Get("https://views.unsplash.com/v?app_id=" + config.BackgroundsUnsplashApplicationID.GetString() + "&photo_id=" + unsplashPhoto.UnsplashID); err != nil {
		log.Errorf("Unsplash Pingback Failed: %s", err.Error())
	}
	log.Debugf("Pinged unsplash for photo %s", unsplashPhoto.UnsplashID)
}
