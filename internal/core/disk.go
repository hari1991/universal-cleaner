package core

// DiskUsage describes the capacity of a filesystem mounted at a given path.
type DiskUsage struct {
	TotalBytes int64 // total capacity
	UsedBytes  int64 // total - free
	FreeBytes  int64 // available to the user
}

// Percent returns the used percentage in the range [0, 100].
func (d DiskUsage) Percent() float64 {
	if d.TotalBytes == 0 {
		return 0
	}
	return float64(d.UsedBytes) / float64(d.TotalBytes) * 100
}

// DiskUsageForPath returns disk space information for the filesystem that
// contains the given path. On error it returns a zero-value struct.
func DiskUsageForPath(path string) DiskUsage {
	return diskUsageForPath(path)
}
