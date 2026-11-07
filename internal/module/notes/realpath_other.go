//go:build !windows

package notes

import "path/filepath"

// resolveRealPath resolves symbolic links on non-Windows platforms.
func resolveRealPath(path string) (string, error) {
	return filepath.EvalSymlinks(path)
}
