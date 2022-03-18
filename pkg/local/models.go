package local

import (
	"gorm.io/gorm"
	"time"
)

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&SavedCoub{},
		&ProfileCoub{},
	)
}

type SavedCoub struct {
	CoubID    int `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Info      []byte `gorm:"type:jsonb;not null"`
	NoAudio   bool   `gorm:"not null"`
}

type ProfileCoub struct {
	gorm.Model
	Profile     string    `gorm:"not null;index:idx_prof_coub,unique"`
	CoubID      int       `gorm:"not null;index:idx_prof_coub,unique"`
	PublishedAt time.Time `gorm:"not null"`
	Info        []byte    `gorm:"type:jsonb;not null"`
}
