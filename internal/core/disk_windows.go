//go:build windows

package core

import "syscall"

func diskUsageForPath(path string) DiskUsage {
	// GetDiskFreeSpaceEx returns free bytes available to the caller, total
	// bytes, and total free bytes (including reserved). We use the caller-
	// available value for FreeBytes.
	var freeBytesAvailable, totalBytes, totalFreeBytes uint64
	pathPtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return DiskUsage{}
	}
	if err := syscall.GetDiskFreeSpaceEx(pathPtr, &freeBytesAvailable, &totalBytes, &totalFreeBytes); err != nil {
		return DiskUsage{}
	}
	total := int64(totalBytes)
	free := int64(freeBytesAvailable)
	used := total - free
	if used < 0 {
		used = 0
	}
	return DiskUsage{
		TotalBytes: total,
		UsedBytes:  used,
		FreeBytes:  free,
	}
}
