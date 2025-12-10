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
	"strings"

	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/modules/avatar"
	"code.vikunja.io/api/pkg/user"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
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

// formatMentionsForEmail replaces mention-user tags with user avatars and names for email display.
// It converts <mention-user data-id="username" data-label="Display Name"> tags to
// <strong><img src="data:..."/> Display Name</strong> with a 20x20 avatar image.
// If data-label is missing, it falls back to data-id. Returns the original HTML unchanged on any error.
func formatMentionsForEmail(s *xorm.Session, htmlText string) string {
	if htmlText == "" {
		return htmlText
	}

	// Create a synthetic body node for fragment parsing
	bodyNode := &html.Node{
		Type:     html.ElementNode,
		Data:     "body",
		DataAtom: atom.Body,
	}

	fragments, err := html.ParseFragment(strings.NewReader(htmlText), bodyNode)
	if err != nil {
		log.Debugf("Failed to parse HTML fragment for mention formatting: %v", err)
		return htmlText
	}



	// If no fragments, return original
	if len(fragments) == 0 {
		return htmlText
	}

	// Extract all usernames first to batch fetch users
	usernames := extractMentionedUsernames(htmlText)
	var usersMap map[int64]*user.User
	var usernameToUser map[string]*user.User

	if len(usernames) == 0 {
		return htmlText
	}

	// Create maps for user data and avatar data URIs
	usernameToAvatarURI := make(map[string]string)

	// Only fetch users if we have a valid session
	if s != nil {
		usersMap, err = user.GetUsersByUsername(s, usernames, true)
		if err != nil {
			log.Debugf("Failed to fetch users for mention formatting: %v", err)
			// Continue without user data - we'll fall back to display names from attributes
		} else {
			// Create username -> user map for easy lookup and fetch avatar data URIs
			usernameToUser = make(map[string]*user.User)
			for _, u := range usersMap {
				usernameToUser[u.Username] = u

				// Fetch avatar data URI for this user
				provider := avatar.GetProvider(u)
				avatarDataURI, err := provider.AsDataURI(u, 20)
				if err == nil && avatarDataURI != "" {
					usernameToAvatarURI[u.Username] = avatarDataURI
				}
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

			// Get pre-fetched avatar data URI for the user
			var avatarDataURI string
			if dataID != "" {
				avatarDataURI = usernameToAvatarURI[dataID]
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

	// Traverse all fragment nodes
	for _, fragment := range fragments {
		traverse(fragment)
	}

	// Apply replacements
	for _, r := range replacements {
		if r.oldNode.Parent != nil {
			r.oldNode.Parent.InsertBefore(r.newNode, r.oldNode)
			r.oldNode.Parent.RemoveChild(r.oldNode)
		}
	}

	// Render each fragment node back to HTML
	var buf bytes.Buffer
	for _, fragment := range fragments {
		err = html.Render(&buf, fragment)
		if err != nil {
			log.Debugf("Failed to render HTML fragment after mention formatting: %v", err)
			return htmlText
		}
	}

	return buf.String()
}
