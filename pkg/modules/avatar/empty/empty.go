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

package empty

import "code.vikunja.io/api/pkg/user"

// Provider represents the empty avatar provider
type Provider struct {
}

// FlushCache is a no-op for the empty provider
func (p *Provider) FlushCache(_ *user.User) error { return nil }

const defaultAvatar string = `<?xml version="1.0" encoding="UTF-8"?>
<svg width="128" height="128" version="1.1" viewBox="0 0 33.867 33.867" xmlns="http://www.w3.org/2000/svg" xmlns:cc="http://creativecommons.org/ns#" xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#" xmlns:xlink="http://www.w3.org/1999/xlink">
<metadata>
<rdf:RDF>
<cc:Work rdf:about="">
<dc:format>image/svg+xml</dc:format>
<dc:type rdf:resource="http://purl.org/dc/dcmitype/StillImage"/>
<dc:title/>
</cc:Work>
</rdf:RDF>
</metadata>
<g transform="matrix(.1916 0 0 .1916 -163.17 -2538.8)" fill="#ccc" shape-rendering="auto">
<circle cx="940.01" cy="13319" r="49.346" color="#000000" color-rendering="auto" image-rendering="auto" solid-color="#000000" style="isolation:auto;mix-blend-mode:normal"/>
<path d="m940.01 13375a88.385 50.907 0 0 0-88.385 50.908 88.385 50.907 0 0 0 0.0948 1.542l176.59-0.596a88.385 50.907 0 0 0 0.071-0.812v-0.303a88.385 50.907 0 0 0-88.375-50.739z" color="#000000" color-rendering="auto" image-rendering="auto" solid-color="#000000" style="isolation:auto;mix-blend-mode:normal"/>
</g>
</svg>`

// GetAvatar implements getting the avatar method
func (p *Provider) GetAvatar(_ *user.User, _ int64) (avatar []byte, mimeType string, err error) {
	return []byte(defaultAvatar), "image/svg+xml", nil
}
