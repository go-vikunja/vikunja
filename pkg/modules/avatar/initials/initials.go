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
	"#e0f8d9",
	"#e3f5f8",
	"#faeefb",
	"#f1efff",
	"#ffecf0",
	"#ffefe4",
}

var avatarTextColors = []string{
	"#005f00",
	"#00548c",
	"#822198",
	"#5d26cd",
	"#9f0850",
	"#9b2200",
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

	// Select background and text colors based on user ID
	colorIndex := int(u.ID) % len(avatarBgColors)
	bgColor := avatarBgColors[colorIndex]
	textColor := avatarTextColors[colorIndex]

	// Convert size to string
	sizeStr := strconv.FormatInt(size, 10)

	// Generate SVG
	svg := fmt.Sprintf(`<svg viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg" width="%s" height="%s">
  <rect width="100" height="100" fill="%s"/>
  <text x="50" y="50" font-family="sans-serif" font-size="50" fill="%s" text-anchor="middle" dominant-baseline="central">%s</text>
</svg>`, sizeStr, sizeStr, bgColor, textColor, initial)

	return []byte(svg), "image/svg+xml", nil
}

// InlineProfilePicture returns the SVG string directly since SVG can be used inline
func (p *Provider) InlineProfilePicture(u *user.User, size int64) (string, error) {
	avatarData, _, err := p.GetAvatar(u, size)
	if err != nil {
		return "", err
	}

	// For SVG, we return the plain SVG string
	return string(avatarData), nil
}
