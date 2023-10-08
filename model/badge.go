package model

import "time"

type badge struct {
	Title       *string   `json:"title" gorm:"not null"`
	Description *string   `json:"description" gorm:"not null"`
	CreatedAt   time.Time `json:"created_at" gorm:"not null"`
}
