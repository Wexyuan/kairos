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
	// CreateEntry generates ID and creates entry in database.
	CreateEntry(ctx context.Context, entry *ClipboardEntry) error
	// DeleteEntry deletes entry by ID.
	DeleteEntry(ctx context.Context, id int64) error
	// DeleteAllEntries deletes all entries.
	DeleteAllEntries(ctx context.Context) error
	// DeleteOldestEntries deletes the oldest entries by last used time.
	DeleteOldestEntries(ctx context.Context, limit int) error
	// UpdateEntryLastUsedAt updates last used time of an entry.
	UpdateEntryLastUsedAt(ctx context.Context, id int64) error
	// GetEntryByID retrieves entry by ID, returns nil if not found.
	GetEntryByID(ctx context.Context, id int64) (*ClipboardEntry, error)
	// GetEntryByHash retrieves entry by content hash, returns nil if not found.
	GetEntryByHash(ctx context.Context, hash string) (*ClipboardEntry, error)
	// CountEntries returns the total number of entries.
	CountEntries(ctx context.Context) (int64, error)
	// ListEntries retrieves entries with keyword filter and pagination.
	ListEntries(ctx context.Context, keyword string, page, pageSize int) ([]*ClipboardEntry, int64, error)
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

// CreateEntry generates ID and creates entry in database.
func (r *clipboardRepository) CreateEntry(ctx context.Context, entry *ClipboardEntry) error {
	entry.ID = r.GenerateID()
	return r.DB(ctx).Create(entry).Error
}

// DeleteEntry deletes entry by ID.
func (r *clipboardRepository) DeleteEntry(ctx context.Context, id int64) error {
	return r.DB(ctx).Delete(&ClipboardEntry{}, id).Error
}

// DeleteAllEntries deletes all entries.
func (r *clipboardRepository) DeleteAllEntries(ctx context.Context) error {
	return r.DB(ctx).Where("1 = 1").Delete(&ClipboardEntry{}).Error
}

// DeleteOldestEntries deletes the oldest entries by last used time.
func (r *clipboardRepository) DeleteOldestEntries(ctx context.Context, limit int) error {
	oldest := r.DB(ctx).Model(&ClipboardEntry{}).Select("id").Order("last_used_at ASC").Limit(limit)
	return r.DB(ctx).Where("id IN (?)", oldest).Delete(&ClipboardEntry{}).Error
}

// UpdateEntryLastUsedAt updates last used time of an entry.
func (r *clipboardRepository) UpdateEntryLastUsedAt(ctx context.Context, id int64) error {
	return r.DB(ctx).Model(&ClipboardEntry{}).Where("id = ?", id).Update("last_used_at", time.Now()).Error
}

// GetEntryByID retrieves entry by ID, returns nil if not found.
func (r *clipboardRepository) GetEntryByID(ctx context.Context, id int64) (*ClipboardEntry, error) {
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

// GetEntryByHash retrieves entry by content hash, returns nil if not found.
func (r *clipboardRepository) GetEntryByHash(ctx context.Context, hash string) (*ClipboardEntry, error) {
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

// CountEntries returns the total number of entries.
func (r *clipboardRepository) CountEntries(ctx context.Context) (int64, error) {
	var total int64
	if err := r.DB(ctx).Model(&ClipboardEntry{}).Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// ListEntries retrieves entries with keyword filter and pagination.
func (r *clipboardRepository) ListEntries(ctx context.Context, keyword string, page, pageSize int) ([]*ClipboardEntry, int64, error) {
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
