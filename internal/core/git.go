package core

import (
	"os/exec"
	"path/filepath"
	"strings"
)

// GitTrackedFiles returns the set of paths that are tracked by git in the given
// repository root. If git is not available or the root is not a git repo, it
// returns nil and no error.
func GitTrackedFiles(root string) map[string]bool {
	// Check if root is inside a git repo.
	cmd := exec.Command("git", "-C", root, "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}
	repoRoot := strings.TrimSpace(string(output))
	if repoRoot == "" {
		return nil
	}

	// List all tracked files.
	cmd = exec.Command("git", "-C", repoRoot, "ls-files")
	output, err = cmd.Output()
	if err != nil {
		return nil
	}
	lines := strings.Split(string(output), "\n")
	tracked := make(map[string]bool, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			tracked[filepath.Join(repoRoot, line)] = true
		}
	}
	return tracked
}

// IsGitRepo reports whether the given directory is inside a git repository.
func IsGitRepo(root string) bool {
	cmd := exec.Command("git", "-C", root, "rev-parse", "--is-inside-work-tree")
	return cmd.Run() == nil
}
