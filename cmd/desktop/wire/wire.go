//go:build wireinject
// +build wireinject

package wire

import (
	"context"

	"github.com/google/wire"
	"github.com/wailsapp/wails/v3/pkg/application"

	"github.com/Wexyuan/kairos"
	"github.com/Wexyuan/kairos/internal/config"
	"github.com/Wexyuan/kairos/internal/constant"
	"github.com/Wexyuan/kairos/internal/module/base"
	"github.com/Wexyuan/kairos/internal/module/clipboard"
	"github.com/Wexyuan/kairos/internal/module/desktop"
	"github.com/Wexyuan/kairos/internal/module/notes"
	"github.com/Wexyuan/kairos/pkg/logger"
)

// InfraProviderSet defines the infrastructure providers.
var InfraProviderSet = wire.NewSet(
	provideDatabase,
	provideSnowflake,
	provideCache,
)

// ModuleProviderSet defines the business module providers.
var ModuleProviderSet = wire.NewSet(
	base.NewRepository,
	desktop.NewWindowService,
	notes.NewNotesService,
	clipboard.NewClipboardRepository,
	clipboard.NewClipboardService,
)

// NewApp creates a new application.
func NewApp(
	cfg *config.Config,
	lg *logger.Logger,
	webviewDataPath string,
	winSvc *desktop.WindowService,
	notesSvc *notes.NotesService,
	clipSvc *clipboard.ClipboardService,
) *application.App {
	watcherCtx, cancelWatcher := context.WithCancel(context.Background())
	clipSvc.StartClipboardWatcher(watcherCtx)

	var app *application.App
	app = application.New(
		application.Options{
			Name: cfg.App.Name,
			Windows: application.WindowsOptions{
				WebviewUserDataPath: webviewDataPath,
			},
			OnShutdown: func() {
				cancelWatcher()
				if err := app.GlobalShortcut.Unregister(constant.ClipboardWindowHotkey); err != nil {
					lg.Errorf("[wire] unregister hotkey %s error: %v", constant.ClipboardWindowHotkey, err)
				}
			},
			Services: []application.Service{
				application.NewService(winSvc),
				application.NewService(notesSvc),
				application.NewService(clipSvc),
			},
			Assets: application.AssetOptions{
				Handler: application.AssetFileServerFS(kairos.Assets),
			},
		},
	)

	if err := app.GlobalShortcut.Register(constant.ClipboardWindowHotkey, func() {
		_ = winSvc.ToggleClipboardWindow(context.Background())
	}); err != nil {
		lg.Errorf("[wire] register hotkey %s error: %v", constant.ClipboardWindowHotkey, err)
	}
	return app
}

// NewWireApp wires and builds the application.
func NewWireApp(cfg *config.Config, lg *logger.Logger) (*application.App, func(), error) {
	panic(wire.Build(
		InfraProviderSet,
		provideWebviewDataPath,
		ModuleProviderSet,
		NewApp,
	))
}
