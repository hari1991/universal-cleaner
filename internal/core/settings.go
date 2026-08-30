package core

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// DefaultSettings returns a sensible starting configuration with all
// non-risky categories enabled and trash usage on.
func DefaultSettings() Settings {
	s := Settings{
		EnabledTypes:  make(map[string]bool),
		ExcludedNames: []string{},
		IncludeRisky:  false,
		UseTrash:      true,
		DarkTheme:     true,
		Accent:        "#2563eb",
	}
	for _, c := range DefaultCategories {
		s.EnabledTypes[c.Name] = !c.Risky
	}
	return s
}

// settingsPath returns the on-disk location of the settings file.
func settingsPath() (string, error) {
	cfg, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(cfg, "universal-cleaner")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "settings.json"), nil
}

// LoadSettings reads persisted settings, falling back to defaults.
func LoadSettings() Settings {
	path, err := settingsPath()
	if err != nil {
		return DefaultSettings()
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return DefaultSettings()
	}
	var s Settings
	if err := json.Unmarshal(data, &s); err != nil {
		return DefaultSettings()
	}
	if s.EnabledTypes == nil {
		s.EnabledTypes = make(map[string]bool)
	}
	if s.Accent == "" {
		s.Accent = "#2563eb"
	}
	return s
}

// SaveSettings persists settings to disk.
func SaveSettings(s Settings) error {
	path, err := settingsPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// LogFilePath returns the path to the persistent activity log file.
func LogFilePath() (string, error) {
	cfg, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(cfg, "universal-cleaner")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "activity.log"), nil
}

// AddRecentFolder adds a folder to the recent-folders list (deduplicated, max 10).
func (s *Settings) AddRecentFolder(path string) {
	for i, f := range s.RecentFolders {
		if f == path {
			s.RecentFolders = append(s.RecentFolders[:i], s.RecentFolders[i+1:]...)
			break
		}
	}
	s.RecentFolders = append([]string{path}, s.RecentFolders...)
	if len(s.RecentFolders) > 10 {
		s.RecentFolders = s.RecentFolders[:10]
	}
}
