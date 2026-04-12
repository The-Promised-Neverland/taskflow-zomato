-- name: CreateProject :exec
INSERT INTO projects (id, name, description, owner_id, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: GetProjectsForUser :many
SELECT DISTINCT p.id, p.name, p.description, p.owner_id, p.created_at, p.updated_at
FROM projects p
LEFT JOIN tasks t ON t.project_id = p.id
WHERE p.owner_id = $1 OR t.assignee_id = $1
ORDER BY p.created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetProjectByID :one
SELECT id, name, description, owner_id, created_at, updated_at
FROM projects
WHERE id = $1;

-- name: UpdateProject :one
UPDATE projects
SET name        = COALESCE(sqlc.narg('name'), name),
    description = COALESCE(sqlc.narg('description'), description),
    updated_at  = $2
WHERE id = $1
RETURNING id, name, description, owner_id, created_at, updated_at;

-- name: DeleteProject :exec
DELETE FROM projects WHERE id = $1;

-- name: GetTaskStatusCountsForProject :many
SELECT status, COUNT(*)::bigint AS count
FROM tasks
WHERE project_id = $1
GROUP BY status
ORDER BY status;

-- name: GetTaskCountsByAssigneeForProject :many
SELECT assignee_id, COUNT(*)::bigint AS count
FROM tasks
WHERE project_id = $1
GROUP BY assignee_id
ORDER BY count DESC;
