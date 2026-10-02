package kairos

import "embed"

// Assets holds the embedded frontend dist.
//
//go:embed all:frontend/dist
var Assets embed.FS

// TrayIcon holds the embedded tray icon.
//
//go:embed build/windows/icon.ico
var TrayIcon []byte
