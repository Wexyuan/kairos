package clipboard

import (
	"time"

	"github.com/Wexyuan/kairos/internal/module/base"
)

// CopyEntryReq is the copy clipboard entry request.
type CopyEntryReq struct {
	ID string `json:"id"` // entry ID
}

// DeleteEntryReq is the delete clipboard entry request.
type DeleteEntryReq struct {
	ID string `json:"id"` // entry ID
}

// ClipboardEntryItem is the clipboard entry item.
type ClipboardEntryItem struct {
	// keep the snowflake ID as string, JavaScript numbers lose precision above 2^53.
	ID         string    `json:"id"`           // entry ID
	Content    string    `json:"content"`      // clipboard content
	LastUsedAt time.Time `json:"last_used_at"` // last used time
}

// ListEntriesReq is the list clipboard entries request.
type ListEntriesReq struct {
	base.PageQuery
	Keyword string `json:"keyword"` // search keyword in content
}
