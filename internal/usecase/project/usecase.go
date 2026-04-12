package project_usecase

import (
	"context"
	"time"

	domain_error "taskflow/internal/domain/errors"
	domain_project "taskflow/internal/domain/project"

	"github.com/google/uuid"
)

type useCase struct {
	projects domain_project.Repository
}

func New(projects domain_project.Repository) domain_project.UseCase {
	return &useCase{projects: projects}
}

func (uc *useCase) Create(ctx context.Context, input domain_project.CreateInput) (*domain_project.Project, error) {
	now := time.Now()
	project := &domain_project.Project{
		ID:          uuid.New(),
		Name:        input.Name,
		Description: input.Description,
		OwnerID:     input.OwnerID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := uc.projects.Create(ctx, project); err != nil {
		return nil, domain_error.Raise(domain_error.CODE_INTERNAL_ERROR, "", err)
	}
	return project, nil
}

func (uc *useCase) List(ctx context.Context, userID uuid.UUID, p domain_project.Pagination) ([]*domain_project.Project, error) {
	projects, err := uc.projects.GetAllForUser(ctx, userID, p)
	if err != nil {
		return nil, domain_error.Raise(domain_error.CODE_INTERNAL_ERROR, "", err)
	}
	return projects, nil
}

func (uc *useCase) Get(ctx context.Context, id uuid.UUID) (*domain_project.Project, error) {
	project, err := uc.projects.GetByID(ctx, id)
	if err != nil {
		return nil, domain_error.Raise(domain_error.CODE_INTERNAL_ERROR, "", err)
	}
	if project == nil {
		return nil, domain_error.Raise(domain_error.CODE_PROJECT_NOT_FOUND, "", nil)
	}
	return project, nil
}

func (uc *useCase) Stats(ctx context.Context, id uuid.UUID, callerID uuid.UUID) (*domain_project.Stats, error) {
	project, err := uc.projects.GetByID(ctx, id)
	if err != nil {
		return nil, domain_error.Raise(domain_error.CODE_INTERNAL_ERROR, "", err)
	}
	if project == nil {
		return nil, domain_error.Raise(domain_error.CODE_PROJECT_NOT_FOUND, "", nil)
	}
	if project.OwnerID != callerID {
		return nil, domain_error.Raise(domain_error.CODE_PROJECT_FORBIDDEN, "", nil)
	}

	stats, err := uc.projects.GetStats(ctx, id)
	if err != nil {
		return nil, domain_error.Raise(domain_error.CODE_INTERNAL_ERROR, "", err)
	}
	return stats, nil
}

func (uc *useCase) Update(ctx context.Context, id uuid.UUID, callerID uuid.UUID, input domain_project.UpdateInput) (*domain_project.Project, error) {
	project, err := uc.projects.GetByID(ctx, id)
	if err != nil {
		return nil, domain_error.Raise(domain_error.CODE_INTERNAL_ERROR, "", err)
	}
	if project == nil {
		return nil, domain_error.Raise(domain_error.CODE_PROJECT_NOT_FOUND, "", nil)
	}
	if project.OwnerID != callerID {
		return nil, domain_error.Raise(domain_error.CODE_PROJECT_FORBIDDEN, "", nil)
	}

	updated, err := uc.projects.Update(ctx, id, input)
	if err != nil {
		return nil, domain_error.Raise(domain_error.CODE_INTERNAL_ERROR, "", err)
	}
	return updated, nil
}

func (uc *useCase) Delete(ctx context.Context, id uuid.UUID, callerID uuid.UUID) error {
	project, err := uc.projects.GetByID(ctx, id)
	if err != nil {
		return domain_error.Raise(domain_error.CODE_INTERNAL_ERROR, "", err)
	}
	if project == nil {
		return domain_error.Raise(domain_error.CODE_PROJECT_NOT_FOUND, "", nil)
	}
	if project.OwnerID != callerID {
		return domain_error.Raise(domain_error.CODE_PROJECT_FORBIDDEN, "", nil)
	}

	if err := uc.projects.Delete(ctx, id); err != nil {
		return domain_error.Raise(domain_error.CODE_INTERNAL_ERROR, "", err)
	}
	return nil
}
