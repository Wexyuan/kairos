package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	configDirName  = "configs"
	configFileName = "config.yaml"
	dataDirName    = "data"
	webviewDirName = "webview"
	moduleFileName = "go.mod"
)

// ResolveConfigPath returns the config file path.
func ResolveConfigPath() (string, error) {
	var candidates []string
	if dir, err := resolveExecutableDir(); err == nil {
		candidates = append(candidates, filepath.Join(dir, configDirName, configFileName))
	}
	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(cwd, configDirName, configFileName))
	}

	for _, candidate := range candidates {
		if isFile(candidate) {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("[config] no config file found in %s", strings.Join(candidates, ", "))
}

// ResolveStorageRoot returns the writable storage root path.
func ResolveStorageRoot(appName string) (string, error) {
	exeDir, exeErr := resolveExecutableDir()

	if exeErr == nil && isDir(filepath.Join(exeDir, dataDirName)) {
		return exeDir, nil
	}
	if cwd, err := os.Getwd(); err == nil && (exeErr != nil || cwd != exeDir) &&
		isFile(filepath.Join(cwd, moduleFileName)) && isFile(filepath.Join(cwd, configDirName, configFileName)) {
		return cwd, nil
	}

	if appName == "" {
		return "", fmt.Errorf("[config] app name is required to resolve the storage root")
	}
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("[config] resolve user cache dir error: %w", err)
	}

	root := filepath.Join(cacheDir, appName)
	if err := os.MkdirAll(root, 0o755); err != nil {
		return "", fmt.Errorf("[config] create root dir %s error: %w", root, err)
	}
	return root, nil
}

// ResolveStoragePath returns the writable storage path.
func ResolveStoragePath(root, value string) (string, error) {
	if value == "" || filepath.IsAbs(value) {
		return value, nil
	}

	resolved := filepath.Join(root, filepath.FromSlash(value))
	if err := os.MkdirAll(filepath.Dir(resolved), 0o755); err != nil {
		return "", fmt.Errorf("[config] create dir for %s error: %w", resolved, err)
	}
	return resolved, nil
}

// ResolveWebviewPath returns the WebView2 cache directory.
func ResolveWebviewPath(root string) (string, error) {
	return ResolveStoragePath(root, filepath.Join(dataDirName, webviewDirName))
}

// resolveExecutableDir returns the directory of the executable.
func resolveExecutableDir() (string, error) {
	path, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("[config] resolve executable path error: %w", err)
	}
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		path = resolved
	}
	return filepath.Dir(path), nil
}

// isDir reports whether the path exists and is a directory.
func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// isFile reports whether the path exists and is a file.
func isFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
