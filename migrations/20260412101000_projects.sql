-- +goose Up
CREATE TABLE IF NOT EXISTS projects (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT        NOT NULL,
    description TEXT        NOT NULL DEFAULT '',
    owner_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
    
CREATE INDEX IF NOT EXISTS idx_projects_owner ON projects(owner_id);

-- +goose Down
DROP INDEX IF EXISTS idx_projects_owner;
DROP TABLE IF EXISTS projects;
