// veans — a beans-shaped CLI for Vikunja.
package main

import (
	"os"

	"code.vikunja.io/veans/internal/commands"
)

// version is overwritten via -ldflags at release time.
var version = "dev"

func main() {
	os.Exit(commands.Execute(version))
}
