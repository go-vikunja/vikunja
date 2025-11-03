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

package marble

import (
	"math"
	"strconv"

	"code.vikunja.io/api/pkg/user"
)

// Provider generates a random avatar based on https://github.com/boringdesigners/boring-avatars
type Provider struct {
}

// FlushCache is a no-op for the marble provider
func (p *Provider) FlushCache(_ *user.User) error { return nil }

const avatarSize = 80

var colors = []string{
	"#A3A948",
	"#EDB92E",
	"#F85931",
	"#CE1836",
	"#009989",
}

type props struct {
	Color      string
	TranslateX int
	TranslateY int
	Rotate     int
	Scale      float64
}

func getUnit(number int, rang, index int) int {
	value := number % rang

	digit := math.Floor(math.Mod(float64(number)/math.Pow(10, float64(index)), 10))

	if index > 0 && (math.Mod(digit, 2) == 0) {
		return -value
	}

	return value
}

func getPropsForUser(u *user.User) []*props {
	ps := []*props{}
	for i := 0; i < 3; i++ {
		f := float64(getUnit(int(u.ID)*(i+1), avatarSize/10, 0))
		ps = append(ps, &props{
			Color:      colors[(int(u.ID)+i)%(len(colors)-1)],
			TranslateX: getUnit(int(u.ID)*(i+1), avatarSize/10, 1),
			TranslateY: getUnit(int(u.ID)*(i+1), avatarSize/10, 2),
			Scale:      1.2 + f/10,
			Rotate:     getUnit(int(u.ID)*(i+1), 360, 1),
		})
	}

	return ps
}

func (p *Provider) GetAvatar(u *user.User, size int64) (avatar []byte, mimeType string, err error) {

	s := strconv.FormatInt(size, 10)
	avatarSizeStr := strconv.Itoa(avatarSize)
	avatarSizeHalf := strconv.Itoa(avatarSize / 2)

	ps := getPropsForUser(u)

	return []byte(`<svg
      viewBox="0 0 ` + avatarSizeStr + ` ` + avatarSizeStr + `"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      width="` + s + `"
      height="` + s + `"
    >
      <mask id="mask__marble" maskUnits="userSpaceOnUse" x="0" y="0" width="` + avatarSizeStr + `" height="` + avatarSizeStr + `">
        <rect width="` + avatarSizeStr + `" height="` + avatarSizeStr + `" rx="` + strconv.Itoa(avatarSize*2) + `" fill="white" />
      </mask>
      <g mask="url(#mask__marble)">
        <rect width="` + avatarSizeStr + `" height="` + avatarSizeStr + `" rx="2" fill="` + ps[0].Color + `" />
        <path
          filter="url(#prefix__filter0_f)"
          d="M32.414 59.35L50.376 70.5H72.5v-71H33.728L26.5 13.381l19.057 27.08L32.414 59.35z"
          fill="` + ps[1].Color + `"
          transform="translate(` + strconv.Itoa(ps[1].TranslateX) + ` ` + strconv.Itoa(ps[1].TranslateY) + `) rotate(` + strconv.Itoa(ps[1].Rotate) + ` ` + avatarSizeHalf + ` ` + avatarSizeHalf + `) scale(` + strconv.FormatFloat(ps[2].Scale, 'f', 2, 64) + `)"
        />
        <path
          filter="url(#prefix__filter0_f)"
          style="mix-blend-mode: overlay;"
          d="M22.216 24L0 46.75l14.108 38.129L78 86l-3.081-59.276-22.378 4.005 12.972 20.186-23.35 27.395L22.215 24z"
          fill="` + ps[2].Color + `"
          transform="translate(` + strconv.Itoa(ps[2].TranslateX) + ` ` + strconv.Itoa(ps[2].TranslateY) + `) rotate(` + strconv.Itoa(ps[2].Rotate) + ` ` + avatarSizeHalf + ` ` + avatarSizeHalf + `) scale(` + strconv.FormatFloat(ps[2].Scale, 'f', 2, 64) + `)"
        />
      </g>
      <defs>
        <filter
          id="prefix__filter0_f"
          filterUnits="userSpaceOnUse"
          colorInterpolationFilters="sRGB"
        >
          <feFlood flood-opacity="0" result="BackgroundImageFix" />
          <feBlend in="SourceGraphic" in2="BackgroundImageFix" result="shape" />
          <feGaussianBlur stdDeviation="7" result="effect1_foregroundBlur" />
        </filter>
      </defs>
    </svg>`), "image/svg+xml", nil
}
