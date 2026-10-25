package clipboard

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/Wexyuan/kairos/internal/module/base"
)

// ClipboardRepository defines clipboard entry data operations.
type ClipboardRepository interface {
	// CreateClipboardEntry generates ID and creates clipboard entry in database.
	CreateClipboardEntry(ctx context.Context, entry *ClipboardEntry) error
	// DeleteClipboardEntry deletes clipboard entry by ID.
	DeleteClipboardEntry(ctx context.Context, id int64) error
	// DeleteAllClipboardEntries deletes all clipboard entries.
	DeleteAllClipboardEntries(ctx context.Context) error
	// DeleteOldestClipboardEntries deletes the oldest clipboard entries by last used time.
	DeleteOldestClipboardEntries(ctx context.Context, limit int) error
	// UpdateClipboardEntryLastUsedAt updates last used time of a clipboard entry.
	UpdateClipboardEntryLastUsedAt(ctx context.Context, id int64) error
	// GetClipboardEntryByID retrieves clipboard entry by ID, returns nil if not found.
	GetClipboardEntryByID(ctx context.Context, id int64) (*ClipboardEntry, error)
	// GetClipboardEntryByHash retrieves clipboard entry by content hash, returns nil if not found.
	GetClipboardEntryByHash(ctx context.Context, hash string) (*ClipboardEntry, error)
	// CountClipboardEntries counts clipboard entries.
	CountClipboardEntries(ctx context.Context) (int64, error)
	// ListClipboardEntries lists clipboard entries with keyword filter and pagination.
	ListClipboardEntries(ctx context.Context, keyword string, page, pageSize int) ([]*ClipboardEntry, int64, error)
}

// clipboardRepository implements ClipboardRepository interface.
type clipboardRepository struct {
	*base.Repository
}

func NewClipboardRepository(repository *base.Repository) ClipboardRepository {
	return &clipboardRepository{
		Repository: repository,
	}
}

// CreateClipboardEntry generates ID and creates clipboard entry in database.
func (r *clipboardRepository) CreateClipboardEntry(ctx context.Context, entry *ClipboardEntry) error {
	entry.ID = r.GenerateID()
	return r.DB(ctx).Create(entry).Error
}

// DeleteClipboardEntry deletes clipboard entry by ID.
func (r *clipboardRepository) DeleteClipboardEntry(ctx context.Context, id int64) error {
	return r.DB(ctx).Delete(&ClipboardEntry{}, id).Error
}

// DeleteAllClipboardEntries deletes all clipboard entries.
func (r *clipboardRepository) DeleteAllClipboardEntries(ctx context.Context) error {
	return r.DB(ctx).Where("1 = 1").Delete(&ClipboardEntry{}).Error
}

// DeleteOldestClipboardEntries deletes the oldest clipboard entries by last used time.
func (r *clipboardRepository) DeleteOldestClipboardEntries(ctx context.Context, limit int) error {
	oldest := r.DB(ctx).Model(&ClipboardEntry{}).Select("id").Order("last_used_at ASC").Limit(limit)
	return r.DB(ctx).Where("id IN (?)", oldest).Delete(&ClipboardEntry{}).Error
}

// UpdateClipboardEntryLastUsedAt updates last used time of a clipboard entry.
func (r *clipboardRepository) UpdateClipboardEntryLastUsedAt(ctx context.Context, id int64) error {
	return r.DB(ctx).Model(&ClipboardEntry{}).Where("id = ?", id).Update("last_used_at", time.Now()).Error
}

// GetClipboardEntryByID retrieves clipboard entry by ID, returns nil if not found.
func (r *clipboardRepository) GetClipboardEntryByID(ctx context.Context, id int64) (*ClipboardEntry, error) {
	var entry ClipboardEntry
	err := r.DB(ctx).Where("id = ?", id).First(&entry).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &entry, nil
}

// GetClipboardEntryByHash retrieves clipboard entry by content hash, returns nil if not found.
func (r *clipboardRepository) GetClipboardEntryByHash(ctx context.Context, hash string) (*ClipboardEntry, error) {
	var entry ClipboardEntry
	err := r.DB(ctx).Where("hash = ?", hash).First(&entry).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &entry, nil
}

// CountClipboardEntries counts clipboard entries.
func (r *clipboardRepository) CountClipboardEntries(ctx context.Context) (int64, error) {
	var total int64
	if err := r.DB(ctx).Model(&ClipboardEntry{}).Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// ListClipboardEntries lists clipboard entries with keyword filter and pagination.
func (r *clipboardRepository) ListClipboardEntries(ctx context.Context, keyword string, page, pageSize int) ([]*ClipboardEntry, int64, error) {
	var entries []*ClipboardEntry
	var total int64

	db := r.DB(ctx).Model(&ClipboardEntry{})
	if keyword != "" {
		db = db.Where("content LIKE ? ESCAPE '\\'", "%"+r.EscapeLike(keyword)+"%")
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := db.Order("last_used_at DESC").Offset(offset).Limit(pageSize).Find(&entries).Error; err != nil {
		return nil, 0, err
	}

	return entries, total, nil
}
