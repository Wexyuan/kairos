package clipboard

import (
	"time"

	"github.com/Wexyuan/kairos/internal/module/base"
)

// CopyClipboardEntryReq is the copy clipboard entry request.
type CopyClipboardEntryReq struct {
	ID string `json:"id"` // clipboard entry ID
}

// DeleteClipboardEntryReq is the delete clipboard entry request.
type DeleteClipboardEntryReq struct {
	ID string `json:"id"` // clipboard entry ID
}

// ClipboardEntryItem is the clipboard entry item.
type ClipboardEntryItem struct {
	// keep the snowflake ID as string, JavaScript numbers lose precision above 2^53.
	ID         string    `json:"id"`           // clipboard entry ID
	Content    string    `json:"content"`      // clipboard content
	LastUsedAt time.Time `json:"last_used_at"` // last used time
}

// ListClipboardEntriesReq is the list clipboard entries request.
type ListClipboardEntriesReq struct {
	base.PageQuery
	Keyword string `json:"keyword"` // search keyword in content
}
