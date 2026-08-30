package core

import "time"

// CleanableItem represents a single file or directory that can be cleaned.
type CleanableItem struct {
	Path         string
	Name         string
	Type         string
	Size         int64
	SizeStr      string
	FileCount    int // number of files in this item (1 for files, >1 for dirs)
	LastModified time.Time
	IsDir        bool
	Risky        bool
}

// SortColumn identifies which column the result list is sorted by.
type SortColumn int

const (
	SortByPath SortColumn = iota
	SortByType
	SortBySize
	SortByModified
)

// SortOrder is the direction of sorting.
type SortOrder int

const (
	SortAsc SortOrder = iota
	SortDesc
)

// Progress is reported by the scanner while walking the filesystem.
type Progress struct {
	CurrentPath string
	Scanned     int
	Found       int
	Done        bool
}

// ScanOptions controls scanner behaviour.
type ScanOptions struct {
	Root          string
	EnabledTypes  map[string]bool // category -> enabled
	ExcludedNames map[string]bool // exact names to never delete
	IncludeRisky  bool            // include targets flagged as risky
	ProgressCb    func(Progress)
	Stop          <-chan struct{} // closed to cancel the scan
	MaxDepth      int             // 0 = unlimited, >0 = max directory depth
}

// Settings is the persisted user configuration.
type Settings struct {
	EnabledTypes  map[string]bool `json:"enabled_types"`
	ExcludedNames []string        `json:"excluded_names"`
	IncludeRisky  bool            `json:"include_risky"`
	UseTrash      bool            `json:"use_trash"`
	DarkTheme     bool            `json:"dark_theme"`
	Accent        string          `json:"accent"`
	RecentFolders []string        `json:"recent_folders"`
	WindowWidth   int             `json:"window_width"`
	WindowHeight  int             `json:"window_height"`
}
