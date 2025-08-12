package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

)

type CleanableItem struct {
	Path        string
	Type        string
	Size        int64
	SizeStr     string
	LastModified time.Time
}

type UniversalCleaner struct {
	window      fyne.Window
	pathLabel   *widget.Label
	scanButton  *widget.Button
	cleanButton *widget.Button
	progressBar *widget.ProgressBar
	itemList    *widget.List
	statusLabel *widget.Label
	
	selectedPath string
	cleanableItems []CleanableItem
	selectedItems map[int]bool
}

// Common directories and files to clean
var cleanTargets = map[string][]string{
	"Node.js": {"node_modules", ".npm", ".yarn", "npm-debug.log", "yarn-error.log"},
	"Java": {"target", "build", ".gradle", ".m2/repository"},
	"Python": {"__pycache__", ".pytest_cache", "*.pyc", "*.pyo", ".tox", "venv", ".venv", "env", ".env"},
	"Rust": {"target", "Cargo.lock"},
	"Go": {"vendor", "bin", "pkg"},
	"C/C++": {"build", "cmake-build-debug", "cmake-build-release", "*.o", "*.obj", "*.exe"},
	"Docker": {".docker", "docker-compose.override.yml"},
	"IDE": {".vscode", ".idea", "*.swp", "*.swo", ".DS_Store", "Thumbs.db"},
	"Build": {"dist", "out", "output", "release", "debug", ".cache", "tmp", "temp"},
}

func main() {
	// Check if running in CLI mode
	if len(os.Args) > 1 {
		runCLI()
		return
	}

	// Run GUI mode
	myApp := app.New()
	myApp.SetIcon(resourceIconPng)
	
	cleaner := &UniversalCleaner{
		selectedItems: make(map[int]bool),
	}
	
	cleaner.window = myApp.NewWindow("Universal Cleaner")
	cleaner.window.Resize(fyne.NewSize(800, 600))
	
	cleaner.setupUI()
	
	cleaner.window.ShowAndRun()
}

func (uc *UniversalCleaner) setupUI() {
	// Header
	title := widget.NewLabel("Universal Cleaner")
	title.TextStyle.Bold = true
	
	// Path selection
	uc.pathLabel = widget.NewLabel("No folder selected")
	selectButton := widget.NewButton("Select Folder", uc.selectFolder)
	
	pathContainer := container.NewHBox(
		widget.NewLabel("Target Folder:"),
		uc.pathLabel,
		selectButton,
	)
	
	// Scan and Clean buttons
	uc.scanButton = widget.NewButton("Scan for Cleanable Items", uc.scanFolder)
	uc.scanButton.Disable()
	
	uc.cleanButton = widget.NewButton("Clean Selected Items", uc.cleanSelected)
	uc.cleanButton.Disable()
	
	buttonContainer := container.NewHBox(uc.scanButton, uc.cleanButton)
	
	// Progress bar
	uc.progressBar = widget.NewProgressBar()
	uc.progressBar.Hide()
	
	// Item list
	uc.itemList = widget.NewList(
		func() int {
			return len(uc.cleanableItems)
		},
		func() fyne.CanvasObject {
			check := widget.NewCheck("", nil)
			label := widget.NewLabel("Template")
			sizeLabel := widget.NewLabel("Size")
			typeLabel := widget.NewLabel("Type")
			
			return container.NewHBox(check, label, widget.NewSeparator(), typeLabel, sizeLabel)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id >= len(uc.cleanableItems) {
				return
			}
			
			item := uc.cleanableItems[id]
			hbox := obj.(*fyne.Container)
			
			check := hbox.Objects[0].(*widget.Check)
			label := hbox.Objects[1].(*widget.Label)
			typeLabel := hbox.Objects[3].(*widget.Label)
			sizeLabel := hbox.Objects[4].(*widget.Label)
			
			check.SetChecked(uc.selectedItems[id])
			check.OnChanged = func(checked bool) {
				uc.selectedItems[id] = checked
				uc.updateCleanButton()
			}
			
			label.SetText(item.Path)
			typeLabel.SetText(item.Type)
			sizeLabel.SetText(item.SizeStr)
		},
	)
	
	// Status label
	uc.statusLabel = widget.NewLabel("Ready to scan")
	
	// Layout
	content := container.NewVBox(
		title,
		widget.NewSeparator(),
		pathContainer,
		buttonContainer,
		uc.progressBar,
		widget.NewLabel("Cleanable Items:"),
		uc.itemList,
		widget.NewSeparator(),
		uc.statusLabel,
	)
	
	uc.window.SetContent(content)
}

func (uc *UniversalCleaner) selectFolder() {
	dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
		if err != nil || uri == nil {
			return
		}
		
		uc.selectedPath = uri.Path()
		uc.pathLabel.SetText(uc.selectedPath)
		uc.scanButton.Enable()
		uc.cleanableItems = nil
		uc.selectedItems = make(map[int]bool)
		uc.itemList.Refresh()
		uc.statusLabel.SetText("Folder selected. Ready to scan.")
	}, uc.window)
}

func (uc *UniversalCleaner) scanFolder() {
	if uc.selectedPath == "" {
		return
	}
	
	uc.progressBar.Show()
	uc.scanButton.Disable()
	uc.cleanButton.Disable()
	uc.statusLabel.SetText("Scanning...")
	
	go func() {
		items := uc.findCleanableItems(uc.selectedPath)
		
		// Update UI on main thread
		uc.cleanableItems = items
		uc.selectedItems = make(map[int]bool)
		
		uc.itemList.Refresh()
		uc.progressBar.Hide()
		uc.scanButton.Enable()
		
		if len(items) > 0 {
			uc.statusLabel.SetText(fmt.Sprintf("Found %d cleanable items", len(items)))
		} else {
			uc.statusLabel.SetText("No cleanable items found")
		}
	}()
}

func (uc *UniversalCleaner) findCleanableItems(rootPath string) []CleanableItem {
	var items []CleanableItem
	
	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue walking even if there's an error
		}
		
		// Skip if it's the root path
		if path == rootPath {
			return nil
		}
		
		name := info.Name()
		
		// Check against all clean targets
		for category, targets := range cleanTargets {
			for _, target := range targets {
				if uc.matchesTarget(name, target, info.IsDir()) {
					size, sizeStr := uc.calculateSize(path, info)
					
					items = append(items, CleanableItem{
						Path:         path,
						Type:         category,
						Size:         size,
						SizeStr:      sizeStr,
						LastModified: info.ModTime(),
					})
					
					// If it's a directory we want to clean, skip walking into it
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
		log.Printf("Error walking directory: %v", err)
	}
	
	return items
}

func (uc *UniversalCleaner) matchesTarget(name, target string, isDir bool) bool {
	// Exact match
	if name == target {
		return true
	}
	
	// Wildcard match for files
	if strings.Contains(target, "*") && !isDir {
		if strings.HasPrefix(target, "*.") {
			ext := strings.TrimPrefix(target, "*")
			return strings.HasSuffix(name, ext)
		}
	}
	
	return false
}

func (uc *UniversalCleaner) calculateSize(path string, info os.FileInfo) (int64, string) {
	if !info.IsDir() {
		return info.Size(), uc.formatSize(info.Size())
	}
	
	var totalSize int64
	filepath.Walk(path, func(subPath string, subInfo os.FileInfo, err error) error {
		if err == nil && !subInfo.IsDir() {
			totalSize += subInfo.Size()
		}
		return nil
	})
	
	return totalSize, uc.formatSize(totalSize)
}

func (uc *UniversalCleaner) formatSize(bytes int64) string {
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

func (uc *UniversalCleaner) updateCleanButton() {
	hasSelected := false
	for _, selected := range uc.selectedItems {
		if selected {
			hasSelected = true
			break
		}
	}
	
	if hasSelected {
		uc.cleanButton.Enable()
	} else {
		uc.cleanButton.Disable()
	}
}

func (uc *UniversalCleaner) cleanSelected() {
	selectedCount := 0
	for _, selected := range uc.selectedItems {
		if selected {
			selectedCount++
		}
	}
	
	if selectedCount == 0 {
		return
	}
	
	// Confirmation dialog
	confirmText := fmt.Sprintf("Are you sure you want to delete %d selected items? This action cannot be undone.", selectedCount)
	
	dialog.ShowConfirm("Confirm Deletion", confirmText, func(confirmed bool) {
		if !confirmed {
			return
		}
		
		uc.performCleanup()
	}, uc.window)
}

func (uc *UniversalCleaner) performCleanup() {
	uc.progressBar.Show()
	uc.cleanButton.Disable()
	uc.scanButton.Disable()
	
	go func() {
		deletedCount := 0
		var failedItems []string
		
		for i, selected := range uc.selectedItems {
			if !selected || i >= len(uc.cleanableItems) {
				continue
			}
			
			item := uc.cleanableItems[i]
			err := os.RemoveAll(item.Path)
			if err != nil {
				failedItems = append(failedItems, item.Path)
				log.Printf("Failed to delete %s: %v", item.Path, err)
			} else {
				deletedCount++
			}
		}
		
		// Update UI
		uc.progressBar.Hide()
		uc.scanButton.Enable()
		
		if len(failedItems) > 0 {
			uc.statusLabel.SetText(fmt.Sprintf("Deleted %d items, %d failed", deletedCount, len(failedItems)))
		} else {
			uc.statusLabel.SetText(fmt.Sprintf("Successfully deleted %d items", deletedCount))
		}
		
		// Refresh the list
		uc.scanFolder()
	}()
}
