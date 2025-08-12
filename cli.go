package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CLI version of the universal cleaner
func runCLI() {
	var (
		scanPath = flag.String("path", ".", "Path to scan for cleanable items")
		dryRun   = flag.Bool("dry-run", false, "Show what would be deleted without actually deleting")
		autoYes  = flag.Bool("yes", false, "Automatically confirm all deletions")
		verbose  = flag.Bool("verbose", false, "Show detailed output")
	)
	flag.Parse()

	fmt.Println("Universal Cleaner CLI")
	fmt.Println("====================")

	// Scan for cleanable items
	fmt.Printf("Scanning directory: %s\n", *scanPath)
	items := findCleanableItemsCLI(*scanPath)

	if len(items) == 0 {
		fmt.Println("No cleanable items found.")
		return
	}

	fmt.Printf("\nFound %d cleanable items:\n", len(items))
	var totalSize int64

	for i, item := range items {
		fmt.Printf("%d. [%s] %s (%s)\n", i+1, item.Type, item.Path, item.SizeStr)
		totalSize += item.Size
		if *verbose {
			fmt.Printf("   Last modified: %s\n", item.LastModified.Format("2006-01-02 15:04:05"))
		}
	}

	fmt.Printf("\nTotal size: %s\n", formatSizeCLI(totalSize))

	if *dryRun {
		fmt.Println("\nDry run mode - no files will be deleted.")
		return
	}

	// Confirm deletion
	if !*autoYes {
		fmt.Print("\nDo you want to delete these items? (y/N): ")
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		if response != "y" && response != "yes" {
			fmt.Println("Operation cancelled.")
			return
		}
	}

	// Perform deletion
	fmt.Println("\nDeleting items...")
	deletedCount := 0
	var failedItems []string

	for _, item := range items {
		if *verbose {
			fmt.Printf("Deleting: %s\n", item.Path)
		}

		err := os.RemoveAll(item.Path)
		if err != nil {
			failedItems = append(failedItems, item.Path)
			if *verbose {
				fmt.Printf("Failed to delete %s: %v\n", item.Path, err)
			}
		} else {
			deletedCount++
		}
	}

	fmt.Printf("\nCompleted: %d items deleted", deletedCount)
	if len(failedItems) > 0 {
		fmt.Printf(", %d failed", len(failedItems))
		if *verbose {
			fmt.Println("\nFailed items:")
			for _, item := range failedItems {
				fmt.Printf("  - %s\n", item)
			}
		}
	}
	fmt.Println()
}

func findCleanableItemsCLI(rootPath string) []CleanableItem {
	var items []CleanableItem

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if path == rootPath {
			return nil
		}

		name := info.Name()

		for category, targets := range cleanTargets {
			for _, target := range targets {
				if matchesTargetCLI(name, target, info.IsDir()) {
					size, sizeStr := calculateSizeCLI(path, info)

					items = append(items, CleanableItem{
						Path:         path,
						Type:         category,
						Size:         size,
						SizeStr:      sizeStr,
						LastModified: info.ModTime(),
					})

					if info.IsDir() {
						return filepath.SkipDir
					}
					break
				}
			}
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error walking directory: %v\n", err)
	}

	return items
}

func matchesTargetCLI(name, target string, isDir bool) bool {
	if name == target {
		return true
	}

	if strings.Contains(target, "*") && !isDir {
		if strings.HasPrefix(target, "*.") {
			ext := strings.TrimPrefix(target, "*")
			return strings.HasSuffix(name, ext)
		}
	}

	return false
}

func calculateSizeCLI(path string, info os.FileInfo) (int64, string) {
	if !info.IsDir() {
		return info.Size(), formatSizeCLI(info.Size())
	}

	var totalSize int64
	filepath.Walk(path, func(subPath string, subInfo os.FileInfo, err error) error {
		if err == nil && !subInfo.IsDir() {
			totalSize += subInfo.Size()
		}
		return nil
	})

	return totalSize, formatSizeCLI(totalSize)
}

func formatSizeCLI(bytes int64) string {
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
