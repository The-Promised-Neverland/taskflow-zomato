-- name: CreateTask :exec
INSERT INTO tasks (id, project_id, title, status, priority, assignee_id, due_date, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);

-- name: GetTasksByProject :many
SELECT id, project_id, title, status, priority, assignee_id, due_date, created_at, updated_at
FROM tasks
WHERE project_id = $1
  AND (sqlc.narg('status')::text IS NULL OR status = sqlc.narg('status'))
  AND (sqlc.narg('assignee_id')::uuid IS NULL OR assignee_id = sqlc.narg('assignee_id'))
ORDER BY created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: GetTaskByID :one
SELECT id, project_id, title, status, priority, assignee_id, due_date, created_at, updated_at
FROM tasks
WHERE id = $1;

-- name: UpdateTask :one
UPDATE tasks
SET title       = COALESCE(sqlc.narg('title'), title),
    status      = COALESCE(sqlc.narg('status'), status),
    priority    = COALESCE(sqlc.narg('priority'), priority),
    assignee_id = COALESCE(sqlc.narg('assignee_id'), assignee_id),
    due_date    = COALESCE(sqlc.narg('due_date'), due_date),
    updated_at  = $2
WHERE id = $1
RETURNING id, project_id, title, status, priority, assignee_id, due_date, created_at, updated_at;

-- name: DeleteTask :exec
DELETE FROM tasks WHERE id = $1;
