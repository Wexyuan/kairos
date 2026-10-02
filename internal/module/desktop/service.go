package desktop

import (
	"context"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/w32"
)

// mainWindowName is the name of the main window.
const mainWindowName = "main"

// WindowService provides desktop window-related methods.
type WindowService struct{}

func NewWindowService() *WindowService {
	return &WindowService{}
}

// SetTheme sets the theme of the main window.
func (s *WindowService) SetTheme(ctx context.Context, req *SetThemeReq) error {
	win, ok := application.Get().Window.GetByName(mainWindowName)
	if !ok {
		return nil
	}

	application.InvokeAsync(func() {
		hwnd := uintptr(win.NativeWindow())
		if hwnd == 0 {
			return
		}
		bg, text := uint32(0x00F5F5F5), uint32(0x001A1A1A)
		if req.Dark {
			bg, text = uint32(0x00101010), uint32(0x00E5E5E5)
		}
		w32.SetTheme(hwnd, req.Dark)
		w32.SetTitleBarColour(hwnd, bg)
		w32.SetTitleTextColour(hwnd, text)
		w32.SetBorderColour(hwnd, bg)
	})
	return nil
}
