package application

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"recipes-desk/internal/modules/auth/application/dto"
	"recipes-desk/internal/modules/auth/application/ports/in"
	"recipes-desk/internal/modules/auth/domain"
)

var ErrSignToken = errors.New("failed to sign token")

// TokenService handles JWT token generation and validation.
type TokenService struct {
	refreshRepo domain.RefreshTokensRepository
	secret      []byte
	accessTTL   time.Duration
	refreshTTL  time.Duration
}

var _ in.TokenService = (*TokenService)(nil)

// NewTokenService creates a new TokenService.
func NewTokenService(
	refreshRepo domain.RefreshTokensRepository,
	secret string,
	accessTTL time.Duration,
	refreshTTL time.Duration,
) *TokenService {
	return &TokenService{
		refreshRepo: refreshRepo,
		secret:      []byte(secret),
		accessTTL:   accessTTL,
		refreshTTL:  refreshTTL,
	}
}

func (s *TokenService) GenerateAccessToken(userID string) (*dto.TokenResult, error) {
	tokenString, expiresAt, err := GenerateToken(
		s.secret,
		jwt.SigningMethodHS256,
		userID,
		s.accessTTL,
	)
	if err != nil {
		return nil, err
	}

	return &dto.TokenResult{
		Token:     tokenString,
		ExpiresAt: expiresAt,
	}, nil
}

func (s *TokenService) ValidateAccessToken(tokenString string) (*dto.Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&dto.Claims{}, //nolint:exhaustruct // JWT library initializes fields
		func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return s.secret, nil
		})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, domain.ErrExpiredToken
		}
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidToken, err)
	}

	if claims, ok := token.Claims.(*dto.Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, domain.ErrInvalidToken
}

func (s *TokenService) GenerateRefreshToken(
	ctx context.Context,
	userID string,
) (*dto.TokenResult, error) {
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return nil, fmt.Errorf("failed to generate random token: %w", err)
	}

	plainToken := base64.URLEncoding.EncodeToString(randomBytes)
	tokenHash := hashToken(plainToken)

	refreshToken := &domain.RefreshToken{ //nolint:exhaustruct // fields set via SetTimestamps
		UserID:    userID,
		TokenHash: tokenHash,
	}

	refreshToken, err := s.refreshRepo.Create(ctx, refreshToken, s.refreshTTL)
	if err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return &dto.TokenResult{
		Token:     plainToken,
		ExpiresAt: refreshToken.ExpiresAt,
	}, nil
}

func (s *TokenService) ValidateRefreshToken(
	ctx context.Context,
	plainToken string,
) (string, error) {
	tokenHash := hashToken(plainToken)

	storedToken, err := s.refreshRepo.FindByHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, domain.ErrTokenNotFound) {
			return "", domain.ErrTokenNotFound
		}
		return "", fmt.Errorf("failed to find refresh token: %w", err)
	}

	// Check if token has expired
	if time.Now().After(storedToken.ExpiresAt) {
		return "", domain.ErrExpiredToken
	}

	return storedToken.UserID, nil
}

func (s *TokenService) RotateRefreshToken(
	ctx context.Context,
	oldPlainToken string,
) (*dto.TokenResult, string, error) {
	userID, err := s.ValidateRefreshToken(ctx, oldPlainToken)
	if err != nil {
		return nil, "", err
	}

	oldTokenHash := hashToken(oldPlainToken)
	if err = s.refreshRepo.DeleteByHash(ctx, oldTokenHash); err != nil {
		return nil, "", fmt.Errorf(
			"failed to delete old refresh token: %w",
			err,
		)
	}

	newToken, err := s.GenerateRefreshToken(ctx, userID)
	if err != nil {
		return nil, "", err
	}

	return newToken, userID, nil
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return base64.StdEncoding.EncodeToString(hash[:])
}

func GenerateToken(
	secret any,
	method jwt.SigningMethod,
	userID string,
	ttl time.Duration,
) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(ttl)
	claims := dto.Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{ //nolint:exhaustruct // only need ExpiresAt, IssuedAt, NotBefore
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(method, claims)
	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("%w: %w", ErrSignToken, err)
	}

	return tokenString, expiresAt, nil
}
