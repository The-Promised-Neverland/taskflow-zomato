package user_repository

import (
	"context"
	"errors"
	"net/netip"
	"time"

	domain_user "taskflow/internal/domain/user"
	db "taskflow/internal/repository/user/driver/postgres"
	postgres "taskflow/utils/database/postgres"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type repository struct {
	q *db.Queries
}

func New(conn *postgres.DBConnector) domain_user.Repository {
	return &repository{q: db.New(conn.Pool)}
}

func (r *repository) Create(ctx context.Context, user *domain_user.User) error {
	return r.q.CreateUser(ctx, db.CreateUserParams{
		ID:           user.ID,
		Name:         user.Name,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		CreatedAt:    user.CreatedAt,
	})
}

func (r *repository) GetByEmail(ctx context.Context, email string) (*domain_user.User, error) {
	row, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return toUser(row), nil
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*domain_user.User, error) {
	row, err := r.q.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return toUser(row), nil
}

func (r *repository) CreateSession(ctx context.Context, session *domain_user.AuthSession) error {
	return r.q.CreateAuthSession(ctx, db.CreateAuthSessionParams{
		ID:                  session.ID,
		UserID:              session.UserID,
		RefreshTokenHash:    session.RefreshTokenHash,
		UserAgent:           session.UserAgent,
		IpAddress:           parseIP(session.IPAddress),
		DeviceName:          session.DeviceName,
		CreatedAt:           session.CreatedAt,
		LastUsedAt:          toTimestamptz(session.LastUsedAt),
		ExpiresAt:           session.ExpiresAt,
		RevokedAt:           toTimestamptz(session.RevokedAt),
		ReplacedBySessionID: session.ReplacedBySessionID,
	})
}

func (r *repository) GetSessionByID(ctx context.Context, id uuid.UUID) (*domain_user.AuthSession, error) {
	row, err := r.q.GetAuthSessionByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return toAuthSession(row), nil
}

func (r *repository) UpdateSessionLastUsedAt(ctx context.Context, id uuid.UUID, lastUsedAt time.Time) error {
	return r.q.UpdateAuthSessionLastUsedAt(ctx, db.UpdateAuthSessionLastUsedAtParams{
		ID:         id,
		LastUsedAt: pgtype.Timestamptz{Time: lastUsedAt, Valid: true},
	})
}

func (r *repository) RevokeSession(ctx context.Context, id uuid.UUID, revokedAt time.Time, replacedBySessionID *uuid.UUID) error {
	return r.q.RevokeAuthSession(ctx, db.RevokeAuthSessionParams{
		ID:                  id,
		RevokedAt:           pgtype.Timestamptz{Time: revokedAt, Valid: true},
		ReplacedBySessionID: replacedBySessionID,
	})
}

func (r *repository) RevokeAllSessionsForUser(ctx context.Context, userID uuid.UUID, revokedAt time.Time) error {
	return r.q.RevokeAllAuthSessionsForUser(ctx, db.RevokeAllAuthSessionsForUserParams{
		UserID:    userID,
		RevokedAt: pgtype.Timestamptz{Time: revokedAt, Valid: true},
	})
}

func toUser(row db.User) *domain_user.User {
	return &domain_user.User{
		ID:           row.ID,
		Name:         row.Name,
		Email:        row.Email,
		PasswordHash: row.PasswordHash,
		CreatedAt:    row.CreatedAt,
	}
}

func toAuthSession(row db.AuthSession) *domain_user.AuthSession {
	return &domain_user.AuthSession{
		ID:                  row.ID,
		UserID:              row.UserID,
		RefreshTokenHash:    row.RefreshTokenHash,
		UserAgent:           row.UserAgent,
		IPAddress:           formatIP(row.IpAddress),
		DeviceName:          row.DeviceName,
		CreatedAt:           row.CreatedAt,
		LastUsedAt:          timestamptzPtr(row.LastUsedAt),
		ExpiresAt:           row.ExpiresAt,
		RevokedAt:           timestamptzPtr(row.RevokedAt),
		ReplacedBySessionID: row.ReplacedBySessionID,
	}
}

func parseIP(ip *string) *netip.Addr {
	if ip == nil || *ip == "" {
		return nil
	}
	addr, err := netip.ParseAddr(*ip)
	if err != nil {
		return nil
	}
	return &addr
}

func formatIP(addr *netip.Addr) *string {
	if addr == nil {
		return nil
	}
	s := addr.String()
	return &s
}

func toTimestamptz(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

func timestamptzPtr(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	v := t.Time
	return &v
}
