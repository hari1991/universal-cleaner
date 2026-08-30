package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"universal-cleaner/internal/core"
)

// CLI version of the universal cleaner.
func runCLI() {
	var (
		scanPath     = flag.String("path", ".", "Path to scan for cleanable items")
		dryRun       = flag.Bool("dry-run", false, "Show what would be deleted without actually deleting")
		autoYes      = flag.Bool("yes", false, "Automatically confirm all deletions")
		verbose      = flag.Bool("verbose", false, "Show detailed output")
		includeRisky = flag.Bool("include-risky", false, "Include risky targets (Cargo.lock, vendor, IDE configs)")
		noTrash      = flag.Bool("no-trash", false, "Delete permanently instead of moving to trash")
	)
	flag.Parse()

	settings := core.LoadSettings()
	if *includeRisky {
		settings.IncludeRisky = true
	}
	useTrash := settings.UseTrash && !*noTrash

	fmt.Println("Universal Cleaner CLI")
	fmt.Println("====================")
	fmt.Printf("Trash: %s\n", trashLabel(useTrash))

	// Show disk usage for the scan path's filesystem.
	du := core.DiskUsageForPath(*scanPath)
	if du.TotalBytes > 0 {
		fmt.Printf("Disk: %s used of %s · %s free (%.1f%%)\n",
			core.FormatSize(du.UsedBytes),
			core.FormatSize(du.TotalBytes),
			core.FormatSize(du.FreeBytes),
			du.Percent())
	}

	fmt.Printf("Scanning directory: %s\n", *scanPath)
	items := core.Scan(core.ScanOptions{
		Root:          *scanPath,
		EnabledTypes:  settings.EnabledTypes,
		ExcludedNames: toExclusionSet(settings.ExcludedNames),
		IncludeRisky:  settings.IncludeRisky,
	})

	if len(items) == 0 {
		fmt.Println("No cleanable items found.")
		return
	}

	fmt.Printf("\nFound %d cleanable items:\n", len(items))
	var totalSize int64
	riskyCount := 0

	for i, item := range items {
		riskyMark := ""
		if item.Risky {
			riskyMark = " [RISKY]"
			riskyCount++
		}
		fmt.Printf("%d. [%s]%s %s (%s)\n", i+1, item.Type, riskyMark, item.Path, item.SizeStr)
		totalSize += item.Size
		if *verbose {
			fmt.Printf("   Last modified: %s\n", item.LastModified.Format("2006-01-02 15:04:05"))
		}
	}

	fmt.Printf("\nTotal size: %s\n", core.FormatSize(totalSize))
	if riskyCount > 0 {
		fmt.Printf("Risky items: %d (deleting these may break projects)\n", riskyCount)
	}

	if *dryRun {
		fmt.Println("\nDry run mode - no files will be deleted.")
		return
	}

	if riskyCount > 0 {
		fmt.Print("\nWARNING: risky targets are included. Continue? (y/N): ")
		if !confirm(autoYes) {
			fmt.Println("Operation cancelled.")
			return
		}
	}

	if !*autoYes {
		fmt.Print("\nDo you want to delete these items? (y/N): ")
		if !confirm(autoYes) {
			fmt.Println("Operation cancelled.")
			return
		}
	}

	fmt.Printf("\nDeleting items (mode: %s)...\n", trashLabel(useTrash))
	deletedCount := 0
	var failedItems []string

	for _, item := range items {
		if *verbose {
			fmt.Printf("Deleting: %s\n", item.Path)
		}
		if err := core.Delete(item, useTrash); err != nil {
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

func trashLabel(useTrash bool) string {
	if useTrash {
		return "move to trash"
	}
	return "permanent delete"
}

func confirm(autoYes *bool) bool {
	if *autoYes {
		return true
	}
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}
