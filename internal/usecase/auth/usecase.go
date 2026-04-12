package auth_usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	domain_error "taskflow/internal/domain/errors"
	domain_user "taskflow/internal/domain/user"
	"taskflow/utils"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	tokenTypeAccess  = "access"
	tokenTypeRefresh = "refresh"
	passwordHashCost = 12
)

type useCase struct {
	users  domain_user.Repository
	config *utils.Config
}

type tokenClaims struct {
	TokenType string `json:"token_type"`
	SessionID string `json:"session_id"`
	jwt.RegisteredClaims
}

func New(config *utils.Config, users domain_user.Repository) domain_user.UseCase {
	return &useCase{users: users, config: config}
}

func (uc *useCase) Register(ctx context.Context, input domain_user.RegisterInput, meta domain_user.SessionMetadata) (*domain_user.AuthResult, error) {
	existing, err := uc.users.GetByEmail(ctx, input.Email)
	if err != nil {
		return nil, domain_error.Raise(domain_error.CODE_INTERNAL_ERROR, "", err)
	}
	if existing != nil {
		return nil, domain_error.Raise(domain_error.CODE_EMAIL_TAKEN, "", nil)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), passwordHashCost)
	if err != nil {
		return nil, domain_error.Raise(domain_error.CODE_INTERNAL_ERROR, "", fmt.Errorf("hash password: %w", err))
	}

	user := &domain_user.User{
		ID:           uuid.New(),
		Name:         input.Name,
		Email:        input.Email,
		PasswordHash: string(hash),
		CreatedAt:    time.Now(),
	}

	if err := uc.users.Create(ctx, user); err != nil {
		return nil, domain_error.Raise(domain_error.CODE_INTERNAL_ERROR, "", err)
	}

	return uc.createSessionResult(ctx, user, meta)
}

func (uc *useCase) Login(ctx context.Context, input domain_user.LoginInput, meta domain_user.SessionMetadata) (*domain_user.AuthResult, error) {
	user, err := uc.users.GetByEmail(ctx, input.Email)
	if err != nil {
		return nil, domain_error.Raise(domain_error.CODE_INTERNAL_ERROR, "", err)
	}
	if user == nil {
		return nil, domain_error.Raise(domain_error.CODE_INVALID_CREDENTIALS, "", nil)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, domain_error.Raise(domain_error.CODE_INVALID_CREDENTIALS, "", nil)
	}

	return uc.createSessionResult(ctx, user, meta)
}

func (uc *useCase) Refresh(ctx context.Context, refreshToken string, meta domain_user.SessionMetadata) (*domain_user.AuthResult, error) {
	claims, err := uc.parseToken(refreshToken, tokenTypeRefresh)
	if err != nil {
		return nil, domain_error.Raise(domain_error.CODE_INVALID_AUTH_TOKEN, "", err)
	}

	sessionID, err := uuid.Parse(claims.SessionID)
	if err != nil {
		return nil, domain_error.Raise(domain_error.CODE_INVALID_AUTH_TOKEN, "", err)
	}

	session, err := uc.users.GetSessionByID(ctx, sessionID)
	if err != nil {
		return nil, domain_error.Raise(domain_error.CODE_INTERNAL_ERROR, "", err)
	}
	if session == nil {
		return nil, domain_error.Raise(domain_error.CODE_INVALID_AUTH_TOKEN, "", nil)
	}

	if session.RevokedAt != nil {
		return nil, domain_error.Raise(domain_error.CODE_INVALID_AUTH_TOKEN, "", nil)
	}
	if time.Now().After(session.ExpiresAt) {
		return nil, domain_error.Raise(domain_error.CODE_INVALID_AUTH_TOKEN, "", nil)
	}
	if session.RefreshTokenHash != hashToken(refreshToken) {
		return nil, domain_error.Raise(domain_error.CODE_INVALID_AUTH_TOKEN, "", nil)
	}

	user, err := uc.users.GetByID(ctx, session.UserID)
	if err != nil {
		return nil, domain_error.Raise(domain_error.CODE_INTERNAL_ERROR, "", err)
	}
	if user == nil {
		return nil, domain_error.Raise(domain_error.CODE_INVALID_AUTH_TOKEN, "", nil)
	}

	now := time.Now()
	newSessionID := uuid.New()
	refreshTokenValue, refreshHash, err := uc.newRefreshToken(user.ID, newSessionID)
	if err != nil {
		return nil, domain_error.Raise(domain_error.CODE_INTERNAL_ERROR, "", err)
	}

	newSession := &domain_user.AuthSession{
		ID:               newSessionID,
		UserID:           user.ID,
		RefreshTokenHash: refreshHash,
		UserAgent:        stringPtr(meta.UserAgent),
		IPAddress:        meta.IPAddress,
		DeviceName:       meta.DeviceName,
		CreatedAt:        now,
		ExpiresAt:        now.Add(uc.config.RefreshTokenExpiration),
	}

	if err := uc.users.CreateSession(ctx, newSession); err != nil {
		return nil, domain_error.Raise(domain_error.CODE_INTERNAL_ERROR, "", err)
	}

	if err := uc.users.UpdateSessionLastUsedAt(ctx, session.ID, now); err != nil {
		return nil, domain_error.Raise(domain_error.CODE_INTERNAL_ERROR, "", err)
	}
	if err := uc.users.RevokeSession(ctx, session.ID, now, &newSessionID); err != nil {
		return nil, domain_error.Raise(domain_error.CODE_INTERNAL_ERROR, "", err)
	}

	accessToken, err := uc.newAccessToken(user.ID, newSessionID)
	if err != nil {
		return nil, domain_error.Raise(domain_error.CODE_INTERNAL_ERROR, "", err)
	}

	return &domain_user.AuthResult{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenValue,
		User:         user,
	}, nil
}

func (uc *useCase) Logout(ctx context.Context, sessionID uuid.UUID) error {
	now := time.Now()
	if err := uc.users.RevokeSession(ctx, sessionID, now, nil); err != nil {
		return domain_error.Raise(domain_error.CODE_INTERNAL_ERROR, "", err)
	}
	return nil
}

func (uc *useCase) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	if err := uc.users.RevokeAllSessionsForUser(ctx, userID, time.Now()); err != nil {
		return domain_error.Raise(domain_error.CODE_INTERNAL_ERROR, "", err)
	}
	return nil
}

func (uc *useCase) AuthenticateAccessToken(ctx context.Context, accessToken string) (*domain_user.AuthenticatedSession, error) {
	claims, err := uc.parseToken(accessToken, tokenTypeAccess)
	if err != nil {
		return nil, domain_error.Raise(domain_error.CODE_INVALID_AUTH_TOKEN, "", err)
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return nil, domain_error.Raise(domain_error.CODE_INVALID_AUTH_TOKEN, "", err)
	}
	sessionID, err := uuid.Parse(claims.SessionID)
	if err != nil {
		return nil, domain_error.Raise(domain_error.CODE_INVALID_AUTH_TOKEN, "", err)
	}

	session, err := uc.users.GetSessionByID(ctx, sessionID)
	if err != nil {
		return nil, domain_error.Raise(domain_error.CODE_INTERNAL_ERROR, "", err)
	}
	if session == nil || session.RevokedAt != nil || time.Now().After(session.ExpiresAt) {
		return nil, domain_error.Raise(domain_error.CODE_INVALID_AUTH_TOKEN, "", nil)
	}
	if session.UserID != userID {
		return nil, domain_error.Raise(domain_error.CODE_INVALID_AUTH_TOKEN, "", nil)
	}

	return &domain_user.AuthenticatedSession{
		UserID:    userID,
		SessionID: sessionID,
	}, nil
}

func (uc *useCase) createSessionResult(ctx context.Context, user *domain_user.User, meta domain_user.SessionMetadata) (*domain_user.AuthResult, error) {
	now := time.Now()
	sessionID := uuid.New()

	refreshToken, refreshHash, err := uc.newRefreshToken(user.ID, sessionID)
	if err != nil {
		return nil, err
	}

	session := &domain_user.AuthSession{
		ID:               sessionID,
		UserID:           user.ID,
		RefreshTokenHash: refreshHash,
		UserAgent:        stringPtr(meta.UserAgent),
		IPAddress:        meta.IPAddress,
		DeviceName:       meta.DeviceName,
		CreatedAt:        now,
		ExpiresAt:        now.Add(uc.config.RefreshTokenExpiration),
	}

	if err := uc.users.CreateSession(ctx, session); err != nil {
		return nil, domain_error.Raise(domain_error.CODE_INTERNAL_ERROR, "", err)
	}

	accessToken, err := uc.newAccessToken(user.ID, sessionID)
	if err != nil {
		return nil, domain_error.Raise(domain_error.CODE_INTERNAL_ERROR, "", err)
	}

	return &domain_user.AuthResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
	}, nil
}

func (uc *useCase) newAccessToken(userID, sessionID uuid.UUID) (string, error) {
	return uc.newToken(userID, sessionID, tokenTypeAccess, uc.config.AccessTokenExpiration)
}

func (uc *useCase) newRefreshToken(userID, sessionID uuid.UUID) (string, string, error) {
	token, err := uc.newToken(userID, sessionID, tokenTypeRefresh, uc.config.RefreshTokenExpiration)
	if err != nil {
		return "", "", err
	}
	return token, hashToken(token), nil
}

func (uc *useCase) newToken(userID, sessionID uuid.UUID, tokenType string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := tokenClaims{
		TokenType: tokenType,
		SessionID: sessionID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(uc.config.JWTSecret))
}

func (uc *useCase) parseToken(tokenStr, expectedType string) (*tokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &tokenClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(uc.config.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(*tokenClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}
	if claims.TokenType != expectedType {
		return nil, fmt.Errorf("invalid token type")
	}
	if claims.SessionID == "" {
		return nil, fmt.Errorf("missing session id")
	}
	return claims, nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func stringPtr(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}
