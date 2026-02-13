package zettel

import (
	"os/exec"
	"path/filepath"
	"strings"
)

// GetGitContext returns the repo name if inside a git project, else empty.
func GetGitContext(cwd string) string {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = cwd
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	// e.g., /home/user/code/my-project -> my-project
	return filepath.Base(strings.TrimSpace(string(out)))
}
