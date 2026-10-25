package clipboard

import (
	"context"
	"syscall"
	"time"

	"github.com/wailsapp/wails/v3/pkg/w32"
)

// procGetClipboardSequenceNumber is the user32 clipboard sequence number syscall.
var procGetClipboardSequenceNumber = syscall.NewLazyDLL("user32.dll").NewProc("GetClipboardSequenceNumber")

// clipboardSequenceNumber returns the system clipboard sequence number, which increments on every clipboard update.
func clipboardSequenceNumber() uint32 {
	ret, _, _ := procGetClipboardSequenceNumber.Call()
	return uint32(ret)
}

// clipboardWatcher polls the system clipboard and saves changed text.
type clipboardWatcher struct {
	service      *ClipboardService
	lastSequence uint32 // sequence number of the last clipboard content
}

// StartClipboardWatcher starts the clipboard watcher.
func (s *ClipboardService) StartClipboardWatcher(ctx context.Context) {
	w := &clipboardWatcher{service: s, lastSequence: clipboardSequenceNumber()}
	go w.run(ctx)
}

// run runs the clipboard polling loop.
func (w *clipboardWatcher) run(ctx context.Context) {
	ticker := time.NewTicker(watcherInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.pollClipboard(ctx)
		}
	}
}

// pollClipboard polls the clipboard and saves the text once the sequence number changed.
func (w *clipboardWatcher) pollClipboard(ctx context.Context) {
	sequence := clipboardSequenceNumber()
	if sequence == w.lastSequence {
		return
	}
	w.lastSequence = sequence

	text, err := w32.GetClipboardText()
	if err != nil {
		return
	}
	if err := w.service.SaveClipboardText(ctx, text); err != nil {
		w.service.lg.Errorf("[clipboard] save clipboard text error: %v", err)
	}
}
