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

package bootstrap

import (
	"bytes"
	"strings"
	"testing"

	"code.vikunja.io/veans/internal/output"
)

func TestSuggestPetname_ShapeAndPrefix(t *testing.T) {
	got := suggestPetname()
	if !strings.HasPrefix(got, "bot-") {
		t.Fatalf("petname missing bot- prefix: %q", got)
	}
	// Two adjective-animal words separated by hyphens means at least
	// three hyphen-delimited segments (bot, word1, word2).
	if parts := strings.Split(got, "-"); len(parts) < 3 {
		t.Fatalf("petname looks malformed: %q", got)
	}
}

func TestIsUsernameTakenErr(t *testing.T) {
	cases := []struct {
		err  *output.Error
		want bool
	}{
		{nil, false},
		{output.New(output.CodeNotFound, "Not found"), false},
		{output.New(output.CodeValidation, "Some other validation issue"), false},
		{output.New(output.CodeValidation, "A user with this username already exists."), true},
		{output.New(output.CodeValidation, "username already exists"), true},
		{output.New(output.CodeValidation, "USERNAME ALREADY EXISTS"), true}, // case-insensitive
	}
	for _, c := range cases {
		if got := isUsernameTakenErr(c.err); got != c.want {
			msg := "<nil>"
			if c.err != nil {
				msg = c.err.Message
			}
			t.Errorf("isUsernameTakenErr(%q) = %v, want %v", msg, got, c.want)
		}
	}
}

// scriptedPrompter returns canned answers to ReadLine, in order.
type scriptedPrompter struct {
	answers []string
	pos     int
}

func (s *scriptedPrompter) ReadLine(_ string) (string, error) {
	if s.pos >= len(s.answers) {
		return "", nil
	}
	a := s.answers[s.pos]
	s.pos++
	return a, nil
}
func (s *scriptedPrompter) ReadPassword(_ string) (string, error) { return "", nil }

func TestConfirmReuse(t *testing.T) {
	yes := []string{"", "y", "Y", "yes", "Yes", "YES", "  yes  "}
	no := []string{"n", "no", "N", "nope", "anything else"}
	var buf bytes.Buffer

	for _, ans := range yes {
		p := &scriptedPrompter{answers: []string{ans}}
		ok, err := confirmReuse(p, &buf, "bot-x")
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			t.Errorf("answer %q should be treated as yes", ans)
		}
	}
	for _, ans := range no {
		p := &scriptedPrompter{answers: []string{ans}}
		ok, err := confirmReuse(p, &buf, "bot-x")
		if err != nil {
			t.Fatal(err)
		}
		if ok {
			t.Errorf("answer %q should be treated as no", ans)
		}
	}
}

func TestPromptForReplacementName_AcceptsDefault(t *testing.T) {
	var buf bytes.Buffer
	p := &scriptedPrompter{answers: []string{""}} // accept default
	name, err := promptForReplacementName(p, &buf, "bot-old", true)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(name, "bot-") {
		t.Errorf("default name missing bot- prefix: %q", name)
	}
	if name == "bot-old" {
		t.Errorf("default name shouldn't equal previous")
	}
}

func TestPromptForReplacementName_AddsPrefix(t *testing.T) {
	var buf bytes.Buffer
	p := &scriptedPrompter{answers: []string{"my-choice"}}
	name, err := promptForReplacementName(p, &buf, "bot-old", true)
	if err != nil {
		t.Fatal(err)
	}
	if name != "bot-my-choice" {
		t.Errorf("got %q, want bot-my-choice", name)
	}
}

func TestPromptForReplacementName_RejectsSameAsPrevious(t *testing.T) {
	var buf bytes.Buffer
	p := &scriptedPrompter{answers: []string{"bot-old"}}
	_, err := promptForReplacementName(p, &buf, "bot-old", true)
	if err == nil {
		t.Fatal("expected error when new name equals previous")
	}
}
