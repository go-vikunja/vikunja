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

package user

import (
	"regexp"
	"strings"

	"code.vikunja.io/api/pkg/i18n"

	"github.com/asaskevich/govalidator"
)

func init() {
	govalidator.TagMap["username"] = func(i string) bool {
		// To avoid making this overly complicated, we only check a few things:
		// 1. No Spaces
		// 2. Should not look like an url
		// 3. Should not contain , (because then it will be impossible to search for)
		// 4. Should not start with link-share-[NUMBER] (reserved for link sharing system)
		if govalidator.HasWhitespace(i) {
			return false
		}

		if govalidator.IsURL(i) {
			return false
		}

		if strings.Contains(i, ",") {
			return false
		}

		// Check if username matches the reserved link-share pattern
		linkSharePattern := regexp.MustCompile(`^link-share-\d+$`)
		return !linkSharePattern.MatchString(i)
	}

	govalidator.TagMap["bcrypt_password"] = func(str string) bool {
		if len(str) < 8 {
			return false
		}

		return len([]byte(str)) < 72
	}

	govalidator.TagMap["language"] = i18n.HasLanguage
}
