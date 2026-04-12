package tasks

import (
	"context"

	domain_task "taskflow/internal/domain/task"

	"github.com/google/uuid"
)

type Stub struct {
	ListFn         func(uuid.UUID, domain_task.ListFilter, domain_task.Pagination) ([]*domain_task.Task, error)
	LastPagination domain_task.Pagination
}

func NewStub() *Stub {
	return &Stub{
		ListFn: func(projectID uuid.UUID, filter domain_task.ListFilter, p domain_task.Pagination) ([]*domain_task.Task, error) {
			return []*domain_task.Task{
				{
					ID:        uuid.New(),
					ProjectID: projectID,
					Title:     "Task 1",
					Status:    domain_task.StatusTodo,
				},
			}, nil
		},
	}
}

func (s *Stub) Create(ctx context.Context, callerID uuid.UUID, input domain_task.CreateInput) (*domain_task.Task, error) {
	return nil, nil
}

func (s *Stub) List(ctx context.Context, projectID uuid.UUID, filter domain_task.ListFilter, p domain_task.Pagination) ([]*domain_task.Task, error) {
	s.LastPagination = p
	return s.ListFn(projectID, filter, p)
}

func (s *Stub) Get(ctx context.Context, id uuid.UUID) (*domain_task.Task, error) {
	return nil, nil
}

func (s *Stub) Update(ctx context.Context, id uuid.UUID, callerID uuid.UUID, input domain_task.UpdateInput) (*domain_task.Task, error) {
	return nil, nil
}

func (s *Stub) Delete(ctx context.Context, id uuid.UUID, callerID uuid.UUID) error {
	return nil
}
