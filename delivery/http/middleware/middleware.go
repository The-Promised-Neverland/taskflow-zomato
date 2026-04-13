package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"taskflow/delivery/http/common"
	domain_error "taskflow/internal/domain/errors"
	domain_user "taskflow/internal/domain/user"
	"taskflow/utils/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestID creates a request ID and adds it to the response and context.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Writer.Header().Set("X-Request-ID", requestID)
		ctx := context.WithValue(c.Request.Context(), common.RequestIDKey, requestID)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

// CORS allows browser requests from the frontend and handles preflight checks.
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Vary", "Origin")
		}
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET,POST,PATCH,DELETE,OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Requested-With, X-Request-ID")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "X-Request-ID")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// Logger attaches a request-scoped logger and emits request/response logs.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		requestID, _ := c.Request.Context().Value(common.RequestIDKey).(string)

		log := logger.FromContext(c.Request.Context()).With(
			"request_id", requestID,
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"remote_addr", c.Request.RemoteAddr,
		)
		ctx := logger.NewContext(c.Request.Context(), log)
		c.Request = c.Request.WithContext(ctx)

		log.InfoContext(ctx, "request")
		c.Next()
		log.InfoContext(ctx, "response",
			"status", c.Writer.Status(),
			"duration", time.Since(start),
		)
	}
}

// Recovery converts panics into a 500 response.
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		logger.FromContext(c.Request.Context()).Error("panic recovered",
			"error", recovered,
		)
		c.JSON(http.StatusInternalServerError, gin.H{"code": "FAILURE", "message": "Internal server error"})
		c.Abort()
	})
}

// Authenticate checks the JWT bearer token and stores user and session IDs in context.
func Authenticate(auth domain_user.UseCase) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		requestID, _ := c.Request.Context().Value(common.RequestIDKey).(string)

		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			slog.WarnContext(c.Request.Context(), "missing or malformed authorization header",
				"path", c.Request.URL.Path,
				"request_id", requestID,
			)
			common.SendAppError(c.Writer, domain_error.Raise(domain_error.CODE_AUTH_TOKEN_MISSING, "", nil))
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		authSession, err := auth.AuthenticateAccessToken(c.Request.Context(), tokenStr)
		if err != nil {
			common.SendAppError(c.Writer, err)
			c.Abort()
			return
		}

		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, common.UserIDKey, authSession.UserID)
		ctx = context.WithValue(ctx, common.SessionIDKey, authSession.SessionID)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

// Pagination reads page and limit query params into request context.
func Pagination() gin.HandlerFunc {
	return func(c *gin.Context) {
		page := 1
		pageSize := 20

		if p := c.Query("page"); p != "" {
			if v, err := strconv.Atoi(p); err == nil && v > 0 {
				page = v
			}
		}
		if limit := c.Query("limit"); limit != "" {
			if v, err := strconv.Atoi(limit); err == nil && v > 0 && v <= 100 {
				pageSize = v
			}
		} else if ps := c.Query("page_size"); ps != "" {
			if v, err := strconv.Atoi(ps); err == nil && v > 0 && v <= 100 {
				pageSize = v
			}
		}

		ctx := context.WithValue(c.Request.Context(), common.PaginationKey, common.Pagination{Page: page, PageSize: pageSize})
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
