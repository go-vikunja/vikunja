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

package richtext

import (
	"fmt"
	"regexp"
	"unicode"
	"unicode/utf8"

	"code.vikunja.io/api/pkg/user"

	"golang.org/x/net/html"
	"xorm.io/xorm"
)

// mentionTokenRegex matches "@username". The username starts/ends with a word
// char so trailing prose punctuation ("@jane.") isn't swallowed. RE2 has no
// look-behind, so the boundary before "@" is checked in code (to reject "a@b").
var mentionTokenRegex = regexp.MustCompile(`@([\p{L}\p{N}_](?:[\p{L}\p{N}._-]*[\p{L}\p{N}_])?)`)

// rebuildMentions replaces "@username" tokens with <mention-user> tags, resolving
// against real users in one batched query. Unknown handles and tokens inside
// code/links are left untouched.
func rebuildMentions(s *xorm.Session, nodes []*html.Node) error {
	var textNodes []*html.Node
	for _, n := range nodes {
		collectMentionTextNodes(n, false, &textNodes)
	}
	if len(textNodes) == 0 {
		return nil
	}

	candidates := map[string]struct{}{}
	for _, tn := range textNodes {
		for _, name := range findMentionCandidates(tn.Data) {
			candidates[name] = struct{}{}
		}
	}
	if len(candidates) == 0 {
		return nil
	}

	usernames := make([]string, 0, len(candidates))
	for name := range candidates {
		usernames = append(usernames, name)
	}

	usersByID, err := user.GetUsersByUsername(s, usernames, false)
	if err != nil {
		return fmt.Errorf("looking up mentioned users: %w", err)
	}

	usersByName := make(map[string]*user.User, len(usersByID))
	for _, u := range usersByID {
		usersByName[u.Username] = u
	}
	if len(usersByName) == 0 {
		return nil
	}

	for _, tn := range textNodes {
		replaceMentionsInTextNode(tn, usersByName)
	}
	return nil
}

// collectMentionTextNodes gathers text nodes outside <code>, <pre>, <a> and
// <mention-user>.
func collectMentionTextNodes(n *html.Node, inSkip bool, out *[]*html.Node) {
	if n.Type == html.TextNode {
		if !inSkip {
			*out = append(*out, n)
		}
		return
	}

	skip := inSkip
	if n.Type == html.ElementNode {
		switch n.Data {
		case "code", "pre", "a", "mention-user":
			skip = true
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		collectMentionTextNodes(c, skip, out)
	}
}

// findMentionCandidates returns the usernames mentioned in text (word-boundary
// "@" only).
func findMentionCandidates(text string) []string {
	var names []string
	for _, m := range mentionTokenRegex.FindAllStringSubmatchIndex(text, -1) {
		if mentionPrecededByWordChar(text, m[0]) {
			continue
		}
		names = append(names, text[m[2]:m[3]])
	}
	return names
}

// replaceMentionsInTextNode splits tn, swapping known @mentions for <mention-user> nodes.
func replaceMentionsInTextNode(tn *html.Node, users map[string]*user.User) {
	text := tn.Data

	var newNodes []*html.Node
	cursor := 0
	for _, m := range mentionTokenRegex.FindAllStringSubmatchIndex(text, -1) {
		start, end := m[0], m[1]
		if mentionPrecededByWordChar(text, start) {
			continue
		}
		u, ok := users[text[m[2]:m[3]]]
		if !ok {
			continue
		}

		if start > cursor {
			newNodes = append(newNodes, &html.Node{Type: html.TextNode, Data: text[cursor:start]})
		}
		newNodes = append(newNodes, newMentionNode(u))
		cursor = end
	}

	if len(newNodes) == 0 {
		return
	}
	if cursor < len(text) {
		newNodes = append(newNodes, &html.Node{Type: html.TextNode, Data: text[cursor:]})
	}

	parent := tn.Parent
	for _, nn := range newNodes {
		parent.InsertBefore(nn, tn)
	}
	parent.RemoveChild(tn)
}

// newMentionNode builds <mention-user data-id="username" data-label="Name">@Name</mention-user>.
// data-id carries the username so extractMentionedUsernames can re-resolve it.
func newMentionNode(u *user.User) *html.Node {
	n := &html.Node{
		Type: html.ElementNode,
		Data: "mention-user",
		Attr: []html.Attribute{
			{Key: "data-id", Val: u.Username},
			{Key: "data-label", Val: u.GetName()},
		},
	}
	n.AppendChild(&html.Node{Type: html.TextNode, Data: "@" + u.GetName()})
	return n
}

// mentionPrecededByWordChar reports whether the rune just before atIndex is a
// letter, digit or underscore — i.e. the "@" is mid-token (an email), not a mention.
func mentionPrecededByWordChar(text string, atIndex int) bool {
	if atIndex == 0 {
		return false
	}
	r, _ := utf8.DecodeLastRuneInString(text[:atIndex])
	return unicode.IsLetter(r) || unicode.IsNumber(r) || r == '_'
}
