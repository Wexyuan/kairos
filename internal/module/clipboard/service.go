package clipboard

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/wailsapp/wails/v3/pkg/w32"

	"github.com/Wexyuan/kairos/internal/ecode"
	"github.com/Wexyuan/kairos/internal/module/base"
	"github.com/Wexyuan/kairos/pkg/logger"
)

const (
	// maxEntries is the max count of unpinned clipboard entries.
	maxEntries = 500
	// maxPinnedEntries is the max count of pinned clipboard entries.
	maxPinnedEntries = 50
	// watcherInterval is the interval of the clipboard watcher.
	watcherInterval = 500 * time.Millisecond
)

// ClipboardService provides clipboard entry business operations.
type ClipboardService struct {
	lg   *logger.Logger
	repo ClipboardRepository
}

func NewClipboardService(lg *logger.Logger, repo ClipboardRepository) *ClipboardService {
	return &ClipboardService{
		lg:   lg,
		repo: repo,
	}
}

// SaveClipboardText saves clipboard text, deduplicating by content hash.
func (s *ClipboardService) SaveClipboardText(ctx context.Context, content string) error {
	if content == "" {
		return nil
	}

	sum := sha256.Sum256([]byte(content))
	hash := hex.EncodeToString(sum[:])

	entry, err := s.repo.GetClipboardEntryByHash(ctx, hash)
	if err != nil {
		return fmt.Errorf("[clipboard] get clipboard entry by hash error: %w", err)
	}
	if entry != nil {
		if err := s.repo.UpdateClipboardEntryLastUsedAt(ctx, entry.ID); err != nil {
			return fmt.Errorf("[clipboard] update clipboard entry last used time error: %w", err)
		}
		return nil
	}

	now := time.Now()
	if err := s.repo.CreateClipboardEntry(ctx, &ClipboardEntry{
		Content:    content,
		Hash:       hash,
		LastUsedAt: now,
	}); err != nil {
		return fmt.Errorf("[clipboard] create clipboard entry error: %w", err)
	}

	if err := s.repo.DeleteExcessClipboardEntries(ctx, maxEntries); err != nil {
		return fmt.Errorf("[clipboard] delete excess clipboard entries error: %w", err)
	}
	return nil
}

// CopyClipboardEntry copies clipboard entry content back to the system clipboard.
func (s *ClipboardService) CopyClipboardEntry(ctx context.Context, req *CopyClipboardEntryReq) error {
	id, err := s.parseClipboardEntryID(req.ID)
	if err != nil {
		return err
	}
	entry, err := s.repo.GetClipboardEntryByID(ctx, id)
	if err != nil {
		return fmt.Errorf("[clipboard] get clipboard entry by ID error: %w", err)
	}
	if entry == nil {
		return ecode.ErrNotFound
	}
	if err := w32.SetClipboardText(entry.Content); err != nil {
		return fmt.Errorf("[clipboard] set clipboard text error: %w", err)
	}
	if err := s.repo.UpdateClipboardEntryLastUsedAt(ctx, entry.ID); err != nil {
		return fmt.Errorf("[clipboard] update clipboard entry last used time error: %w", err)
	}
	return nil
}

// SetClipboardEntryPinned pins or unpins a clipboard entry by the target state.
func (s *ClipboardService) SetClipboardEntryPinned(ctx context.Context, req *SetClipboardEntryPinnedReq) error {
	id, err := s.parseClipboardEntryID(req.ID)
	if err != nil {
		return err
	}
	entry, err := s.repo.GetClipboardEntryByID(ctx, id)
	if err != nil {
		return fmt.Errorf("[clipboard] get clipboard entry by ID error: %w", err)
	}
	if entry == nil {
		return ecode.ErrNotFound
	}
	// Return early on the same state, otherwise re-pinning would refresh the pinned order.
	if (entry.PinnedAt != nil) == req.Pinned {
		return nil
	}
	if !req.Pinned {
		if err := s.repo.UpdateClipboardEntryPinnedAt(ctx, id, nil); err != nil {
			return fmt.Errorf("[clipboard] clear clipboard entry pinned time error: %w", err)
		}
		return nil
	}

	pinned, err := s.repo.CountPinnedClipboardEntries(ctx)
	if err != nil {
		return fmt.Errorf("[clipboard] count pinned clipboard entries error: %w", err)
	}
	if pinned >= maxPinnedEntries {
		return ecode.ErrClipboardPinLimit
	}
	now := time.Now()
	if err := s.repo.UpdateClipboardEntryPinnedAt(ctx, id, &now); err != nil {
		return fmt.Errorf("[clipboard] set clipboard entry pinned time error: %w", err)
	}
	return nil
}

// DeleteClipboardEntry deletes clipboard entry by ID.
func (s *ClipboardService) DeleteClipboardEntry(ctx context.Context, req *DeleteClipboardEntryReq) error {
	id, err := s.parseClipboardEntryID(req.ID)
	if err != nil {
		return err
	}
	if err := s.repo.DeleteClipboardEntry(ctx, id); err != nil {
		return fmt.Errorf("[clipboard] delete clipboard entry error: %w", err)
	}
	return nil
}

// ClearClipboardEntries deletes all clipboard entries.
func (s *ClipboardService) ClearClipboardEntries(ctx context.Context) error {
	if err := s.repo.DeleteAllClipboardEntries(ctx); err != nil {
		return fmt.Errorf("[clipboard] delete all clipboard entries error: %w", err)
	}
	return nil
}

// ListClipboardEntries lists clipboard entries with keyword filter and pagination.
func (s *ClipboardService) ListClipboardEntries(ctx context.Context, req *ListClipboardEntriesReq) (*base.PageResult[ClipboardEntryItem], error) {
	query := req.Clamp()
	entries, total, err := s.repo.ListClipboardEntries(ctx, req.Keyword, query.Page, query.PageSize)
	if err != nil {
		return nil, fmt.Errorf("[clipboard] list clipboard entries error: %w", err)
	}

	list := make([]*ClipboardEntryItem, 0, len(entries))
	for _, entry := range entries {
		list = append(list, &ClipboardEntryItem{
			ID:         strconv.FormatInt(entry.ID, 10),
			Content:    entry.Content,
			Pinned:     entry.PinnedAt != nil,
			LastUsedAt: entry.LastUsedAt,
		})
	}
	return &base.PageResult[ClipboardEntryItem]{
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
		List:     list,
	}, nil
}

// parseClipboardEntryID parses the string clipboard entry ID sent by clients.
// Clients cannot pass the int64 snowflake ID as a JSON number without losing precision.
func (s *ClipboardService) parseClipboardEntryID(raw string) (int64, error) {
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, ecode.ErrNotFound
	}
	return id, nil
}
