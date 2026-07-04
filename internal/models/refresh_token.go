package models

import (
	"time"

	"github.com/google/uuid"
)

// RefreshToken stores server-side refresh tokens for rotation and revocation.
type RefreshToken struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index"                       json:"user_id"`
	Token     string    `gorm:"uniqueIndex;not null"                           json:"-"`
	ExpiresAt time.Time `gorm:"not null"                                       json:"expires_at"`
	Revoked   bool      `gorm:"default:false"                                  json:"revoked"`
	CreatedAt time.Time `                                                      json:"created_at"`
}
