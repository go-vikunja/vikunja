// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2020 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

// +build mage

package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/magefile/mage/mg"
	"golang.org/x/sync/errgroup"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	PACKAGE = `code.vikunja.io/api`
	DIST    = `dist`
)

var (
	Goflags = []string{
		"-v",
	}
	Executable    = "vikunja"
	Ldflags       = ""
	Tags          = ""
	VersionNumber = "dev"
	Version       = "master" // This holds the built version, master by default, when building from a tag or release branch, their name
	BinLocation   = ""
	PkgVersion    = "master"
	ApiPackages   = []string{}
	RootPath      = ""
	GoFiles       = []string{}

	// Aliases are mage aliases of targets
	Aliases = map[string]interface{}{
		"do-the-swag":        DoTheSwag,
		"check:go-sec":       Check.GoSec,
		"check:got-swag":     Check.GotSwag,
		"release:os-package": Release.OsPackage,
	}
)

func setVersion() {
	versionCmd := exec.Command("git", "describe", "--tags", "--always", "--abbrev=10")
	version, err := versionCmd.Output()
	if err != nil {
		fmt.Printf("Error getting version: %s\n", err)
		os.Exit(1)
	}
	VersionNumber = strings.Trim(string(version), "\n")
	VersionNumber = strings.Replace(VersionNumber, "-", "+", 1)
	VersionNumber = strings.Replace(VersionNumber, "-g", "-", 1)

	if os.Getenv("DRONE_TAG") != "" {
		Version = os.Getenv("DRONE_TAG")
	} else if os.Getenv("DRONE_BRANCH") != "" {
		Version = strings.Replace(os.Getenv("DRONE_BRANCH"), "release/v", "", 1)
	}
}

func setBinLocation() {
	if os.Getenv("DRONE_WORKSPACE") != "" {
		BinLocation = DIST + `/binaries/` + Executable + `-` + Version + `-linux-amd64`
	} else {
		BinLocation = Executable
	}
}

func setPkgVersion() {
	if Version == "master" {
		PkgVersion = VersionNumber
	}
}

func setExecutable() {
	if runtime.GOOS == "windows" {
		Executable += ".exe"
	}
}

func setApiPackages() {
	cmd := exec.Command("go", "list", "all")
	pkgs, err := cmd.Output()
	if err != nil {
		fmt.Printf("Error getting packages: %s\n", err)
		os.Exit(1)
	}
	for _, p := range strings.Split(string(pkgs), "\n") {
		if strings.Contains(p, "code.vikunja.io/api") && !strings.Contains(p, "code.vikunja.io/api/pkg/integrations") {
			ApiPackages = append(ApiPackages, p)
		}
	}
}

func setRootPath() {
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting pwd: %s\n", err)
		os.Exit(1)
	}
	if err := os.Setenv("VIKUNJA_SERVICE_ROOTPATH", pwd); err != nil {
		fmt.Printf("Error setting root path: %s\n", err)
		os.Exit(1)
	}
	RootPath = pwd
}

func setGoFiles() {
	// GOFILES := $(shell find . -name "*.go" -type f ! -path "*/bindata.go")
	cmd := exec.Command("find", ".", "-name", "*.go", "-type", "f", "!", "-path", "*/bindata.go")
	files, err := cmd.Output()
	if err != nil {
		fmt.Printf("Error getting go files: %s\n", err)
		os.Exit(1)
	}
	for _, f := range strings.Split(string(files), "\n") {
		if strings.HasSuffix(f, ".go") {
			GoFiles = append(GoFiles, RootPath+strings.TrimLeft(f, "."))
		}
	}
}

// Some variables can always get initialized, so we do just that.
func init() {
	setExecutable()
	setRootPath()
}

// Some variables have external dependencies (like git) which may not always be available.
func initVars() {
	Tags = os.Getenv("TAGS")
	setVersion()
	setBinLocation()
	setPkgVersion()
	setApiPackages()
	setGoFiles()
	Ldflags = `-X "` + PACKAGE + `/pkg/version.Version=` + VersionNumber + `" -X "main.Tags=` + Tags + `"`
}

func runAndStreamOutput(cmd string, args ...string) {
	c := exec.Command(cmd, args...)

	c.Env = os.Environ()
	c.Dir = RootPath

	fmt.Printf("%s\n\n", c.String())

	stdout, _ := c.StdoutPipe()
	errbuf := bytes.Buffer{}
	c.Stderr = &errbuf
	c.Start()

	reader := bufio.NewReader(stdout)
	line, err := reader.ReadString('\n')
	for err == nil {
		fmt.Print(line)
		line, err = reader.ReadString('\n')
	}

	if err := c.Wait(); err != nil {
		fmt.Printf(errbuf.String())
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
}

// Will check if the tool exists and if not install it from the provided import path
// If any errors occur, it will exit with a status code of 1.
func checkAndInstallGoTool(tool, importPath string) {
	if err := exec.Command(tool).Run(); err != nil && strings.Contains(err.Error(), "executable file not found") {
		fmt.Printf("%s not installed, installing %s...\n", tool, importPath)
		if err := exec.Command("go", "install", Goflags[0], importPath).Run(); err != nil {
			fmt.Printf("Error installing %s\n", tool)
			os.Exit(1)
		}
		fmt.Println("Installed.")
	}
}

// Calculates a hash of a file
func calculateSha256FileHash(path string) (hash string, err error) {
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

// Copy the src file to dst. Any existing file will be overwritten and will not
// copy file attributes.
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

	_, err = io.Copy(out, in)
	if err != nil {
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

// os.Rename has issues with moving files between docker volumes.
// Because of this limitaion, it fails in drone.
// Source: https://gist.github.com/var23rav/23ae5d0d4d830aff886c3c970b8f6c6b
func moveFile(sourcePath, destPath string) error {
	inputFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("couldn't open source file: %s", err)
	}
	outputFile, err := os.Create(destPath)
	if err != nil {
		inputFile.Close()
		return fmt.Errorf("couldn't open dest file: %s", err)
	}
	defer outputFile.Close()
	_, err = io.Copy(outputFile, inputFile)
	inputFile.Close()
	if err != nil {
		return fmt.Errorf("writing to output file failed: %s", err)
	}
	// The copy was successful, so now delete the original file
	err = os.Remove(sourcePath)
	if err != nil {
		return fmt.Errorf("failed removing original file: %s", err)
	}
	return nil
}

// Formats the code using go fmt
func Fmt() {
	mg.Deps(initVars)
	args := append([]string{"-s", "-w"}, GoFiles...)
	runAndStreamOutput("gofmt", args...)
}

// Generates the swagger docs from the code annotations
func DoTheSwag() {
	mg.Deps(initVars)
	checkAndInstallGoTool("swag", "github.com/swaggo/swag/cmd/swag")
	runAndStreamOutput("swag", "init", "-g", "./pkg/routes/routes.go", "--parseDependency", "-d", RootPath, "-o", RootPath+"/pkg/swagger")
}

type Test mg.Namespace

// Runs all tests except integration tests
func (Test) Unit() {
	mg.Deps(initVars)
	// We run everything sequentially and not in parallel to prevent issues with real test databases
	args := append([]string{"test", Goflags[0], "-p", "1"}, ApiPackages...)
	runAndStreamOutput("go", args...)
}

// Runs the tests and builds the coverage html file from coverage output
func (Test) Coverage() {
	mg.Deps(initVars)
	mg.Deps(Test.Unit)
	runAndStreamOutput("go", "tool", "cover", "-html=cover.out", "-o", "cover.html")
}

// Runs the integration tests
func (Test) Integration() {
	mg.Deps(initVars)
	// We run everything sequentially and not in parallel to prevent issues with real test databases
	runAndStreamOutput("go", "test", Goflags[0], "-p", "1", PACKAGE+"/pkg/integrations")
}

type Check mg.Namespace

// Checks if the code is properly formatted with go fmt
func (Check) Fmt() error {
	mg.Deps(initVars)
	args := append([]string{"-s", "-d"}, GoFiles...)
	c := exec.Command("gofmt", args...)
	out, err := c.Output()
	if err != nil {
		return err
	}

	if len(out) > 0 {
		fmt.Println("Code is not properly gofmt'ed.")
		fmt.Println("Please run 'mage fmt' and commit the result:")
		fmt.Print(string(out))
		os.Exit(1)
	}

	return nil
}

// Runs golint on all packages
func (Check) Lint() {
	mg.Deps(initVars)
	checkAndInstallGoTool("golint", "golang.org/x/lint/golint")
	args := append([]string{"-set_exit_status"}, ApiPackages...)
	runAndStreamOutput("golint", args...)
}

// Checks if the swagger docs need to be re-generated from the code annotations
func (Check) GotSwag() {
	mg.Deps(initVars)
	// The check is pretty cheaply done: We take the hash of the swagger.json file, generate the docs,
	// hash the file again and compare the two hashes to see if anything changed. If that's the case,
	// regenerating the docs is necessary.
	// swag is not capable of just outputting the generated docs to stdout, therefore we need to do it this way.
	// Another drawback of this is obviously it will only work once - we're not resetting the newly generated
	// docs after the check. This behaviour is good enough for ci though.
	oldHash, err := calculateSha256FileHash(RootPath + "/pkg/swagger/swagger.json")
	if err != nil {
		fmt.Printf("Error getting old hash of the swagger docs: %s", err)
		os.Exit(1)
	}

	DoTheSwag()

	newHash, err := calculateSha256FileHash(RootPath + "/pkg/swagger/swagger.json")
	if err != nil {
		fmt.Printf("Error getting new hash of the swagger docs: %s", err)
		os.Exit(1)
	}

	if oldHash != newHash {
		fmt.Println("Swagger docs are not up to date.")
		fmt.Println("Please run 'mage do-the-swag' and commit the result.")
		os.Exit(1)
	}
}

// Checks the source code for misspellings
func (Check) Misspell() {
	mg.Deps(initVars)
	checkAndInstallGoTool("misspell", "github.com/client9/misspell/cmd/misspell")
	runAndStreamOutput("misspell", append([]string{"-error"}, GoFiles...)...)
}

// Checks the source code for ineffectual assigns
func (Check) Ineffassign() {
	mg.Deps(initVars)
	checkAndInstallGoTool("ineffassign", "github.com/gordonklaus/ineffassign")
	runAndStreamOutput("ineffassign", GoFiles...)
}

// Checks for the cyclomatic complexity of the source code
func (Check) Gocyclo() {
	mg.Deps(initVars)
	checkAndInstallGoTool("gocyclo", "github.com/fzipp/gocyclo")
	runAndStreamOutput("gocyclo", append([]string{"-over", "49"}, GoFiles...)...)
}

// Statically analyzes the source code about a range of different problems
func (Check) Static() {
	mg.Deps(initVars)
	checkAndInstallGoTool("staticcheck", "honnef.co/go/tools/cmd/staticcheck")
	runAndStreamOutput("staticcheck", ApiPackages...)
}

// Checks the source code for potential security issues
func (Check) GoSec() {
	mg.Deps(initVars)
	if err := exec.Command("gosec").Run(); err != nil && strings.Contains(err.Error(), "executable file not found") {
		fmt.Println("Please manually install gosec by running")
		fmt.Println("curl -sfL https://raw.githubusercontent.com/securego/gosec/master/install.sh | bash -s -- -b $GOPATH/bin v2.2.0")
		os.Exit(1)
	}
	runAndStreamOutput("gosec", "./...")
}

// Checks for repeated strings that could be replaced by a constant
func (Check) Goconst() {
	mg.Deps(initVars)
	checkAndInstallGoTool("goconst", "github.com/jgautheron/goconst/cmd/goconst")
	runAndStreamOutput("goconst", ApiPackages...)
}

// Runs fmt-check, lint, got-swag, misspell-check, ineffasign-check, gocyclo-check, static-check, gosec-check, goconst-check all in parallel
func (Check) All() {
	mg.Deps(initVars)
	mg.Deps(
		Check.Fmt,
		Check.Lint,
		Check.GotSwag,
		Check.Misspell,
		Check.Ineffassign,
		Check.Gocyclo,
		Check.Static,
		Check.GoSec,
		Check.Goconst,
	)
}

type Build mg.Namespace

// Cleans all build, executable and bindata files
func (Build) Clean() error {
	mg.Deps(initVars)
	if err := exec.Command("go", "clean", "./...").Run(); err != nil {
		return err
	}
	if err := os.Remove(Executable); err != nil && !os.IsNotExist(err) {
		return err
	}
	if err := os.RemoveAll(DIST); err != nil && !os.IsNotExist(err) {
		return err
	}
	if err := os.RemoveAll(BinLocation); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// Generates static content into the final binary
func (Build) Generate() {
	mg.Deps(initVars)
	runAndStreamOutput("go", "generate", PACKAGE+"/pkg/static")
}

// Builds a vikunja binary, ready to run
func (Build) Build() {
	mg.Deps(initVars)
	mg.Deps(Build.Generate)
	runAndStreamOutput("go", "build", Goflags[0], "-tags", Tags, "-ldflags", "-s -w "+Ldflags, "-o", Executable)
}

type Release mg.Namespace

// Runs all steps in the right order to create release packages for various platforms
func (Release) Release(ctx context.Context) error {
	mg.Deps(initVars)
	mg.Deps(Build.Generate, Release.Dirs)
	mg.Deps(Release.Windows, Release.Linux, Release.Darwin)

	// Run compiling in parallel to speed it up
	errs, _ := errgroup.WithContext(ctx)
	errs.Go((Release{}).Windows)
	errs.Go((Release{}).Linux)
	errs.Go((Release{}).Darwin)
	if err := errs.Wait(); err != nil {
		return err
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
	if err := (Release{}).Zip(); err != nil {
		return err
	}

	return nil
}

// Creates all directories needed to release vikunja
func (Release) Dirs() error {
	for _, d := range []string{"binaries", "release", "zip"} {
		if err := os.MkdirAll(RootPath+"/"+DIST+"/"+d, 0755); err != nil {
			return err
		}
	}
	return nil
}

func runXgo(targets string) error {
	mg.Deps(initVars)
	checkAndInstallGoTool("xgo", "src.techknowlogick.com/xgo")

	extraLdflags := `-linkmode external -extldflags "-static" `

	// See https://github.com/techknowlogick/xgo/issues/79
	if strings.HasPrefix(targets, "darwin") {
		extraLdflags = ""
	}

	runAndStreamOutput("xgo",
		"-dest", RootPath+"/"+DIST+"/binaries",
		"-tags", "netgo "+Tags,
		"-ldflags", extraLdflags+Ldflags,
		"-targets", targets,
		"-out", Executable+"-"+Version,
		RootPath)
	if os.Getenv("DRONE_WORKSPACE") != "" {
		return filepath.Walk("/build/", func(path string, info os.FileInfo, err error) error {
			// Skip directories
			if info.IsDir() {
				return nil
			}

			return moveFile(path, RootPath+"/"+DIST+"/binaries/"+info.Name())
		})
	}
	return nil
}

// Builds binaries for windows
func (Release) Windows() error {
	return runXgo("windows/*")
}

// Builds binaries for linux
func (Release) Linux() error {
	return runXgo("linux/*")
}

// Builds binaries for darwin
func (Release) Darwin() error {
	return runXgo("darwin/*")
}

// Compresses the built binaries in dist/binaries/ to reduce their filesize
func (Release) Compress(ctx context.Context) error {
	// $(foreach file,$(filter-out $(wildcard $(wildcard $(DIST)/binaries/$(EXECUTABLE)-*mips*)),$(wildcard $(DIST)/binaries/$(EXECUTABLE)-*)), upx -9 $(file);)

	errs, _ := errgroup.WithContext(ctx)

	filepath.Walk(RootPath+"/"+DIST+"/binaries/", func(path string, info os.FileInfo, err error) error {
		// Only executable files
		if !strings.Contains(info.Name(), Executable) {
			return nil
		}
		// No mips or s390x for you today
		if strings.Contains(info.Name(), "mips") || strings.Contains(info.Name(), "s390x") {
			return nil
		}

		// Runs compressing in parallel since upx is single-threaded
		errs.Go(func() error {
			runAndStreamOutput("upx", "-9", path)
			return nil
		})

		return nil
	})

	return errs.Wait()
}

// Copies all built binaries to dist/release/ in preparation for creating the os packages
func (Release) Copy() error {
	return filepath.Walk(RootPath+"/"+DIST+"/binaries/", func(path string, info os.FileInfo, err error) error {
		// Only executable files
		if !strings.Contains(info.Name(), Executable) {
			return nil
		}

		return copyFile(path, RootPath+"/"+DIST+"/release/"+info.Name())
	})
}

// Creates sha256 checksum files for each binary in dist/release/
func (Release) Check() error {
	p := RootPath + "/" + DIST + "/release/"
	return filepath.Walk(p, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		f, err := os.Create(p + info.Name() + ".sha256")
		if err != nil {
			return err
		}

		hash, err := calculateSha256FileHash(path)
		if err != nil {
			return err
		}

		_, err = f.WriteString(hash + "  " + info.Name())
		if err != nil {
			return err
		}

		return f.Close()
	})
}

// Creates a folder for each
func (Release) OsPackage() error {
	p := RootPath + "/" + DIST + "/release/"

	// We first put all files in a map to then iterate over it since the walk function would otherwise also iterate
	// over the newly created files, creating some kind of endless loop.
	bins := make(map[string]os.FileInfo)
	if err := filepath.Walk(p, func(path string, info os.FileInfo, err error) error {
		if strings.Contains(info.Name(), ".sha256") || info.IsDir() {
			return nil
		}
		bins[path] = info
		return nil
	}); err != nil {
		return err
	}

	for path, info := range bins {
		folder := p + info.Name() + "-full/"
		if err := os.Mkdir(folder, 0755); err != nil {
			return err
		}
		if err := moveFile(p+info.Name()+".sha256", folder+info.Name()+".sha256"); err != nil {
			return err
		}
		if err := moveFile(path, folder+info.Name()); err != nil {
			return err
		}
		if err := copyFile(RootPath+"/config.yml.sample", folder+"config.yml.sample"); err != nil {
			return err
		}
		if err := copyFile(RootPath+"/LICENSE", folder+"LICENSE"); err != nil {
			return err
		}
	}
	return nil
}

// Creates a zip file from all os-package folders in dist/release
func (Release) Zip() error {
	p := RootPath + "/" + DIST + "/release/"
	if err := filepath.Walk(p, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() || info.Name() == "release" {
			return nil
		}

		fmt.Printf("Zipping %s...\n", info.Name())

		c := exec.Command("zip", "-r", RootPath+"/"+DIST+"/zip/"+info.Name(), ".", "-i", "*")
		c.Dir = path
		out, err := c.Output()
		fmt.Print(string(out))
		return err
	}); err != nil {
		return err
	}

	return nil
}

// Creates a debian package from a built binary
func (Release) Deb() {
	runAndStreamOutput(
		"fpm",
		"-s", "dir",
		"-t", "deb",
		"--url", "https://vikunja.io",
		"-n", "vikunja",
		"-v", PkgVersion,
		"--license", "GPLv3",
		"--directories", "/opt/vikunja",
		"--after-install", "./build/after-install.sh",
		"--description", "'Vikunja is an open-source todo application, written in Go. It lets you create lists,tasks and share them via teams or directly between users.'",
		"-m", "maintainers@vikunja.io",
		"-p", RootPath+"/"+Executable+"-"+Version+"_amd64.deb",
		"./"+BinLocation+"=/opt/vikunja/vikunja",
		"./config.yml.sample=/etc/vikunja/config.yml",
	)
}

// Creates a debian repo structure
func (Release) Reprepro() {
	runAndStreamOutput("reprepro_expect", "debian", "includedeb", "strech", RootPath+"/"+Executable+"-"+Version+"_amd64.deb")
}
