# Taskflow Backend

## 1. Overview

Taskflow is a Go backend for managing users, projects, tasks, auth sessions, and project stats.

This repo is mostly backend-only. A Lovable frontend was used for visual testing, but the backend here is still the source of truth.

It is also deployed on EC2 for review and live testing.

Tech stack:

- Go
- Gin
- PostgreSQL
- Goose for migrations
- sqlc for SQL access
- Docker and Docker Compose
- JWT access tokens plus refresh tokens

## 2. Architecture Decisions

- HTTP handlers stay thin and business rules live in use cases.
- The code is split by layer instead of crammed into one package.
- sqlc is used instead of an ORM so the SQL stays explicit and easy to reason about.
- Goose is used for migrations so schema changes are versioned and repeatable.
- Auth uses access tokens plus refresh tokens and session rows. That gives us:
  - login on multiple devices
  - logout from the current device
  - logout from all devices
  - refresh token rotation
  - basic session audit data like IP and user agent
- Seeder data is part of the startup flow so the reviewer can log in right away.

Auth flow in short:

- Login and register create a session row in the database.
- The API returns an access token and a refresh token.
- The access token is used on protected requests.
- The backend checks the session row in the database on every protected request.
- The refresh token is only used when the frontend calls the refresh endpoint.
- Logout revokes one session.
- Logout all revokes every session for that user.

The session table stores basic metadata like IP address and user agent. With more time, this could be extended with proper device fingerprinting so sessions are easier to recognize in the UI and audit logs.

Tradeoffs:

- More files and packages, but the code is easier to follow.
- Less magic than an ORM, but the SQL is clear.
- Seeder runs automatically, which is convenient for review, but it means startup is doing a bit more work.

## 3. Running Locally

Assume Docker is installed and nothing else.

```bash
git clone <your-repo-url>
cd taskflow-abhijit
cp .env.example .env
docker compose up --build
```

When it is up, open this in a browser:

```text
http://localhost:8080/api/health
```

If you are hitting the EC2 deploy directly, the base URL is:

```text
http://ec2-13-126-105-149.ap-south-1.compute.amazonaws.com:8080/
```

If you want the stack in the background instead of the terminal, use:

```bash
docker compose up -d --build
```

Local env values already work in `.env.example`. The important ones are:

- `DB_HOST=localhost`
- `DB_PORT=5433`
- `DB_NAME=taskflow`
- `DB_USER=postgres`
- `DB_PASSWORD=postgres`
- `DB_SSLMODE=disable`
- `ACCESS_TOKEN_EXPIRATION=25h`
- `REFRESH_TOKEN_EXPIRATION=720h`

## 4. Running Migrations

Migrations run automatically when you start the stack with Docker Compose.

If you want to run them by themselves, use:

```bash
docker compose up -d db
docker compose up migrate
```

Seeder data also runs automatically before the API starts.

If you want to trigger seeding on its own after migrations are done:

```bash
docker compose up seeder
```

## 5. Test Credentials

Use these credentials right after the seed job runs:

- Email: `test@example.com`
- Password: `password123`
- Name: `Test User`

## 6. API Reference

Common rules:

- Protected routes need `Authorization: Bearer <access_token>`.
- Send `Content-Type: application/json` for JSON requests.
- List endpoints support `page` and `limit`.
- `limit` defaults to `20` and maxes out at `100`.
- Task `status` must be one of `todo`, `in_progress`, `done`.
- Task `priority` must be one of `low`, `medium`, `high`.
- Task `due_date` must be `YYYY-MM-DD`.

Common success response:

```json
{
  "data": {},
  "code": "SUCCESS"
}
```

Common error response:

```json
{
  "error_code": "TKF-REST-00",
  "message": "invalid request payload",
  "data": null
}
```

### Health

| Method | Path | Auth | Request | Response |
| --- | --- | --- | --- | --- |
| GET | `/api/health` | No | None | `200` with `{ "data": { "status": "ok", "service": "taskflow", "timestamp": 1234567890 }, "code": "SUCCESS" }` |

### Auth

| Method | Path | Auth | Request body | Response |
| --- | --- | --- | --- | --- |
| POST | `/api/v1/auth/register` | No | `{ "name": "Jane", "email": "jane@example.com", "password": "secret123" }` | `201` with `{ "data": { "access_token": "...", "refresh_token": "...", "user": {...} }, "code": "SUCCESS" }` |
| POST | `/api/v1/auth/login` | No | `{ "email": "jane@example.com", "password": "secret123" }` | `200` with the same token/user shape |
| POST | `/api/v1/auth/refresh` | No | `{ "refresh_token": "..." }` | `200` with the same token/user shape |
| POST | `/api/v1/auth/logout` | Yes | None | `204 No Content` |
| POST | `/api/v1/auth/logout-all` | Yes | None | `204 No Content` |

Notes:

- `name` must be at least 2 characters.
- `email` must be valid.
- `password` must be at least 6 characters on register.
- Logout routes use the current access token to find the active session.

### Projects

| Method | Path | Auth | Request body / query | Response |
| --- | --- | --- | --- | --- |
| GET | `/api/v1/projects?page=1&limit=20` | Yes | Query params only | `200` with `{ "data": { "projects": [...] }, "code": "SUCCESS" }` |
| POST | `/api/v1/projects` | Yes | `{ "name": "Project A", "description": "Optional description" }` | `201` with the created project |
| GET | `/api/v1/projects/:id` | Yes | `:id` is a project UUID | `200` with the project |
| PATCH | `/api/v1/projects/:id` | Yes | `{ "name": "New name", "description": "New description" }` or just the fields you want to change | `200` with the updated project |
| DELETE | `/api/v1/projects/:id` | Yes | `:id` is a project UUID | `204 No Content` |
| GET | `/api/v1/projects/:id/stats` | Yes | `:id` is a project UUID | `200` with `{ "data": { "by_status": {...}, "by_assignee": [...] }, "code": "SUCCESS" }` |
| GET | `/api/v1/projects/:id/tasks?page=1&limit=20&status=todo&assignee=<uuid>` | Yes | Query params only | `200` with `{ "data": { "tasks": [...] }, "code": "SUCCESS" }` |
| POST | `/api/v1/projects/:id/tasks` | Yes | `{ "title": "Task title", "priority": "medium", "assignee_id": "<uuid>", "due_date": "2026-04-30" }` | `201` with the created task |

Notes:

- Project `name` is required.
- Project `description` is optional.
- Update requests only need the fields that changed.
- Project stats are only visible to the project owner.

### Tasks

| Method | Path | Auth | Request body | Response |
| --- | --- | --- | --- | --- |
| PATCH | `/api/v1/tasks/:id` | Yes | `{ "title": "New title", "status": "done", "priority": "high", "assignee_id": "<uuid>", "due_date": "2026-04-30" }` | `200` with the updated task |
| DELETE | `/api/v1/tasks/:id` | Yes | None | `204 No Content` |

Notes:

- `title` is required on create.
- `priority` is optional on create and defaults to `medium` if omitted.
- `assignee_id` is optional.
- `due_date` is optional and must be `YYYY-MM-DD` if provided.
- Task updates are allowed for the project owner or the assigned user.

## 7. What I’d Do With More Time

- Add Swagger/OpenAPI docs so the API is easier to explore.
- Add load testing with K6 and Grafana to check how the API behaves under stress.
- Add more integration tests against a real Postgres container.
- Improve auth with things like password reset, active session management, and device fingerprinting.
- Add better logging, rate limiting, and observability around migrations and failures.
