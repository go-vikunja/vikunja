// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2021 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

//go:build mage
// +build mage

package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/iancoleman/strcase"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/magefile/mage/mg"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v3"
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
	Version       = "unstable" // This holds the built version, unstable by default, when building from a tag or release branch, their name
	BinLocation   = ""
	PkgVersion    = "unstable"
	ApiPackages   = []string{}
	RootPath      = ""
	GoFiles       = []string{}

	// Aliases are mage aliases of targets
	Aliases = map[string]interface{}{
		"build":                 Build.Build,
		"do-the-swag":           DoTheSwag,
		"check:got-swag":        Check.GotSwag,
		"release:os-package":    Release.OsPackage,
		"dev:make-migration":    Dev.MakeMigration,
		"dev:make-event":        Dev.MakeEvent,
		"dev:make-listener":     Dev.MakeListener,
		"dev:make-notification": Dev.MakeNotification,
		"generate-docs":         GenerateDocs,
		"check:golangci-fix":    Check.GolangciFix,
	}
)

func runCmdWithOutput(name string, arg ...string) (output []byte, err error) {
	cmd := exec.Command(name, arg...)
	output, err = cmd.Output()
	if err != nil {
		ee := err.(*exec.ExitError)
		return nil, fmt.Errorf("error running command: %s, %s", string(ee.Stderr), err)
	}

	return output, nil
}

func setVersion() {
	version, err := runCmdWithOutput("git", "describe", "--tags", "--always", "--abbrev=10")
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

	if Version == "main" {
		Version = "unstable"
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
	if Version == "unstable" {
		PkgVersion = VersionNumber
	}
}

func setExecutable() {
	if runtime.GOOS == "windows" {
		Executable += ".exe"
	}
}

func setApiPackages() {
	pkgs, err := runCmdWithOutput("go", "list", "all")
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
	files, err := runCmdWithOutput("find", ".", "-name", "*.go", "-type", "f", "!", "-path", "*/bindata.go")
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
func moveFile(src, dst string) error {
	inputFile, err := os.Open(src)
	defer inputFile.Close()
	if err != nil {
		return fmt.Errorf("couldn't open source file: %s", err)
	}

	outputFile, err := os.Create(dst)
	defer outputFile.Close()
	if err != nil {
		return fmt.Errorf("couldn't open dest file: %s", err)
	}

	_, err = io.Copy(outputFile, inputFile)
	if err != nil {
		return fmt.Errorf("writing to output file failed: %s", err)
	}

	// Make sure to copy copy the permissions of the original file as well
	si, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.Chmod(dst, si.Mode()); err != nil {
		return err
	}

	// The copy was successful, so now delete the original file
	err = os.Remove(src)
	if err != nil {
		return fmt.Errorf("failed removing original file: %s", err)
	}
	return nil
}

func appendToFile(filename, content string) error {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.WriteString(content)
	return err
}

const InfoColor = "\033[1;32m%s\033[0m"

func printSuccess(text string, args ...interface{}) {
	text = fmt.Sprintf(text, args...)
	fmt.Printf(InfoColor+"\n", text)
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
	setApiPackages()
	// We run everything sequentially and not in parallel to prevent issues with real test databases
	args := append([]string{"test", Goflags[0], "-p", "1", "-coverprofile", "cover.out", "-timeout", "20m"}, ApiPackages...)
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
	runAndStreamOutput("go", "test", Goflags[0], "-p", "1", "-timeout", "20m", PACKAGE+"/pkg/integrations")
}

type Check mg.Namespace

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

func checkGolangCiLintInstalled() {
	mg.Deps(initVars)
	if err := exec.Command("golangci-lint").Run(); err != nil && strings.Contains(err.Error(), "executable file not found") {
		fmt.Println("Please manually install golangci-lint by running")
		fmt.Println("curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.31.0")
		os.Exit(1)
	}
}

func (Check) Golangci() {
	checkGolangCiLintInstalled()
	runAndStreamOutput("golangci-lint", "run")
}

func (Check) GolangciFix() {
	checkGolangCiLintInstalled()
	runAndStreamOutput("golangci-lint", "run", "--fix")
}

// Runs golangci and the swagger test in parralel
func (Check) All() {
	mg.Deps(initVars)
	mg.Deps(
		Check.Golangci,
		Check.GotSwag,
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

// Builds a vikunja binary, ready to run
func (Build) Build() {
	mg.Deps(initVars)
	runAndStreamOutput("go", "build", Goflags[0], "-tags", Tags, "-ldflags", "-s -w "+Ldflags, "-o", Executable)
}

type Release mg.Namespace

// Runs all steps in the right order to create release packages for various platforms
func (Release) Release(ctx context.Context) error {
	mg.Deps(initVars)
	mg.Deps(Release.Dirs)

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
	return runXgo("darwin-10.15/*")
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
			runAndStreamOutput("chmod", "+x", path) // Make sure all binaries are executable. Sometimes the CI does weired things and they're not.
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

// Creates a debian repo structure
func (Release) Reprepro() {
	mg.Deps(setVersion, setBinLocation)
	runAndStreamOutput("reprepro_expect", "debian", "includedeb", "buster", RootPath+"/"+DIST+"/os-packages/"+Executable+"_"+strings.ReplaceAll(VersionNumber, "v0", "0")+"_amd64.deb")
}

// Creates deb, rpm and apk packages
func (Release) Packages() error {
	mg.Deps(initVars)
	var err error
	binpath := "nfpm"
	err = exec.Command(binpath).Run()
	if err != nil && strings.Contains(err.Error(), "executable file not found") {
		binpath = "/nfpm"
		err = exec.Command(binpath).Run()
	}
	if err != nil && strings.Contains(err.Error(), "executable file not found") {
		fmt.Println("Please manually install nfpm by running")
		fmt.Println("curl -sfL https://install.goreleaser.com/github.com/goreleaser/nfpm.sh | sh -s -- -b $(go env GOPATH)/bin")
		os.Exit(1)
	}

	// Because nfpm does not  support templating, we replace the values in the config file and restore it after running
	nfpmConfigPath := RootPath + "/nfpm.yaml"
	nfpmconfig, err := ioutil.ReadFile(nfpmConfigPath)
	if err != nil {
		return err
	}

	fixedConfig := strings.ReplaceAll(string(nfpmconfig), "<version>", VersionNumber)
	fixedConfig = strings.ReplaceAll(fixedConfig, "<binlocation>", BinLocation)
	if err := ioutil.WriteFile(nfpmConfigPath, []byte(fixedConfig), 0); err != nil {
		return err
	}

	releasePath := RootPath + "/" + DIST + "/os-packages/"
	if err := os.MkdirAll(releasePath, 0755); err != nil {
		return err
	}

	runAndStreamOutput(binpath, "pkg", "--packager", "deb", "--target", releasePath)
	runAndStreamOutput(binpath, "pkg", "--packager", "rpm", "--target", releasePath)
	runAndStreamOutput(binpath, "pkg", "--packager", "apk", "--target", releasePath)

	return ioutil.WriteFile(nfpmConfigPath, nfpmconfig, 0)
}

type Dev mg.Namespace

// Creates a new bare db migration skeleton in pkg/migration with the current date
func (Dev) MakeMigration() error {

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter the name of the struct: ")
	str, _ := reader.ReadString('\n')
	str = strings.Trim(str, "\n")

	date := time.Now().Format("20060102150405")

	migration := `// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2021 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package migration

import (
	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

type ` + str + date + ` struct {
}

func (` + str + date + `) TableName() string {
	return "` + str + `"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "` + date + `",
		Description: "",
		Migrate: func(tx *xorm.Engine) error {
			return tx.Sync2(` + str + date + `{})
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
`
	filename := "/pkg/migration/" + date + ".go"
	f, err := os.Create(RootPath + filename)
	defer f.Close()
	if err != nil {
		return err
	}

	if _, err := f.WriteString(migration); err != nil {
		return err
	}

	printSuccess("Migration has been created at %s!", filename)

	return nil
}

// Create a new event. Takes the name of the event as the first argument and the module where the event should be created as the second argument. Events will be appended to the pkg/<module>/events.go file.
func (Dev) MakeEvent(name, module string) error {

	name = strcase.ToCamel(name)

	if !strings.HasSuffix(name, "Event") {
		name += "Event"
	}

	eventName := strings.ReplaceAll(strcase.ToDelimited(name, '.'), ".event", "")

	newEventCode := `
// ` + name + ` represents a ` + name + ` event
type ` + name + ` struct {
}

// Name defines the name for ` + name + `
func (t *` + name + `) Name() string {
	return "` + eventName + `"
}
`
	filename := "./pkg/" + module + "/events.go"
	if err := appendToFile(filename, newEventCode); err != nil {
		return err
	}

	printSuccess("The new event has been created successfully! Head over to %s and adjust its content.", filename)

	return nil
}

// Create a new listener for an event. Takes the name of the listener, the name of the event to listen to and the module where everything should be placed as parameters.
func (Dev) MakeListener(name, event, module string) error {
	name = strcase.ToCamel(name)
	listenerName := strcase.ToDelimited(name, '.')
	listenerCode := `
// ` + name + `  represents a listener
type ` + name + ` struct {
}

// Name defines the name for the ` + name + ` listener
func (s *` + name + `) Name() string {
	return "` + listenerName + `"
}

// Handle is executed when the event ` + name + ` listens on is fired
func (s *` + name + `) Handle(msg *message.Message) (err error) {
	event := &` + event + `{}
	err = json.Unmarshal(msg.Payload, event)
	if err != nil {
		return err
	}

	return nil
}
`
	filename := "./pkg/" + module + "/listeners.go"

	//////
	// Register the listener

	file, err := os.Open(filename)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(file)
	var idx int64 = 0
	for scanner.Scan() {
		if scanner.Text() == "}" {
			//idx -= int64(len(scanner.Text()))
			break
		}
		idx += int64(len(scanner.Bytes()) + 1)
	}
	file.Close()

	registerListenerCode := `	events.RegisterListener((&` + event + `{}).Name(), &` + name + `{})
`

	f, err := os.OpenFile(filename, os.O_RDWR, 0600)
	if err != nil {
		return err
	}

	defer f.Close()

	if _, err := f.Seek(idx, 0); err != nil {
		return err
	}
	remainder, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	f.Seek(idx, 0)
	f.Write([]byte(registerListenerCode))
	f.Write(remainder)

	///////
	// Append the listener code
	if err := appendToFile(filename, listenerCode); err != nil {
		return err
	}

	printSuccess("The new listener has been created successfully! Head over to %s and adjust its content.", filename)

	return nil
}

// Create a new notification. Takes the name of the notification as the first argument and the module where the notification should be created as the second argument. Notifications will be appended to the pkg/<module>/notifications.go file.
func (Dev) MakeNotification(name, module string) error {

	name = strcase.ToCamel(name)

	if !strings.HasSuffix(name, "Notification") {
		name += "Notification"
	}

	notficationName := strings.ReplaceAll(strcase.ToDelimited(name, '.'), ".notification", "")

	newNotificationCode := `
// ` + name + ` represents a ` + name + ` notification
type ` + name + ` struct {
}

// ToMail returns the mail notification for ` + name + `
func (n *` + name + `) ToMail() *notifications.Mail {
	return notifications.NewMail().
		Subject("").
		Greeting("Hi ").
		Line("").
		Action("", "")
}

// ToDB returns the ` + name + ` notification in a format which can be saved in the db
func (n *` + name + `) ToDB() interface{} {
	return nil
}

// Name returns the name of the notification
func (n *` + name + `) Name() string {
	return "` + notficationName + `"
}

`
	filename := "./pkg/" + module + "/notifications.go"
	if err := appendToFile(filename, newNotificationCode); err != nil {
		return err
	}

	printSuccess("The new notification has been created successfully! Head over to %s and adjust its content.", filename)

	return nil
}

type configOption struct {
	key          string
	description  string
	defaultValue string

	children []*configOption
}

func parseYamlConfigNode(node *yaml.Node) (config *configOption) {
	config = &configOption{
		key:         node.Value,
		description: strings.ReplaceAll(node.HeadComment, "# ", ""),
	}

	valMap := make(map[string]*configOption)

	var lastOption *configOption

	for i, n2 := range node.Content {
		coo := &configOption{
			key:         n2.Value,
			description: strings.ReplaceAll(n2.HeadComment, "# ", ""),
		}

		// If there's a key in valMap for the current key we should use that to append etc
		// Else we just create a new configobject
		co, exists := valMap[n2.Value]
		if exists {
			co.description = coo.description
		} else {
			valMap[n2.Value] = coo
			config.children = append(config.children, coo)
		}

		//		fmt.Println(i, coo.key, coo.description, n2.Value)

		if i%2 == 0 {
			lastOption = coo
			continue
		} else {
			lastOption.defaultValue = n2.Value
		}

		if i-1 >= 0 && i-1 <= len(node.Content) && node.Content[i-1].Value != "" {
			coo.defaultValue = n2.Value
			coo.key = node.Content[i-1].Value
		}

		if len(n2.Content) > 0 {
			for _, n := range n2.Content {
				coo.children = append(coo.children, parseYamlConfigNode(n))
			}
		}
	}

	return config
}

func printConfig(config []*configOption, level int, parent string) (rendered string) {

	// Keep track of what we already printed to prevent printing things twice
	printed := make(map[string]bool)

	for _, option := range config {

		// FIXME: Not a good solution. Ideally this would work without the level check, but since generating config
		// for more than two levels is currently broken anyway, I'll fix this after moving the config generation
		// to a better format than yaml.
		if level == 0 && option.key != "" {
			parent = option.key
		}

		if option.key != "" {

			// Filter out all config objects where the default value == key
			// Yaml is weired: It gives you a slice with an entry each for the key and their value.
			if printed[option.key] {
				continue
			}

			if level == 0 {
				rendered += "---\n\n"
			}

			rendered += "#"
			for i := 0; i <= level; i++ {
				rendered += "#"
			}
			rendered += " " + option.key + "\n\n"

			if option.description != "" {
				rendered += option.description + "\n\n"
			}

			// Top level config values never have a default value
			if level > 0 {
				rendered += "Default: `" + option.defaultValue
				if option.defaultValue == "" {
					rendered += "<empty>"
				}
				rendered += "`\n\n"

				fullPath := parent + "." + option.key

				rendered += "Full path: `" + fullPath + "`\n\n"
				rendered += "Environment path: `VIKUNJA_" + strcase.ToScreamingSnake(fullPath) + "`\n\n"
			}
		}

		printed[option.key] = true
		rendered += "\n" + printConfig(option.children, level+1, parent)
	}

	return
}

const (
	configDocPath       = `docs/content/doc/setup/config.md`
	configInjectComment = `<!-- Generated config will be injected here -->`
)

// Generates the config docs from a commented config.yml.sample file in the repo root.
func GenerateDocs() error {

	config, err := ioutil.ReadFile("config.yml.sample")
	if err != nil {
		return err
	}

	var d yaml.Node
	err = yaml.Unmarshal(config, &d)
	if err != nil {
		return err
	}

	conf := []*configOption{}

	for _, node := range d.Content {
		for _, n := range node.Content {
			co := parseYamlConfigNode(n)
			conf = append(conf, co)
		}
	}

	renderedConfig := printConfig(conf, 0, "")

	// Rebuild the config
	file, err := os.OpenFile(configDocPath, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	defer file.Close()

	// We read the config doc up until the marker, then stop and append our generated config
	fullConfig := ""

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		t := scanner.Text()
		fullConfig += t + "\n"

		if t == configInjectComment {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	fullConfig += "\n" + renderedConfig

	// We write the full file to prevent old content leftovers at the end
	// I know, there are probably better ways to do this.
	if err := ioutil.WriteFile(configDocPath, []byte(fullConfig), 0); err != nil {
		return err
	}

	return nil
}
