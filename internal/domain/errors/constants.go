package domain_error

import "net/http"

const SVC = "TKF"

// Error codes follow the SVC-DOMAIN-NN pattern.
const (
	// Core system
	CODE_UNKNOWN_ERROR  ErrorCode = "TKF-SYS-00"
	CODE_INTERNAL_ERROR ErrorCode = "TKF-SYS-01"

	// Authentication
	CODE_AUTH_TOKEN_MISSING  ErrorCode = "TKF-AUTH-00"
	CODE_INVALID_AUTH_TOKEN  ErrorCode = "TKF-AUTH-01"
	CODE_EMAIL_TAKEN         ErrorCode = "TKF-AUTH-02"
	CODE_INVALID_CREDENTIALS ErrorCode = "TKF-AUTH-03"

	// REST/API
	CODE_INVALID_PAYLOAD   ErrorCode = "TKF-REST-00"
	CODE_VALIDATION_FAILED ErrorCode = "TKF-REST-01"

	// Project domain
	CODE_PROJECT_NOT_FOUND ErrorCode = "TKF-PRJ-00"
	CODE_PROJECT_FORBIDDEN ErrorCode = "TKF-PRJ-01"

	// Task domain
	CODE_TASK_NOT_FOUND         ErrorCode = "TKF-TSK-00"
	CODE_TASK_FORBIDDEN         ErrorCode = "TKF-TSK-01"
	CODE_TASK_PROJECT_NOT_FOUND ErrorCode = "TKF-TSK-02"
)

var msgMap = map[ErrorCode]string{
	CODE_UNKNOWN_ERROR:  "unknown error occurred",
	CODE_INTERNAL_ERROR: "internal server error",

	CODE_AUTH_TOKEN_MISSING:  "authorization token is missing",
	CODE_INVALID_AUTH_TOKEN:  "invalid authorization token",
	CODE_EMAIL_TAKEN:         "email is already registered",
	CODE_INVALID_CREDENTIALS: "invalid email or password",

	CODE_INVALID_PAYLOAD:   "invalid request payload",
	CODE_VALIDATION_FAILED: "validation failed",

	CODE_PROJECT_NOT_FOUND: "project not found",
	CODE_PROJECT_FORBIDDEN: "you do not have permission to modify this project",

	CODE_TASK_NOT_FOUND:         "task not found",
	CODE_TASK_FORBIDDEN:         "you do not have permission to modify this task",
	CODE_TASK_PROJECT_NOT_FOUND: "project not found",
}

// ErrCodeToStatusMap binds error codes to HTTP status codes.
// Codes outside the map default to 400 Bad Request.
var ErrCodeToStatusMap = map[ErrorCode]int{
	CODE_UNKNOWN_ERROR:  http.StatusInternalServerError,
	CODE_INTERNAL_ERROR: http.StatusInternalServerError,

	CODE_AUTH_TOKEN_MISSING:  http.StatusUnauthorized,
	CODE_INVALID_AUTH_TOKEN:  http.StatusUnauthorized,
	CODE_EMAIL_TAKEN:         http.StatusConflict,
	CODE_INVALID_CREDENTIALS: http.StatusUnauthorized,

	CODE_INVALID_PAYLOAD:   http.StatusBadRequest,
	CODE_VALIDATION_FAILED: http.StatusBadRequest,

	CODE_PROJECT_NOT_FOUND: http.StatusNotFound,
	CODE_PROJECT_FORBIDDEN: http.StatusForbidden,

	CODE_TASK_NOT_FOUND:         http.StatusNotFound,
	CODE_TASK_FORBIDDEN:         http.StatusForbidden,
	CODE_TASK_PROJECT_NOT_FOUND: http.StatusNotFound,
}
