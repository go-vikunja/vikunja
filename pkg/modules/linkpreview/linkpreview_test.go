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

package linkpreview

import (
	"net/url"
	"strings"
	"testing"

	"github.com/otiai10/opengraph/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func parse(t *testing.T, doc, base string) *Preview {
	t.Helper()
	u, err := url.Parse(base)
	require.NoError(t, err)
	og := opengraph.New(base)
	require.NoError(t, og.Parse(strings.NewReader(doc)))
	preview := &Preview{URL: base}
	fillPreview(preview, og, u)
	return preview
}

func TestFillPreviewOpenGraph(t *testing.T) {
	const doc = `<html><head>
		<title>Fallback Title</title>
		<meta property="og:title" content="OG Title">
		<meta property="og:description" content="OG description here">
		<meta property="og:site_name" content="Example">
		<meta property="og:image" content="/img/card.png">
		<link rel="icon" href="https://example.com/favicon.ico">
	</head><body>hi</body></html>`

	p := parse(t, doc, "https://example.com/page")
	assert.Equal(t, "OG Title", p.Title)
	assert.Equal(t, "OG description here", p.Description)
	assert.Equal(t, "Example", p.SiteName)
	assert.Equal(t, "https://example.com/img/card.png", p.Image, "relative og:image should resolve against the base")
	assert.Equal(t, "https://example.com/favicon.ico", p.Favicon)
}

func TestResolveURLDropsDangerousSchemes(t *testing.T) {
	base, _ := url.Parse("https://example.com/")
	assert.Empty(t, resolveURL(base, "javascript:alert(1)"))
	assert.Empty(t, resolveURL(base, "data:image/png;base64,AAAA"))
	assert.Equal(t, "https://example.com/a.png", resolveURL(base, "/a.png"))
	assert.Equal(t, "https://cdn.example.com/x.png", resolveURL(base, "https://cdn.example.com/x.png"))
}

func TestTruncate(t *testing.T) {
	assert.Equal(t, "abc", truncate("  abc  ", 300))
	assert.Equal(t, strings.Repeat("x", 300)+"…", truncate(strings.Repeat("x", 400), 300))
}
