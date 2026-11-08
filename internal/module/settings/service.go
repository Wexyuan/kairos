package settings

import (
	"context"
	"fmt"

	"github.com/Wexyuan/kairos/internal/config"
	"github.com/Wexyuan/kairos/internal/constant"
)

// SettingsService provides settings operations for the frontend.
type SettingsService struct {
	cfg *config.Config
}

func NewSettingsService(cfg *config.Config) (*SettingsService, error) {
	if cfg == nil {
		return nil, fmt.Errorf("[settings] config is required")
	}
	return &SettingsService{cfg: cfg}, nil
}

// GetSettings returns the config values for the frontend.
func (s *SettingsService) GetSettings(ctx context.Context) (*GetSettingsResp, error) {
	return &GetSettingsResp{
		Name:                      s.cfg.App.Name,
		Version:                   constant.Version,
		ClipboardShortcut:         s.cfg.Clipboard.Shortcut,
		ClipboardMaxEntries:       s.cfg.Clipboard.MaxEntries,
		ClipboardMaxPinnedEntries: s.cfg.Clipboard.MaxPinnedEntries,
		NoteMaxImageSize:          s.cfg.Notes.MaxImageSize,
	}, nil
}
