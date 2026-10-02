package main

import (
	"log"

	"github.com/wailsapp/wails/v3/pkg/application"

	"github.com/Wexyuan/kairos"
	"github.com/Wexyuan/kairos/cmd/desktop/wire"
	appconfig "github.com/Wexyuan/kairos/internal/config"
	"github.com/Wexyuan/kairos/internal/module/desktop"
	"github.com/Wexyuan/kairos/pkg/config"
	"github.com/Wexyuan/kairos/pkg/database"
	"github.com/Wexyuan/kairos/pkg/logger"
)

// models is the list of models to migrate.
var models = []any{}

// main starts the application.
func main() {
	// 1. Resolve the config path.
	cfgPath, err := appconfig.ResolveConfigPath()
	if err != nil {
		log.Fatalf("[main] resolve config path error: %v", err)
	}

	// 2. Load the config file.
	var cfg appconfig.Config
	if err := config.Load(cfgPath, &cfg); err != nil {
		log.Fatalf("[main] load config %s error: %v", cfgPath, err)
	}

	// 3. Resolve the writable storage root path.
	root, err := appconfig.ResolveStorageRoot(cfg.App.Name)
	if err != nil {
		log.Fatalf("[main] resolve app storage root path error: %v", err)
	}

	// 4. Create the logger.
	logFile, err := appconfig.ResolveStoragePath(root, cfg.Logger.File)
	if err != nil {
		log.Fatalf("[main] resolve log storage path error: %v", err)
	}
	lg, err := logger.New(logger.Options{
		Level: cfg.Logger.Level,
		File:  logFile,
	})
	if err != nil {
		log.Fatalf("[main] create logger error: %v", err)
	}
	defer func() { _ = lg.Sync() }()

	// 5. Migrate the database.
	dsn, err := appconfig.ResolveStoragePath(root, cfg.Database.DSN)
	if err != nil {
		lg.Fatalf("[main] resolve database storage path error: %v", err)
	}
	db, err := database.New(database.Options{
		DSN:             dsn,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
	})
	if err != nil {
		lg.Fatalf("[main] init database error: dsn=%s: %v", cfg.Database.DSN, err)
	}
	if err := db.AutoMigrate(models...); err != nil {
		lg.Fatalf("[main] migrate models error: %v", err)
	}
	if err := db.Close(); err != nil {
		lg.Errorf("close migration database error: %v", err)
	}

	// 6. Build the application.
	app, err := wire.NewWireApp(&cfg)
	if err != nil {
		lg.Fatalf("[main] build application error: %v", err)
	}

	// 7. Create the main window.
	win := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:             "main",
		Title:            cfg.App.Name,
		Width:            1440,
		Height:           900,
		MinWidth:         1280,
		MinHeight:        800,
		BackgroundColour: application.NewRGB(255, 255, 255),
		URL:              "/",
	})

	// 8. Run the system tray.
	desktop.NewTray(win, cfg.App.Name, kairos.TrayIcon).RunTray()

	// 9. Run the application.
	lg.Infof("[main] application starting: config=%s root=%s", cfgPath, root)
	if err := app.Run(); err != nil {
		lg.Fatalf("[main] run application error: %v", err)
	}
}
