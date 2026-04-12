package common

type ContextKey string

type Pagination struct {
	Page     int
	PageSize int
}

const (
	RequestIDKey  ContextKey = "request_id"
	UserIDKey     ContextKey = "user_id"
	SessionIDKey  ContextKey = "session_id"
	PaginationKey ContextKey = "pagination"
)
