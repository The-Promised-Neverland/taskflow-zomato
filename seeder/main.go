package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	passwordHashCost = 12

	seedUserEmail    = "test@example.com"
	seedUserPassword = "password123"

	seedUserName       = "Test User"
	seedProjectName    = "Seed Project"
	seedProjectDesc    = "Project created by the local seed script"
	seedTaskOneTitle   = "Design the task flow"
	seedTaskTwoTitle   = "Implement the API"
	seedTaskThreeTitle = "Validate the release"
)

type config struct {
	DBConnectionStr string `env:"DB_CONNECTION_STR"`
	DBHost          string `env:"DB_HOST"`
	DBPort          string `env:"DB_PORT"`
	DBName          string `env:"DB_NAME"`
	DBUser          string `env:"DB_USER"`
	DBPassword      string `env:"DB_PASSWORD"`
	DBSSLMode       string `env:"DB_SSLMODE"`
}

func main() {
	if err := run(); err != nil {
		slog.Error("seeding failed", "error", err)
		os.Exit(1)
	}

	slog.Info("seeding completed successfully")
}

func run() error {
	_ = godotenv.Load(".env")

	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	pool, err := newPool(cfg)
	if err != nil {
		return err
	}
	defer pool.Close()

	ctx := context.Background()
	return seed(ctx, pool)
}

func loadConfig() (*config, error) {
	cfg := &config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	if cfg.DBConnectionStr == "" {
		if cfg.DBHost == "" || cfg.DBPort == "" || cfg.DBName == "" || cfg.DBUser == "" || cfg.DBPassword == "" || cfg.DBSSLMode == "" {
			return nil, fmt.Errorf("database config is incomplete: set DB_CONNECTION_STR or DB_HOST, DB_PORT, DB_NAME, DB_USER, DB_PASSWORD, DB_SSLMODE")
		}

		cfg.DBConnectionStr = fmt.Sprintf(
			"postgresql://%s:%s@%s:%s/%s?sslmode=%s",
			cfg.DBUser,
			cfg.DBPassword,
			cfg.DBHost,
			cfg.DBPort,
			cfg.DBName,
			cfg.DBSSLMode,
		)
	}

	return cfg, nil
}

func newPool(cfg *config) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.DBConnectionStr)
	if err != nil {
		return nil, fmt.Errorf("parse pg dsn: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create pgx pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping pg: %w", err)
	}

	return pool, nil
}

func seed(ctx context.Context, pool *pgxpool.Pool) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	now := time.Now().UTC()

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(seedUserPassword), passwordHashCost)
	if err != nil {
		return fmt.Errorf("hash seed password: %w", err)
	}

	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	projectID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	taskOneID := uuid.MustParse("33333333-3333-3333-3333-333333333331")
	taskTwoID := uuid.MustParse("33333333-3333-3333-3333-333333333332")
	taskThreeID := uuid.MustParse("33333333-3333-3333-3333-333333333333")

	if _, err := tx.Exec(ctx, `
		INSERT INTO users (id, name, email, password_hash, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE
		SET name = EXCLUDED.name,
			email = EXCLUDED.email,
			password_hash = EXCLUDED.password_hash,
			created_at = EXCLUDED.created_at
	`, userID, seedUserName, seedUserEmail, string(passwordHash), now); err != nil {
		return fmt.Errorf("seed user: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO projects (id, name, description, owner_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE
		SET name = EXCLUDED.name,
			description = EXCLUDED.description,
			owner_id = EXCLUDED.owner_id,
			created_at = EXCLUDED.created_at,
			updated_at = EXCLUDED.updated_at
	`, projectID, seedProjectName, seedProjectDesc, userID, now, now); err != nil {
		return fmt.Errorf("seed project: %w", err)
	}

	taskRows := []struct {
		id       uuid.UUID
		title    string
		status   string
		priority string
	}{
		{id: taskOneID, title: seedTaskOneTitle, status: "todo", priority: "low"},
		{id: taskTwoID, title: seedTaskTwoTitle, status: "in_progress", priority: "medium"},
		{id: taskThreeID, title: seedTaskThreeTitle, status: "done", priority: "high"},
	}

	for _, task := range taskRows {
		if _, err := tx.Exec(ctx, `
			INSERT INTO tasks (id, project_id, title, status, priority, assignee_id, due_date, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (id) DO UPDATE
			SET project_id = EXCLUDED.project_id,
				title = EXCLUDED.title,
				status = EXCLUDED.status,
				priority = EXCLUDED.priority,
				assignee_id = EXCLUDED.assignee_id,
				due_date = EXCLUDED.due_date,
				created_at = EXCLUDED.created_at,
				updated_at = EXCLUDED.updated_at
		`, task.id, projectID, task.title, task.status, task.priority, userID, nil, now, now); err != nil {
			return fmt.Errorf("seed task %s: %w", task.id, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit seed tx: %w", err)
	}

	slog.Info("seed data inserted",
		"user_email", seedUserEmail,
		"project_id", projectID.String(),
	)
	return nil
}
