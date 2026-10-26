package wire

import (
	"fmt"

	"github.com/Wexyuan/kairos/internal/config"
	"github.com/Wexyuan/kairos/pkg/cache"
	"github.com/Wexyuan/kairos/pkg/database"
	"github.com/Wexyuan/kairos/pkg/id"
	"github.com/Wexyuan/kairos/pkg/logger"
)

func provideDatabase(cfg *config.Config, log *logger.Logger) (*database.Database, func(), error) {
	root, err := config.ResolveStorageRoot(cfg.App.Name)
	if err != nil {
		return nil, nil, fmt.Errorf("[wire] resolve storage root path error: %w", err)
	}
	dsn, err := config.ResolveStoragePath(root, cfg.Database.DSN)
	if err != nil {
		return nil, nil, fmt.Errorf("[wire] resolve database dsn path error: %w", err)
	}

	db, err := database.New(database.Options{
		DSN:             dsn,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("[wire] init database error: dsn=%s: %w", dsn, err)
	}
	return db, func() {
		if err := db.Close(); err != nil {
			log.Errorf("close database error: %v", err)
		}
	}, nil
}

func provideSnowflake(cfg *config.Config) (*id.Snowflake, error) {
	return id.New(cfg.App.SnowflakeNode)
}

func provideCache(cfg *config.Config, log *logger.Logger) (cache.Cache, func(), error) {
	store, err := cache.New(cache.Options{
		TTL: cfg.Cache.TTL,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("[wire] init cache error: ttl=%d: %w", cfg.Cache.TTL, err)
	}
	return store, func() {
		if err := store.Close(); err != nil {
			log.Errorf("close cache error: %v", err)
		}
	}, nil
}

func provideWebviewDataPath(cfg *config.Config) (string, error) {
	root, err := config.ResolveStorageRoot(cfg.App.Name)
	if err != nil {
		return "", fmt.Errorf("[wire] resolve storage root path error: %w", err)
	}

	path, err := config.ResolveWebviewPath(root)
	if err != nil {
		return "", fmt.Errorf("[wire] resolve webview path error: %w", err)
	}
	return path, nil
}
