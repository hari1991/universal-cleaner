package core

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// TrashDir returns the operating-system trash folder used by MoveToTrash.
func TrashDir() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, ".Trash"), nil
	case "windows":
		// Windows doesn't have a simple trash folder; the Recycle Bin is a
		// virtual folder managed by the shell. We return empty to signal that
		// the caller should use the PowerShell/SHFileOperation path or fall
		// back to permanent deletion.
		return "", nil
	default: // linux and other unix (FreeDesktop.org Trash spec)
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, ".local", "share", "Trash", "files"), nil
	}
}

// trashInfoDir returns the directory for .trashinfo sidecar files (Linux only).
func trashInfoDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share", "Trash", "info"), nil
}

// MoveToTrash moves path into the OS trash folder. If no trash folder is
// available (e.g. Windows) it falls back to a permanent removal (returned via
// fallback=true). On Linux it also writes the .trashinfo sidecar so the DE
// recognizes the item as trashable/restoreable.
func MoveToTrash(path string) (fallback bool, err error) {
	trash, err := TrashDir()
	if err != nil || trash == "" {
		return true, os.RemoveAll(path)
	}
	if mkErr := os.MkdirAll(trash, 0o755); mkErr != nil {
		return true, os.RemoveAll(path)
	}

	base := filepath.Base(path)
	dst := filepath.Join(trash, base)
	if _, statErr := os.Stat(dst); statErr == nil {
		// Name collision: append a timestamp to avoid overwriting.
		dst = filepath.Join(trash, fmt.Sprintf("%s %s", base, time.Now().Format("2006-01-02 15-04-05")))
	}

	// On Linux, write the .trashinfo sidecar before moving.
	if runtime.GOOS != "darwin" && runtime.GOOS != "windows" {
		if infoErr := writeTrashInfo(base, path); infoErr != nil {
			// Non-fatal: continue without sidecar.
			_ = infoErr
		}
	}

	if renameErr := os.Rename(path, dst); renameErr != nil {
		// Rename across volumes can fail; fall back to removal.
		return true, os.RemoveAll(path)
	}
	return false, nil
}

// writeTrashInfo writes a FreeDesktop.org .trashinfo sidecar file so the DE
// can restore the item from trash.
func writeTrashInfo(base, srcPath string) error {
	infoDir, err := trashInfoDir()
	if err != nil {
		return err
	}
	if mkErr := os.MkdirAll(infoDir, 0o755); mkErr != nil {
		return mkErr
	}
	absSrc, err := filepath.Abs(srcPath)
	if err != nil {
		absSrc = srcPath
	}
	// Escape the path as a file URI per the trash spec.
	uri := &url.URL{Scheme: "file", Path: absSrc}
	infoPath := filepath.Join(infoDir, base+".trashinfo")
	content := fmt.Sprintf("[Trash Info]\nPath=%s\nDeletionDate=%s\n",
		uri.String(), time.Now().Format("2006-01-02T15:04:05"))
	return os.WriteFile(infoPath, []byte(content), 0o644)
}

// Delete removes an item either via trash or permanently based on useTrash.
func Delete(item CleanableItem, useTrash bool) error {
	if useTrash {
		_, err := MoveToTrash(item.Path)
		return err
	}
	return os.RemoveAll(item.Path)
}
