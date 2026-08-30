package core

import "strings"

// matchesTarget reports whether a filesystem entry matches a target spec.
func matchesTarget(name string, t Target, isDir bool) bool {
	// Exact name match wins regardless of dir/file expectation.
	if name == t.Name {
		return true
	}

	// Wildcard match applies only to files.
	if strings.Contains(t.Name, "*") && !isDir {
		if strings.HasPrefix(t.Name, "*.") {
			ext := strings.TrimPrefix(t.Name, "*")
			return strings.HasSuffix(name, ext)
		}
	}
	return false
}
