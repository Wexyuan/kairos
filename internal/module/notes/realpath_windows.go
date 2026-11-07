//go:build windows

package notes

import (
	"fmt"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows"
)

// resolveRealPath resolves symlinks, junctions and other reparse points on Windows.
func resolveRealPath(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("[notes] resolve absolute path %s error: %w", path, err)
	}
	name, err := windows.UTF16PtrFromString(abs)
	if err != nil {
		return "", fmt.Errorf("[notes] encode path %s error: %w", path, err)
	}
	handle, err := windows.CreateFile(
		name,
		0,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_FLAG_BACKUP_SEMANTICS,
		0,
	)
	if err != nil {
		return "", fmt.Errorf("[notes] open path %s error: %w", path, err)
	}
	defer windows.CloseHandle(handle)

	size := uint32(512)
	for {
		buffer := make([]uint16, size)
		n, err := windows.GetFinalPathNameByHandle(handle, &buffer[0], size, 0)
		if err != nil {
			return "", fmt.Errorf("[notes] resolve final path %s error: %w", path, err)
		}
		if n < size {
			return normalizeWindowsFinalPath(windows.UTF16ToString(buffer[:n])), nil
		}
		size = n + 1
	}
}

func normalizeWindowsFinalPath(path string) string {
	const uncPrefix = `\\?\UNC\`
	if strings.HasPrefix(path, uncPrefix) {
		return `\\` + strings.TrimPrefix(path, uncPrefix)
	}
	return strings.TrimPrefix(path, `\\?\`)
}
