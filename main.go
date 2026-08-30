package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"universal-cleaner/internal/core"
)

// UniversalCleaner holds all GUI state.
type UniversalCleaner struct {
	window fyne.Window
	app    fyne.App

	// settings
	settings core.Settings

	// scan state
	allItems     []core.CleanableItem
	displayItems []core.CleanableItem
	selected     map[string]bool // keyed by item.Path
	sortCol      core.SortColumn
	sortOrder    core.SortOrder
	searchQuery  string

	// widgets
	pathLabel    *widget.Label
	diskLabel    *widget.Label
	scanBtn      *widget.Button
	dryRunBtn    *widget.Button
	cleanBtn     *widget.Button
	stopBtn      *widget.Button
	selectAllBtn *widget.Button
	noneBtn      *widget.Button
	categorySel  *widget.Select
	searchEntry  *widget.Entry
	table        *widget.Table

	progressBar       *widget.ProgressBar
	progressBarInfinite *widget.ProgressBarInfinite
	progressLbl       *widget.Label
	statusLabel *widget.Label
	totalLabel  *widget.Label
	logList     *widget.List
	emptyLabel  *widget.Label
	themeBtn    *widget.Button
	recentSel   *widget.Select
	exportBtn   *widget.Button
	logFile     *os.File

	logLines []string

	selectedPath string
	scanCancel   chan struct{}
	scanning     bool
}

func main() {
	if len(os.Args) > 1 {
		runCLI()
		return
	}

	myApp := app.NewWithID("com.universalcleaner.app")
	settings := core.LoadSettings()
	if resourceIconPng != nil && len(resourceIconPng.StaticContent) > 0 {
		myApp.SetIcon(resourceIconPng)
	}
	applyTheme(myApp, settings)

	cleaner := &UniversalCleaner{
		app:       myApp,
		settings:  settings,
		selected:  make(map[string]bool),
		sortCol:   core.SortBySize,
		sortOrder: core.SortDesc,
	}

	cleaner.window = myApp.NewWindow("Universal Cleaner")
	w, h := 980, 680
	if settings.WindowWidth > 0 {
		w = settings.WindowWidth
	}
	if settings.WindowHeight > 0 {
		h = settings.WindowHeight
	}
	cleaner.window.Resize(fyne.NewSize(float32(w), float32(h)))
	cleaner.setupUI()
	cleaner.bindShortcuts()
	cleaner.setupMenu()
	cleaner.openLogFile()

	cleaner.window.SetCloseIntercept(func() {
		// Persist window geometry.
		size := cleaner.window.Canvas().Size()
		cleaner.settings.WindowWidth = int(size.Width)
		cleaner.settings.WindowHeight = int(size.Height)
		core.SaveSettings(cleaner.settings)
		if cleaner.logFile != nil {
			cleaner.logFile.Close()
		}
		cleaner.window.Close()
	})

	cleaner.window.ShowAndRun()
}

func applyTheme(a fyne.App, s core.Settings) {
	a.Settings().SetTheme(newCustomTheme(s.Accent, s.DarkTheme))
}

func (uc *UniversalCleaner) setupUI() {
	uc.buildHeader()
	uc.buildTable()
	uc.buildLog()

	// Status widgets must exist before buildToolbar, because the category
	// Select's OnChanged fires during SetSelected and reaches rebuildDisplay.
	uc.progressBar = widget.NewProgressBar()
	uc.progressBar.Hide()
	uc.progressBarInfinite = widget.NewProgressBarInfinite()
	uc.progressBarInfinite.Hide()
	uc.progressLbl = widget.NewLabel("")
	uc.progressLbl.Truncation = fyne.TextTruncateEllipsis
	uc.statusLabel = widget.NewLabel("Ready")
	uc.totalLabel = widget.NewLabel("0 items selected · 0 B")

	// Empty-state placeholder must exist before buildToolbar for the same
	// reason — rebuildDisplay toggles its visibility. Wrap in a centered
	// container so it appears in the middle of the table area, not top-left.
	uc.emptyLabel = widget.NewLabel("No items to display.\nSelect a folder and click Scan to find cleanable items.")
	uc.emptyLabel.Alignment = fyne.TextAlignCenter
	uc.emptyLabel.Hide()

	uc.buildToolbar()

	// Progress row: stack determinate + infinite bars (only one visible at a time).
	progressStack := container.NewStack(uc.progressBar, uc.progressBarInfinite)
	progressRow := container.NewBorder(nil, nil, nil, uc.stopBtn, progressStack)

	// Toolbar row 1: action buttons + category filter
	// Toolbar row 2: search entry (full width)
	tableTop := container.NewVBox(
		uc.toolbarContainer(),
		uc.searchRow(),
		uc.pathRow(),
		progressRow,
		uc.progressLbl,
	)
	// Stack the table and empty label; the empty label is centered via
	// layout.NewSpacer() padding so it appears in the middle of the area.
	emptyCentered := container.NewVBox(
		layout.NewSpacer(),
		container.NewHBox(layout.NewSpacer(), uc.emptyLabel, layout.NewSpacer()),
		layout.NewSpacer(),
	)
	tableStack := container.NewStack(uc.table, emptyCentered)
	tableArea := container.NewBorder(tableTop, nil, nil, nil, tableStack)

	// Bottom split: table (top, large) + log (bottom, small).
	split := container.NewVSplit(tableArea, uc.logContainer())

	// Footer status bar: status label expands, total label is right-aligned.
	footer := container.NewBorder(nil, nil, uc.statusLabel, uc.totalLabel)

	content := container.NewBorder(uc.headerContainer(), footer, nil, nil, split)
	uc.window.SetContent(content)
	uc.refreshActions()

	// Show disk info for the home directory on startup so the user sees
	// disk space even before selecting a folder.
	uc.updateDiskInfo()

	// Support dragging a folder onto the window to select it.
	uc.window.SetOnDropped(func(_ fyne.Position, uris []fyne.URI) {
		for _, u := range uris {
			if u.Scheme() == "file" {
				uc.setFolder(u.Path())
				break
			}
		}
	})

	// Defer split offset until the split has been laid out at least once.
	split.SetOffset(0.78)
}

// ---- Header ----

func (uc *UniversalCleaner) buildHeader() {
	uc.pathLabel = widget.NewLabel("No folder selected")
	uc.pathLabel.Truncation = fyne.TextTruncateEllipsis
	uc.diskLabel = widget.NewLabel("")
	uc.diskLabel.TextStyle = fyne.TextStyle{Italic: true}
}

func (uc *UniversalCleaner) headerContainer() fyne.CanvasObject {
	title := widget.NewLabel("Universal Cleaner")
	title.TextStyle.Bold = true
	title.Wrapping = fyne.TextWrapOff

	subtitle := widget.NewLabel("Reclaim disk space from build artifacts & dependencies")
	subtitle.Wrapping = fyne.TextWrapOff

	uc.themeBtn = widget.NewButtonWithIcon("", theme.ColorPaletteIcon(), uc.toggleTheme)
	uc.themeBtn.Importance = widget.LowImportance
	uc.updateThemeIcon()

	settingsBtn := widget.NewButtonWithIcon("Settings", theme.SettingsIcon(), uc.openSettings)
	settingsBtn.Importance = widget.LowImportance

	left := container.NewVBox(title, subtitle)
	right := container.NewHBox(uc.themeBtn, settingsBtn)
	return container.NewBorder(nil, nil, left, right)
}

// toggleTheme switches between dark and light themes.
func (uc *UniversalCleaner) toggleTheme() {
	uc.settings.DarkTheme = !uc.settings.DarkTheme
	applyTheme(uc.app, uc.settings)
	uc.updateThemeIcon()
	core.SaveSettings(uc.settings)
}

// updateThemeIcon sets the theme button icon based on the current theme.
func (uc *UniversalCleaner) updateThemeIcon() {
	if uc.themeBtn == nil {
		return
	}
	if uc.settings.DarkTheme {
		uc.themeBtn.SetIcon(theme.ColorAchromaticIcon()) // sun → switch to light
	} else {
		uc.themeBtn.SetIcon(theme.ColorPaletteIcon()) // palette → switch to dark
	}
}

// ---- Toolbar ----

func (uc *UniversalCleaner) buildToolbar() {
	uc.scanBtn = widget.NewButtonWithIcon("Scan", theme.ViewRefreshIcon(), uc.scanFolder)
	uc.dryRunBtn = widget.NewButtonWithIcon("Dry Run", theme.SearchIcon(), uc.dryRun)
	uc.cleanBtn = widget.NewButtonWithIcon("Clean", theme.DeleteIcon(), uc.cleanSelected)
	uc.cleanBtn.Importance = widget.HighImportance
	uc.exportBtn = widget.NewButtonWithIcon("Export", theme.DownloadIcon(), uc.exportResults)

	uc.stopBtn = widget.NewButtonWithIcon("Stop", theme.CancelIcon(), uc.stopScan)
	uc.stopBtn.Importance = widget.DangerImportance
	uc.stopBtn.Hide()

	uc.selectAllBtn = widget.NewButton("Select All", uc.selectAll)
	uc.noneBtn = widget.NewButton("None", uc.selectNone)

	categories := []string{"All Categories"}
	for _, c := range core.DefaultCategories {
		categories = append(categories, c.Name)
	}
	uc.categorySel = widget.NewSelect(categories, func(string) {
		uc.rebuildDisplay()
	})
	uc.categorySel.SetSelected("All Categories")

	uc.searchEntry = widget.NewEntry()
	uc.searchEntry.SetPlaceHolder("Filter by path, type or name...")
	uc.searchEntry.OnChanged = func(s string) {
		uc.searchQuery = s
		uc.rebuildDisplay()
	}
}

func (uc *UniversalCleaner) toolbarContainer() fyne.CanvasObject {
	return container.NewHBox(
		uc.scanBtn, uc.dryRunBtn, uc.cleanBtn, uc.exportBtn,
		widget.NewSeparator(),
		uc.selectAllBtn, uc.noneBtn,
		widget.NewSeparator(),
		widget.NewLabel("Category:"), uc.categorySel,
	)
}

// searchRow returns the filter/search entry on its own row, full-width.
func (uc *UniversalCleaner) searchRow() fyne.CanvasObject {
	return container.NewBorder(nil, nil, widget.NewLabel("Filter:"), nil, uc.searchEntry)
}

func (uc *UniversalCleaner) pathRow() fyne.CanvasObject {
	selectBtn := widget.NewButtonWithIcon("Select Folder", theme.FolderOpenIcon(), uc.selectFolder)
	selectBtn.Importance = widget.MediumImportance

	// Recent folders dropdown.
	recentOpts := []string{"Recent…"}
	recentOpts = append(recentOpts, uc.settings.RecentFolders...)
	uc.recentSel = widget.NewSelect(recentOpts, func(val string) {
		if val == "Recent…" || val == "" {
			return
		}
		uc.selectedPath = val
		uc.pathLabel.SetText(val)
		uc.allItems = nil
		uc.displayItems = nil
		uc.selected = make(map[string]bool)
		uc.rebuildDisplay()
		uc.updateDiskInfo()
		uc.setStatus("Folder selected from recent. Ready to scan.")
		uc.refreshActions()
	})
	uc.recentSel.SetSelected("Recent…")

	// Use Border layout so the path label expands to fill available space.
	pathAndDisk := container.NewVBox(uc.pathLabel, uc.diskLabel)
	return container.NewBorder(nil, nil,
		widget.NewLabel("Target:"),
		container.NewHBox(uc.recentSel, selectBtn),
		pathAndDisk,
	)
}

// updateDiskInfo queries the filesystem for the selected path and updates
// the disk info label with total/used/free space.
func (uc *UniversalCleaner) updateDiskInfo() {
	if uc.diskLabel == nil {
		return
	}
	path := uc.selectedPath
	if path == "" {
		// Fall back to home dir so we show something useful on startup.
		if home, err := os.UserHomeDir(); err == nil {
			path = home
		}
	}
	if path == "" {
		uc.diskLabel.SetText("")
		return
	}
	du := core.DiskUsageForPath(path)
	if du.TotalBytes == 0 {
		uc.diskLabel.SetText("")
		return
	}
	uc.diskLabel.SetText(fmt.Sprintf("Disk: %s used of %s · %s free (%.1f%%)",
		core.FormatSize(du.UsedBytes),
		core.FormatSize(du.TotalBytes),
		core.FormatSize(du.FreeBytes),
		du.Percent()))
}

// ---- Actions ----

func (uc *UniversalCleaner) selectFolder() {
	dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
		if err != nil || uri == nil {
			return
		}
		uc.setFolder(uri.Path())
	}, uc.window)
}

// setFolder sets the scan target, updates recents, and refreshes the UI.
// This is called from the Fyne main thread (dialog callback or drop handler),
// so it uses direct UI calls instead of fyne.Do-wrapped helpers.
func (uc *UniversalCleaner) setFolder(path string) {
	uc.selectedPath = path
	uc.pathLabel.SetText(path)
	uc.settings.AddRecentFolder(path)
	uc.refreshRecentDropdown()
	core.SaveSettings(uc.settings)
	uc.allItems = nil
	uc.displayItems = nil
	uc.selected = make(map[string]bool)
	uc.rebuildDisplay()
	uc.updateDiskInfo()
	uc.statusLabel.SetText("Folder selected. Ready to scan.")
	uc.refreshActions()
	// Log directly to file + UI (we're on the main thread).
	uc.log("Folder selected: %s", path)
}

// refreshRecentDropdown rebuilds the recent-folders dropdown options.
func (uc *UniversalCleaner) refreshRecentDropdown() {
	if uc.recentSel == nil {
		return
	}
	opts := []string{"Recent…"}
	opts = append(opts, uc.settings.RecentFolders...)
	uc.recentSel.Options = opts
	uc.recentSel.SetSelected("Recent…")
}

// updateCategoryCounts rebuilds the category dropdown with item counts and
// total sizes per category.
func (uc *UniversalCleaner) updateCategoryCounts() {
	if uc.categorySel == nil {
		return
	}
	counts := make(map[string]int)
	sizes := make(map[string]int64)
	for _, it := range uc.allItems {
		counts[it.Type]++
		sizes[it.Type] += it.Size
	}
	opts := []string{fmt.Sprintf("All Categories (%d)", len(uc.allItems))}
	for _, c := range core.DefaultCategories {
		n := counts[c.Name]
		if n > 0 {
			opts = append(opts, fmt.Sprintf("%s (%d · %s)", c.Name, n, core.FormatSize(sizes[c.Name])))
		}
	}
	uc.categorySel.Options = opts
	uc.categorySel.SetSelected(opts[0])
}

func (uc *UniversalCleaner) scanFolder() {
	if uc.selectedPath == "" {
		return
	}
	if uc.scanning {
		return // a scan is already in progress
	}
	uc.scanning = true
	uc.scanCancel = make(chan struct{})
	uc.progressBarInfinite.Show()
	uc.progressBarInfinite.Start()
	uc.progressLbl.Show()
	uc.stopBtn.Show()
	uc.scanBtn.Disable()
	uc.dryRunBtn.Disable()
	uc.cleanBtn.Disable()
	uc.setStatus("Scanning...")
	uc.log("Scan started: %s", uc.selectedPath)

	go func() {
		items := core.Scan(core.ScanOptions{
			Root:          uc.selectedPath,
			EnabledTypes:  uc.settings.EnabledTypes,
			ExcludedNames: toExclusionSet(uc.settings.ExcludedNames),
			IncludeRisky:  uc.settings.IncludeRisky,
			Stop:          uc.scanCancel,
			ProgressCb: func(p core.Progress) {
				if p.Done {
					return
				}
				if p.CurrentPath != "" {
					uc.setProgressLabel(fmt.Sprintf("Scanned %d entries — %s", p.Scanned, p.CurrentPath))
				}
			},
		})

		stopped := false
		select {
		case <-uc.scanCancel:
			stopped = true
		default:
		}

		fyne.Do(func() {
			uc.allItems = items
			uc.selected = make(map[string]bool)
			uc.updateCategoryCounts()
			uc.rebuildDisplay()
			uc.progressBarInfinite.Stop()
			uc.progressBarInfinite.Hide()
			uc.progressLbl.Hide()
			uc.stopBtn.Hide()
			uc.scanBtn.Enable()
			uc.scanning = false
			uc.refreshActions()

			if stopped {
				uc.statusLabel.SetText(fmt.Sprintf("Scan stopped — %d items found so far", len(items)))
			} else if len(items) > 0 {
				var total int64
				risky := 0
				for _, it := range items {
					total += it.Size
					if it.Risky {
						risky++
					}
				}
				uc.statusLabel.SetText(fmt.Sprintf("Found %d items · %s", len(items), core.FormatSize(total)))
				if risky > 0 {
					uc.log("Warning: %d risky items found (deleting may break projects)", risky)
				}
			} else {
				uc.statusLabel.SetText("No cleanable items found")
			}
		})
		if stopped {
			uc.log("Scan stopped by user — %d items found", len(items))
		} else {
			uc.log("Scan complete: %d items", len(items))
		}
	}()
}

func (uc *UniversalCleaner) stopScan() {
	if uc.scanCancel != nil {
		select {
		case <-uc.scanCancel:
		default:
			close(uc.scanCancel)
		}
	}
	uc.progressBarInfinite.Stop()
	uc.progressBarInfinite.Hide()
	uc.stopBtn.Hide()
	uc.setStatus("Scan stopped")
	uc.log("Scan stopped by user")
}

func (uc *UniversalCleaner) dryRun() {
	if len(uc.allItems) == 0 {
		dialog.ShowInformation("Dry Run", "Scan a folder first to preview cleanable items.", uc.window)
		return
	}
	var b strings.Builder
	var total int64
	risky := 0
	for _, it := range uc.displayItems {
		mark := ""
		if it.Risky {
			mark = "  [RISKY]"
			risky++
		}
		fmt.Fprintf(&b, "[%s]%s %s  (%s)\n", it.Type, mark, it.Path, it.SizeStr)
		total += it.Size
	}
	header := fmt.Sprintf("Dry Run — %d items would be removable (%s)", len(uc.displayItems), core.FormatSize(total))
	if risky > 0 {
		header += fmt.Sprintf("\n%d risky items included.", risky)
	}
	dialog.ShowInformation("Dry Run Preview", header+"\n\n"+b.String(), uc.window)
	uc.log("Dry run: %d items, %s", len(uc.displayItems), core.FormatSize(total))
}

func (uc *UniversalCleaner) cleanSelected() {
	var picked []core.CleanableItem
	for _, it := range uc.allItems {
		if uc.selected[it.Path] {
			picked = append(picked, it)
		}
	}
	if len(picked) == 0 {
		return
	}

	var total int64
	risky := 0
	for _, it := range picked {
		total += it.Size
		if it.Risky {
			risky++
		}
	}
	mode := "move to Trash"
	if !uc.settings.UseTrash {
		mode = "permanently delete"
	}
	msg := fmt.Sprintf("Delete %d items (%s)?\nMode: %s", len(picked), core.FormatSize(total), mode)
	if risky > 0 {
		msg += fmt.Sprintf("\n\nWARNING: %d risky items are selected and may break projects.", risky)
	}
	if !uc.settings.UseTrash {
		msg += "\n\nThis action CANNOT be undone."
	}

	dialog.ShowConfirm("Confirm Deletion", msg, func(ok bool) {
		if !ok {
			return
		}
		uc.performCleanup(picked)
	}, uc.window)
}

func (uc *UniversalCleaner) performCleanup(picked []core.CleanableItem) {
	uc.progressBarInfinite.Hide()
	uc.progressBar.Min = 0
	uc.progressBar.Max = float64(len(picked))
	uc.progressBar.Show()
	uc.progressLbl.Show()
	uc.cleanBtn.Disable()
	uc.scanBtn.Disable()
	uc.dryRunBtn.Disable()
	uc.setStatus("Cleaning...")

	go func() {
		deleted, failed := 0, 0
		for i, it := range picked {
			uc.setProgress(float64(i) / float64(len(picked)))
			uc.setProgressLabel(it.Path)
			if err := core.Delete(it, uc.settings.UseTrash); err != nil {
				failed++
				uc.log("Failed: %s (%v)", it.Path, err)
			} else {
				deleted++
				uc.log("Deleted: %s", it.Path)
			}
		}
		fyne.Do(func() {
			uc.progressBar.Hide()
			uc.progressLbl.Hide()
			uc.scanBtn.Enable()
			uc.refreshActions()
			uc.updateCategoryCounts()
			uc.updateDiskInfo()
			switch {
			case failed > 0:
				uc.statusLabel.SetText(fmt.Sprintf("Deleted %d, %d failed", deleted, failed))
			default:
				uc.statusLabel.SetText(fmt.Sprintf("Deleted %d items", deleted))
			}
			uc.selected = make(map[string]bool)
			// Remove deleted items from the list instead of auto-rescanning.
			uc.allItems = removeDeleted(uc.allItems, picked)
			uc.rebuildDisplay()
		})
		uc.log("Cleanup done: %d deleted, %d failed", deleted, failed)
	}()
}

// removeDeleted returns items with any successfully-deleted paths removed.
// Since we can't know per-item success here, we remove all picked paths that
// no longer exist on disk.
func removeDeleted(items, picked []core.CleanableItem) []core.CleanableItem {
	deletedSet := make(map[string]bool, len(picked))
	for _, p := range picked {
		if _, err := os.Stat(p.Path); os.IsNotExist(err) {
			deletedSet[p.Path] = true
		}
	}
	if len(deletedSet) == 0 {
		return items
	}
	out := items[:0:0]
	for _, it := range items {
		if !deletedSet[it.Path] {
			out = append(out, it)
		}
	}
	return out
}

// ---- Selection ----

func (uc *UniversalCleaner) selectAll() {
	for _, it := range uc.displayItems {
		uc.selected[it.Path] = true
	}
	uc.table.Refresh()
	uc.updateSelectionSummary()
	uc.refreshActions()
}

func (uc *UniversalCleaner) selectNone() {
	for k := range uc.selected {
		delete(uc.selected, k)
	}
	uc.table.Refresh()
	uc.updateSelectionSummary()
	uc.refreshActions()
}

func (uc *UniversalCleaner) toggleSelect(path string) {
	if uc.selected[path] {
		delete(uc.selected, path)
	} else {
		uc.selected[path] = true
	}
	uc.table.Refresh()
	uc.updateSelectionSummary()
	uc.refreshActions()
}

func (uc *UniversalCleaner) updateSelectionSummary() {
	if uc.totalLabel == nil {
		return
	}
	count, size := 0, int64(0)
	for _, it := range uc.allItems {
		if uc.selected[it.Path] {
			count++
			size += it.Size
		}
	}
	uc.totalLabel.SetText(fmt.Sprintf("%d items selected · %s", count, core.FormatSize(size)))
}

func (uc *UniversalCleaner) refreshActions() {
	if uc.scanBtn == nil {
		return
	}
	hasItems := len(uc.allItems) > 0
	hasSelection := false
	for _, v := range uc.selected {
		if v {
			hasSelection = true
			break
		}
	}
	uc.scanBtn.Enable()
	setBtnEnabled(uc.dryRunBtn, hasItems)
	setBtnEnabled(uc.cleanBtn, hasSelection)
	setBtnEnabled(uc.selectAllBtn, hasItems)
	setBtnEnabled(uc.noneBtn, hasSelection)
	setBtnEnabled(uc.exportBtn, hasItems)
}

// ---- Display rebuild (filter + sort) ----

func (uc *UniversalCleaner) rebuildDisplay() {
	if uc.table == nil {
		return // UI not fully built yet
	}
	items := uc.allItems
	if cat := uc.categorySel.Selected; cat != "" && !strings.HasPrefix(cat, "All Categories") {
		// Extract the category name before the parenthesis.
		catName := cat
		if idx := strings.Index(cat, " ("); idx > 0 {
			catName = cat[:idx]
		}
		filtered := items[:0:0]
		for _, it := range items {
			if it.Type == catName {
				filtered = append(filtered, it)
			}
		}
		items = filtered
	}
	items = core.FilterItems(items, uc.searchQuery)
	core.SortItems(items, uc.sortCol, uc.sortOrder)
	uc.displayItems = items
	uc.table.Refresh()
	// Show empty-state placeholder when there are no items.
	if len(items) == 0 {
		uc.emptyLabel.Show()
	} else {
		uc.emptyLabel.Hide()
	}
	uc.updateSelectionSummary()
	uc.refreshActions()
}

// ---- Helpers ----

// exportResults opens a file save dialog and writes the current results to
// CSV or JSON depending on the chosen extension.
func (uc *UniversalCleaner) exportResults() {
	if len(uc.allItems) == 0 {
		dialog.ShowInformation("Export", "Scan a folder first to export results.", uc.window)
		return
	}
	dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil || writer == nil {
			return
		}
		path := writer.URI().Path()
		writer.Close()

		ext := strings.ToLower(filepath.Ext(path))
		var exportErr error
		switch ext {
		case ".json":
			exportErr = core.ExportJSON(uc.displayItems, path)
		default:
			exportErr = core.ExportCSV(uc.displayItems, path)
		}
		if exportErr != nil {
			uc.log("Export failed: %v", exportErr)
			dialog.ShowError(exportErr, uc.window)
		} else {
			uc.log("Exported %d items to %s", len(uc.displayItems), path)
			uc.setStatus(fmt.Sprintf("Exported %d items", len(uc.displayItems)))
		}
	}, uc.window)
}

// openLogFile opens the persistent activity log file for appending.
func (uc *UniversalCleaner) openLogFile() {
	path, err := core.LogFilePath()
	if err != nil {
		return
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	uc.logFile = f
	uc.log("Session started")
}

// setStatus updates the status bar, safe to call from any goroutine.
func (uc *UniversalCleaner) setStatus(s string) {
	fyne.Do(func() { uc.statusLabel.SetText(s) })
}

// setProgress updates the progress bar, safe to call from any goroutine.
func (uc *UniversalCleaner) setProgress(v float64) {
	fyne.Do(func() { uc.progressBar.SetValue(v) })
}

// setProgressLabel updates the current-path label, safe from any goroutine.
func (uc *UniversalCleaner) setProgressLabel(s string) {
	fyne.Do(func() { uc.progressLbl.SetText(s) })
}

// log appends a timestamped line to the activity log, safe from any goroutine.
// It also persists the line to the on-disk log file if one is open.
func (uc *UniversalCleaner) log(format string, args ...any) {
	line := time.Now().Format("15:04:05") + "  " + fmt.Sprintf(format, args...)
	// Write to the persistent log file (best-effort, no UI dependency).
	if uc.logFile != nil {
		fmt.Fprintln(uc.logFile, line)
	}
	fyne.Do(func() {
		uc.logLines = append(uc.logLines, line)
		if len(uc.logLines) > 500 {
			uc.logLines = uc.logLines[len(uc.logLines)-500:]
		}
		uc.logList.Refresh()
		uc.logList.ScrollToBottom()
	})
}

func (uc *UniversalCleaner) buildLog() {
	uc.logList = widget.NewList(
		func() int { return len(uc.logLines) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(uc.logLines[i])
		},
	)
}

func (uc *UniversalCleaner) logContainer() fyne.CanvasObject {
	header := widget.NewLabel("Activity Log")
	header.TextStyle.Bold = true
	return container.NewBorder(header, nil, nil, nil, uc.logList)
}

// setupMenu creates the application menu bar (File / View / Help).
func (uc *UniversalCleaner) setupMenu() {
	fileMenu := fyne.NewMenu("File",
		fyne.NewMenuItem("Open Folder…", uc.selectFolder),
		fyne.NewMenuItem("Export Results…", uc.exportResults),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Quit", func() {
			uc.window.Close()
		}),
	)
	viewMenu := fyne.NewMenu("View",
		fyne.NewMenuItem("Toggle Theme", uc.toggleTheme),
		fyne.NewMenuItem("Settings…", uc.openSettings),
	)
	helpMenu := fyne.NewMenu("Help",
		fyne.NewMenuItem("About", func() {
			dialog.ShowInformation("About Universal Cleaner",
				"Universal Cleaner\n\nReclaim disk space from build artifacts and dependencies.\n\nBuilt with Go and Fyne.",
				uc.window)
		}),
	)
	uc.window.SetMainMenu(fyne.NewMainMenu(fileMenu, viewMenu, helpMenu))
}

// keyboardShortcut implements fyne.Shortcut for custom key bindings.
type keyboardShortcut struct {
	key fyne.KeyName
	mod fyne.KeyModifier
}

func (s keyboardShortcut) ShortcutName() string { return string(s.key) }
func (s keyboardShortcut) Key() fyne.KeyName    { return s.key }
func (s keyboardShortcut) Mod() fyne.KeyModifier { return s.mod }

func (uc *UniversalCleaner) bindShortcuts() {
	rescan := keyboardShortcut{key: fyne.KeyR, mod: fyne.KeyModifierSuper}
	search := keyboardShortcut{key: fyne.KeyF, mod: fyne.KeyModifierSuper}
	uc.window.Canvas().AddShortcut(rescan, func(_ fyne.Shortcut) {
		if uc.selectedPath != "" {
			uc.scanFolder()
		}
	})
	uc.window.Canvas().AddShortcut(search, func(_ fyne.Shortcut) {
		uc.window.Canvas().Focus(uc.searchEntry)
	})
}

// setBtnEnabled toggles a button's enabled state without panicking when called
// repeatedly with the same value.
func setBtnEnabled(b *widget.Button, enabled bool) {
	if enabled {
		b.Enable()
	} else {
		b.Disable()
	}
}

// toExclusionSet converts a slice of names to a set, returning nil when empty.
func toExclusionSet(names []string) map[string]bool {
	if len(names) == 0 {
		return nil
	}
	m := make(map[string]bool, len(names))
	for _, n := range names {
		m[n] = true
	}
	return m
}

// relPath returns item.Path relative to the scan root when possible.
func (uc *UniversalCleaner) relPath(p string) string {
	if uc.selectedPath == "" {
		return p
	}
	rel, err := filepath.Rel(uc.selectedPath, p)
	if err != nil {
		return p
	}
	return rel
}
