//go:build darwin || linux || freebsd || netbsd || openbsd

package core

import "syscall"

func diskUsageForPath(path string) DiskUsage {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return DiskUsage{}
	}
	total := int64(stat.Blocks) * int64(stat.Bsize)
	// Bavail = blocks available to unprivileged users; Bfree = all free blocks.
	free := int64(stat.Bavail) * int64(stat.Bsize)
	used := total - int64(stat.Bfree)*int64(stat.Bsize)
	// Guard against edge cases where Bavail > Bfree (reserved blocks).
	if used < 0 {
		used = total - free
	}
	return DiskUsage{
		TotalBytes: total,
		UsedBytes:  used,
		FreeBytes:  free,
	}
}
