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

package initials

import (
	"fmt"
	"html"
	"strconv"
	"strings"

	"code.vikunja.io/api/pkg/user"
)

// Provider represents the provider implementation of the initials provider
type Provider struct {
}

// FlushCache is a no-op for the initials provider since SVG generation is lightweight
func (p *Provider) FlushCache(_ *user.User) error { return nil }

var avatarBgColors = []string{
	"#45bdf3", // rgb(69, 189, 243)
	"#e08f70", // rgb(224, 143, 112)
	"#4db6ac", // rgb(77, 182, 172)
	"#9575cd", // rgb(149, 117, 205)
	"#b0855e", // rgb(176, 133, 94)
	"#f06292", // rgb(240, 98, 146)
	"#a3d36c", // rgb(163, 211, 108)
	"#7986cb", // rgb(121, 134, 203)
	"#f1b91d", // rgb(241, 185, 29)
}

// GetAvatar returns an initials avatar for a user as SVG
func (p *Provider) GetAvatar(u *user.User, size int64) (avatar []byte, mimeType string, err error) {
	// Get the text to display
	avatarText := u.Name
	if avatarText == "" {
		avatarText = u.Username
	}
	if avatarText == "" {
		return nil, "", fmt.Errorf("user has no name or username")
	}

	// Get the first character and convert to uppercase
	firstRune := []rune(strings.ToUpper(avatarText))[0]
	initial := html.EscapeString(string(firstRune))

	// Select background color based on user ID
	bgColor := avatarBgColors[int(u.ID)%len(avatarBgColors)]

	// Convert size to string
	sizeStr := strconv.FormatInt(size, 10)

	// Generate SVG
	svg := fmt.Sprintf(`<svg viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg" width="%s" height="%s">
  <rect width="100" height="100" fill="%s"/>
  <text x="50" y="50" font-family="sans-serif" font-size="50" fill="white" text-anchor="middle" dominant-baseline="central">%s</text>
</svg>`, sizeStr, sizeStr, bgColor, initial)

	return []byte(svg), "image/svg+xml", nil
}
