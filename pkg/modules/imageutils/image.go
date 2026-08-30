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

package imageutils

import (
	"image"
	"io"
)

// MaxPixels bounds allocations from attacker-controlled dimensions.
const MaxPixels = 50_000_000 // 50 megapixels

// Division avoids overflow from multiplying hostile dimensions.
func ValidateConfig(cfg image.Config) error {
	if cfg.Width <= 0 || cfg.Height <= 0 {
		return &ErrImageTooLarge{Width: cfg.Width, Height: cfg.Height}
	}
	if cfg.Width > MaxPixels/cfg.Height {
		return &ErrImageTooLarge{Width: cfg.Width, Height: cfg.Height}
	}
	return nil
}

func ValidateReader(r io.Reader) (image.Config, error) {
	cfg, _, err := image.DecodeConfig(r)
	if err != nil {
		return image.Config{}, err
	}
	if err := ValidateConfig(cfg); err != nil {
		return image.Config{}, err
	}
	return cfg, nil
}
