package core

import (
	"fmt"
	"os"
	"path/filepath"
)

// FormatSize renders a byte count as a human-readable string.
func FormatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// CalculateSize returns the total size in bytes and file count of a file or
// directory tree, plus a formatted string representation of the size.
func CalculateSize(path string, isDir bool) (int64, int, string) {
	if !isDir {
		info, err := os.Stat(path)
		if err != nil {
			return 0, 0, FormatSize(0)
		}
		return info.Size(), 1, FormatSize(info.Size())
	}

	var total int64
	var count int
	_ = filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			total += info.Size()
			count++
		}
		return nil
	})
	return total, count, FormatSize(total)
}
