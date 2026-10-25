package base

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/Wexyuan/kairos/internal/constant"
	"github.com/Wexyuan/kairos/pkg/cache"
	"github.com/Wexyuan/kairos/pkg/database"
	"github.com/Wexyuan/kairos/pkg/id"
)

// Repository provides shared data access methods.
type Repository struct {
	db *database.Database // underlying database
	sf *id.Snowflake      // snowflake ID generator
	c  cache.Cache        // shared cache
}

func NewRepository(db *database.Database, sf *id.Snowflake, c cache.Cache) *Repository {
	return &Repository{
		db: db,
		sf: sf,
		c:  c,
	}
}

// GenerateID generates a unique ID.
func (r *Repository) GenerateID() int64 {
	return r.sf.NextID()
}

// DB returns the underlying database.
func (r *Repository) DB(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(constant.TxKey).(*gorm.DB); ok {
		return tx
	}
	return r.db.WithContext(ctx)
}

// Transaction runs transaction with the given function.
func (r *Repository) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(context.WithValue(ctx, constant.TxKey, tx))
	})
	if err != nil {
		return fmt.Errorf("[base] run transaction error: %w", err)
	}
	return nil
}

// Cache returns the shared cache.
func (r *Repository) Cache() cache.Cache {
	return r.c
}

// EscapeLike escapes LIKE wildcards in the keyword.
func (r *Repository) EscapeLike(keyword string) string {
	replacer := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return replacer.Replace(keyword)
}
