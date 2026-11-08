package settings

// GetSettingsResp is the get settings response.
type GetSettingsResp struct {
	Name                      string `json:"name"`                         // application name
	Version                   string `json:"version"`                      // application version
	ClipboardShortcut         string `json:"clipboard_shortcut"`           // global hotkey toggling the clipboard window
	ClipboardMaxEntries       int    `json:"clipboard_max_entries"`        // max unpinned clipboard entries
	ClipboardMaxPinnedEntries int    `json:"clipboard_max_pinned_entries"` // max pinned clipboard entries
	NoteMaxImageSize          int64  `json:"note_max_image_size"`          // max note image bytes
}
