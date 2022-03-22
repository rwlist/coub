package local

import (
	"time"

	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&SavedCoub{},
		&ProfileCoub{},
		&LikedCoub{},
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

type LikedCoub struct {
	gorm.Model
	Profile string `gorm:"not null;index:idx_liked_coub,unique"`
	CoubID  int    `gorm:"not null;index:idx_liked_coub,unique"`
	Info    []byte `gorm:"type:jsonb;not null"`
}
