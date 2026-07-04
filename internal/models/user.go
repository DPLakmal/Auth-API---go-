package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents a registered account in the system.
type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name         string    `gorm:"not null"                                       json:"name"`
	Email        string    `gorm:"uniqueIndex;not null"                           json:"email"`
	PasswordHash string    `gorm:"not null"                                       json:"-"` // never serialised
	CreatedAt    time.Time `                                                      json:"created_at"`
	UpdatedAt    time.Time `                                                      json:"updated_at"`
}
