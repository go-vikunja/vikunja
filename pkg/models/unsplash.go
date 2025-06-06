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

package models

import (
	"code.vikunja.io/api/pkg/files"
	"xorm.io/xorm"
)

// Unsplash requires us to do pingbacks to their site and also name the image author.
// To do this properly, we need to save these information somewhere.

// UnsplashPhoto is an unsplash photo in the db
type UnsplashPhoto struct {
	ID         int64  `xorm:"autoincr not null unique pk" json:"id,omitempty"`
	FileID     int64  `xorm:"not null" json:"-"`
	UnsplashID string `xorm:"varchar(50)" json:"unsplash_id"`
	Author     string `xorm:"text" json:"author"`
	AuthorName string `xorm:"text" json:"author_name"`
}

// TableName contains the table name for an unsplash photo
func (u *UnsplashPhoto) TableName() string {
	return "unsplash_photos"
}

// Save persists an unsplash photo to the db
func (u *UnsplashPhoto) Save(s *xorm.Session) error {
	_, err := s.Insert(u)
	return err
}

// GetUnsplashPhotoByFileID returns an unsplash photo by its saved file id
func GetUnsplashPhotoByFileID(s *xorm.Session, fileID int64) (u *UnsplashPhoto, err error) {
	u = &UnsplashPhoto{}
	exists, err := s.Where("file_id = ?", fileID).Get(u)
	if err != nil {
		return
	}
	if !exists {
		return nil, files.ErrFileIsNotUnsplashFile{FileID: fileID}
	}
	return
}

// RemoveUnsplashPhoto removes an unsplash photo from the db
func RemoveUnsplashPhoto(s *xorm.Session, fileID int64) (err error) {
	// This is intentionally "fire and forget" which is why we don't check if we have an
	// unsplash entry for that file at all. If there is one, it will be deleted.
	// We do this to keep the function simple.
	_, err = s.Where("file_id = ?", fileID).Delete(&UnsplashPhoto{})
	return
}
