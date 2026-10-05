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
	bg, text := uint32(0x00F5F5F5), uint32(0x001A1A1A)
	if req.Dark {
		bg, text = uint32(0x00101010), uint32(0x00E5E5E5)
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
