package auth

import (
	"context"

	domain_user "taskflow/internal/domain/user"

	"github.com/google/uuid"
)

type Stub struct {
	RegisterFn     func(domain_user.RegisterInput, domain_user.SessionMetadata) (*domain_user.AuthResult, error)
	LoginFn        func(domain_user.LoginInput, domain_user.SessionMetadata) (*domain_user.AuthResult, error)
	RefreshFn      func(string, domain_user.SessionMetadata) (*domain_user.AuthResult, error)
	LogoutFn       func(uuid.UUID) error
	LogoutAllFn    func(uuid.UUID) error
	AuthenticateFn func(string) (*domain_user.AuthenticatedSession, error)
}

func NewStub() *Stub {
	userID := uuid.New()
	sessionID := uuid.New()

	return &Stub{
		RegisterFn: func(input domain_user.RegisterInput, meta domain_user.SessionMetadata) (*domain_user.AuthResult, error) {
			return &domain_user.AuthResult{
				AccessToken:  "access-token",
				RefreshToken: "refresh-token",
				User: &domain_user.User{
					ID:    userID,
					Name:  input.Name,
					Email: input.Email,
				},
			}, nil
		},
		LoginFn: func(input domain_user.LoginInput, meta domain_user.SessionMetadata) (*domain_user.AuthResult, error) {
			return &domain_user.AuthResult{
				AccessToken:  "access-token",
				RefreshToken: "refresh-token",
				User: &domain_user.User{
					ID:    userID,
					Name:  "Jane",
					Email: input.Email,
				},
			}, nil
		},
		RefreshFn: func(refreshToken string, meta domain_user.SessionMetadata) (*domain_user.AuthResult, error) {
			return &domain_user.AuthResult{
				AccessToken:  "new-access-token",
				RefreshToken: "new-refresh-token",
				User: &domain_user.User{
					ID:   userID,
					Name: "Jane",
				},
			}, nil
		},
		LogoutFn:    func(sessionID uuid.UUID) error { return nil },
		LogoutAllFn: func(userID uuid.UUID) error { return nil },
		AuthenticateFn: func(accessToken string) (*domain_user.AuthenticatedSession, error) {
			return &domain_user.AuthenticatedSession{UserID: userID, SessionID: sessionID}, nil
		},
	}
}

func (s *Stub) Register(ctx context.Context, input domain_user.RegisterInput, meta domain_user.SessionMetadata) (*domain_user.AuthResult, error) {
	return s.RegisterFn(input, meta)
}

func (s *Stub) Login(ctx context.Context, input domain_user.LoginInput, meta domain_user.SessionMetadata) (*domain_user.AuthResult, error) {
	return s.LoginFn(input, meta)
}

func (s *Stub) Refresh(ctx context.Context, refreshToken string, meta domain_user.SessionMetadata) (*domain_user.AuthResult, error) {
	return s.RefreshFn(refreshToken, meta)
}

func (s *Stub) Logout(ctx context.Context, sessionID uuid.UUID) error {
	return s.LogoutFn(sessionID)
}

func (s *Stub) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	return s.LogoutAllFn(userID)
}

func (s *Stub) AuthenticateAccessToken(ctx context.Context, accessToken string) (*domain_user.AuthenticatedSession, error) {
	return s.AuthenticateFn(accessToken)
}
