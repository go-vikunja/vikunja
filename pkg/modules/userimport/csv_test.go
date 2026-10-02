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

package userimport

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const csvHeader = "id,displayName,userPrincipalName,mail,jobTitle,department,accountEnabled"

func parse(t *testing.T, content string) *ParseResult {
	t.Helper()
	res, err := ParseCSV(strings.NewReader(content))
	require.NoError(t, err)
	return res
}

func TestParseCSV(t *testing.T) {
	t.Run("full row", func(t *testing.T) {
		res := parse(t, csvHeader+"\n"+
			"aaa-1,Jane Doe,jane.doe@corp.com,Jane.Doe@corp.com,Engineer,IT,true\n")
		require.Len(t, res.People, 1)
		p := res.People[0]
		assert.Equal(t, "aaa-1", p.ID)
		assert.Equal(t, "Jane Doe", p.DisplayName)
		assert.Equal(t, "jane.doe@corp.com", p.UserPrincipalName)
		assert.Equal(t, "Jane.Doe@corp.com", p.Mail, "the original case is kept, only comparisons lowercase")
		assert.Equal(t, "Engineer", p.JobTitle)
		assert.Equal(t, "IT", p.Department)
		assert.True(t, p.Enabled)
	})
	t.Run("byte order mark and CRLF line endings", func(t *testing.T) {
		res := parse(t, "\xEF\xBB\xBF"+csvHeader+"\r\naaa-1,Jane,j@corp.com,j@corp.com,,,true\r\naaa-2,Joe,o@corp.com,o@corp.com,,,true\r\n")
		assert.Len(t, res.People, 2)
	})
	t.Run("quoted fields with commas and newlines", func(t *testing.T) {
		res := parse(t, csvHeader+"\n"+
			`aaa-1,"Doe, Jane",j@corp.com,j@corp.com,"Head of
Engineering",IT,true`+"\n")
		require.Len(t, res.People, 1)
		assert.Equal(t, "Doe, Jane", res.People[0].DisplayName)
		assert.Equal(t, "Head of\nEngineering", res.People[0].JobTitle)
	})
	t.Run("columns are found by name, case-insensitively, in any order, extra columns ignored", func(t *testing.T) {
		res := parse(t, "MAIL,Extra,DisplayName,ID\nj@corp.com,ignored,Jane,aaa-1\n")
		require.Len(t, res.People, 1)
		assert.Equal(t, "aaa-1", res.People[0].ID)
		assert.Equal(t, "Jane", res.People[0].DisplayName)
		assert.True(t, res.People[0].Enabled, "no accountEnabled column means enabled")
	})
	t.Run("missing required column", func(t *testing.T) {
		for _, header := range []string{"id,displayName", "id,mail", "mail,displayName"} {
			_, err := ParseCSV(strings.NewReader(header + "\n"))
			require.Error(t, err, header)
		}
	})
	t.Run("empty file", func(t *testing.T) {
		_, err := ParseCSV(strings.NewReader(""))
		require.Error(t, err)
	})
	t.Run("header only has no people", func(t *testing.T) {
		res := parse(t, csvHeader+"\n")
		assert.Empty(t, res.People)
	})
	t.Run("rows without a valid id or mail are skipped and counted", func(t *testing.T) {
		res := parse(t, csvHeader+"\n"+
			",No Id,,n@corp.com,,,true\n"+
			"aaa-2,No Mail,,,,,true\n"+
			"aaa-3,Bad Mail,,not-an-email,,,true\n"+
			"aaa-4,Spaces,,a b@corp.com,,,true\n"+
			"aaa-5,Display Form,,Jane <j@corp.com>,,,true\n"+
			"aaa-6,Good,,g@corp.com,,,true\n")
		require.Len(t, res.People, 1)
		assert.Equal(t, "aaa-6", res.People[0].ID)
		assert.Equal(t, 5, res.SkippedInvalid)
	})
	t.Run("overlong id and mail are invalid", func(t *testing.T) {
		longID := strings.Repeat("a", maxIDLength+1)
		longMail := strings.Repeat("a", maxMailLength) + "@corp.com"
		res := parse(t, csvHeader+"\n"+
			longID+",A,,a@corp.com,,,true\n"+
			"aaa-1,B,,"+longMail+",,,true\n")
		assert.Empty(t, res.People)
		assert.Equal(t, 2, res.SkippedInvalid)
	})
	t.Run("duplicate mail (case-insensitive) and duplicate id, first wins", func(t *testing.T) {
		res := parse(t, csvHeader+"\n"+
			"aaa-1,First,,j@corp.com,,,true\n"+
			"aaa-2,Second,,J@CORP.com,,,true\n"+
			"aaa-1,Third,,other@corp.com,,,true\n"+
			"aaa-4,Fourth,,fourth@corp.com,,,true\n")
		require.Len(t, res.People, 2)
		assert.Equal(t, "First", res.People[0].DisplayName)
		assert.Equal(t, "Fourth", res.People[1].DisplayName)
		assert.Equal(t, 2, res.SkippedDuplicate)
	})
	t.Run("accountEnabled: only an explicit false disables", func(t *testing.T) {
		res := parse(t, csvHeader+"\n"+
			"a1,A,,a@corp.com,,,true\n"+
			"a2,B,,b@corp.com,,,FALSE\n"+
			"a3,C,,c@corp.com,,,0\n"+
			"a4,D,,d@corp.com,,,\n"+
			"a5,E,,e@corp.com,,,garbage\n"+
			"a6,F,,f@corp.com,,,no\n")
		require.Len(t, res.People, 6)
		got := make([]bool, 0, 6)
		for _, p := range res.People {
			got = append(got, p.Enabled)
		}
		assert.Equal(t, []bool{true, false, false, true, true, false}, got)
	})
	t.Run("ragged rows are tolerated", func(t *testing.T) {
		res := parse(t, csvHeader+"\n"+
			"a1,Short,,s@corp.com\n"+
			"a2,Long,,l@corp.com,T,D,true,extra,columns\n")
		require.Len(t, res.People, 2)
		assert.Empty(t, res.People[0].JobTitle)
		assert.True(t, res.People[0].Enabled)
		assert.Equal(t, "T", res.People[1].JobTitle)
	})
	t.Run("non-ASCII names", func(t *testing.T) {
		res := parse(t, csvHeader+"\n"+"a1,陳大文 José Ærø,,c@corp.com,經理,資訊科技部,true\n")
		require.Len(t, res.People, 1)
		assert.Equal(t, "陳大文 José Ærø", res.People[0].DisplayName)
		assert.Equal(t, "經理", res.People[0].JobTitle)
	})
	t.Run("over-long text is truncated on a character boundary", func(t *testing.T) {
		long := strings.Repeat("é", maxFieldLength+20)
		res := parse(t, csvHeader+"\n"+"a1,"+long+",,a@corp.com,"+long+","+long+",true\n")
		require.Len(t, res.People, 1)
		assert.Len(t, []rune(res.People[0].DisplayName), maxFieldLength)
		assert.Len(t, []rune(res.People[0].JobTitle), maxFieldLength)
		assert.Len(t, []rune(res.People[0].Department), maxFieldLength)
	})
	t.Run("invalid UTF-8 and NUL bytes are removed", func(t *testing.T) {
		res := parse(t, csvHeader+"\n"+"a1,Ren\xe9e\x00,,a@corp.com,,,true\n")
		require.Len(t, res.People, 1)
		assert.Equal(t, "Renee", res.People[0].DisplayName)
	})
	t.Run("surrounding whitespace is trimmed", func(t *testing.T) {
		res := parse(t, csvHeader+"\n"+"  a1 , Jane , , j@corp.com , Eng , IT , true \n")
		require.Len(t, res.People, 1)
		assert.Equal(t, "a1", res.People[0].ID)
		assert.Equal(t, "j@corp.com", res.People[0].Mail)
		assert.Equal(t, "Jane", res.People[0].DisplayName)
	})
	t.Run("malformed quoting is an error, not a partial import", func(t *testing.T) {
		_, err := ParseCSV(strings.NewReader(csvHeader + "\n" + `a1,"Unclosed,,j@corp.com,,,true` + "\n"))
		require.Error(t, err)
	})
}
