package repository

import (
	"github.com/google/uuid"
	"github.com/pubudulakmal/auth-api/internal/models"
	"gorm.io/gorm"
)

// TokenRepository handles all database operations for refresh tokens.
type TokenRepository struct {
	db *gorm.DB
}

// NewTokenRepository creates a new TokenRepository.
func NewTokenRepository(db *gorm.DB) *TokenRepository {
	return &TokenRepository{db: db}
}

// Save persists a new refresh token record.
func (r *TokenRepository) Save(token *models.RefreshToken) error {
	return r.db.Create(token).Error
}

// FindByToken retrieves a refresh token record by its opaque token string.
func (r *TokenRepository) FindByToken(tokenStr string) (*models.RefreshToken, error) {
	var rt models.RefreshToken
	if err := r.db.Where("token = ?", tokenStr).First(&rt).Error; err != nil {
		return nil, err
	}
	return &rt, nil
}

// Revoke marks a single refresh token as revoked.
func (r *TokenRepository) Revoke(tokenStr string) error {
	return r.db.Model(&models.RefreshToken{}).
		Where("token = ?", tokenStr).
		Update("revoked", true).Error
}

// RevokeAllForUser marks all refresh tokens for a given user as revoked.
// Useful when changing passwords or forcing a global logout.
func (r *TokenRepository) RevokeAllForUser(userID uuid.UUID) error {
	return r.db.Model(&models.RefreshToken{}).
		Where("user_id = ? AND revoked = false", userID).
		Update("revoked", true).Error
}
