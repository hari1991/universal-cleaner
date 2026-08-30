package core

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ExportCSV writes the given items to a CSV file at the specified path.
func ExportCSV(items []CleanableItem, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	if err := w.Write([]string{"Path", "Type", "Size", "Files", "Modified", "Risky"}); err != nil {
		return err
	}
	for _, it := range items {
		risky := "no"
		if it.Risky {
			risky = "yes"
		}
		if err := w.Write([]string{
			it.Path,
			it.Type,
			it.SizeStr,
			fmt.Sprintf("%d", it.FileCount),
			it.LastModified.Format("2006-01-02 15:04:05"),
			risky,
		}); err != nil {
			return err
		}
	}
	return nil
}

// ExportJSON writes the given items to a JSON file at the specified path.
func ExportJSON(items []CleanableItem, path string) error {
	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// ExportFilename returns a timestamped filename with the given extension.
func ExportFilename(ext string) string {
	return fmt.Sprintf("universal-cleaner-export.%s", strings.TrimPrefix(filepath.Ext(ext), "."))
}
