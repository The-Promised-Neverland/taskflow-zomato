-- name: CreateUser :exec
INSERT INTO users (id, name, email, password_hash, created_at)
VALUES ($1, $2, $3, $4, $5);

-- name: GetUserByEmail :one
SELECT id, name, email, password_hash, created_at
FROM users
WHERE email = $1;

-- name: GetUserByID :one
SELECT id, name, email, password_hash, created_at
FROM users
WHERE id = $1;

-- name: CreateAuthSession :exec
INSERT INTO auth_sessions (
    id, user_id, refresh_token_hash, user_agent, ip_address, device_name,
    created_at, last_used_at, expires_at, revoked_at, replaced_by_session_id
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);

-- name: GetAuthSessionByID :one
SELECT id, user_id, refresh_token_hash, user_agent, ip_address, device_name,
       created_at, last_used_at, expires_at, revoked_at, replaced_by_session_id
FROM auth_sessions
WHERE id = $1;

-- name: UpdateAuthSessionLastUsedAt :exec
UPDATE auth_sessions
SET last_used_at = $2
WHERE id = $1;

-- name: RevokeAuthSession :exec
UPDATE auth_sessions
SET revoked_at = $2,
    replaced_by_session_id = $3
WHERE id = $1;

-- name: RevokeAllAuthSessionsForUser :exec
UPDATE auth_sessions
SET revoked_at = $2
WHERE user_id = $1
  AND revoked_at IS NULL;
