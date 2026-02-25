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

package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"math"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"syscall"
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
	Executable    = "vikunja"
	Ldflags       = ""
	Tags          = ""
	VersionNumber = "dev"
	Version       = "unstable" // This holds the built version, unstable by default, when building from a tag or release branch, their name
	BinLocation   = ""
	PkgVersion    = "unstable"

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
		"dev:prepare-worktree":        Dev.PrepareWorktree,
		"dev:tag-release":             Dev.TagRelease,
		"test:e2e":                    Test.E2E,
		"plugins:build":               Plugins.Build,
		"lint":                        Check.Golangci,
		"lint:fix":                    Check.GolangciFix,
		"generate:config-yaml":        Generate.ConfigYAML,
		"generate:swagger-docs":       Generate.SwaggerDocs,
	}
)

func goDetectVerboseFlag() string {
	return fmt.Sprintf("-v=%t", mg.Verbose())
}

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

func setVersion() error {
	versionNumber, err := getRawVersionNumber()
	if err != nil {
		return err
	}
	VersionNumber = strings.Trim(versionNumber, "\n")
	VersionNumber = strings.Replace(VersionNumber, "-g", "-", 1)

	version, err := getRawVersionString()
	if err != nil {
		return fmt.Errorf("error getting version: %w", err)
	}
	Version = version
	return nil
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

// Some variables can always get initialized, so we do just that.
func init() {
	setExecutable()
}

// Some variables have external dependencies (like git) which may not always be available.
func initVars() error {
	// Always include osusergo to use pure Go os/user implementation instead of CGO.
	// This prevents SIGFPE crashes when running under systemd without HOME set,
	// caused by glibc's getpwuid_r() failing in certain environments.
	// See: https://github.com/go-vikunja/vikunja/issues/2170
	Tags = "osusergo " + strings.ReplaceAll(os.Getenv("TAGS"), ",", " ")
	if err := setVersion(); err != nil {
		return err
	}
	setBinLocation()
	setPkgVersion()
	Ldflags = `-X "` + PACKAGE + `/pkg/version.Version=` + VersionNumber + `" -X "main.Tags=` + Tags + `"`
	return nil
}

func runAndStreamOutput(cmd string, args ...string) error {
	c := exec.Command(cmd, args...)

	c.Env = os.Environ()
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	fmt.Printf("%s\n\n", c.String())
	return c.Run()
}

// Will check if the tool exists and if not install it from the provided import path
// If any errors occur, it will exit with a status code of 1.
func checkAndInstallGoTool(tool, importPath string) {
	if err := exec.Command(tool).Run(); err != nil && strings.Contains(err.Error(), "executable file not found") {
		fmt.Printf("%s not installed, installing %s...\n", tool, importPath)
		if err := exec.Command("go", "install", goDetectVerboseFlag(), importPath).Run(); err != nil {
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
	if err != nil {
		return fmt.Errorf("couldn't open source file: %s", err)
	}
	defer inputFile.Close()

	outputFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("couldn't open dest file: %s", err)
	}
	defer outputFile.Close()

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
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o600)
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

// getE2EPort returns the port from the given env var, or a random available port.
func getE2EPort(envVar string) (int, error) {
	if v := os.Getenv(envVar); v != "" {
		return strconv.Atoi(v)
	}
	return getRandomPort()
}

// getRandomPort finds a random available TCP port.
func getRandomPort() (int, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

// setProcessGroup configures a command to run in its own process group,
// so that all child processes can be killed together.
func setProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

// killProcessGroup sends a signal to the entire process group of the given command.
func killProcessGroup(cmd *exec.Cmd) {
	if cmd.Process != nil {
		pgid, err := syscall.Getpgid(cmd.Process.Pid)
		if err == nil {
			syscall.Kill(-pgid, syscall.SIGTERM)
		}
		cmd.Wait()
	}
}

// waitForHTTP polls a URL until it returns a 200 status or the timeout expires.
func waitForHTTP(url string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	client := &http.Client{Timeout: 2 * time.Second}
	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("timed out waiting for %s after %s", url, timeout)
}

// Fmt formats the code using go fmt
func Fmt() error {
	mg.Deps(initVars)
	var goFiles []string
	err := filepath.Walk(".", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".go" {
			goFiles = append(goFiles, path)
		}
		return nil
	})
	if err != nil {
		return err
	}
	args := append([]string{"-s", "-w"}, goFiles...)
	return runAndStreamOutput("gofmt", args...)
}

type Test mg.Namespace

// Feature runs the feature tests
func (Test) Feature() error {
	mg.Deps(initVars)
	// We run everything sequentially and not in parallel to prevent issues with real test databases
	return runAndStreamOutput("go", "test", goDetectVerboseFlag(), "-p", "1", "-coverprofile", "cover.out", "-timeout", "45m", "-short", "./...")
}

// Coverage runs the tests and builds the coverage html file from coverage output
func (Test) Coverage() error {
	mg.Deps(initVars)
	mg.Deps(Test.Feature)
	return runAndStreamOutput("go", "tool", "cover", "-html=cover.out", "-o", "cover.html")
}

// Web runs the web tests
func (Test) Web() error {
	mg.Deps(initVars)
	// We run everything sequentially and not in parallel to prevent issues with real test databases
	args := []string{"test", goDetectVerboseFlag(), "-p", "1", "-timeout", "45m", "./pkg/webtests"}
	return runAndStreamOutput("go", args...)
}

func (Test) Filter(filter string) error {
	mg.Deps(initVars)
	// We run everything sequentially and not in parallel to prevent issues with real test databases
	return runAndStreamOutput("go", "test", goDetectVerboseFlag(), "-p", "1", "-timeout", "45m", "-run", filter, "-short", "./...")
}

func (Test) All() {
	mg.Deps(initVars)
	mg.Deps(Test.Feature, Test.Web)
}

// E2E builds the API, starts it with an in-memory database and the frontend dev server,
// runs the Playwright e2e tests against them, then tears everything down.
// This does not touch your local database.
//
// Any arguments are passed through to Playwright. Examples:
//
//	mage test:e2e ""                                     # run all tests
//	mage test:e2e "tests/e2e/misc/menu.spec.ts"         # run a specific test file
//	mage test:e2e "--grep menu"                          # filter by test name
//	mage test:e2e "--headed"                             # run in headed browser mode
//	mage test:e2e "--headed tests/e2e/misc/menu.spec.ts" # combine flags
//
// Environment variable overrides:
//   - VIKUNJA_E2E_API_PORT: API port (default: random)
//   - VIKUNJA_E2E_FRONTEND_PORT: Frontend port (default: random)
//   - VIKUNJA_E2E_TESTING_TOKEN: Testing token for seed endpoints (default: random)
//   - VIKUNJA_E2E_SKIP_BUILD: Set to "true" to skip rebuilding the API binary (default: false)
func (Test) E2E(args string) error {
	mg.Deps(initVars)

	// Determine ports
	apiPort, err := getE2EPort("VIKUNJA_E2E_API_PORT")
	if err != nil {
		return fmt.Errorf("could not get API port: %w", err)
	}
	frontendPort, err := getE2EPort("VIKUNJA_E2E_FRONTEND_PORT")
	if err != nil {
		return fmt.Errorf("could not get frontend port: %w", err)
	}

	// Generate a random testing token
	testingToken := os.Getenv("VIKUNJA_E2E_TESTING_TOKEN")
	if testingToken == "" {
		testingToken = fmt.Sprintf("e2e-test-token-%d", time.Now().UnixNano())
	}

	fmt.Printf("E2E test configuration:\n")
	fmt.Printf("  API port:      %d\n", apiPort)
	fmt.Printf("  Frontend port: %d\n", frontendPort)
	fmt.Printf("  Testing token: %s\n", testingToken)

	// Build the API binary (unless skipped)
	if os.Getenv("VIKUNJA_E2E_SKIP_BUILD") != "true" {
		fmt.Println("\n--- Building API binary ---")
		if err := (Build{}).Build(); err != nil {
			return fmt.Errorf("failed to build API: %w", err)
		}
	}

	// Create temp directory for file uploads and rootpath
	tmpDir, err := os.MkdirTemp("", "vikunja-e2e-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer func() {
		fmt.Println("\n--- Cleaning up temp directory ---")
		os.RemoveAll(tmpDir)
	}()

	if err := os.MkdirAll(filepath.Join(tmpDir, "files"), 0o755); err != nil {
		return fmt.Errorf("failed to create files dir: %w", err)
	}

	// Start the API server — all config via env vars, no config file
	// Uses in-memory SQLite (no DB file on disk)
	fmt.Println("\n--- Starting API server ---")
	apiCmd := exec.Command("./vikunja", "web")
	apiCmd.Env = append(os.Environ(),
		fmt.Sprintf("VIKUNJA_SERVICE_INTERFACE=:%d", apiPort),
		fmt.Sprintf("VIKUNJA_SERVICE_PUBLICURL=http://127.0.0.1:%d/", apiPort),
		fmt.Sprintf("VIKUNJA_SERVICE_TESTINGTOKEN=%s", testingToken),
		fmt.Sprintf("VIKUNJA_SERVICE_ROOTPATH=%s", tmpDir),
		"VIKUNJA_SERVICE_JWTSECRET=e2e-test-jwt-secret-do-not-use-in-production",
		"VIKUNJA_DATABASE_TYPE=sqlite",
		"VIKUNJA_DATABASE_PATH=memory",
		fmt.Sprintf("VIKUNJA_FILES_BASEPATH=%s", filepath.Join(tmpDir, "files")),
		"VIKUNJA_LOG_LEVEL=WARNING",
		"VIKUNJA_MAILER_ENABLED=false",
		"VIKUNJA_REDIS_ENABLED=false",
		"VIKUNJA_RATELIMIT_NOAUTHLIMIT=1000",
	)
	apiCmd.Stdout = os.Stdout
	apiCmd.Stderr = os.Stderr
	setProcessGroup(apiCmd)
	if err := apiCmd.Start(); err != nil {
		return fmt.Errorf("failed to start API: %w", err)
	}
	defer func() {
		fmt.Println("\n--- Stopping API server ---")
		killProcessGroup(apiCmd)
	}()

	// Wait for API to be ready
	apiBase := fmt.Sprintf("http://127.0.0.1:%d/api/v1", apiPort)
	fmt.Printf("Waiting for API at %s ...\n", apiBase)
	if err := waitForHTTP(apiBase+"/info", 30*time.Second); err != nil {
		return fmt.Errorf("API failed to start: %w", err)
	}
	printSuccess("API is ready!")

	// Build the frontend
	fmt.Println("\n--- Building frontend ---")
	buildFrontendCmd := exec.Command("pnpm", "build:dev")
	buildFrontendCmd.Dir = "frontend"
	buildFrontendCmd.Stdout = os.Stdout
	buildFrontendCmd.Stderr = os.Stderr
	if err := buildFrontendCmd.Run(); err != nil {
		return fmt.Errorf("failed to build frontend: %w", err)
	}
	printSuccess("Frontend built!")

	// Serve the built frontend with vite preview (static, no file watchers)
	fmt.Println("\n--- Starting frontend preview server ---")
	frontendCmd := exec.Command("pnpm", "preview:dev", "--port", strconv.Itoa(frontendPort))
	frontendCmd.Dir = "frontend"
	frontendCmd.Stdout = os.Stdout
	frontendCmd.Stderr = os.Stderr
	setProcessGroup(frontendCmd)
	if err := frontendCmd.Start(); err != nil {
		return fmt.Errorf("failed to start frontend: %w", err)
	}
	defer func() {
		fmt.Println("\n--- Stopping frontend preview server ---")
		killProcessGroup(frontendCmd)
	}()

	// Wait for frontend to be ready
	frontendBase := fmt.Sprintf("http://127.0.0.1:%d", frontendPort)
	fmt.Printf("Waiting for frontend at %s ...\n", frontendBase)
	if err := waitForHTTP(frontendBase, 60*time.Second); err != nil {
		return fmt.Errorf("frontend failed to start: %w", err)
	}
	printSuccess("Frontend is ready!")

	// Run Playwright tests
	fmt.Println("\n--- Running Playwright e2e tests ---")
	playwrightArgs := []string{"test:e2e"}
	if strings.TrimSpace(args) != "" {
		playwrightArgs = append(playwrightArgs, strings.Fields(args)...)
	}
	playwrightCmd := exec.Command("pnpm", playwrightArgs...)
	playwrightCmd.Dir = "frontend"
	playwrightCmd.Env = append(os.Environ(),
		fmt.Sprintf("API_URL=%s/", apiBase),
		fmt.Sprintf("BASE_URL=%s", frontendBase),
		fmt.Sprintf("VIKUNJA_SERVICE_TESTINGTOKEN=%s", testingToken),
		fmt.Sprintf("TEST_SECRET=%s", testingToken),
	)
	playwrightCmd.Stdout = os.Stdout
	playwrightCmd.Stderr = os.Stderr

	testErr := playwrightCmd.Run()

	if testErr != nil {
		return fmt.Errorf("e2e tests failed: %w", testErr)
	}

	printSuccess("All e2e tests passed!")
	return nil
}

type Check mg.Namespace

// GotSwag checks if the swagger docs need to be re-generated from the code annotations
func (Check) GotSwag() {
	mg.Deps(initVars)
	// The check is pretty cheaply done: We take the hash of the swagger.json file, generate the docs,
	// hash the file again and compare the two hashes to see if anything changed. If that's the case,
	// regenerating the docs is necessary.
	// swag is not capable of just outputting the generated docs to stdout, therefore we need to do it this way.
	// Another drawback of this is obviously it will only work once - we're not resetting the newly generated
	// docs after the check. This behaviour is good enough for ci though.
	oldHash, err := calculateSha256FileHash("./pkg/swagger/swagger.json")
	if err != nil {
		fmt.Printf("Error getting old hash of the swagger docs: %s", err)
		os.Exit(1)
	}

	(Generate{}).SwaggerDocs()

	newHash, err := calculateSha256FileHash("./pkg/swagger/swagger.json")
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

// Translations checks if all translation keys used in the code exist in the English translation file
func (Check) Translations() {
	mg.Deps(initVars)
	fmt.Println("Checking for missing translation keys...")

	// Load translations from the English translation file
	translationFile := "./pkg/i18n/lang/en.json"
	translations, err := loadTranslations(translationFile)
	if err != nil {
		fmt.Printf("Error loading translations: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Loaded %d translation keys from %s\n", len(translations), translationFile)

	// Extract keys from codebase
	keys, err := walkCodebaseForTranslationKeys(".")
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
		fmt.Println("curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.4.0")
		os.Exit(1)
	}
}

func (Check) Golangci() error {
	checkGolangCiLintInstalled()
	return runAndStreamOutput("golangci-lint", "run")
}

func (Check) GolangciFix() error {
	checkGolangCiLintInstalled()
	return runAndStreamOutput("golangci-lint", "run", "--fix")
}

// All runs golangci and the swagger test in parallel
func (Check) All() {
	mg.Deps(initVars)
	mg.Deps(
		Check.Golangci,
		Check.GotSwag,
		Check.Translations,
	)
}

type Build mg.Namespace

// Clean cleans all build, executable and bindata files
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

// Build builds a vikunja binary, ready to run
func (Build) Build() error {
	mg.Deps(initVars)
	// Check if the frontend dist folder exists
	distPath := filepath.Join("frontend", "dist")
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

	return runAndStreamOutput("go", "build", goDetectVerboseFlag(), "-tags", Tags, "-ldflags", "-s -w "+Ldflags, "-o", Executable)
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

// Release runs all steps in the right order to create release packages for various platforms
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

// Dirs creates all directories needed to release vikunja
func (Release) Dirs() error {
	for _, d := range []string{"binaries", "release", "zip"} {
		if err := os.MkdirAll("./"+DIST+"/"+d, 0o755); err != nil {
			return err
		}
	}
	return nil
}

func prepareXgo() error {
	mg.Deps(initVars)
	checkAndInstallGoTool("xgo", "src.techknowlogick.com/xgo")

	fmt.Println("Pulling latest xgo docker image...")
	return runAndStreamOutput("docker", "pull", "ghcr.io/techknowlogick/xgo:latest")
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

	if err := runAndStreamOutput("xgo",
		"-dest", "./"+DIST+"/binaries",
		"-tags", "netgo "+Tags,
		"-ldflags", extraLdflags+Ldflags,
		"-targets", targets,
		"-out", outName,
		"."); err != nil {
		return err
	}
	if os.Getenv("DRONE_WORKSPACE") != "" {
		return filepath.Walk("/build/", func(path string, info os.FileInfo, err error) error {
			// Skip directories
			if info.IsDir() {
				return nil
			}

			return moveFile(path, "./"+DIST+"/binaries/"+info.Name())
		})
	}
	return nil
}

// Windows builds binaries for windows
func (Release) Windows() error {
	return runXgo("windows/*")
}

// Linux builds binaries for linux
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

// Darwin builds binaries for darwin
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

// Compress compresses the built binaries in dist/binaries/ to reduce their filesize
func (Release) Compress(ctx context.Context) error {
	// $(foreach file,$(filter-out $(wildcard $(wildcard $(DIST)/binaries/$(EXECUTABLE)-*mips*)),$(wildcard $(DIST)/binaries/$(EXECUTABLE)-*)), upx -9 $(file);)

	errs, _ := errgroup.WithContext(ctx)

	filepath.Walk("./"+DIST+"/binaries/", func(path string, info os.FileInfo, err error) error {
		// Only executable files
		if !strings.Contains(info.Name(), Executable) {
			return nil
		}
		if strings.Contains(info.Name(), "mips") ||
			strings.Contains(info.Name(), "s390x") ||
			strings.Contains(info.Name(), "riscv64") ||
			strings.Contains(info.Name(), "darwin") ||
			(strings.Contains(info.Name(), "windows") && strings.Contains(info.Name(), "arm64")) {
			// not supported by upx
			return nil
		}

		// Runs compressing in parallel since upx is single-threaded
		errs.Go(func() error {
			if err := runAndStreamOutput("chmod", "+x", path); err != nil { // Make sure all binaries are executable. Sometimes the CI does weird things and they're not.
				return err
			}
			return runAndStreamOutput("upx", "-9", path)
		})

		return nil
	})

	return errs.Wait()
}

// Copy copies all built binaries to dist/release/ in preparation for creating the os packages
func (Release) Copy() error {
	return filepath.Walk("./"+DIST+"/binaries/", func(path string, info os.FileInfo, err error) error {
		// Only executable files
		if !strings.Contains(info.Name(), Executable) {
			return nil
		}

		return copyFile(path, "./"+DIST+"/release/"+info.Name())
	})
}

// Check creates sha256 checksum files for each binary in dist/release/
func (Release) Check() error {
	p := "./" + DIST + "/release/"
	return filepath.Walk(p, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
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

// OsPackage creates a folder for each
func (Release) OsPackage() error {
	p := "./" + DIST + "/release/"

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

	generateConfigYAMLFromJSON("./"+DefaultConfigYAMLSamplePath, true)

	for path, info := range bins {
		folder := p + info.Name() + "-full/"
		if err := os.Mkdir(folder, 0o755); err != nil {
			return err
		}
		if err := moveFile(p+info.Name()+".sha256", folder+info.Name()+".sha256"); err != nil {
			return err
		}
		if err := moveFile(path, folder+info.Name()); err != nil {
			return err
		}
		if err := copyFile("./"+DefaultConfigYAMLSamplePath, folder+DefaultConfigYAMLSamplePath); err != nil {
			return err
		}
		if err := copyFile("./LICENSE", folder+"LICENSE"); err != nil {
			return err
		}
	}
	return nil
}

// Zip creates a zip file from all os-package folders in dist/release
func (Release) Zip() error {
	rootDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("could not get working directory: %w", err)
	}

	p := "./" + DIST + "/release/"
	if err := filepath.Walk(p, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() || info.Name() == "release" {
			return nil
		}

		fmt.Printf("Zipping %s...\n", info.Name())

		zipFile := filepath.Join(rootDir, DIST, "zip", info.Name()+".zip")
		c := exec.Command("zip", "-r", zipFile, ".", "-i", "*")
		c.Dir = path
		out, err := c.Output()
		fmt.Print(string(out))
		return err
	}); err != nil {
		return err
	}

	return nil
}

// Reprepro creates a debian repo structure
func (Release) Reprepro() error {
	mg.Deps(setVersion, setBinLocation)
	return runAndStreamOutput("reprepro_expect", "debian", "includedeb", "buster", "./"+DIST+"/os-packages/"+Executable+"_"+strings.ReplaceAll(VersionNumber, "v0", "0")+"_amd64.deb")
}

// PrepareNFPMConfig prepares the nfpm config
func (Release) PrepareNFPMConfig() error {
	mg.Deps(initVars)
	var err error

	// Because nfpm does not support templating, we replace the values in the config file and restore it after running
	nfpmConfigPath := "./nfpm.yaml"
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

// Packages creates deb, rpm and apk packages
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

	releasePath := "./" + DIST + "/os-packages/"
	if err := os.MkdirAll(releasePath, 0o755); err != nil {
		return err
	}

	if err := runAndStreamOutput(binpath, "pkg", "--packager", "deb", "--target", releasePath); err != nil {
		return err
	}
	if err := runAndStreamOutput(binpath, "pkg", "--packager", "rpm", "--target", releasePath); err != nil {
		return err
	}
	if err := runAndStreamOutput(binpath, "pkg", "--packager", "apk", "--target", releasePath); err != nil {
		return err
	}

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
	filename := "./pkg/migration/" + date + ".go"
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.WriteString(migration); err != nil {
		return err
	}

	printSuccess("Migration has been created at %s!", filename)

	return nil
}

// MakeEvent create a new event. Takes the name of the event as the first argument and the module where the event should be created as the second argument. Events will be appended to the pkg/<module>/events.go file.
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

// MakeListener create a new listener for an event. Takes the name of the listener, the name of the event to listen to and the module where everything should be placed as parameters.
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
			// idx -= int64(len(scanner.Text()))
			break
		}
		idx += int64(len(scanner.Bytes()) + 1)
	}
	file.Close()

	registerListenerCode := `	events.RegisterListener((&` + event + `{}).Name(), &` + name + `{})
`

	f, err := os.OpenFile(filename, os.O_RDWR, 0o600)
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

// MakeNotification create a new notification. Takes the name of the notification as the first argument and the module where the notification should be created as the second argument. Notifications will be appended to the pkg/<module>/notifications.go file.
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

// SwaggerDocs generates the swagger docs from the code annotations
func (Generate) SwaggerDocs() error {
	mg.Deps(initVars)

	checkAndInstallGoTool("swag", "github.com/swaggo/swag/cmd/swag")
	return runAndStreamOutput("swag", "init", "-g", "./pkg/routes/routes.go", "--parseDependency", "-d", ".", "-o", "./pkg/swagger")
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

	err = os.WriteFile(yamlPath, []byte(yamlData), 0o644)
	if err != nil {
		fmt.Println("Error writing YAML file:", err)
		return
	}

	fmt.Println("Successfully generated " + yamlPath)
}

// ConfigYAML create a yaml config file from the config-raw.json definition
func (Generate) ConfigYAML(commented bool) {
	generateConfigYAMLFromJSON(DefaultConfigYAMLSamplePath, commented)
}

// PrepareWorktree creates a new git worktree for development.
// The first argument is the name, which becomes both the folder name and branch name.
// The second argument is a path to a plan file that will be copied to the new worktree (pass "" to skip).
// The worktree is created in the parent directory (../).
// It also copies the current config.yml with an updated rootpath, and initializes the frontend.
func (Dev) PrepareWorktree(name string, planPath string) error {
	if name == "" {
		return fmt.Errorf("name is required: mage dev:prepare-worktree <name> <plan-path>")
	}

	// Get the parent directory path
	worktreePath := filepath.Join("..", name)

	fmt.Printf("Creating worktree at %s with branch %s...\n", worktreePath, name)

	// Create the git worktree
	cmd := exec.Command("git", "worktree", "add", worktreePath, "-b", name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create worktree: %w", err)
	}
	printSuccess("Worktree created successfully!")

	// Copy and modify config.yml
	configSrc := "config.yml"
	configDst := filepath.Join(worktreePath, "config.yml")

	if _, err := os.Stat(configSrc); err == nil {
		configContent, err := os.ReadFile(configSrc)
		if err != nil {
			return fmt.Errorf("failed to read config.yml: %w", err)
		}

		// Replace the rootpath value
		re := regexp.MustCompile(`(?m)^(\s*rootpath:\s*)"[^"]*"`)
		newConfig := re.ReplaceAllString(string(configContent), `${1}"`+worktreePath+`"`)

		// Also handle unquoted rootpath values
		re2 := regexp.MustCompile(`(?m)^(\s*rootpath:\s*)(/[^\s\n]+)`)
		newConfig = re2.ReplaceAllString(newConfig, `${1}"`+worktreePath+`"`)

		if err := os.WriteFile(configDst, []byte(newConfig), 0o644); err != nil {
			return fmt.Errorf("failed to write config.yml: %w", err)
		}
		printSuccess("Config copied with updated rootpath!")
	} else {
		fmt.Println("Warning: config.yml not found, skipping config copy")
	}

	// Copy .claude/settings.local.json if it exists
	claudeSettingsSrc := filepath.Join(".claude", "settings.local.json")
	if _, err := os.Stat(claudeSettingsSrc); err == nil {
		claudeDir := filepath.Join(worktreePath, ".claude")
		if err := os.MkdirAll(claudeDir, 0o755); err != nil {
			return fmt.Errorf("failed to create .claude directory: %w", err)
		}
		claudeSettingsDst := filepath.Join(claudeDir, "settings.local.json")
		if err := copyFile(claudeSettingsSrc, claudeSettingsDst); err != nil {
			return fmt.Errorf("failed to copy .claude/settings.local.json: %w", err)
		}
		printSuccess("Claude settings copied!")
	}

	// Copy plan file if provided
	if planPath != "" {
		planPath = strings.TrimSpace(planPath)
		if planPath != "" {
			// Create plans directory in the new worktree
			plansDir := filepath.Join(worktreePath, "plans")
			if err := os.MkdirAll(plansDir, 0o755); err != nil {
				return fmt.Errorf("failed to create plans directory: %w", err)
			}

			// Determine source path (relative to current directory or absolute)
			srcPlanPath := planPath
			if !filepath.IsAbs(planPath) {
				srcPlanPath = planPath
			}

			if _, err := os.Stat(srcPlanPath); err != nil {
				return fmt.Errorf("plan file not found: %s", srcPlanPath)
			}

			dstPlanPath := filepath.Join(plansDir, filepath.Base(planPath))
			if err := copyFile(srcPlanPath, dstPlanPath); err != nil {
				return fmt.Errorf("failed to copy plan file: %w", err)
			}
			printSuccess("Plan file copied to %s!", dstPlanPath)
		}
	}

	// Initialize frontend
	fmt.Println("Initializing frontend...")
	frontendDir := filepath.Join(worktreePath, "frontend")

	// Run pnpm install
	pnpmCmd := exec.Command("pnpm", "i")
	pnpmCmd.Dir = frontendDir
	pnpmCmd.Stdout = os.Stdout
	pnpmCmd.Stderr = os.Stderr
	if err := pnpmCmd.Run(); err != nil {
		return fmt.Errorf("failed to run pnpm install: %w", err)
	}

	// Run patch-sass-embedded (shell alias from devenv)
	patchCmd := exec.Command("bash", "-ic", "patch-sass-embedded")
	patchCmd.Dir = frontendDir
	patchCmd.Stdout = os.Stdout
	patchCmd.Stderr = os.Stderr
	if err := patchCmd.Run(); err != nil {
		// patch-sass-embedded might not be critical, just warn
		fmt.Printf("Warning: patch-sass-embedded failed: %v\n", err)
	}

	printSuccess("Frontend initialized!")
	printSuccess("\nWorktree ready at: %s", worktreePath)
	printSuccess("Branch: %s", name)
	fmt.Println("\nTo start working:")
	fmt.Printf("  cd %s\n", worktreePath)

	return nil
}

// TagRelease creates a new release tag with changelog.
// It updates the version badge in README.md, generates changelog using git-cliff,
// commits the changes, and creates an annotated tag.
func (Dev) TagRelease(version string) error {
	if version == "" {
		return fmt.Errorf("version is required: mage dev:tag-release <version>")
	}

	// Ensure version starts with 'v'
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}

	fmt.Printf("Creating release %s...\n", version)

	// Get the last tag
	lastTagBytes, err := runCmdWithOutput("git", "describe", "--tags", "--abbrev=0")
	if err != nil {
		return fmt.Errorf("failed to get last tag: %w", err)
	}
	lastTag := strings.TrimSpace(string(lastTagBytes))
	fmt.Printf("Last tag: %s\n", lastTag)

	// Generate changelog using git cliff
	fmt.Println("Generating changelog...")
	changelogBytes, err := runCmdWithOutput("git", "cliff", lastTag+"..HEAD", "--tag", version)
	if err != nil {
		return fmt.Errorf("failed to generate changelog: %w", err)
	}
	changelog := string(changelogBytes)

	// Clean up the changelog
	changelog = cleanupChangelog(changelog)

	// Update README.md version badge
	fmt.Println("Updating README.md version badge...")
	if err := updateReadmeBadge(version); err != nil {
		return fmt.Errorf("failed to update README badge: %w", err)
	}

	// Prepend changelog to CHANGELOG.md
	fmt.Println("Updating CHANGELOG.md...")
	if err := prependChangelog(changelog); err != nil {
		return fmt.Errorf("failed to update CHANGELOG.md: %w", err)
	}

	// Commit the changes
	fmt.Println("Committing changes...")
	commitMsg := fmt.Sprintf("chore: %s release preparations", version)
	cmd := exec.Command("git", "add", "README.md", "CHANGELOG.md")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stage files: %w", err)
	}

	cmd = exec.Command("git", "commit", "-m", commitMsg)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	// Prepare tag message (remove markdown header formatting)
	tagMessage := prepareTagMessage(changelog)

	// Create the annotated tag
	fmt.Printf("Creating tag %s...\n", version)
	cmd = exec.Command("git", "tag", "-a", version, "-m", tagMessage)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create tag: %w", err)
	}

	printSuccess("Release %s created successfully!", version)
	fmt.Println("\nNext steps:")
	fmt.Println("  git push origin main")
	fmt.Printf("  git push origin %s\n", version)

	return nil
}

// cleanupChangelog cleans up the generated changelog by:
// - Removing duplicate lines
// - Fixing entries that span multiple lines
// - Ensuring each change is on a single line
func cleanupChangelog(changelog string) string {
	lines := strings.Split(changelog, "\n")
	var cleanedLines []string
	seenLines := make(map[string]bool)
	var currentEntry strings.Builder

	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Check if this is a new entry (starts with * or - or is a header)
		isNewEntry := strings.HasPrefix(trimmedLine, "* ") ||
			strings.HasPrefix(trimmedLine, "- ") ||
			strings.HasPrefix(trimmedLine, "## ") ||
			strings.HasPrefix(trimmedLine, "### ") ||
			trimmedLine == ""

		if isNewEntry {
			// Flush the current entry if any
			if currentEntry.Len() > 0 {
				entryStr := strings.TrimSpace(currentEntry.String())
				if !seenLines[entryStr] && entryStr != "" {
					cleanedLines = append(cleanedLines, entryStr)
					seenLines[entryStr] = true
				}
				currentEntry.Reset()
			}

			// Start a new entry or add empty line/header
			if trimmedLine == "" {
				// Only add empty line if the previous line wasn't empty
				if len(cleanedLines) > 0 && cleanedLines[len(cleanedLines)-1] != "" {
					cleanedLines = append(cleanedLines, "")
				}
			} else if strings.HasPrefix(trimmedLine, "## ") || strings.HasPrefix(trimmedLine, "### ") {
				// Headers are never duplicates
				cleanedLines = append(cleanedLines, trimmedLine)
			} else {
				currentEntry.WriteString(trimmedLine)
			}
		} else if currentEntry.Len() > 0 {
			// This is a continuation of the current entry
			currentEntry.WriteString(" ")
			currentEntry.WriteString(trimmedLine)
		} else if trimmedLine != "" {
			// Standalone line that's not part of an entry
			if !seenLines[trimmedLine] {
				cleanedLines = append(cleanedLines, trimmedLine)
				seenLines[trimmedLine] = true
			}
		}

		// Handle last line
		if i == len(lines)-1 && currentEntry.Len() > 0 {
			entryStr := strings.TrimSpace(currentEntry.String())
			if !seenLines[entryStr] && entryStr != "" {
				cleanedLines = append(cleanedLines, entryStr)
			}
		}
	}

	return strings.Join(cleanedLines, "\n")
}

// updateReadmeBadge updates the version badge in README.md
func updateReadmeBadge(version string) error {
	readmePath := "README.md"
	content, err := os.ReadFile(readmePath)
	if err != nil {
		return fmt.Errorf("failed to read README.md: %w", err)
	}

	// Convert version for badge (e.g., v1.0.0-rc3 -> v1.0.0rc3 for the badge display)
	badgeVersion := strings.ReplaceAll(version, "-", "")

	// Update the badge - match the pattern: download-vX.X.X...-brightgreen
	re := regexp.MustCompile(`(download-)(v[0-9a-zA-Z.]+)(-brightgreen)`)
	newContent := re.ReplaceAllString(string(content), "${1}"+badgeVersion+"${3}")

	if err := os.WriteFile(readmePath, []byte(newContent), 0o644); err != nil {
		return fmt.Errorf("failed to write README.md: %w", err)
	}

	return nil
}

// prependChangelog prepends the new changelog entries to CHANGELOG.md
func prependChangelog(newChangelog string) error {
	changelogPath := "CHANGELOG.md"
	existingContent, err := os.ReadFile(changelogPath)
	if err != nil {
		return fmt.Errorf("failed to read CHANGELOG.md: %w", err)
	}

	// Find where to insert the new changelog (after the header section)
	content := string(existingContent)
	headerEnd := strings.Index(content, "\n## ")
	if headerEnd == -1 {
		// No existing version sections, append at the end
		headerEnd = len(content)
	}

	// Build new content: header + new changelog + existing versions
	header := content[:headerEnd]
	existingVersions := ""
	if headerEnd < len(content) {
		existingVersions = content[headerEnd:]
	}

	// Ensure there's proper spacing
	newContent := strings.TrimRight(header, "\n") + "\n\n" +
		strings.TrimSpace(newChangelog) + "\n" +
		existingVersions

	if err := os.WriteFile(changelogPath, []byte(newContent), 0o644); err != nil {
		return fmt.Errorf("failed to write CHANGELOG.md: %w", err)
	}

	return nil
}

// prepareTagMessage removes markdown header formatting from the changelog for use as a tag message
func prepareTagMessage(changelog string) string {
	lines := strings.Split(changelog, "\n")
	var result []string

	for _, line := range lines {
		// Remove ## and ### prefixes
		if strings.HasPrefix(line, "### ") {
			result = append(result, strings.TrimPrefix(line, "### "))
		} else if strings.HasPrefix(line, "## ") {
			result = append(result, strings.TrimPrefix(line, "## "))
		} else {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
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

	out := filepath.Join("plugins", filepath.Base(pathToSourceFiles)+".so")
	return runAndStreamOutput("go", "build", "-buildmode=plugin", "-tags", Tags, "-o", out, pathToSourceFiles)
}
