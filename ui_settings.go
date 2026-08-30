package main

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"universal-cleaner/internal/core"
)

// openSettings shows a dialog for editing categories, exclusions, deletion
// mode, theme and accent color.
func (uc *UniversalCleaner) openSettings() {
	// Work on a copy so cancellation leaves state untouched.
	draft := uc.settings
	draft.EnabledTypes = copyBoolMap(uc.settings.EnabledTypes)
	draft.ExcludedNames = append([]string{}, uc.settings.ExcludedNames...)

	// --- Categories ---
	catBoxes := buildCategoryChecks(&draft)
	catContainer := container.NewVBox(catBoxes...)

	// --- Deletion mode ---
	trashCheck := widget.NewCheck("Move deleted items to Trash (recoverable)", func(b bool) { draft.UseTrash = b })
	trashCheck.SetChecked(draft.UseTrash)

	// --- Theme ---
	darkCheck := widget.NewCheck("Dark theme", func(b bool) {
		draft.DarkTheme = b
	})
	darkCheck.SetChecked(draft.DarkTheme)

	accentEntry := widget.NewEntry()
	accentEntry.SetPlaceHolder("#2563eb")
	accentEntry.SetText(draft.Accent)

	// --- Exclusions ---
	exclEntry := widget.NewMultiLineEntry()
	exclEntry.SetPlaceHolder("One name per line, e.g.\nCargo.lock\nnode_modules")
	exclEntry.SetText(strings.Join(draft.ExcludedNames, "\n"))

	form := container.NewVBox(
		sectionLabel("Categories to scan"),
		container.NewScroll(catContainer),
		widget.NewSeparator(),
		sectionLabel("Deletion"),
		trashCheck,
		widget.NewSeparator(),
		sectionLabel("Appearance"),
		darkCheck,
		container.NewHBox(widget.NewLabel("Accent color:"), accentEntry),
		widget.NewSeparator(),
		sectionLabel("Excluded names (never delete)"),
		exclEntry,
	)

	dlg := dialog.NewCustomConfirm("Settings", "Save", "Cancel", form, func(save bool) {
		if !save {
			return
		}
		draft.Accent = strings.TrimSpace(accentEntry.Text)
		draft.ExcludedNames = parseLines(exclEntry.Text)
		uc.settings = draft
		applyTheme(uc.app, draft)
		core.SaveSettings(draft)
		uc.log("Settings updated")
		// Re-scan if a folder is already selected so the new filters apply.
		if uc.selectedPath != "" {
			uc.scanFolder()
		} else {
			uc.refreshActions()
		}
	}, uc.window)
	dlg.Resize(fyne.NewSize(520, 560))
	dlg.Show()
}

func sectionLabel(s string) fyne.CanvasObject {
	l := widget.NewLabel(s)
	l.TextStyle = fyne.TextStyle{Bold: true}
	return l
}

func buildCategoryChecks(draft *core.Settings) []fyne.CanvasObject {
	var objs []fyne.CanvasObject
	for _, c := range core.DefaultCategories {
		c := c
		label := c.Name
		if c.Risky {
			label += "  (risky)"
		}
		check := widget.NewCheck(label, func(b bool) {
			draft.EnabledTypes[c.Name] = b
		})
		check.SetChecked(draft.EnabledTypes[c.Name])
		desc := widget.NewLabel(c.Description)
		desc.Importance = widget.LowImportance
		objs = append(objs, container.NewVBox(check, desc))
	}
	return objs
}

func parseLines(s string) []string {
	var out []string
	for _, line := range strings.Split(s, "\n") {
		l := strings.TrimSpace(line)
		if l != "" {
			out = append(out, l)
		}
	}
	return out
}

func copyBoolMap(m map[string]bool) map[string]bool {
	out := make(map[string]bool, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

// keep theme import referenced (icons used elsewhere).
var _ = theme.SettingsIcon
