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
	"bytes"
	"encoding/base64"
	"fmt"
	"strings"

	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/modules/avatar"
	"code.vikunja.io/api/pkg/modules/avatar/empty"
	"code.vikunja.io/api/pkg/modules/avatar/gravatar"
	"code.vikunja.io/api/pkg/modules/avatar/initials"
	"code.vikunja.io/api/pkg/modules/avatar/ldap"
	"code.vikunja.io/api/pkg/modules/avatar/marble"
	"code.vikunja.io/api/pkg/modules/avatar/openid"
	"code.vikunja.io/api/pkg/modules/avatar/upload"
	"code.vikunja.io/api/pkg/user"

	"golang.org/x/net/html"
	"xorm.io/xorm"
)

func FindMentionedUsersInText(s *xorm.Session, text string) (users map[int64]*user.User, err error) {
	usernames := extractMentionedUsernames(text)
	if len(usernames) == 0 {
		return
	}

	return user.GetUsersByUsername(s, usernames, true)
}

// extractMentionedUsernames parses HTML content and extracts usernames from mention spans.
// It looks for <mention-user data-id="username"> elements and returns the usernames.
func extractMentionedUsernames(htmlText string) []string {
	doc, err := html.Parse(strings.NewReader(htmlText))
	if err != nil {
		return nil
	}

	usernames := []string{}
	seen := make(map[string]bool) // Deduplicate usernames

	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "mention-user" {
			var dataID string

			// Extract data-id attribute
			for _, attr := range n.Attr {
				if attr.Key == "data-id" {
					dataID = attr.Val
				}
			}

			if dataID != "" {
				if !seen[dataID] {
					usernames = append(usernames, dataID)
					seen[dataID] = true
				}
			}
		}

		// Traverse child nodes
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			traverse(child)
		}
	}

	traverse(doc)
	return usernames
}

// getAvatarProvider returns the appropriate avatar provider for a user
func getAvatarProvider(u *user.User) avatar.Provider {
	switch u.AvatarProvider {
	case "gravatar":
		return &gravatar.Provider{}
	case "initials":
		return &initials.Provider{}
	case "upload":
		return &upload.Provider{}
	case "marble":
		return &marble.Provider{}
	case "ldap":
		return &ldap.Provider{}
	case "openid":
		return &openid.Provider{}
	default:
		return &empty.Provider{}
	}
}

// getUserAvatarAsBase64 fetches the user's avatar and returns it as a data URI string
// suitable for embedding in emails. For SVG avatars, it converts them to data URIs.
// Returns an empty string if the avatar cannot be fetched.
func getUserAvatarAsBase64(u *user.User, size int) string {
	log.Debugf("getUserAvatarAsBase64 called for user %s (ID: %d, AvatarProvider: %s)", u.Username, u.ID, u.AvatarProvider)
	provider := getAvatarProvider(u)
	log.Debugf("Using avatar provider: %T", provider)

	// Use the new InlineProfilePicture method
	inlineData, err := provider.InlineProfilePicture(u, int64(size))
	if err != nil {
		log.Debugf("Failed to get inline profile picture for user %s: %v", u.Username, err)
		return ""
	}

	log.Debugf("InlineProfilePicture returned data of length %d for user %s", len(inlineData), u.Username)
	if len(inlineData) > 50 {
		log.Debugf("First 50 chars of inline data: %s", inlineData[:50])
	} else {
		log.Debugf("Full inline data: %s", inlineData)
	}

	// If it's already a data URI (base64 encoded), return as-is
	if strings.HasPrefix(inlineData, "data:") {
		log.Debugf("Detected data URI for user %s, returning as-is", u.Username)
		return inlineData
	}

	// If it's SVG content, convert to data URI
	// SVG can start with either <svg or <?xml (for XML declaration)
	if strings.HasPrefix(inlineData, "<svg") || strings.HasPrefix(inlineData, "<?xml") {
		log.Debugf("Detected SVG content for user %s, converting to data URI", u.Username)
		svgDataURI := fmt.Sprintf("data:image/svg+xml;base64,%s",
			base64.StdEncoding.EncodeToString([]byte(inlineData)))
		log.Debugf("SVG data URI length: %d", len(svgDataURI))
		return svgDataURI
	}

	// Fallback: if it's neither data URI nor SVG, log and return empty
	maxLen := 50
	if len(inlineData) < maxLen {
		maxLen = len(inlineData)
	}
	log.Debugf("Unexpected inline profile picture format for user %s: %s", u.Username, inlineData[:maxLen])
	return ""
}

// formatMentionsForEmail replaces mention-user tags with user avatars and names for email display.
// It converts <mention-user data-id="username" data-label="Display Name"> tags to
// <strong><img src="data:..."/> Display Name</strong> with a 20x20 avatar image.
// If data-label is missing, it falls back to data-id. Returns the original HTML unchanged on any error.
func formatMentionsForEmail(s *xorm.Session, htmlText string) string {
	if htmlText == "" {
		return htmlText
	}

	doc, err := html.Parse(strings.NewReader(htmlText))
	if err != nil {
		log.Debugf("Failed to parse HTML for mention formatting: %v", err)
		return htmlText
	}

	// Extract all usernames first to batch fetch users
	usernames := extractMentionedUsernames(htmlText)
	var usersMap map[int64]*user.User
	var usernameToUser map[string]*user.User

	if len(usernames) > 0 && s != nil {
		usersMap, err = user.GetUsersByUsername(s, usernames, true)
		if err != nil {
			log.Debugf("Failed to fetch users for mention formatting: %v", err)
			// Continue without user data - we'll fall back to display names from attributes
		} else {
			// Create username -> user map for easy lookup
			usernameToUser = make(map[string]*user.User)
			for _, u := range usersMap {
				usernameToUser[u.Username] = u
			}
		}
	}

	// Track nodes to replace (can't modify while traversing)
	type replacement struct {
		oldNode *html.Node
		newNode *html.Node
	}
	replacements := []replacement{}

	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "mention-user" {
			var dataLabel, dataID string

			// Extract data-label and data-id attributes
			for _, attr := range n.Attr {
				switch attr.Key {
				case "data-label":
					dataLabel = attr.Val
				case "data-id":
					dataID = attr.Val
				}
			}

			// Determine what to display
			displayName := dataLabel
			if displayName == "" {
				displayName = dataID
			}

			// If still empty and has text content (old format), use that
			if displayName == "" && n.FirstChild != nil && n.FirstChild.Type == html.TextNode {
				displayName = strings.TrimPrefix(n.FirstChild.Data, "@")
			}

			if displayName == "" {
				log.Debugf("Mention node has no data-label, data-id, or text content, skipping")
				// Continue traversing children in case there are nested elements
				for child := n.FirstChild; child != nil; child = child.NextSibling {
					traverse(child)
				}
				return
			}

			// Create <strong> wrapper
			strongNode := &html.Node{
				Type: html.ElementNode,
				Data: "strong",
			}

			// Try to get avatar for the user
			var avatarDataURI string
			if usernameToUser != nil && dataID != "" {
				if u, ok := usernameToUser[dataID]; ok {
					log.Debugf("Getting avatar for user %s (ID: %d, AvatarProvider: %s)", u.Username, u.ID, u.AvatarProvider)
					avatarDataURI = getUserAvatarAsBase64(u, 20)
					if avatarDataURI == "" {
						log.Debugf("getUserAvatarAsBase64 returned empty string for user %s", u.Username)
					} else {
						log.Debugf("getUserAvatarAsBase64 returned data URI of length %d for user %s", len(avatarDataURI), u.Username)
					}
				} else {
					log.Debugf("User with dataID '%s' not found in usernameToUser map", dataID)
				}
			} else {
				if usernameToUser == nil {
					log.Debugf("usernameToUser map is nil")
				}
				if dataID == "" {
					log.Debugf("dataID is empty")
				}
			}

			// If we have an avatar, add the img element
			if avatarDataURI != "" {
				imgNode := &html.Node{
					Type: html.ElementNode,
					Data: "img",
					Attr: []html.Attribute{
						{Key: "src", Val: avatarDataURI},
						{Key: "width", Val: "20"},
						{Key: "height", Val: "20"},
						{Key: "style", Val: "border-radius: 50%; vertical-align: middle; margin-right: 4px;"},
						{Key: "alt", Val: displayName},
					},
				}
				strongNode.AppendChild(imgNode)

				// Add display name without @ since we have the avatar
				textNode := &html.Node{
					Type: html.TextNode,
					Data: displayName,
				}
				strongNode.AppendChild(textNode)
			} else {
				// Fall back to @DisplayName without avatar
				textNode := &html.Node{
					Type: html.TextNode,
					Data: "@" + displayName,
				}
				strongNode.AppendChild(textNode)
			}

			// Schedule replacement
			replacements = append(replacements, replacement{
				oldNode: n,
				newNode: strongNode,
			})

			// Don't traverse children of mention-user since we're replacing it
			return
		}

		// Traverse child nodes
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			traverse(child)
		}
	}

	traverse(doc)

	// Apply replacements
	for _, r := range replacements {
		if r.oldNode.Parent != nil {
			r.oldNode.Parent.InsertBefore(r.newNode, r.oldNode)
			r.oldNode.Parent.RemoveChild(r.oldNode)
		}
	}

	// Render back to HTML
	var buf bytes.Buffer
	err = html.Render(&buf, doc)
	if err != nil {
		log.Debugf("Failed to render HTML after mention formatting: %v", err)
		return htmlText
	}

	// The html.Parse wraps content in <html><head></head><body>...</body></html>
	// We need to extract just the body content
	result := buf.String()

	// Remove the wrapper tags
	// html.Parse adds: <html><head></head><body>CONTENT</body></html>
	result = strings.TrimPrefix(result, "<html><head></head><body>")
	result = strings.TrimSuffix(result, "</body></html>")

	return result
}
