package services

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/pubudulakmal/auth-api/internal/models"
	"github.com/pubudulakmal/auth-api/internal/repository"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// TokenPair holds both tokens returned to the client on successful auth.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// AuthService contains the core authentication business logic.
type AuthService struct {
	userRepo   *repository.UserRepository
	tokenRepo  *repository.TokenRepository
	jwtService *JWTService
}

// NewAuthService constructs an AuthService with all required dependencies.
func NewAuthService(
	userRepo *repository.UserRepository,
	tokenRepo *repository.TokenRepository,
	jwtService *JWTService,
) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		tokenRepo:  tokenRepo,
		jwtService: jwtService,
	}
}

// Register creates a new user account and returns a token pair.
// Returns an error if the email is already in use.
func (s *AuthService) Register(name, email, password string) (*TokenPair, *models.User, error) {
	// Check for duplicate email
	if _, err := s.userRepo.FindByEmail(email); err == nil {
		return nil, nil, errors.New("email already registered")
	}

	// Hash password with bcrypt cost 12
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return nil, nil, err
	}

	user := &models.User{
		ID:           uuid.New(),
		Name:         name,
		Email:        email,
		PasswordHash: string(hash),
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, nil, err
	}

	pair, err := s.issueTokenPair(user.ID, user.Email)
	if err != nil {
		return nil, nil, err
	}

	return pair, user, nil
}

// Login verifies credentials and issues a fresh token pair.
// All existing refresh tokens for the user are revoked before issuing new ones
// to prevent session proliferation.
func (s *AuthService) Login(email, password string) (*TokenPair, *models.User, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		// Return a generic error to prevent email enumeration
		return nil, nil, errors.New("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, nil, errors.New("invalid email or password")
	}

	// Revoke all prior refresh tokens before issuing a new pair
	_ = s.tokenRepo.RevokeAllForUser(user.ID)

	pair, err := s.issueTokenPair(user.ID, user.Email)
	if err != nil {
		return nil, nil, err
	}

	return pair, user, nil
}

// Refresh validates the provided refresh token, rotates it, and returns a new token pair.
// The old refresh token is immediately revoked (rotation strategy).
func (s *AuthService) Refresh(refreshTokenStr string) (*TokenPair, error) {
	rt, err := s.tokenRepo.FindByToken(refreshTokenStr)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("refresh token not found")
		}
		return nil, err
	}

	if rt.Revoked {
		return nil, errors.New("refresh token has been revoked")
	}

	if time.Now().After(rt.ExpiresAt) {
		return nil, errors.New("refresh token has expired")
	}

	// Revoke the used token (rotation)
	if err := s.tokenRepo.Revoke(refreshTokenStr); err != nil {
		return nil, err
	}

	// Fetch the user to embed fresh claims
	user, err := s.userRepo.FindByID(rt.UserID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	pair, err := s.issueTokenPair(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	return pair, nil
}

// Logout revokes the provided refresh token, ending that session.
func (s *AuthService) Logout(refreshTokenStr string) error {
	rt, err := s.tokenRepo.FindByToken(refreshTokenStr)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("refresh token not found")
		}
		return err
	}

	if rt.Revoked {
		return errors.New("refresh token already revoked")
	}

	return s.tokenRepo.Revoke(refreshTokenStr)
}

// issueTokenPair is a helper that generates both tokens and persists the refresh token.
func (s *AuthService) issueTokenPair(userID uuid.UUID, email string) (*TokenPair, error) {
	accessToken, err := s.jwtService.GenerateAccessToken(userID, email)
	if err != nil {
		return nil, err
	}

	refreshTokenStr, expiresAt, err := s.jwtService.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	rt := &models.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		Token:     refreshTokenStr,
		ExpiresAt: expiresAt,
	}

	if err := s.tokenRepo.Save(rt); err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenStr,
	}, nil
}
