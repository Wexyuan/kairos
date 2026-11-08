package config

const (
	// defaultClipboardShortcut is the default shortcut for the clipboard window.
	defaultClipboardShortcut = "Ctrl+Shift+V"
	// defaultClipboardMaxEntries is the default max clipboard unpinned entries.
	defaultClipboardMaxEntries = 500
	// defaultClipboardMaxPinnedEntries is the default max clipboard pinned entries.
	defaultClipboardMaxPinnedEntries = 50
	// defaultNoteMaxImageSize is the default max note image bytes.
	defaultNoteMaxImageSize = 10 << 20
)

// SetDefaults sets the default values for the config.
func (c *Config) SetDefaults() {
	if c.Clipboard.Shortcut == "" {
		c.Clipboard.Shortcut = defaultClipboardShortcut
	}
	if c.Clipboard.MaxEntries <= 0 {
		c.Clipboard.MaxEntries = defaultClipboardMaxEntries
	}
	if c.Clipboard.MaxPinnedEntries <= 0 {
		c.Clipboard.MaxPinnedEntries = defaultClipboardMaxPinnedEntries
	}
	if c.Notes.MaxImageSize <= 0 {
		c.Notes.MaxImageSize = defaultNoteMaxImageSize
	}
}
