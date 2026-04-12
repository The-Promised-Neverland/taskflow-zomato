package domain_user

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

type RegisterInput struct {
	Name     string
	Email    string
	Password string
}

type LoginInput struct {
	Email    string
	Password string
}

type AuthResult struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         *User  `json:"user"`
}

type SessionMetadata struct {
	UserAgent  string
	IPAddress  *string
	DeviceName *string
}

type AuthSession struct {
	ID                  uuid.UUID  `json:"id"`
	UserID              uuid.UUID  `json:"user_id"`
	RefreshTokenHash    string     `json:"-"`
	UserAgent           *string    `json:"user_agent,omitempty"`
	IPAddress           *string    `json:"ip_address,omitempty"`
	DeviceName          *string    `json:"device_name,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	LastUsedAt          *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt           time.Time  `json:"expires_at"`
	RevokedAt           *time.Time `json:"revoked_at,omitempty"`
	ReplacedBySessionID *uuid.UUID `json:"replaced_by_session_id,omitempty"`
}

type AuthenticatedSession struct {
	UserID    uuid.UUID
	SessionID uuid.UUID
}

type Repository interface {
	Create(ctx context.Context, user *User) error
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	CreateSession(ctx context.Context, session *AuthSession) error
	GetSessionByID(ctx context.Context, id uuid.UUID) (*AuthSession, error)
	UpdateSessionLastUsedAt(ctx context.Context, id uuid.UUID, lastUsedAt time.Time) error
	RevokeSession(ctx context.Context, id uuid.UUID, revokedAt time.Time, replacedBySessionID *uuid.UUID) error
	RevokeAllSessionsForUser(ctx context.Context, userID uuid.UUID, revokedAt time.Time) error
}

type UseCase interface {
	Register(ctx context.Context, input RegisterInput, meta SessionMetadata) (*AuthResult, error)
	Login(ctx context.Context, input LoginInput, meta SessionMetadata) (*AuthResult, error)
	Refresh(ctx context.Context, refreshToken string, meta SessionMetadata) (*AuthResult, error)
	Logout(ctx context.Context, sessionID uuid.UUID) error
	LogoutAll(ctx context.Context, userID uuid.UUID) error
	AuthenticateAccessToken(ctx context.Context, accessToken string) (*AuthenticatedSession, error)
}
