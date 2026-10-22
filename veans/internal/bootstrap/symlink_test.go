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
	"os"
	"path/filepath"
	"testing"

	"code.vikunja.io/veans/internal/config"
)

func openRoot(t *testing.T, dir string) *os.Root {
	t.Helper()
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = root.Close() })
	return root
}

func hostileRepo(t *testing.T) (repo, outside string) {
	t.Helper()
	base := t.TempDir()
	repo = filepath.Join(base, "repo")
	outside = filepath.Join(base, "home")
	for _, d := range []string{repo, outside} {
		if err := os.Mkdir(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	return repo, outside
}

func symlink(t *testing.T, target, link string) {
	t.Helper()
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
}

func assertMissing(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Lstat(path); !os.IsNotExist(err) {
		t.Fatalf("%s was created outside the repo (err=%v)", path, err)
	}
}

func TestConfigDest_SymlinkOutsideRepoRefused(t *testing.T) {
	repo, outside := hostileRepo(t)
	victim := filepath.Join(outside, ".bashrc")
	if err := os.WriteFile(victim, []byte("original"), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(repo, config.Filename)
	symlink(t, victim, link)

	if _, err := newConfigDest(openRoot(t, repo), repo, link); err == nil {
		t.Fatal("expected an error for a config symlink escaping the repo")
	}
	if buf, _ := os.ReadFile(victim); string(buf) != "original" {
		t.Fatalf("victim changed: %q", buf)
	}
}

func TestConfigDest_DanglingSymlinkOutsideRepoRefused(t *testing.T) {
	repo, outside := hostileRepo(t)
	victim := filepath.Join(outside, "new-file")
	link := filepath.Join(repo, config.Filename)
	symlink(t, victim, link)

	dest, err := newConfigDest(openRoot(t, repo), repo, link)
	if err == nil {
		err = dest.save(&config.Config{Server: "https://x"})
	}
	if err == nil {
		t.Fatal("expected an error for a dangling symlink escaping the repo")
	}
	assertMissing(t, victim)
}

func TestConfigDest_InRepoSymlinkWorks(t *testing.T) {
	repo, _ := hostileRepo(t)
	if err := os.Mkdir(filepath.Join(repo, "conf"), 0o755); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(repo, config.Filename)
	symlink(t, filepath.Join("conf", "veans.yml"), link)

	dest, err := newConfigDest(openRoot(t, repo), repo, link)
	if err != nil {
		t.Fatal(err)
	}
	if !dest.exists() {
		t.Fatal("dangling in-repo link should count as existing")
	}
	if err := dest.save(&config.Config{Server: "https://x"}); err != nil {
		t.Fatal(err)
	}
	if _, err := config.Load(filepath.Join(repo, "conf", "veans.yml")); err != nil {
		t.Fatalf("in-repo link target not written: %v", err)
	}
}

func TestConfigDest_OutsideRepoPathIsPlainWrite(t *testing.T) {
	repo, outside := hostileRepo(t)
	path := filepath.Join(outside, "custom.yml")

	dest, err := newConfigDest(openRoot(t, repo), repo, path)
	if err != nil {
		t.Fatal(err)
	}
	if err := dest.save(&config.Config{Server: "https://x"}); err != nil {
		t.Fatal(err)
	}
	if _, err := config.Load(path); err != nil {
		t.Fatalf("--config outside the repo not written: %v", err)
	}
}

func TestConfigDest_SymlinkedParentDirRefused(t *testing.T) {
	repo, outside := hostileRepo(t)
	symlink(t, outside, filepath.Join(repo, "sub"))
	path := filepath.Join(repo, "sub", config.Filename)

	dest, err := newConfigDest(openRoot(t, repo), repo, path)
	if err == nil {
		err = dest.save(&config.Config{Server: "https://x"})
	}
	if err == nil {
		t.Fatal("expected an error for a config path through a symlinked dir")
	}
	assertMissing(t, filepath.Join(outside, config.Filename))
}

func TestInstallClaudeCodeHook_SymlinkedDirRefused(t *testing.T) {
	repo, outside := hostileRepo(t)
	symlink(t, outside, filepath.Join(repo, ".claude"))

	if _, _, err := installClaudeCodeHook(openRoot(t, repo)); err == nil {
		t.Fatal("expected an error for .claude escaping the repo")
	}
	assertMissing(t, filepath.Join(outside, "settings.json"))
}

func TestInstallClaudeCodeHook_SymlinkedFileRefused(t *testing.T) {
	repo, outside := hostileRepo(t)
	victim := filepath.Join(outside, "settings.json")
	if err := os.WriteFile(victim, []byte(`{"model":"x"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(repo, ".claude"), 0o755); err != nil {
		t.Fatal(err)
	}
	symlink(t, victim, filepath.Join(repo, ".claude", "settings.json"))

	if _, _, err := installClaudeCodeHook(openRoot(t, repo)); err == nil {
		t.Fatal("expected an error for settings.json escaping the repo")
	}
	if buf, _ := os.ReadFile(victim); string(buf) != `{"model":"x"}` {
		t.Fatalf("victim changed: %q", buf)
	}
}

func TestInstallClaudeCodeHook_InRepoSymlinkWorks(t *testing.T) {
	repo, _ := hostileRepo(t)
	if err := os.Mkdir(filepath.Join(repo, "agent"), 0o755); err != nil {
		t.Fatal(err)
	}
	symlink(t, "agent", filepath.Join(repo, ".claude"))

	if _, _, err := installClaudeCodeHook(openRoot(t, repo)); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(repo, "agent", "settings.json")); err != nil {
		t.Fatalf("in-repo link target not written: %v", err)
	}
}

func TestInstallOpenCodeHook_SymlinkedDirRefused(t *testing.T) {
	repo, outside := hostileRepo(t)
	symlink(t, outside, filepath.Join(repo, ".opencode"))

	if _, _, err := installOpenCodeHook(openRoot(t, repo)); err == nil {
		t.Fatal("expected an error for .opencode escaping the repo")
	}
	assertMissing(t, filepath.Join(outside, "plugin"))
}
