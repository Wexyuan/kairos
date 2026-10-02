package desktop

import (
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

// Tray provides system tray related methods.
type Tray struct {
	win  application.Window // main window
	name string             // tray name
	icon []byte             // tray icon bytes
}

func NewTray(win application.Window, name string, icon []byte) *Tray {
	return &Tray{
		win:  win,
		name: name,
		icon: icon,
	}
}

// RunTray runs the system tray.
func (t *Tray) RunTray() {
	// hide window when tray is clicked rather than close it directly.
	t.win.RegisterHook(events.Common.WindowClosing, func(event *application.WindowEvent) {
		event.Cancel()
		t.win.Hide()
	})

	tray := application.Get().SystemTray.New()
	tray.SetIcon(t.icon)
	tray.SetTooltip(t.name)
	tray.SetMenu(t.buildMenu())
	tray.OnClick(t.showWindow)
	tray.Run()
}

// buildMenu builds the system tray menu.
func (t *Tray) buildMenu() *application.Menu {
	menu := application.NewMenu()
	menu.Add("显示主窗口").OnClick(func(*application.Context) { t.showWindow() })
	menu.Add("退出").OnClick(func(*application.Context) { application.Get().Quit() })
	return menu
}

// showWindow shows the main window.
func (t *Tray) showWindow() {
	t.win.Restore()
	t.win.Show()
	t.win.Focus()
}
