package base

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel holds the fields shared by all models.
type BaseModel struct {
	ID        int64          `gorm:"primarykey;type:bigint" json:"id"` // primary key
	CreatedAt time.Time      `json:"created_at"`                       // create time
	UpdatedAt time.Time      `json:"updated_at"`                       // last update time
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`          // soft delete time
}
