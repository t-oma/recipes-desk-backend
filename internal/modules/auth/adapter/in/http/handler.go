package httphandler

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"recipes-desk/internal/modules/auth/application"
	"recipes-desk/internal/modules/auth/application/dto"
	"recipes-desk/internal/modules/auth/application/ports/in"
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

// handleError maps application errors to HTTP status codes.
func handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, application.ErrValidation):
		c.JSON(http.StatusBadRequest, err.Error())
	case errors.Is(err, application.ErrUserNotFound):
		c.JSON(http.StatusNotFound, err.Error())
	case errors.Is(err, application.ErrUnauthorized):
		c.JSON(http.StatusUnauthorized, err.Error())
	case errors.Is(err, application.ErrUserAlreadyExists):
		c.JSON(http.StatusConflict, gin.H{"error": "user already exists"})
	case errors.Is(err, application.ErrConflict):
		c.JSON(http.StatusConflict, err.Error())
	case errors.Is(err, application.ErrInvalidCredentials):
		c.JSON(http.StatusBadRequest, err.Error())

	case errors.Is(err, application.ErrServiceUnavailable):
		c.JSON(http.StatusServiceUnavailable, err.Error())
	default:
		c.JSON(http.StatusInternalServerError, err.Error())
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

	c.JSON(http.StatusOK, toUserResponse(user))
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
