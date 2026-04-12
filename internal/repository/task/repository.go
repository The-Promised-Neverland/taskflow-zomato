package task_repository

import (
	"context"
	"errors"
	"time"

	domain_task "taskflow/internal/domain/task"
	db "taskflow/internal/repository/task/driver/postgres"
	postgres "taskflow/utils/database/postgres"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type repository struct {
	q *db.Queries
}

func New(conn *postgres.DBConnector) domain_task.Repository {
	return &repository{q: db.New(conn.Pool)}
}

func (r *repository) Create(ctx context.Context, task *domain_task.Task) error {
	return r.q.CreateTask(ctx, db.CreateTaskParams{
		ID:         task.ID,
		ProjectID:  task.ProjectID,
		Title:      task.Title,
		Status:     string(task.Status),
		Priority:   string(task.Priority),
		AssigneeID: task.AssigneeID,
		DueDate:    task.DueDate,
		CreatedAt:  task.CreatedAt,
		UpdatedAt:  task.UpdatedAt,
	})
}

func (r *repository) GetByProject(ctx context.Context, projectID uuid.UUID, filter domain_task.ListFilter, p domain_task.Pagination) ([]*domain_task.Task, error) {
	var status *string
	if filter.Status != nil {
		s := string(*filter.Status)
		status = &s
	}

	rows, err := r.q.GetTasksByProject(ctx, db.GetTasksByProjectParams{
		ProjectID:  projectID,
		Status:     status,
		AssigneeID: filter.AssigneeID,
		Limit:      int32(p.PageSize),
		Offset:     int32((p.Page - 1) * p.PageSize),
	})
	if err != nil {
		return nil, err
	}

	tasks := make([]*domain_task.Task, len(rows))
	for i, row := range rows {
		tasks[i] = toTask(row)
	}
	return tasks, nil
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*domain_task.Task, error) {
	row, err := r.q.GetTaskByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return toTask(row), nil
}

func (r *repository) Update(ctx context.Context, id uuid.UUID, input domain_task.UpdateInput) (*domain_task.Task, error) {
	var status *string
	if input.Status != nil {
		s := string(*input.Status)
		status = &s
	}

	var priority *string
	if input.Priority != nil {
		p := string(*input.Priority)
		priority = &p
	}

	row, err := r.q.UpdateTask(ctx, db.UpdateTaskParams{
		ID:         id,
		UpdatedAt:  time.Now(),
		Title:      input.Title,
		Status:     status,
		Priority:   priority,
		AssigneeID: input.AssigneeID,
		DueDate:    input.DueDate,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return toTask(row), nil
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.q.DeleteTask(ctx, id)
}

func toTask(row db.Task) *domain_task.Task {
	return &domain_task.Task{
		ID:         row.ID,
		ProjectID:  row.ProjectID,
		Title:      row.Title,
		Status:     domain_task.Status(row.Status),
		Priority:   domain_task.Priority(row.Priority),
		AssigneeID: row.AssigneeID,
		DueDate:    row.DueDate,
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
	}
}
