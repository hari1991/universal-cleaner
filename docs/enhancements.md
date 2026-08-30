# Universal Cleaner — Enhancement Roadmap

This document tracks all identified enhancements for the Universal Cleaner
application, grouped by priority and category. Items are implemented in
priority order (P0 → P3).

Status legend: `[ ]` pending · `[~]` in progress · `[x]` done

---

## P0 — Bugs / Correctness (fix now)

### #1 Stop button is non-functional
- **Problem:** `stopScan()` closes `uc.scanCancel`, but `core.Scan` never
  receives on it. The scanner runs to completion regardless of the Stop click.
- **Fix:** Pass a `context.Context` (or a `stop` channel) into `Scan` and check
  it inside the `filepath.Walk` callback; return `filepath.SkipDir`/stop on
  cancel.
- **Files:** `internal/core/scanner.go`, `main.go`
- **Status:** [ ]

### #2 Keyboard shortcut hijacks text entry
- **Problem:** `bindShortcuts` binds `KeyR` → rescan and `KeyF` → focus search
  at the canvas level. Typing "r" or "f" in the search box triggers a rescan
  instead of filtering.
- **Fix:** Check whether the focused element is an editable widget (Entry) and
  skip the shortcut in that case. Use `Cmd/Ctrl+R` instead of bare `R`.
- **Files:** `main.go`
- **Status:** [ ]

### #3 Progress bar is fake
- **Problem:** `float64(p.Scanned%100) / 100.0` cycles 0→1 repeatedly and does
  not reflect real progress.
- **Fix:** Use an indeterminate progress bar (`Min=0, Max=0`) during scan, or
  report a meaningful ratio (e.g., bytes scanned / estimated total). For
  cleanup, the per-item ratio is already correct.
- **Files:** `main.go`
- **Status:** [ ]

### #4 Auto-rescan after cleanup is surprising
- **Problem:** `performCleanup` calls `uc.scanFolder()` at the end, starting a
  new scan without user action.
- **Fix:** Remove the automatic rescan; instead clear the deleted items from
  the list and update the status. Offer a "Rescan" prompt.
- **Files:** `main.go`
- **Status:** [ ]

### #5 Double-scan race condition
- **Problem:** `scanFolder` doesn't check if a scan is already running.
  `performCleanup`'s auto-rescan can overlap with a manual scan click.
- **Fix:** Add a `scanning bool` guard; ignore `scanFolder` calls while a scan
  is in progress (or queue them).
- **Files:** `main.go`
- **Status:** [ ]

### #6 Windows trash is broken
- **Problem:** `os.Getenv("RECYCLER")` returns empty (that env var doesn't
  exist). Falls back to permanent delete silently.
- **Fix:** Use `SHFileOperation` via `syscall`/`ole` to move to Recycle Bin on
  Windows, or use a third-party library. At minimum, document the limitation.
- **Files:** `internal/core/trash.go`
- **Status:** [ ]

### #7 Linux trash missing `.trashinfo` sidecar
- **Problem:** Moving files to `~/.local/share/Trash/files` without the
  corresponding `.trashinfo` file means the DE doesn't recognize them as
  trashable/restoreable.
- **Fix:** Write a matching `.trashinfo` file in `~/.local/share/Trash/info/`
  with `[Trash Info]` path and deletion date.
- **Files:** `internal/core/trash.go`
- **Status:** [ ]

---

## P1 — High Value

### #8 Size calculation blocks the scan
- **Problem:** `CalculateSize` walks the entire subtree of every match during
  the scan. A single `node_modules` with 100k files pauses the scan for
  seconds.
- **Fix:** Defer size calculation — collect matches first, then compute sizes
  in parallel (worker pool) after the walk, or lazily on display.
- **Files:** `internal/core/scanner.go`, `internal/core/size.go`
- **Status:** [ ]

### #9 `toLower`/`indexOf` reinvent the stdlib
- **Problem:** Hand-rolled byte loops are slower and less clear than
  `strings.Contains(strings.ToLower(s), q)`.
- **Fix:** Replace with stdlib string functions.
- **Files:** `internal/core/scanner.go`
- **Status:** [ ]

### #10 No empty state
- **Problem:** When the table has zero items, it's just blank.
- **Fix:** Show a placeholder message ("No items found — select a folder and
  scan") when `displayItems` is empty.
- **Files:** `main.go`, `ui_table.go`
- **Status:** [ ]

### #11 No "Open in Finder/Explorer"
- **Problem:** Can't jump to a scanned item's location to inspect it before
  deleting.
- **Fix:** Add a right-click or double-click action on a table row that opens
  the item's parent folder in the OS file manager (`open`, `explorer`, `xdg-open`).
- **Files:** `main.go`, `internal/core/open.go` (new)
- **Status:** [ ]

### #12 No item file count
- **Problem:** For directories, only size is shown — "45 MB · 12,000 files" is
  far more informative.
- **Fix:** Count files during `CalculateSize` and add a `FileCount` field to
  `CleanableItem`; display it in the table.
- **Files:** `internal/core/types.go`, `internal/core/size.go`,
  `internal/core/scanner.go`, `ui_table.go`
- **Status:** [ ]

### #13 No per-category counts in dropdown
- **Problem:** "Node.js (12)" helps users jump to the biggest category.
- **Fix:** After scan, build category options with counts and total sizes.
- **Files:** `main.go`
- **Status:** [ ]

### #14 No export of results
- **Problem:** Can't save the scan list to CSV/JSON for records or scripting.
- **Fix:** Add an "Export" button that writes the current results to CSV/JSON
  via a file save dialog.
- **Files:** `main.go`, `internal/core/export.go` (new)
- **Status:** [ ]

### #15 No recent folders
- **Problem:** Re-selecting the same folder each session is tedious.
- **Fix:** Store recent folder paths in settings; show a dropdown or menu of
  recent folders next to the path row.
- **Files:** `internal/core/settings.go`, `main.go`
- **Status:** [ ]

### #16 No drag-and-drop
- **Problem:** Can't drag a folder onto the window to select it.
- **Fix:** Use `window.SetOnDropped` (Fyne v2.6) to accept folder URIs.
- **Files:** `main.go`
- **Status:** [ ]

### #17 Theme button has no state indication
- **Problem:** Can't tell whether dark or light is active from the icon.
- **Fix:** Swap the icon between `theme.ColorPaletteIcon()` (dark) and
  `theme.ColorAchromaticIcon()` (light) based on current state.
- **Files:** `main.go`
- **Status:** [ ]

### #18 No menu bar
- **Problem:** Desktop apps conventionally have File/View/Help menus.
- **Fix:** Add a `AppMainMenu` with File (Open Folder, Export, Quit), View
  (Toggle Theme, Settings), and Help (About) items.
- **Files:** `main.go`
- **Status:** [ ]

### #19 No window geometry persistence
- **Problem:** Window size/position resets each launch.
- **Fix:** Save window size in settings; restore on startup.
- **Files:** `internal/core/settings.go`, `main.go`
- **Status:** [ ]

### #20 No log persistence
- **Problem:** Activity log is lost on restart; useful for auditing deletions.
- **Fix:** Append log lines to a log file in the config dir; optionally load
  recent lines on startup.
- **Files:** `internal/core/settings.go`, `main.go`
- **Status:** [ ]

### #21 No size column right-alignment
- **Problem:** Numbers look better right-aligned.
- **Fix:** Use `fyne.TextAlignTrailing` for the size column labels.
- **Files:** `ui_table.go`
- **Status:** [ ]

### #22 No sort-direction indicator
- **Problem:** The header doesn't show ▲/▼ for the current sort.
- **Fix:** Append ▲/▼ to the active column's header label.
- **Files:** `ui_table.go`
- **Status:** [ ]

---

## P2 — Safety / Smartness

### #23 Generic names match too broadly
- **Problem:** `target`, `build`, `bin`, `out`, `tmp`, `env`, `vendor` match
  *any* directory with that name, even non-build ones.
- **Fix:** Add heuristics — e.g., only match `target` if `Cargo.toml` exists in
  an ancestor; only match `build` if `CMakeLists.txt` or `build.gradle` nearby.
- **Files:** `internal/core/scanner.go`, `internal/core/targets.go`
- **Status:** [ ]

### #24 No gitignore / git-awareness
- **Problem:** Cleaning a `bin/` that's tracked in git removes committed files.
- **Fix:** Optionally check `.gitignore` or run `git ls-files` to skip
  git-tracked paths.
- **Files:** `internal/core/scanner.go` (new `git.go`)
- **Status:** [ ]

### #25 `.m2` is a global cache, not project-local
- **Problem:** `.m2` lives in the home directory, not in projects.
- **Fix:** Remove `.m2` from the Java category, or only match it when the scan
  root is the home directory.
- **Files:** `internal/core/targets.go`
- **Status:** [ ]

### #26 No max-depth option
- **Problem:** Scanning a home directory recursively can descend into huge
  subtrees with no bound.
- **Fix:** Add a `MaxDepth` field to `ScanOptions` and enforce it in the walk.
- **Files:** `internal/core/scanner.go`, `internal/core/types.go`
- **Status:** [ ]

---

## P3 — Code Quality

### #27 Dead code cleanup
- **Problem:** `selectedCount()`, `selectedSize()`, `shortPath()` in
  main.go/ui_table.go are never called; `IsRiskyName()`, `CategoryByName()` in
  core are unused; `DeleteResult` type is unused; `var _ = fmt.Sprintf` is a
  hack.
- **Fix:** Remove unused functions and types; remove the `fmt` import hack.
- **Files:** `main.go`, `ui_table.go`, `internal/core/*.go`
- **Status:** [ ]

### #28 Redundant `fyne.Do` in dialog callbacks
- **Problem:** `selectFolder`'s dialog callback runs on the Fyne thread, but
  `uc.setStatus`/`uc.log` called there wrap in `fyne.Do` redundantly.
- **Fix:** Minor — acceptable overhead; document or use direct calls in
  main-thread contexts.
- **Files:** `main.go`
- **Status:** [ ]

### #29 No tests for core package
- **Problem:** `core` has testable pure functions (`matchesTarget`,
  `SortItems`, `FilterItems`, `FormatSize`) with zero coverage.
- **Fix:** Add unit tests for core functions (when prompted).
- **Files:** `internal/core/*_test.go`
- **Status:** [ ]
