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

package migration

import (
	"errors"
	"image"

	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/log"
	"github.com/bbrks/go-blurhash"
	"golang.org/x/image/draw"
	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

type lists20211212210054 struct {
	ID                 int64  `xorm:"bigint autoincr not null unique pk" json:"id" param:"list"`
	BackgroundFileID   int64  `xorm:"null" json:"-"`
	BackgroundBlurHash string `xorm:"varchar(50) null" json:"background_blur_hash"`
}

func (lists20211212210054) TableName() string {
	return "lists"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20211212210054",
		Description: "Add blurHash to list backgrounds.",
		Migrate: func(tx *xorm.Engine) error {
			err := tx.Sync2(lists20211212210054{})
			if err != nil {
				return err
			}

			lists := []*lists20211212210054{}
			err = tx.Where("background_file_id is not null AND background_file_id != ?", 0).Find(&lists)
			if err != nil {
				return err
			}

			log.Infof("Creating BlurHash for %d list backgrounds, this might take a while...", len(lists))

			for _, l := range lists {
				bgFile := &files.File{
					ID: l.BackgroundFileID,
				}
				if err := bgFile.LoadFileByID(); err != nil {
					return err
				}

				src, _, err := image.Decode(bgFile.File)
				if err != nil && !errors.Is(err, image.ErrFormat) {
					return err
				}
				if err != nil && errors.Is(err, image.ErrFormat) {
					log.Warningf("Could not generate a blur hash of list %d's background image: %s", l.ID, err)
				}

				dst := image.NewRGBA(image.Rect(0, 0, 32, 32))
				draw.NearestNeighbor.Scale(dst, dst.Rect, src, src.Bounds(), draw.Over, nil)

				hash, err := blurhash.Encode(4, 3, dst)
				if err != nil {
					return err
				}

				l.BackgroundBlurHash = hash
				_, err = tx.Where("id = ?", l.ID).
					Cols("background_blur_hash").
					Update(l)
				if err != nil {
					return err
				}
				log.Debugf("Created BlurHash for list %d", l.ID)
			}

			return nil
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
