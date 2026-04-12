package projects

import (
	"context"

	domain_project "taskflow/internal/domain/project"

	"github.com/google/uuid"
)

type Stub struct {
	StatsFn func(uuid.UUID, uuid.UUID) (*domain_project.Stats, error)
}

func NewStub() *Stub {
	return &Stub{
		StatsFn: func(id uuid.UUID, callerID uuid.UUID) (*domain_project.Stats, error) {
			return &domain_project.Stats{
				ByStatus: map[string]int64{
					"todo":        2,
					"in_progress": 1,
				},
				ByAssignee: []domain_project.AssigneeTaskCount{
					{AssigneeID: &callerID, Count: 2},
				},
			}, nil
		},
	}
}

func (s *Stub) Create(ctx context.Context, input domain_project.CreateInput) (*domain_project.Project, error) {
	return nil, nil
}

func (s *Stub) List(ctx context.Context, userID uuid.UUID, p domain_project.Pagination) ([]*domain_project.Project, error) {
	return nil, nil
}

func (s *Stub) Get(ctx context.Context, id uuid.UUID) (*domain_project.Project, error) {
	return &domain_project.Project{ID: id, OwnerID: uuid.Nil}, nil
}

func (s *Stub) Stats(ctx context.Context, id uuid.UUID, callerID uuid.UUID) (*domain_project.Stats, error) {
	return s.StatsFn(id, callerID)
}

func (s *Stub) Update(ctx context.Context, id uuid.UUID, callerID uuid.UUID, input domain_project.UpdateInput) (*domain_project.Project, error) {
	return nil, nil
}

func (s *Stub) Delete(ctx context.Context, id uuid.UUID, callerID uuid.UUID) error {
	return nil
}
