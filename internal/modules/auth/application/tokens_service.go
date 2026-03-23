package application

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog"

	"recipes-desk/internal/modules/auth/application/dto"
	"recipes-desk/internal/modules/auth/application/ports/in"
	"recipes-desk/internal/modules/auth/domain"
	"recipes-desk/internal/modules/auth/domain/entity"
	"recipes-desk/internal/modules/auth/domain/ports"
	"recipes-desk/internal/modules/auth/domain/valueobject"
)

// TokenService handles JWT token generation and validation.
type TokenService struct {
	refreshRepo ports.RefreshTokenRepository
	idGen       ports.IDGenerator
	uow         ports.UnitOfWork
	log         *zerolog.Logger
	secret      []byte
	accessTTL   time.Duration
	refreshTTL  time.Duration
}

var _ in.TokenService = (*TokenService)(nil)

// NewTokenService creates a new TokenService.
func NewTokenService(
	refreshRepo ports.RefreshTokenRepository,
	idGen ports.IDGenerator,
	uow ports.UnitOfWork,
	log *zerolog.Logger,
	secret string,
	accessTTL time.Duration,
	refreshTTL time.Duration,
) *TokenService {
	return &TokenService{
		refreshRepo: refreshRepo,
		idGen:       idGen,
		uow:         uow,
		log:         log,
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
		return nil, s.mapError(err, "GenerateAccessToken")
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
				return nil, ErrTokenInvalid
			}
			return s.secret, nil
		})
	if err != nil {
		return nil, s.mapError(err, "ValidateAccessToken")
	}

	if claims, ok := token.Claims.(*dto.Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, domain.ErrTokenInvalid
}

func (s *TokenService) GenerateRefreshToken(
	ctx context.Context,
	userID string,
) (*dto.TokenResult, error) {
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return nil, s.mapError(err, "GenerateRefreshToken")
	}

	plainToken := base64.URLEncoding.EncodeToString(randomBytes)
	tokenHash := hashToken(plainToken)

	userIDVO, err := valueobject.NewUserID(userID)
	if err != nil {
		return nil, err
	}

	tokenIDVO, err := valueobject.NewRefreshTokenID(s.idGen.Generate())
	if err != nil {
		return nil, err
	}
	tokenHashVO := valueobject.TokenHash(tokenHash)
	refreshToken := entity.NewRefreshToken(tokenIDVO, userIDVO, tokenHashVO)

	refreshToken, err = s.refreshRepo.Create(ctx, refreshToken, s.refreshTTL)
	if err != nil {
		return nil, s.mapError(err, "GenerateRefreshToken")
	}

	return &dto.TokenResult{
		Token:     plainToken,
		ExpiresAt: refreshToken.ExpiresAt(),
	}, nil
}

func (s *TokenService) ValidateRefreshToken(
	ctx context.Context,
	plainToken string,
) (string, error) {
	tokenHash := hashToken(plainToken)

	storedToken, err := s.refreshRepo.FindByHash(ctx, tokenHash)
	if err != nil {
		return "", s.mapError(err, "ValidateRefreshToken")
	}

	// Check if token has expired
	if time.Now().After(storedToken.ExpiresAt()) {
		return "", domain.ErrTokenExpired
	}

	return storedToken.UserID().String(), nil
}

func (s *TokenService) RotateRefreshToken(
	ctx context.Context,
	oldPlainToken string,
) (*dto.TokenResult, string, error) {
	var userID string
	var newToken *dto.TokenResult
	err := s.uow.Execute(ctx, func(sesCtx context.Context) error {
		var err error
		userID, err = s.ValidateRefreshToken(sesCtx, oldPlainToken)
		if err != nil {
			return err
		}

		oldTokenHash := hashToken(oldPlainToken)
		if err = s.refreshRepo.DeleteByHash(sesCtx, oldTokenHash); err != nil {
			return err
		}

		newToken, err = s.GenerateRefreshToken(sesCtx, userID)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, "", s.mapError(err, "RotateRefreshToken")
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
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

// mapError maps domain and infrastructure errors to application-level errors.
// This ensures that only safe, client-facing errors are exposed.
func (s *TokenService) mapError(err error, operation string) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, jwt.ErrTokenExpired):
		return ErrTokenExpired
	case errors.Is(err, jwt.ErrTokenNotValidYet),
		errors.Is(err, jwt.ErrTokenMalformed),
		errors.Is(err, jwt.ErrTokenInvalidIssuer),
		errors.Is(err, jwt.ErrTokenInvalidAudience),
		errors.Is(err, jwt.ErrTokenInvalidSubject),
		errors.Is(err, jwt.ErrTokenUnverifiable),
		errors.Is(err, jwt.ErrTokenSignatureInvalid):
		s.log.Error().Err(err).Str("operation", operation).Msg("Invalid token")
		return ErrTokenInvalid

	// Domain errors that are safe to pass through
	case errors.Is(err, domain.ErrValidation):
		return err
	case errors.Is(err, domain.ErrNotFound):
		return err
	case errors.Is(err, domain.ErrForbidden):
		return err
	case errors.Is(err, domain.ErrConflict):
		return err
	case errors.Is(err, domain.ErrUnauthorized):
		return err

	// Infrastructure errors - map to safe versions
	case errors.Is(err, jwt.ErrInvalidKeyType):
		s.log.Error().Err(err).Str("operation", operation).Msg("Invalid key type")
		return ErrInternal
	case errors.Is(err, domain.ErrTimeout):
		s.log.Error().Err(err).Str("operation", operation).Msg("Database timeout")
		return ErrServiceUnavailable
	case errors.Is(err, domain.ErrDatabase):
		s.log.Error().Err(err).Str("operation", operation).Msg("Database error")
		return ErrInternal
	default:
		s.log.Error().Err(err).Str("operation", operation).Msg("Unexpected error")
		return ErrInternal
	}
}
