package core

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// ErrScanStopped is returned (via items being partial) when a scan is cancelled.
var ErrScanStopped = errors.New("scan stopped by user")

// Scan walks rootPath and returns cleanable items honouring the given options.
// It reports progress via opts.ProgressCb (which may be nil). If opts.Stop is
// closed, the scan aborts early and returns the items found so far.
//
// Size calculation is deferred: the walk collects matches quickly, then sizes
// are computed in parallel by a small worker pool. This prevents a single
// large directory (e.g. node_modules) from blocking the entire scan.
func Scan(opts ScanOptions) []CleanableItem {
	var items []CleanableItem

	// Build the set of active targets from enabled categories.
	type activeTarget struct {
		category string
		target   Target
		risky    bool
	}
	var active []activeTarget
	for _, c := range DefaultCategories {
		if c.Risky {
			if !opts.IncludeRisky && !categoryEnabled(opts, c.Name) {
				continue
			}
		} else if !categoryEnabled(opts, c.Name) {
			continue
		}
		for _, t := range c.Targets {
			active = append(active, activeTarget{category: c.Name, target: t, risky: c.Risky})
		}
	}

	type rawMatch struct {
		path     string
		name     string
		category string
		risky    bool
		isDir    bool
		modTime  time.Time
	}
	var matches []rawMatch

	scanned := 0
	lastReport := time.Now()
	rootDepth := len(strings.Split(filepath.Clean(opts.Root), string(filepath.Separator)))

	err := filepath.Walk(opts.Root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if path == opts.Root {
			return nil
		}

		// Check for cancellation.
		if opts.Stop != nil {
			select {
			case <-opts.Stop:
				return ErrScanStopped
			default:
			}
		}

		// Enforce max depth.
		if opts.MaxDepth > 0 {
			depth := len(strings.Split(filepath.Clean(path), string(filepath.Separator))) - rootDepth
			if depth > opts.MaxDepth {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		scanned++
		if opts.ProgressCb != nil && time.Since(lastReport) > 30*time.Millisecond {
			opts.ProgressCb(Progress{CurrentPath: path, Scanned: scanned, Found: len(matches)})
			lastReport = time.Now()
		}

		name := info.Name()

		// Honour exclusions (exact name match).
		if opts.ExcludedNames != nil && opts.ExcludedNames[name] {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		for _, a := range active {
			if matchesTarget(name, a.target, info.IsDir()) {
				// Check hint files: if the target specifies HintFiles, only
				// match when one of them exists in the item's parent or an
				// ancestor up to the scan root.
				if len(a.target.HintFiles) > 0 && !hasHintFile(path, opts.Root, a.target.HintFiles) {
					continue
				}
				matches = append(matches, rawMatch{
					path:     path,
					name:     name,
					category: a.category,
					risky:    a.risky,
					isDir:    info.IsDir(),
					modTime:  info.ModTime(),
				})
				if info.IsDir() {
					return filepath.SkipDir
				}
				break
			}
		}
		return nil
	})

	if err != nil && !errors.Is(err, ErrScanStopped) {
		if opts.ProgressCb != nil {
			opts.ProgressCb(Progress{CurrentPath: err.Error(), Scanned: scanned, Found: len(matches)})
		}
	}

	// Phase 2: compute sizes in parallel.
	items = make([]CleanableItem, len(matches))
	const workers = 8
	type sizeResult struct {
		index     int
		size      int64
		fileCount int
	}
	jobs := make(chan int, len(matches))
	results := make(chan sizeResult, len(matches))

	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				m := matches[i]
				size, count, _ := CalculateSize(m.path, m.isDir)
				results <- sizeResult{index: i, size: size, fileCount: count}
			}
		}()
	}
	for i := range matches {
		jobs <- i
	}
	close(jobs)

	// Collect results as they arrive.
	type sizeInfo struct {
		size  int64
		count int
	}
	resultMap := make(map[int]sizeInfo, len(matches))
	go func() {
		for r := range results {
			resultMap[r.index] = sizeInfo{size: r.size, count: r.fileCount}
		}
	}()
	wg.Wait()
	close(results)

	for i, m := range matches {
		si := resultMap[i]
		items[i] = CleanableItem{
			Path:         m.path,
			Name:         m.name,
			Type:         m.category,
			Size:         si.size,
			SizeStr:      FormatSize(si.size),
			FileCount:    si.count,
			LastModified: m.modTime,
			IsDir:        m.isDir,
			Risky:        m.risky,
		}
	}

	if opts.ProgressCb != nil {
		opts.ProgressCb(Progress{Scanned: scanned, Found: len(items), Done: true})
	}

	return items
}

func categoryEnabled(opts ScanOptions, name string) bool {
	if opts.EnabledTypes == nil {
		return true
	}
	enabled, ok := opts.EnabledTypes[name]
	if !ok {
		return true
	}
	return enabled
}

// SortItems sorts a slice of cleanable items in place by the given column/order.
func SortItems(items []CleanableItem, col SortColumn, order SortOrder) {
	switch col {
	case SortByType:
		sort.Slice(items, func(i, j int) bool { return cmpString(items[i].Type, items[j].Type, order) })
	case SortBySize:
		sort.Slice(items, func(i, j int) bool { return cmpInt64(items[i].Size, items[j].Size, order) })
	case SortByModified:
		sort.Slice(items, func(i, j int) bool { return cmpTime(items[i].LastModified, items[j].LastModified, order) })
	default:
		sort.Slice(items, func(i, j int) bool { return cmpString(items[i].Path, items[j].Path, order) })
	}
}

func cmpString(a, b string, order SortOrder) bool {
	if order == SortAsc {
		return a < b
	}
	return a > b
}

func cmpInt64(a, b int64, order SortOrder) bool {
	if order == SortAsc {
		return a < b
	}
	return a > b
}

func cmpTime(a, b time.Time, order SortOrder) bool {
	if order == SortAsc {
		return a.Before(b)
	}
	return a.After(b)
}

// FilterItems returns the subset of items whose path/type/name contain query
// (case-insensitive). An empty query returns the slice unchanged.
func FilterItems(items []CleanableItem, query string) []CleanableItem {
	if query == "" {
		return items
	}
	q := strings.ToLower(query)
	var out []CleanableItem
	for _, it := range items {
		if strings.Contains(strings.ToLower(it.Path), q) ||
			strings.Contains(strings.ToLower(it.Type), q) ||
			strings.Contains(strings.ToLower(it.Name), q) {
			out = append(out, it)
		}
	}
	return out
}

// hasHintFile checks whether any of the hint files exist in the item's parent
// directory or any ancestor up to (but not including) the scan root. This
// prevents generic directory names like "target" or "build" from matching
// outside of actual project directories.
func hasHintFile(itemPath, root string, hints []string) bool {
	parent := filepath.Dir(itemPath)
	rootClean := filepath.Clean(root)
	for {
		for _, h := range hints {
			if _, err := os.Stat(filepath.Join(parent, h)); err == nil {
				return true
			}
		}
		parentClean := filepath.Clean(parent)
		if parentClean == rootClean || parent == "/" || parent == "." {
			break
		}
		parent = filepath.Dir(parent)
	}
	return false
}
