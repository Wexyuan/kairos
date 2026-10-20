package desktop

import (
	"context"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/w32"

	"github.com/Wexyuan/kairos/internal/constant"
)

// WindowService provides desktop window-related methods.
type WindowService struct{}

func NewWindowService() *WindowService {
	return &WindowService{}
}

// SetTheme sets the theme of all windows.
func (s *WindowService) SetTheme(ctx context.Context, req *SetThemeReq) error {
	bg, text := w32.RGB(0xFF, 0xFF, 0xFF), w32.RGB(0x18, 0x22, 0x33)
	if req.Dark {
		bg, text = w32.RGB(0x1A, 0x22, 0x2B), w32.RGB(0xE8, 0xEE, 0xF8)
	}

	application.InvokeAsync(func() {
		for _, name := range []string{
			constant.MainWindow,
			constant.ClipboardWindow,
		} {
			win, ok := application.Get().Window.GetByName(name)
			if !ok {
				continue
			}
			hwnd := uintptr(win.NativeWindow())
			if hwnd == 0 {
				continue
			}
			w32.SetTheme(hwnd, req.Dark)
			w32.SetTitleBarColour(hwnd, bg)
			w32.SetTitleTextColour(hwnd, text)
			w32.SetBorderColour(hwnd, bg)
		}
	})
	return nil
}
