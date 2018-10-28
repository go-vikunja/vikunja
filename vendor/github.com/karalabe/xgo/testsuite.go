// Go CGO cross compiler
// Copyright (c) 2016 Péter Szilágyi. All rights reserved.
//
// Released under the MIT license.

// This is a manual test suite to run the cross compiler against various known
// projects, codebases and repositories to ensure at least a baseline guarantee
// that things work as they supposed to.
//
// Run as: go run testsuite.go

// +build ignore

package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

// layers defines all the docker layers needed for the final xgo image. The last
// one will be used to run the test suite against.
var layers = []struct {
	tag string
	dir string
}{
	{"karalabe/xgo-base", "base"},
	{"karalabe/xgo-1.6.2", "go-1.6.2"},
	{"karalabe/xgo-1.6.x", "go-1.6.x"},
	{"karalabe/xgo-latest", "go-latest"},
	//{"karalabe/xgo-latest-ios", "go-latest-ios"}, // Non-public layer (XCode licensing)
}

// tests defaines all the input test cases and associated arguments the cross
// compiler should be ran for and with which arguments.
var tests = []struct {
	path string
	args []string
}{
	// Tiny test cases to smoke test cross compilations
	{"github.com/karalabe/xgo/tests/embedded_c", nil},
	{"github.com/karalabe/xgo/tests/embedded_cpp", nil},

	// Baseline projects to ensure minimal requirements
	//{"github.com/project-iris/iris", nil}, // Deps failed, disable
	{"github.com/ethereum/go-ethereum/cmd/geth", []string{"--branch", "develop"}},

	// Third party projects using xgo, smoke test that they don't break
	{"github.com/rwcarlsen/cyan/cmd/cyan", nil},
	{"github.com/cockroachdb/cockroach", []string{"--targets", "darwin-10.11/amd64"}},
}

func main() {
	// Retrieve the current working directory to locate the dockerfiles
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to retrieve local working directory: %v", err)
	}
	if _, err := os.Stat(filepath.Join(pwd, "docker", "base")); err != nil {
		log.Fatalf("Failed to locate docker image: %v", err)
	}
	// Assemble the multi-layered xgo docker image
	for _, layer := range layers {
		cmd := exec.Command("docker", "build", "--tag", layer.tag, filepath.Join(pwd, "docker", layer.dir))

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			log.Fatalf("Failed to build xgo layer: %v", err)
		}
	}
	// Iterate over each of the test cases and run them
	for i, test := range tests {
		cmd := exec.Command("docker", append([]string{"run", "--entrypoint", "xgo", layers[len(layers)-1].tag, "-v"}, append(test.args, test.path)...)...)

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			log.Fatalf("Test #%d: cross compilation failed: %v", i, err)
		}
	}
}
