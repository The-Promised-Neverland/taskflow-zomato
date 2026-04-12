package domain_task

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Status string
type Priority string

const (
	StatusTodo       Status = "todo"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"

	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
)

type Task struct {
	ID         uuid.UUID  `json:"id"`
	ProjectID  uuid.UUID  `json:"project_id"`
	Title      string     `json:"title"`
	Status     Status     `json:"status"`
	Priority   Priority   `json:"priority"`
	AssigneeID *uuid.UUID `json:"assignee_id,omitempty"`
	DueDate    *time.Time `json:"due_date,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type CreateInput struct {
	ProjectID  uuid.UUID
	Title      string
	Priority   Priority
	AssigneeID *uuid.UUID
	DueDate    *time.Time
}

type UpdateInput struct {
	Title      *string
	Status     *Status
	Priority   *Priority
	AssigneeID *uuid.UUID
	DueDate    *time.Time
}

type ListFilter struct {
	Status     *Status
	AssigneeID *uuid.UUID
}

type Pagination struct {
	Page     int
	PageSize int
}

type Repository interface {
	Create(ctx context.Context, task *Task) error
	GetByProject(ctx context.Context, projectID uuid.UUID, filter ListFilter, p Pagination) ([]*Task, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Task, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateInput) (*Task, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type UseCase interface {
	Create(ctx context.Context, callerID uuid.UUID, input CreateInput) (*Task, error)
	List(ctx context.Context, projectID uuid.UUID, filter ListFilter, p Pagination) ([]*Task, error)
	Get(ctx context.Context, id uuid.UUID) (*Task, error)
	Update(ctx context.Context, id uuid.UUID, callerID uuid.UUID, input UpdateInput) (*Task, error)
	Delete(ctx context.Context, id uuid.UUID, callerID uuid.UUID) error
}
