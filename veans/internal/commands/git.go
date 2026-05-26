package commands

import (
	"os/exec"
	"strings"
)

// runGit runs `git <args...>` in the current working directory and returns
// trimmed stdout. Errors are returned to the caller so they can decide
// whether silence or escalation is appropriate.
func runGit(args ...string) (string, error) {
	out, err := exec.Command("git", args...).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimRight(string(out), "\r\n"), nil
}
