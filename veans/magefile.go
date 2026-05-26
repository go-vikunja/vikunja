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

//go:build mage

// Mage targets for the veans CLI. Patterned after the parent monorepo's
// magefile (Build/Test/Lint namespaces), but scoped to this submodule.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// Build compiles the veans binary into ./veans (or ./veans.exe on Windows).
func Build() error {
	out := "./veans"
	if runtime.GOOS == "windows" {
		out = "./veans.exe"
	}
	return sh.RunV("go", "build", "-o", out, "./cmd/veans")
}

// Clean removes built artifacts.
func Clean() error {
	for _, p := range []string{"./veans", "./veans.exe"} {
		if _, err := os.Stat(p); err == nil {
			if err := os.Remove(p); err != nil {
				return err
			}
		}
	}
	return nil
}

// Fmt runs goimports across the module.
func Fmt() error {
	return sh.RunV("go", "fmt", "./...")
}

// Test namespace.
type Test mg.Namespace

// All runs unit tests across the module.
func (Test) All() error {
	return sh.RunV("go", "test", "./...")
}

// Filter runs `go test -run <expr> ./...` — pass the expression as an argument.
func (Test) Filter(expr string) error {
	if expr == "" {
		return fmt.Errorf("test:filter requires a regexp argument")
	}
	return sh.RunV("go", "test", "-run", expr, "./...")
}

// E2E runs the e2e suite. Requires VEANS_E2E_API_URL to point at a running
// Vikunja instance and either VEANS_E2E_ADMIN_TOKEN or
// VEANS_E2E_ADMIN_USER + VEANS_E2E_ADMIN_PASS for the admin/seed identity.
//
// Set VEANS_E2E_SKIP_BUILD=true to reuse a previously-built binary.
func (Test) E2E() error {
	if os.Getenv("VEANS_E2E_API_URL") == "" {
		return fmt.Errorf("VEANS_E2E_API_URL is not set — start a Vikunja instance and export the URL")
	}
	if os.Getenv("VEANS_E2E_ADMIN_TOKEN") == "" {
		if os.Getenv("VEANS_E2E_ADMIN_USER") == "" || os.Getenv("VEANS_E2E_ADMIN_PASS") == "" {
			return fmt.Errorf("set either VEANS_E2E_ADMIN_TOKEN or VEANS_E2E_ADMIN_USER + VEANS_E2E_ADMIN_PASS")
		}
	}
	if os.Getenv("VEANS_E2E_SKIP_BUILD") == "" {
		if err := Build(); err != nil {
			return err
		}
	}
	abs, err := filepath.Abs("./veans")
	if err != nil {
		return err
	}
	return sh.RunWithV(map[string]string{"VEANS_BINARY": abs}, "go", "test", "-count=1", "./e2e/...")
}

// Lint namespace.
type Lint mg.Namespace

// All runs golangci-lint over the module.
func (Lint) All() error {
	if _, err := exec.LookPath("golangci-lint"); err != nil {
		return fmt.Errorf("golangci-lint not installed: %w", err)
	}
	return sh.RunV("golangci-lint", "run", "./...")
}

// Fix runs golangci-lint with --fix.
func (Lint) Fix() error {
	if _, err := exec.LookPath("golangci-lint"); err != nil {
		return fmt.Errorf("golangci-lint not installed: %w", err)
	}
	return sh.RunV("golangci-lint", "run", "--fix", "./...")
}

// Aliases lets `mage test` resolve to `Test.All` (and the others) without
// having to spell out the namespace. Mirrors the parent magefile's pattern.
var Aliases = map[string]any{
	"test":        Test.All,
	"test:filter": Test.Filter,
	"test:e2e":    Test.E2E,
	"lint":        Lint.All,
	"lint:fix":    Lint.Fix,
}

// trimLast is a tiny helper for prettier path printing in error messages.
func trimLast(p string) string {
	return strings.TrimSuffix(p, "/")
}

var _ = trimLast
