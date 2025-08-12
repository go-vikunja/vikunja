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
// +build mage

package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/magefile/mage/mg"
	"golang.org/x/sync/errgroup"
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
		"build":                       Build.Build,
		"check:got-swag":              Check.GotSwag,
		"release":                     Release.Release,
		"release:os-package":          Release.OsPackage,
		"release:prepare-nfpm-config": Release.PrepareNFPMConfig,
		"dev:make-migration":          Dev.MakeMigration,
		"dev:make-event":              Dev.MakeEvent,
		"dev:make-listener":           Dev.MakeListener,
		"dev:make-notification":       Dev.MakeNotification,
		"plugins:build":               Plugins.Build,
		"lint":                        Check.Golangci,
		"lint:fix":                    Check.GolangciFix,
		"generate:config-yaml":        Generate.ConfigYAML,
		"generate:swagger-docs":       Generate.SwaggerDocs,
	}
)

func runCmdWithOutput(name string, arg ...string) (output []byte, err error) {
	cmd := exec.Command(name, arg...)
	output, err = cmd.Output()
	if err != nil {
		if ee, is := err.(*exec.ExitError); is {
			return nil, fmt.Errorf("error running command: %s, %s", string(ee.Stderr), err)
		}
		return nil, fmt.Errorf("error running command: %s", err)
	}

	return output, nil
}

func getRawVersionString() (version string, err error) {
	version, err = getRawVersionNumber()
	if err != nil {
		return
	}

	if version == "main" {
		version = "unstable"
	}

	if version != "" && version != "unstable" {
		return
	}

	return
}

func getRawVersionNumber() (version string, err error) {
	versionEnv := os.Getenv("RELEASE_VERSION")
	if versionEnv != "" {
		return versionEnv, nil
	}

	if os.Getenv("DRONE_TAG") != "" {
		return os.Getenv("DRONE_TAG"), nil
	}

	if os.Getenv("DRONE_BRANCH") != "" {
		return strings.Replace(os.Getenv("DRONE_BRANCH"), "release/v", "", 1), nil
	}

	versionBytes, err := runCmdWithOutput("git", "describe", "--tags", "--always", "--abbrev=10")
	return string(versionBytes), err
}

func setVersion() {
	versionNumber, err := getRawVersionNumber()
	VersionNumber = strings.Trim(versionNumber, "\n")
	VersionNumber = strings.Replace(VersionNumber, "-g", "-", 1)

	version, err := getRawVersionString()
	if err != nil {
		fmt.Printf("Error getting version: %s\n", err)
		os.Exit(1)
	}
	Version = version
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
		if strings.Contains(p, "code.vikunja.io/api") && !strings.Contains(p, "code.vikunja.io/api/pkg/webtests") {
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
	files, err := runCmdWithOutput("find", "./pkg", "-name", "*.go", "-type", "f", "!", "-path", "*/bindata.go")
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
	err := c.Start()
	if err != nil {
		fmt.Printf("Could not start: %s\n", err)
		os.Exit(1)
	}

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
// Because of this limitation, it fails in drone.
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

type Test mg.Namespace

// Runs the feature tests
func (Test) Feature() {
	mg.Deps(initVars)
	setApiPackages()
	// We run everything sequentially and not in parallel to prevent issues with real test databases
	args := append([]string{"test", Goflags[0], "-p", "1", "-coverprofile", "cover.out", "-timeout", "45m"}, ApiPackages...)
	runAndStreamOutput("go", args...)
}

// Runs the tests and builds the coverage html file from coverage output
func (Test) Coverage() {
	mg.Deps(initVars)
	mg.Deps(Test.Feature)
	runAndStreamOutput("go", "tool", "cover", "-html=cover.out", "-o", "cover.html")
}

// Runs the web tests
func (Test) Web() {
	mg.Deps(initVars)
	// We run everything sequentially and not in parallel to prevent issues with real test databases
	args := []string{"test", Goflags[0], "-p", "1", "-timeout", "45m", PACKAGE + "/pkg/webtests"}
	runAndStreamOutput("go", args...)
}

func (Test) Filter(filter string) {
	mg.Deps(initVars)
	setApiPackages()
	// We run everything sequentially and not in parallel to prevent issues with real test databases
	args := append([]string{"test", Goflags[0], "-p", "1", "-timeout", "45m", "-run", filter}, ApiPackages...)
	runAndStreamOutput("go", args...)
}

func (Test) All() {
	mg.Deps(initVars)
	mg.Deps(Test.Feature, Test.Web)
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

	(Generate{}).SwaggerDocs()

	newHash, err := calculateSha256FileHash(RootPath + "/pkg/swagger/swagger.json")
	if err != nil {
		fmt.Printf("Error getting new hash of the swagger docs: %s", err)
		os.Exit(1)
	}

	if oldHash != newHash {
		fmt.Println("Swagger docs are not up to date.")
		fmt.Println("Please run 'mage generate:swagger-docs' and commit the result.")
		os.Exit(1)
	}
}

// Checks if all translation keys used in the code exist in the English translation file
func (Check) Translations() {
	mg.Deps(initVars)
	fmt.Println("Checking for missing translation keys...")

	// Load translations from the English translation file
	translationFile := RootPath + "/pkg/i18n/lang/en.json"
	translations, err := loadTranslations(translationFile)
	if err != nil {
		fmt.Printf("Error loading translations: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Loaded %d translation keys from %s\n", len(translations), translationFile)

	// Extract keys from codebase
	keys, err := walkCodebaseForTranslationKeys(RootPath)
	if err != nil {
		fmt.Printf("Error walking codebase: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Found %d translation keys in the codebase\n", len(keys))

	// Check for missing keys
	missingKeys := make(map[string][]TranslationKey)
	for _, key := range keys {
		if !translations[key.Key] {
			missingKeys[key.Key] = append(missingKeys[key.Key], key)
		}
	}

	// Print results
	if len(missingKeys) > 0 {
		fmt.Printf("\nFound %d missing translation keys:\n", len(missingKeys))
		for key, occurrences := range missingKeys {
			fmt.Printf("\nKey: %s\n", key)
			for _, occurrence := range occurrences {
				fmt.Printf("  - %s:%d\n", occurrence.FilePath, occurrence.Line)
			}
		}
		os.Exit(1)
	} else {
		printSuccess("All translation keys are present in the translation file!")
	}
}

// TranslationKey represents a translation key found in the code
type TranslationKey struct {
	Key      string
	FilePath string
	Line     int
}

// loadTranslations loads the English translation file and returns a flattened map
func loadTranslations(filePath string) (map[string]bool, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading translation file: %v", err)
	}

	var translationsMap map[string]interface{}
	if err := json.Unmarshal(data, &translationsMap); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %v", err)
	}

	// Flatten the nested structure
	flattenedMap := make(map[string]bool)
	flattenTranslations("", translationsMap, flattenedMap)

	return flattenedMap, nil
}

// flattenTranslations recursively flattens a nested map structure into a flat map with dot-separated keys
func flattenTranslations(prefix string, src map[string]interface{}, dest map[string]bool) {
	for k, v := range src {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}

		switch vv := v.(type) {
		case string:
			dest[key] = true
		case map[string]interface{}:
			flattenTranslations(key, vv, dest)
		}
	}
}

// walkCodebaseForTranslationKeys walks the codebase and extracts all translation keys
func walkCodebaseForTranslationKeys(rootDir string) ([]TranslationKey, error) {
	var allKeys []TranslationKey

	pkgDir := filepath.Join(rootDir, "pkg")

	err := filepath.Walk(pkgDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories (starting with .)
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") {
			return filepath.SkipDir
		}

		// Only process Go files
		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			keys, err := extractTranslationKeysFromFile(path)
			if err != nil {
				fmt.Printf("Warning: %v\n", err)
				return nil
			}
			allKeys = append(allKeys, keys...)
		}

		return nil
	})

	return allKeys, err
}

// extractTranslationKeysFromFile extracts all i18n.T calls from a file
func extractTranslationKeysFromFile(filePath string) ([]TranslationKey, error) {
	// Read the file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file %s: %v", filePath, err)
	}

	var keys []TranslationKey

	// Regex to match i18n.T calls
	re := regexp.MustCompile(`i18n\.(T)\([^,]+,\s*"([^"]+)"`)
	matches := re.FindAllSubmatchIndex(content, -1)

	for _, match := range matches {
		if len(match) >= 4 {
			// Extract the key from the match
			keyStart, keyEnd := match[4], match[5]
			key := string(content[keyStart:keyEnd])

			// Count lines to determine the line number
			beforeMatch := content[:keyStart]
			lineCount := bytes.Count(beforeMatch, []byte("\n")) + 1

			keys = append(keys, TranslationKey{
				Key:      key,
				FilePath: filePath,
				Line:     lineCount,
			})
		}
	}

	return keys, nil
}

func checkGolangCiLintInstalled() {
	mg.Deps(initVars)
	if err := exec.Command("golangci-lint").Run(); err != nil && strings.Contains(err.Error(), "executable file not found") {
		fmt.Println("Please manually install golangci-lint by running")
		fmt.Println("curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.56.2")
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

// Runs golangci and the swagger test in parallel
func (Check) All() {
	mg.Deps(initVars)
	mg.Deps(
		Check.Golangci,
		Check.GotSwag,
		Check.Translations,
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
	// Check if the frontend dist folder exists
	distPath := filepath.Join(RootPath, "frontend", "dist")
	if _, err := os.Stat(distPath); os.IsNotExist(err) {
		if err := os.MkdirAll(distPath, 0o755); err != nil {
			fmt.Printf("Error creating %s: %s\n", distPath, err)
			os.Exit(1)
		}
	}

	indexFile := filepath.Join(distPath, "index.html")
	if _, err := os.Stat(indexFile); os.IsNotExist(err) {
		f, err := os.Create(indexFile)
		if err != nil {
			fmt.Printf("Error creating %s: %s\n", indexFile, err)
			os.Exit(1)
		}
		f.Close()
		fmt.Printf("Warning: %s not found, created empty file\n", indexFile)
	}

	runAndStreamOutput("go", "build", Goflags[0], "-tags", Tags, "-ldflags", "-s -w "+Ldflags, "-o", Executable)
}

func (Build) SaveVersionToFile() error {
	// Open the file for writing. If the file doesn't exist, create it.
	// If it exists, truncate it.
	file, err := os.Create("VERSION")
	if err != nil {
		return fmt.Errorf("error creating VERSION file: %w", err)
	}
	defer file.Close()

	// Write the version number to the file
	_, err = file.WriteString(VersionNumber)
	if err != nil {
		return fmt.Errorf("error writing to VERSION file: %w", err)
	}

	fmt.Println("Version number saved successfully to VERSION file")

	return nil
}

type Release mg.Namespace

// Runs all steps in the right order to create release packages for various platforms
func (Release) Release(ctx context.Context) error {
	mg.Deps(initVars)
	mg.Deps(Release.Dirs, prepareXgo)

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

func prepareXgo() {
	mg.Deps(initVars)
	checkAndInstallGoTool("xgo", "src.techknowlogick.com/xgo")

	fmt.Println("Pulling latest xgo docker image...")
	runAndStreamOutput("docker", "pull", "ghcr.io/techknowlogick/xgo:latest")
}

func runXgo(targets string) error {
	mg.Deps(initVars)
	checkAndInstallGoTool("xgo", "src.techknowlogick.com/xgo")

	extraLdflags := `-linkmode external -extldflags "-static" `

	// See https://github.com/techknowlogick/xgo/issues/79
	if strings.HasPrefix(targets, "darwin") {
		extraLdflags = ""
	}
	outName := os.Getenv("XGO_OUT_NAME")
	if outName == "" {
		outName = Executable + "-" + Version
	}

	runAndStreamOutput("xgo",
		"-dest", RootPath+"/"+DIST+"/binaries",
		"-tags", "netgo "+Tags,
		"-ldflags", extraLdflags+Ldflags,
		"-targets", targets,
		"-out", outName,
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
	return runXgo(strings.Join(targets, ","))
}

// Builds binaries for darwin
func (Release) Darwin() error {
	return runXgo("darwin-10.15/*")
}

func (Release) Xgo(target string) error {
	parts := strings.Split(target, "/")
	if len(parts) < 2 {
		return fmt.Errorf("invalid target")
	}

	variant := ""
	if len(parts) > 2 && parts[2] != "" {
		variant = "-" + strings.ReplaceAll(parts[2], "v", "")
	}

	return runXgo(parts[0] + "/" + parts[1] + variant)
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
		if strings.Contains(info.Name(), "mips") ||
			strings.Contains(info.Name(), "s390x") ||
			strings.Contains(info.Name(), "riscv64") ||
			strings.Contains(info.Name(), "darwin") {
			// not supported by upx
			return nil
		}

		// Runs compressing in parallel since upx is single-threaded
		errs.Go(func() error {
			runAndStreamOutput("chmod", "+x", path) // Make sure all binaries are executable. Sometimes the CI does weird things and they're not.
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

	generateConfigYAMLFromJSON(RootPath+"/"+DefaultConfigYAMLSamplePath, true)

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
		if err := copyFile(RootPath+"/"+DefaultConfigYAMLSamplePath, folder+DefaultConfigYAMLSamplePath); err != nil {
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

		c := exec.Command("zip", "-r", RootPath+"/"+DIST+"/zip/"+info.Name()+".zip", ".", "-i", "*")
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

// Prepares the nfpm config
func (Release) PrepareNFPMConfig() error {
	mg.Deps(initVars)
	var err error

	// Because nfpm does not support templating, we replace the values in the config file and restore it after running
	nfpmConfigPath := RootPath + "/nfpm.yaml"
	nfpmconfig, err := os.ReadFile(nfpmConfigPath)
	if err != nil {
		return err
	}

	fixedConfig := strings.ReplaceAll(string(nfpmconfig), "<version>", VersionNumber)
	fixedConfig = strings.ReplaceAll(fixedConfig, "<binlocation>", BinLocation)
	if err := os.WriteFile(nfpmConfigPath, []byte(fixedConfig), 0); err != nil {
		return err
	}

	generateConfigYAMLFromJSON(DefaultConfigYAMLSamplePath, true)

	return nil
}

// Creates deb, rpm and apk packages
func (Release) Packages() error {
	mg.Deps(initVars)

	var err error
	binpath := os.Getenv("NFPM_BIN_PATH")
	if binpath == "" {
		binpath = "nfpm"
	}
	err = exec.Command(binpath).Run()
	if err != nil && strings.Contains(err.Error(), "executable file not found") {
		binpath = "/usr/bin/nfpm"
		err = exec.Command(binpath).Run()
	}
	if err != nil && strings.Contains(err.Error(), "executable file not found") {
		fmt.Println("Please manually install nfpm by running")
		fmt.Println("curl -sfL https://install.goreleaser.com/github.com/goreleaser/nfpm.sh | sh -s -- -b $(go env GOPATH)/bin")
		os.Exit(1)
	}

	err = (Release{}).PrepareNFPMConfig()
	if err != nil {
		return err
	}

	releasePath := RootPath + "/" + DIST + "/os-packages/"
	if err := os.MkdirAll(releasePath, 0755); err != nil {
		return err
	}

	runAndStreamOutput(binpath, "pkg", "--packager", "deb", "--target", releasePath)
	runAndStreamOutput(binpath, "pkg", "--packager", "rpm", "--target", releasePath)
	runAndStreamOutput(binpath, "pkg", "--packager", "apk", "--target", releasePath)

	return nil
}

type Dev mg.Namespace

// MakeMigration creates a new bare db migration skeleton in pkg/migration.
// If you pass the struct name as an argument, the prompt will be skipped.
func (Dev) MakeMigration(name string) error {
	str := strings.TrimSpace(name)
	if str == "" {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter the name of the struct: ")
		s, _ := reader.ReadString('\n')
		str = strings.TrimSpace(s)
	}

	date := time.Now().Format("20060102150405")

	migration := `// Vikunja is a to-do list application to facilitate your life.
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
			return tx.Sync(` + str + date + `{})
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
	remainder, err := io.ReadAll(f)
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
func (n *` + name + `) ToMail(lang string) *notifications.Mail {
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

type Generate mg.Namespace

const DefaultConfigYAMLSamplePath = "config.yml.sample"

// Generates the swagger docs from the code annotations
func (Generate) SwaggerDocs() {
	mg.Deps(initVars)

	checkAndInstallGoTool("swag", "github.com/swaggo/swag/cmd/swag")
	runAndStreamOutput("swag", "init", "-g", "./pkg/routes/routes.go", "--parseDependency", "-d", RootPath, "-o", RootPath+"/pkg/swagger")
}

type ConfigNode struct {
	Key      string        `json:"key,omitempty"`
	Value    interface{}   `json:"default_value,omitempty"`
	Comment  string        `json:"comment,omitempty"`
	Children []*ConfigNode `json:"children,omitempty"`
}

func convertConfigJSONToYAML(node *ConfigNode, indent int, isTopLevel bool, parentKey string, commentOut bool) string {
	var result strings.Builder

	writeComment := func(comment string, indent int) {
		indent = int(math.Max(float64(indent), 0))
		if comment != "" {
			commentLines := strings.Split(comment, "\n")
			for _, line := range commentLines {
				result.WriteString(strings.Repeat("  ", indent))
				result.WriteString("# " + line + "\n")
			}
		}
	}

	writeLine := func(line string, indent int) {
		indent = int(math.Max(float64(indent), 0))
		if commentOut {
			result.WriteString(strings.Repeat("  ", indent) + "# " + line + "\n")
		} else {
			result.WriteString(strings.Repeat("  ", indent) + line + "\n")
		}
	}

	if isTopLevel {
		writeComment(node.Comment, indent)
	}

	if node.Key != "" {
		if !isTopLevel {
			writeComment(node.Comment, indent)
		}
		line := node.Key + ":"
		if node.Value != nil {
			value := node.Value
			if value == nil {
				value = node.Value
			}
			line += " " + formatValue(value)
		}
		writeLine(line, indent)
	}

	if len(node.Children) > 0 {
		isProviders := node.Key == "providers" && parentKey == "openid"
		isArray := len(node.Children) > 0 && node.Children[0].Key == ""
		for i, child := range node.Children {
			if isProviders {
				writeComment(child.Comment, indent+1)
				writeLine("-", indent+1)
				result.WriteString(convertConfigJSONToYAML(child, indent+1, false, node.Key, commentOut))
			} else if isArray {
				writeComment(child.Comment, indent+1)
				writeLine("- "+formatValue(child.Value), indent+1)
			} else {
				result.WriteString(convertConfigJSONToYAML(child, indent+1, false, node.Key, commentOut))
			}
			if i == len(node.Children)-1 && !isProviders && !isArray {
				writeLine("", indent)
			}
		}
	}

	return result.String()
}

func formatValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		if intValue, err := strconv.Atoi(v); err == nil {
			return fmt.Sprintf("%d", intValue)
		}
		if floatValue, err := strconv.ParseFloat(v, 64); err == nil {
			return fmt.Sprintf("%g", floatValue)
		}
		if boolValue, err := strconv.ParseBool(v); err == nil {
			return fmt.Sprintf("%v", boolValue)
		}
		return fmt.Sprintf("%q", v)
	case float64:
		return fmt.Sprintf("%g", v)
	case int:
		return fmt.Sprintf("%d", v)
	case bool:
		return fmt.Sprintf("%v", v)
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", v)
	}
}

func generateConfigYAMLFromJSON(yamlPath string, commented bool) {
	jsonData, err := os.ReadFile("config-raw.json")
	if err != nil {
		fmt.Println("Error reading JSON file:", err)
		return
	}

	var root ConfigNode
	err = json.Unmarshal(jsonData, &root)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return
	}

	yamlData := convertConfigJSONToYAML(&root, -1, true, "", commented)

	err = os.WriteFile(yamlPath, []byte(yamlData), 0644)
	if err != nil {
		fmt.Println("Error writing YAML file:", err)
		return
	}

	fmt.Println("Successfully generated " + yamlPath)
}

// Create a yaml config file from the config-raw.json definition
func (Generate) ConfigYAML(commented bool) {
	generateConfigYAMLFromJSON(DefaultConfigYAMLSamplePath, commented)
}

type Plugins mg.Namespace

// Build compiles a Go plugin at the provided path.
func (Plugins) Build(pathToSourceFiles string) error {
	mg.Deps(initVars)
	if pathToSourceFiles == "" {
		return fmt.Errorf("please provide a plugin path")
	}

	// Convert relative path to absolute path
	if !strings.HasPrefix(pathToSourceFiles, "/") {
		absPath, err := filepath.Abs(pathToSourceFiles)
		if err != nil {
			return fmt.Errorf("failed to resolve absolute path: %v", err)
		}
		pathToSourceFiles = absPath
	}

	out := filepath.Join(RootPath, "plugins", filepath.Base(pathToSourceFiles)+".so")
	runAndStreamOutput("go", "build", "-buildmode=plugin", "-o", out, pathToSourceFiles)
	return nil
}
