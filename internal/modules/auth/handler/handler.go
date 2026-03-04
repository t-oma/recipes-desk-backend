package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"recipes-desk/internal/modules/auth/application/dto"
	"recipes-desk/internal/modules/auth/application/ports/in"
	"recipes-desk/internal/modules/auth/domain"
)

const (
	accessTokenCookieName  = "access_token"
	refreshTokenCookieName = "refresh_token"
	cookiePath             = "/"
)

// Handler handles HTTP requests for authentication.
type Handler struct {
	service in.AuthService
	log     *zerolog.Logger
}

// NewHandler creates a new auth handler.
func NewHandler(service in.AuthService, log *zerolog.Logger) *Handler {
	return &Handler{
		service: service,
		log:     log,
	}
}

func handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
	case errors.Is(err, domain.ErrAlreadyExists):
		c.JSON(http.StatusConflict, gin.H{"error": "user already exists"})
	case errors.Is(err, domain.ErrValidation):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrInvalidToken), errors.Is(err, domain.ErrExpiredToken),
		errors.Is(err, domain.ErrTokenNotFound):
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}

func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Debug().Err(err).Msg("Invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.service.Register(c.Request.Context(), dto.RegisterInput{
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Password:  req.Password,
	})
	if err != nil {
		handleError(c, err)
		return
	}

	// Set cookies
	setAuthCookie(c, accessTokenCookieName, result.AccessToken, result.AccessExpiresAt)
	setAuthCookie(c, refreshTokenCookieName, result.RefreshToken, result.RefreshExpiresAt)

	c.JSON(http.StatusCreated, toAuthResponse(result))
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Debug().Err(err).Msg("Invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.log.Debug().Str("user_email", req.Email).Msg("Logging in user")
	result, err := h.service.Login(c.Request.Context(), dto.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		handleError(c, err)
		return
	}

	// Set cookies
	setAuthCookie(c, accessTokenCookieName, result.AccessToken, result.AccessExpiresAt)
	setAuthCookie(c, refreshTokenCookieName, result.RefreshToken, result.RefreshExpiresAt)

	c.JSON(http.StatusOK, toAuthResponse(result))
}

func (h *Handler) Refresh(c *gin.Context) {
	refreshToken, err := c.Cookie(refreshTokenCookieName)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token required"})
		return
	}

	result, err := h.service.RefreshTokens(c.Request.Context(), dto.RefreshTokensInput{
		RefreshToken: refreshToken,
	})
	if err != nil {
		handleError(c, err)
		return
	}

	// Set new cookies
	setAuthCookie(c, accessTokenCookieName, result.AccessToken, result.AccessExpiresAt)
	setAuthCookie(c, refreshTokenCookieName, result.RefreshToken, result.RefreshExpiresAt)

	c.JSON(http.StatusOK, nil)
}

// Logout handles user logout by clearing the authentication cookies.
func (h *Handler) Logout(c *gin.Context) {
	_ = c.GetString("userID")

	clearAuthCookie(c, accessTokenCookieName)
	clearAuthCookie(c, refreshTokenCookieName)
	c.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
}

// Me returns the current authenticated user.
func (h *Handler) Me(c *gin.Context) {
	userID := c.GetString("userID")

	user, err := h.service.GetByID(c.Request.Context(), userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, user)
}

// setAuthCookie sets an httpOnly cookie with the given name and token.
func setAuthCookie(c *gin.Context, name, token string, expiresAt time.Time) {
	maxAge := int(time.Until(expiresAt).Seconds())
	if maxAge < 0 {
		maxAge = 0
	}

	cookie := &http.Cookie{
		Name:     name,
		Value:    token,
		Path:     cookiePath,
		MaxAge:   maxAge,
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	}

	c.SetCookieData(cookie)
}

// clearAuthCookie clears the authentication cookie.
func clearAuthCookie(c *gin.Context, name string) {
	c.SetCookie(name, "", -1, cookiePath, "", false, true)
}
