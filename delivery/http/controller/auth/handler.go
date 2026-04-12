package auth_controller

import (
	"net"
	"net/http"
	"strings"

	"taskflow/delivery/http/common"
	domain_error "taskflow/internal/domain/errors"
	domain_user "taskflow/internal/domain/user"
	"taskflow/utils/validator"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Controller struct {
	auth domain_user.UseCase
}

func New(auth domain_user.UseCase) *Controller {
	return &Controller{auth: auth}
}

// Register handles user sign-up.
// POST /auth/register
func (c *Controller) Register(ctx *gin.Context) {
	var req registerRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.SendAppError(ctx.Writer, domain_error.Raise(domain_error.CODE_INVALID_PAYLOAD, "", err))
		return
	}
	if err := validator.Validate(ctx.Request.Context(), req); err != nil {
		common.SendAppError(ctx.Writer, err)
		return
	}

	result, err := c.auth.Register(ctx.Request.Context(), domain_user.RegisterInput{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	}, requestMetadata(ctx))
	if err != nil {
		common.SendAppError(ctx.Writer, err)
		return
	}

	common.SendJSON(ctx.Writer, http.StatusCreated, map[string]interface{}{
		"access_token":  result.AccessToken,
		"refresh_token": result.RefreshToken,
		"user":          result.User,
	})
}

// Login handles user sign-in.
// POST /auth/login
func (c *Controller) Login(ctx *gin.Context) {
	var req loginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.SendAppError(ctx.Writer, domain_error.Raise(domain_error.CODE_INVALID_PAYLOAD, "", err))
		return
	}
	if err := validator.Validate(ctx.Request.Context(), req); err != nil {
		common.SendAppError(ctx.Writer, err)
		return
	}

	result, err := c.auth.Login(ctx.Request.Context(), domain_user.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	}, requestMetadata(ctx))
	if err != nil {
		common.SendAppError(ctx.Writer, err)
		return
	}

	common.SendJSON(ctx.Writer, http.StatusOK, map[string]interface{}{
		"access_token":  result.AccessToken,
		"refresh_token": result.RefreshToken,
		"user":          result.User,
	})
}

// Refresh issues a new token pair from a refresh token.
// POST /auth/refresh
func (c *Controller) Refresh(ctx *gin.Context) {
	var req refreshRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.SendAppError(ctx.Writer, domain_error.Raise(domain_error.CODE_INVALID_PAYLOAD, "", err))
		return
	}
	if err := validator.Validate(ctx.Request.Context(), req); err != nil {
		common.SendAppError(ctx.Writer, err)
		return
	}

	result, err := c.auth.Refresh(ctx.Request.Context(), req.RefreshToken, requestMetadata(ctx))
	if err != nil {
		common.SendAppError(ctx.Writer, err)
		return
	}

	common.SendJSON(ctx.Writer, http.StatusOK, map[string]interface{}{
		"access_token":  result.AccessToken,
		"refresh_token": result.RefreshToken,
		"user":          result.User,
	})
}

// Logout revokes the active session.
// POST /auth/logout
func (c *Controller) Logout(ctx *gin.Context) {
	sessionID, ok := ctx.Request.Context().Value(common.SessionIDKey).(uuid.UUID)
	if !ok {
		common.SendAppError(ctx.Writer, domain_error.Raise(domain_error.CODE_AUTH_TOKEN_MISSING, "", nil))
		return
	}

	if err := c.auth.Logout(ctx.Request.Context(), sessionID); err != nil {
		common.SendAppError(ctx.Writer, err)
		return
	}

	ctx.Status(http.StatusNoContent)
}

// LogoutAll revokes every session for the user.
// POST /auth/logout-all
func (c *Controller) LogoutAll(ctx *gin.Context) {
	userID, ok := ctx.Request.Context().Value(common.UserIDKey).(uuid.UUID)
	if !ok {
		common.SendAppError(ctx.Writer, domain_error.Raise(domain_error.CODE_AUTH_TOKEN_MISSING, "", nil))
		return
	}

	if err := c.auth.LogoutAll(ctx.Request.Context(), userID); err != nil {
		common.SendAppError(ctx.Writer, err)
		return
	}

	ctx.Status(http.StatusNoContent)
}

func requestMetadata(ctx *gin.Context) domain_user.SessionMetadata {
	return domain_user.SessionMetadata{
		UserAgent:  ctx.Request.UserAgent(),
		IPAddress:  clientIP(ctx.Request),
		DeviceName: nil,
	}
}

func clientIP(r *http.Request) *string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		if len(parts) > 0 {
			ip := strings.TrimSpace(parts[0])
			if ip != "" {
				return &ip
			}
		}
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		addr := strings.TrimSpace(r.RemoteAddr)
		if addr == "" {
			return nil
		}
		return &addr
	}
	if host == "" {
		return nil
	}
	return &host
}
