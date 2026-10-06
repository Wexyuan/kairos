package clipboard

import (
	"time"

	"github.com/Wexyuan/kairos/internal/module/base"
)

// ClipboardEntry is the clipboard entry entity stored in the database.
type ClipboardEntry struct {
	base.BaseModel
	Content    string    `gorm:"type:text;not null;comment:clipboard content" json:"content"`
	Hash       string    `gorm:"type:varchar(64);not null;index:idx_clipboard_entries_hash;comment:sha256 hash of the content" json:"hash"`
	LastUsedAt time.Time `gorm:"not null;index:idx_last_used_at;comment:last used time" json:"last_used_at"`
}
