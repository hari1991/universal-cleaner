package core

import (
	"os/exec"
	"runtime"
)

// RevealInFileManager opens the given path's parent directory in the OS file
// manager, selecting the item if possible.
func RevealInFileManager(path string) error {
	switch runtime.GOOS {
	case "darwin":
		// `open -R` reveals the file in Finder.
		return exec.Command("open", "-R", path).Start()
	case "windows":
		// `explorer /select,<path>` reveals the file in Explorer.
		return exec.Command("explorer", "/select,"+path).Start()
	default:
		// Linux: open the parent directory since xdg-open can't select.
		return exec.Command("xdg-open", parentDir(path)).Start()
	}
}

func parentDir(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			if i == 0 {
				return "/"
			}
			return path[:i]
		}
	}
	return "."
}
