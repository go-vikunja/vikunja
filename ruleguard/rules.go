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

//go:build ruleguard

package gorules

import "github.com/quasilyte/go-ruleguard/dsl"

func rawHeadersInLog(m dsl.Matcher) {
	m.Import("code.vikunja.io/api/pkg/log")
	// $*_ on both sides of $h doesn't backtrack when Where fails, so spell out positions.
	m.Match(`log.$f($h, $*_)`, `log.$f($_, $h, $*_)`, `log.$f($_, $_, $h, $*_)`, `log.$f($_, $_, $_, $h, $*_)`).
		Where(m["h"].Type.Is("http.Header") && !m["f"].Text.Matches(`^SafeHeaders$`)).
		Report("http.Header carries credentials, wrap it in log.SafeHeaders")
}

func envInLog(m dsl.Matcher) {
	m.Import("code.vikunja.io/api/pkg/log")
	m.Match(`log.$f($*_, os.Environ(), $*_)`, `log.$f($*_, os.Getenv($_), $*_)`).
		Report("environment variables may hold secrets, don't log them")
}
