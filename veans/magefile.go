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
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

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
	for _, p := range []string{"./veans", "./veans.exe", "./" + releaseDist} {
		if _, err := os.Stat(p); err == nil {
			if err := os.RemoveAll(p); err != nil {
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

// All runs unit tests across the module. Passes `-short` so the e2e
// package self-skips via its TestMain — the parent monorepo's
// pkg/webtests follows the same convention.
func (Test) All() error {
	return sh.RunV("go", "test", "-short", "./...")
}

// Filter runs `go test -short -run <expr> ./...` — pass the expression as
// an argument. `-short` is included so e2e doesn't run accidentally; use
// `mage test:e2e` for those.
func (Test) Filter(expr string) error {
	if expr == "" {
		return fmt.Errorf("test:filter requires a regexp argument")
	}
	return sh.RunV("go", "test", "-short", "-run", expr, "./...")
}

// E2E runs the e2e suite without `-short` so TestMain lets it through.
// Requires VEANS_E2E_API_URL to point at a running Vikunja instance and
// either VEANS_E2E_TESTING_TOKEN (matching the API's VIKUNJA_SERVICE_TESTINGTOKEN
// — the harness will seed its own admin via /api/v1/test/users) or
// VEANS_E2E_ADMIN_TOKEN (a pre-existing JWT for the admin to use as-is).
//
// Set VEANS_E2E_SKIP_BUILD=true to reuse a previously-built binary.
func (Test) E2E() error {
	if os.Getenv("VEANS_E2E_API_URL") == "" {
		return fmt.Errorf("VEANS_E2E_API_URL is not set — start a Vikunja instance and export the URL")
	}
	if os.Getenv("VEANS_E2E_ADMIN_TOKEN") == "" && os.Getenv("VEANS_E2E_TESTING_TOKEN") == "" {
		return fmt.Errorf("set VEANS_E2E_ADMIN_TOKEN, or VEANS_E2E_TESTING_TOKEN (matching the API's VIKUNJA_SERVICE_TESTINGTOKEN) so the suite can seed its own admin")
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

// -----------------------------------------------------------------------------
// Release
//
// Cross-compiles the veans binary for every OS/arch the parent vikunja binary
// targets, runs upx where supported, bundles each into a zip with the LICENSE
// and a sha256, and templates nfpm.yaml so the CI can build deb/rpm/apk/
// archlinux packages from the same artifacts. Everything lands under
// `<veans>/dist/`. The CI workflow uploads dist/zip/* to S3 /veans/<ver>/ and
// hands dist/binaries/* off to the nfpm job.

const releaseDist = "dist"

type Release mg.Namespace

var (
	releaseVersionNumber string
	releaseVersionString string
	releaseLdflags       string
	releaseTags          = "netgo osusergo"
	releaseInitOnce      sync.Once
	releaseInitErr       error
)

func releaseInitVars(ctx context.Context) error {
	releaseInitOnce.Do(func() {
		num := os.Getenv("RELEASE_VERSION")
		if num == "" {
			out, err := exec.CommandContext(ctx, "git", "describe", "--tags", "--always", "--abbrev=10").Output()
			if err != nil {
				releaseInitErr = fmt.Errorf("git describe: %w", err)
				return
			}
			num = strings.TrimSpace(string(out))
		}
		releaseVersionNumber = strings.Replace(strings.Trim(num, "\n"), "-g", "-", 1)
		switch releaseVersionNumber {
		case "", "main":
			releaseVersionString = "unstable"
		default:
			releaseVersionString = releaseVersionNumber
		}
		releaseLdflags = fmt.Sprintf(`-X main.version=%s`, releaseVersionNumber)
	})
	return releaseInitErr
}

// Release runs all release steps end-to-end: dirs → xgo (windows/linux/darwin
// in parallel) → upx → copy → sha256 → per-target bundle dirs → zip.
func (Release) Release(ctx context.Context) error {
	mg.Deps(releaseInitVars)
	if err := (Release{}).Dirs(); err != nil {
		return err
	}
	if err := releasePrepareXgo(ctx); err != nil {
		return err
	}

	// Run cross-compilation per OS in parallel; xgo serializes targets
	// inside the docker container so each OS still gets full CPU.
	var wg sync.WaitGroup
	var (
		mu       sync.Mutex
		firstErr error
	)
	record := func(err error) {
		if err == nil {
			return
		}
		mu.Lock()
		if firstErr == nil {
			firstErr = err
		}
		mu.Unlock()
	}
	for _, fn := range []func(context.Context) error{
		(Release{}).Windows,
		(Release{}).Linux,
		(Release{}).Darwin,
	} {
		wg.Add(1)
		go func(f func(context.Context) error) {
			defer wg.Done()
			record(f(ctx))
		}(fn)
	}
	wg.Wait()
	if firstErr != nil {
		return firstErr
	}

	if err := (Release{}).Compress(ctx); err != nil {
		return err
	}
	if err := (Release{}).Copy(); err != nil {
		return err
	}
	if err := (Release{}).Check(); err != nil {
		return err
	}
	if err := (Release{}).OsPackage(); err != nil {
		return err
	}
	return (Release{}).Zip(ctx)
}

// Dirs creates all directories needed to release veans.
func (Release) Dirs() error {
	for _, d := range []string{"binaries", "release", "zip"} {
		if err := os.MkdirAll(filepath.Join(releaseDist, d), 0o755); err != nil {
			return err
		}
	}
	return nil
}

func releasePrepareXgo(_ context.Context) error {
	if _, err := exec.LookPath("xgo"); err != nil {
		fmt.Println("xgo not found, installing src.techknowlogick.com/xgo...")
		if err := sh.RunV("go", "install", "src.techknowlogick.com/xgo@latest"); err != nil {
			return fmt.Errorf("installing xgo: %w", err)
		}
	}
	fmt.Println("Pulling latest xgo docker image...")
	return sh.RunV("docker", "pull", "ghcr.io/techknowlogick/xgo:latest")
}

func runXgo(ctx context.Context, targets string) error {
	mg.Deps(releaseInitVars)
	if err := releasePrepareXgo(ctx); err != nil {
		return err
	}

	extraLdflags := `-linkmode external -extldflags "-static" `
	// xgo's darwin builds can't use the static external linker.
	if strings.HasPrefix(targets, "darwin") {
		extraLdflags = ""
	}

	outName := os.Getenv("XGO_OUT_NAME")
	if outName == "" {
		outName = "veans-" + releaseVersionString
	}

	return sh.RunV("xgo",
		"-dest", filepath.Join(releaseDist, "binaries"),
		"-tags", releaseTags,
		"-ldflags", extraLdflags+releaseLdflags,
		"-targets", targets,
		"-out", outName,
		"./cmd/veans",
	)
}

// Windows builds binaries for windows. Same target set as parent vikunja.
func (Release) Windows(ctx context.Context) error {
	return runXgo(ctx, "windows/*")
}

// Linux builds binaries for linux. Same target set as parent vikunja.
func (Release) Linux(ctx context.Context) error {
	targets := []string{
		"linux/amd64",
		"linux/arm-5",
		"linux/arm-6",
		"linux/arm-7",
		"linux/arm64",
		"linux/mips",
		"linux/mipsle",
		"linux/mips64",
		"linux/mips64le",
		"linux/riscv64",
	}
	return runXgo(ctx, strings.Join(targets, ","))
}

// Darwin builds binaries for macOS. Same minimum (10.15) as parent.
func (Release) Darwin(ctx context.Context) error {
	return runXgo(ctx, "darwin-10.15/*")
}

// Xgo cross-compiles a single os/arch[-variant] target.
func (Release) Xgo(ctx context.Context, target string) error {
	parts := strings.Split(target, "/")
	if len(parts) < 2 {
		return fmt.Errorf("invalid target %q (expected os/arch[/variant])", target)
	}
	variant := ""
	if len(parts) > 2 && parts[2] != "" {
		variant = "-" + strings.ReplaceAll(parts[2], "v", "")
	}
	return runXgo(ctx, parts[0]+"/"+parts[1]+variant)
}

// Compress runs upx -9 over each built binary that upx can actually handle.
// Skip list matches the parent vikunja magefile.
func (Release) Compress(_ context.Context) error {
	var wg sync.WaitGroup
	var (
		mu       sync.Mutex
		firstErr error
	)
	record := func(err error) {
		if err == nil {
			return
		}
		mu.Lock()
		if firstErr == nil {
			firstErr = err
		}
		mu.Unlock()
	}

	walkErr := filepath.Walk(filepath.Join(releaseDist, "binaries"), func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		name := info.Name()
		if !strings.Contains(name, "veans") {
			return nil
		}
		if strings.Contains(name, "mips") ||
			strings.Contains(name, "s390x") ||
			strings.Contains(name, "riscv64") ||
			strings.Contains(name, "darwin") ||
			(strings.Contains(name, "windows") && strings.Contains(name, "arm64")) {
			// upx can't compress these targets.
			return nil
		}
		wg.Add(1)
		go func(p string) {
			defer wg.Done()
			if err := sh.RunV("chmod", "+x", p); err != nil {
				record(err)
				return
			}
			record(sh.RunV("upx", "-9", p))
		}(path)
		return nil
	})
	if walkErr != nil {
		return walkErr
	}
	wg.Wait()
	return firstErr
}

// Copy copies all built binaries to dist/release/ as the staging area for
// per-target bundles and nfpm.
func (Release) Copy() error {
	return filepath.Walk(filepath.Join(releaseDist, "binaries"), func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		if !strings.Contains(info.Name(), "veans") {
			return nil
		}
		return copyFile(path, filepath.Join(releaseDist, "release", info.Name()))
	})
}

// Check writes a sha256 file next to each binary in dist/release/.
func (Release) Check() error {
	p := filepath.Join(releaseDist, "release")
	return filepath.Walk(p, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		if strings.HasSuffix(info.Name(), ".sha256") {
			return nil
		}
		sum, err := sha256File(path)
		if err != nil {
			return err
		}
		return os.WriteFile(path+".sha256", []byte(sum+"  "+info.Name()+"\n"), 0o644)
	})
}

// OsPackage creates one folder per binary in dist/release/, populated with
// the binary, its sha256, and the LICENSE so the bundle is self-contained.
func (Release) OsPackage() error {
	p := filepath.Join(releaseDist, "release")

	// Snapshot first so we don't walk into the newly-created folders.
	bins := map[string]os.FileInfo{}
	if err := filepath.Walk(p, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		if strings.HasSuffix(info.Name(), ".sha256") {
			return nil
		}
		bins[path] = info
		return nil
	}); err != nil {
		return err
	}

	licensePath, err := licenseSource()
	if err != nil {
		return err
	}

	for binPath, info := range bins {
		folder := filepath.Join(p, info.Name()+"-full") + string(os.PathSeparator)
		if err := os.MkdirAll(folder, 0o755); err != nil {
			return err
		}
		if err := moveFile(binPath+".sha256", filepath.Join(folder, info.Name()+".sha256")); err != nil {
			return err
		}
		if err := moveFile(binPath, filepath.Join(folder, info.Name())); err != nil {
			return err
		}
		if err := copyFile(licensePath, filepath.Join(folder, "LICENSE")); err != nil {
			return err
		}
	}
	return nil
}

// Zip turns each per-target folder under dist/release/<name>-full/ into
// dist/zip/<name>-full.zip. Uses the system `zip` so we get the same on-wire
// format as the parent's release artifacts.
func (Release) Zip(ctx context.Context) error {
	rootDir, err := os.Getwd()
	if err != nil {
		return err
	}
	p := filepath.Join(releaseDist, "release")
	return filepath.Walk(p, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() || info.Name() == "release" {
			return nil
		}
		fmt.Printf("Zipping %s...\n", info.Name())
		zipFile := filepath.Join(rootDir, releaseDist, "zip", info.Name()+".zip")
		//nolint:gosec // mage build helper; arguments are derived from the local fs walk above.
		c := exec.CommandContext(ctx, "zip", "-r", zipFile, ".", "-i", "*")
		c.Dir = path
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	})
}

// PrepareNFPMConfig templates ./nfpm.yaml in place, substituting <version>,
// <arch> and <binlocation> the same way the parent magefile does. Set
// NFPM_ARCH to the nfpm arch name (amd64, arm64, arm7, 386) before calling.
// The substituted file is meant to be consumed by `nfpm pkg` immediately
// after; this is destructive and intentional (the CI checks the repo out
// fresh per job).
func (Release) PrepareNFPMConfig() error {
	mg.Deps(releaseInitVars)
	cfgPath := "./nfpm.yaml"
	raw, err := os.ReadFile(cfgPath)
	if err != nil {
		return err
	}

	var nfpmArch string
	switch os.Getenv("NFPM_ARCH") {
	case "arm64":
		nfpmArch = "arm64"
	case "arm7":
		nfpmArch = "arm7"
	case "386":
		nfpmArch = "386"
	default:
		nfpmArch = "amd64"
	}

	// nfpm resolves <binlocation> relative to its working directory. In CI the
	// nfpm action runs from $GITHUB_WORKSPACE while the veans source already
	// occupies ./veans, so the CI stages the binary at ./veans/veans-bin and
	// passes NFPM_BIN_PATH=./veans/veans-bin. Outside CI the default works for
	// a local `mage build && mage release:prepare-nfpm-config && nfpm pkg
	// --config nfpm.yaml` from inside veans/.
	binLocation := os.Getenv("NFPM_BIN_PATH")
	if binLocation == "" {
		binLocation = "./veans"
	}

	fixed := strings.ReplaceAll(string(raw), "<version>", releaseVersionNumber)
	fixed = strings.ReplaceAll(fixed, "<arch>", nfpmArch)
	fixed = strings.ReplaceAll(fixed, "<binlocation>", binLocation)
	return os.WriteFile(cfgPath, []byte(fixed), 0o600)
}

// -----------------------------------------------------------------------------
// helpers

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	si, err := os.Stat(src)
	if err != nil {
		return err
	}
	if err := os.Chmod(dst, si.Mode()); err != nil {
		return err
	}
	return out.Close()
}

func moveFile(src, dst string) error {
	if err := copyFile(src, dst); err != nil {
		return err
	}
	return os.Remove(src)
}

func sha256File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// licenseSource resolves the AGPLv3 LICENSE file. veans intentionally doesn't
// vendor its own copy — the parent repo's LICENSE applies to both. Look in
// ../LICENSE (the normal layout when running from veans/) and fall back to
// ./LICENSE for unusual checkouts.
func licenseSource() (string, error) {
	for _, p := range []string{"../LICENSE", "./LICENSE"} {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("could not find LICENSE in ../ or ./")
}

// Aliases lets `mage test` resolve to `Test.All` (and the others) without
// having to spell out the namespace. Mirrors the parent magefile's pattern.
var Aliases = map[string]any{
	"test":                        Test.All,
	"test:filter":                 Test.Filter,
	"test:e2e":                    Test.E2E,
	"lint":                        Lint.All,
	"lint:fix":                    Lint.Fix,
	"release":                     Release.Release,
	"release:xgo":                 Release.Xgo,
	"release:prepare-nfpm-config": Release.PrepareNFPMConfig,
}
