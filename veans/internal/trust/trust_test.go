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

package trust

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

func TestOrigin(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"https://vikunja.example.com", "https://vikunja.example.com:443"},
		{"HTTPS://Vikunja.Example.com/sub/path/", "https://vikunja.example.com:443"},
		{"http://127.0.0.1:3456", "http://127.0.0.1:3456"},
		{"http://localhost", "http://localhost:80"},
		{"https://[::1]:8443/x", "https://[::1]:8443"},
	}
	for _, c := range cases {
		got, err := Origin(c.in)
		if err != nil {
			t.Errorf("Origin(%q): %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("Origin(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestOrigin_Rejects(t *testing.T) {
	for _, in := range []string{
		"https://trusted.example@evil.example/",
		"https://user:pass@evil.example",
		"ftp://vikunja.example.com",
		"vikunja.example.com",
		"https://",
		"https://vіkunja.example.com",
		"",
	} {
		if got, err := Origin(in); err == nil {
			t.Errorf("Origin(%q) = %q, want error", in, got)
		}
	}
}

type scriptedPrompter struct {
	answer  string
	prompts []string
}

func (p *scriptedPrompter) ReadLine(prompt string) (string, error) {
	p.prompts = append(p.prompts, prompt)
	return p.answer, nil
}

func newGate(t *testing.T, p Prompter) *Gate {
	t.Helper()
	t.Setenv("VEANS_SERVER", "")
	return &Gate{Store: NewStore(filepath.Join(t.TempDir(), "trusted.json")), Prompter: p}
}

func TestGate_NonInteractiveFailsClosed(t *testing.T) {
	g := newGate(t, nil)
	if err := g.Check("https://vikunja.example.com"); !errors.Is(err, ErrUntrusted) {
		t.Fatalf("got %v, want ErrUntrusted", err)
	}
}

func TestGate_PromptYesRecords(t *testing.T) {
	p := &scriptedPrompter{answer: "y"}
	g := newGate(t, p)
	if err := g.Check("https://vikunja.example.com/"); err != nil {
		t.Fatalf("Check: %v", err)
	}
	if len(p.prompts) != 1 {
		t.Fatalf("expected one prompt, got %d", len(p.prompts))
	}

	g.Prompter = nil
	if err := g.Check("https://VIKUNJA.example.com:443/other"); err != nil {
		t.Fatalf("recorded origin not trusted: %v", err)
	}
	if err := g.Check("http://vikunja.example.com"); !errors.Is(err, ErrUntrusted) {
		t.Fatalf("http origin: got %v, want ErrUntrusted", err)
	}
}

func TestGate_PromptShowsParsedOrigin(t *testing.T) {
	p := &scriptedPrompter{answer: "y"}
	g := newGate(t, p)
	if err := g.Check("http://127.0.0.1:3456/sub"); err != nil {
		t.Fatal(err)
	}
	if want := "http://127.0.0.1:3456"; len(p.prompts) != 1 || !strings.Contains(p.prompts[0], want) {
		t.Fatalf("prompt %q does not show %q", p.prompts, want)
	}
}

func TestGate_PromptNoRefuses(t *testing.T) {
	for _, ans := range []string{"", "n", "nope"} {
		g := newGate(t, &scriptedPrompter{answer: ans})
		if err := g.Check("https://vikunja.example.com"); !errors.Is(err, ErrUntrusted) {
			t.Fatalf("answer %q: got %v, want ErrUntrusted", ans, err)
		}
		g.Prompter = nil
		if err := g.Check("https://vikunja.example.com"); !errors.Is(err, ErrUntrusted) {
			t.Fatalf("answer %q was recorded: %v", ans, err)
		}
	}
}

func TestGate_UserinfoRejectedWithoutPrompt(t *testing.T) {
	p := &scriptedPrompter{answer: "y"}
	g := newGate(t, p)
	if err := g.Check("https://trusted.example@evil.example/"); err == nil {
		t.Fatal("userinfo server passed the gate")
	}
	if len(p.prompts) != 0 {
		t.Fatalf("prompted for a userinfo URL: %q", p.prompts)
	}
}

func TestGate_EnvServer(t *testing.T) {
	g := newGate(t, nil)
	t.Setenv("VEANS_SERVER", "HTTPS://vikunja.example.com/")
	if err := g.Check("https://vikunja.example.com:443"); err != nil {
		t.Fatalf("matching VEANS_SERVER: %v", err)
	}
	if err := g.Check("https://evil.example"); !errors.Is(err, ErrUntrusted) {
		t.Fatalf("mismatching VEANS_SERVER: got %v, want ErrUntrusted", err)
	}
	t.Setenv("VEANS_SERVER", "https://vikunja.example.com.evil.example")
	if err := g.Check("https://vikunja.example.com"); !errors.Is(err, ErrUntrusted) {
		t.Fatalf("prefix VEANS_SERVER: got %v, want ErrUntrusted", err)
	}
}

func TestGate_Record(t *testing.T) {
	g := newGate(t, nil)
	if err := g.Record("https://vikunja.example.com"); err != nil {
		t.Fatal(err)
	}
	if err := g.Check("https://vikunja.example.com"); err != nil {
		t.Fatalf("recorded origin not trusted: %v", err)
	}
}

func TestDefaultStore_IgnoresXDG(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	if got, want := DefaultStore().path, filepath.Join(home, ".config", "veans", "trusted.json"); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
