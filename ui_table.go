package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	"universal-cleaner/internal/core"
)

const (
	colSel      = 0
	colPath     = 1
	colType     = 2
	colSize     = 3
	colFiles    = 4
	colModified = 5
	colRisk     = 6
	numCols     = 7
)

// headerLabels returns the column headers, with a sort-direction arrow
// appended to the active sort column.
func (uc *UniversalCleaner) headerLabels() [numCols]string {
	var h [numCols]string
	h[colSel] = "✓"
	h[colPath] = "Path"
	h[colType] = "Type"
	h[colSize] = "Size"
	h[colFiles] = "Files"
	h[colModified] = "Modified"
	h[colRisk] = "Risk"

	arrow := " ▲"
	if uc.sortOrder == core.SortDesc {
		arrow = " ▼"
	}
	switch uc.sortCol {
	case core.SortByPath:
		h[colPath] += arrow
	case core.SortByType:
		h[colType] += arrow
	case core.SortBySize:
		h[colSize] += arrow
	case core.SortByModified:
		h[colModified] += arrow
	}
	return h
}

// buildTable constructs the results table with a header row and sortable columns.
func (uc *UniversalCleaner) buildTable() {
	uc.table = widget.NewTable(
		func() (rows, cols int) { return len(uc.displayItems) + 1, numCols },
		func() fyne.CanvasObject {
			l := widget.NewLabel("")
			l.Truncation = fyne.TextTruncateEllipsis
			return l
		},
		func(id widget.TableCellID, o fyne.CanvasObject) {
			l := o.(*widget.Label)
			if id.Row == 0 {
				l.TextStyle = fyne.TextStyle{Bold: true}
				headers := uc.headerLabels()
				l.SetText(headers[id.Col])
				l.Alignment = alignForCol(id.Col)
				return
			}
			l.TextStyle = fyne.TextStyle{}
			l.Alignment = alignForCol(id.Col)
			idx := id.Row - 1
			if idx >= len(uc.displayItems) {
				l.SetText("")
				return
			}
			it := uc.displayItems[idx]
			switch id.Col {
			case colSel:
				if uc.selected[it.Path] {
					l.SetText("☑")
				} else {
					l.SetText("☐")
				}
			case colPath:
				l.SetText(uc.relPath(it.Path))
			case colType:
				l.SetText(it.Type)
			case colSize:
				l.SetText(it.SizeStr)
			case colFiles:
				if it.IsDir && it.FileCount > 0 {
					l.SetText(fmt.Sprintf("%d", it.FileCount))
				} else {
					l.SetText("—")
				}
			case colModified:
				l.SetText(it.LastModified.Format("2006-01-02 15:04"))
			case colRisk:
				if it.Risky {
					l.SetText("⚠")
				} else {
					l.SetText("")
				}
			}
		},
	)

	uc.table.SetColumnWidth(colSel, 40)
	uc.table.SetColumnWidth(colPath, 360)
	uc.table.SetColumnWidth(colType, 80)
	uc.table.SetColumnWidth(colSize, 90)
	uc.table.SetColumnWidth(colFiles, 70)
	uc.table.SetColumnWidth(colModified, 130)
	uc.table.SetColumnWidth(colRisk, 50)

	uc.table.OnSelected = func(id widget.TableCellID) {
		defer uc.table.UnselectAll()
		if id.Row == 0 {
			uc.handleSort(id.Col)
			return
		}
		idx := id.Row - 1
		if idx < 0 || idx >= len(uc.displayItems) {
			return
		}
		item := uc.displayItems[idx]
		switch id.Col {
		case colSel:
			uc.toggleSelect(item.Path)
		case colPath:
			// Double-purpose: reveal in file manager on path click.
			uc.revealItem(item.Path)
		}
	}
}

// alignForCol returns the text alignment for a given column.
func alignForCol(col int) fyne.TextAlign {
	if col == colSize || col == colFiles {
		return fyne.TextAlignTrailing
	}
	return fyne.TextAlignLeading
}

// revealItem opens the item's location in the OS file manager.
func (uc *UniversalCleaner) revealItem(path string) {
	if err := core.RevealInFileManager(path); err != nil {
		uc.log("Failed to reveal %s: %v", path, err)
	} else {
		uc.log("Revealed in file manager: %s", path)
	}
}

// handleSort re-sorts the display by the given column, toggling direction when
// the same column is clicked again.
func (uc *UniversalCleaner) handleSort(col int) {
	var target core.SortColumn
	switch col {
	case colPath:
		target = core.SortByPath
	case colType:
		target = core.SortByType
	case colSize:
		target = core.SortBySize
	case colModified:
		target = core.SortByModified
	default:
		return // selection / risk columns are not sortable
	}
	if target == uc.sortCol {
		if uc.sortOrder == core.SortAsc {
			uc.sortOrder = core.SortDesc
		} else {
			uc.sortOrder = core.SortAsc
		}
	} else {
		uc.sortCol = target
		uc.sortOrder = core.SortAsc
	}
	uc.rebuildDisplay()
	uc.log("Sorted by %s %s", sortLabel(uc.sortCol), dirLabel(uc.sortOrder))
}

func sortLabel(c core.SortColumn) string {
	switch c {
	case core.SortByType:
		return "type"
	case core.SortBySize:
		return "size"
	case core.SortByModified:
		return "modified"
	default:
		return "path"
	}
}

func dirLabel(o core.SortOrder) string {
	if o == core.SortAsc {
		return "ascending"
	}
	return "descending"
}
