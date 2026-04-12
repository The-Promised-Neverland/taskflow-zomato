package task_usecase

import (
	"context"
	"time"

	domain_error "taskflow/internal/domain/errors"
	domain_project "taskflow/internal/domain/project"
	domain_task "taskflow/internal/domain/task"

	"github.com/google/uuid"
)

type useCase struct {
	tasks    domain_task.Repository
	projects domain_project.Repository
}

func New(tasks domain_task.Repository, projects domain_project.Repository) domain_task.UseCase {
	return &useCase{tasks: tasks, projects: projects}
}

func (uc *useCase) Create(ctx context.Context, callerID uuid.UUID, input domain_task.CreateInput) (*domain_task.Task, error) {
	project, err := uc.projects.GetByID(ctx, input.ProjectID)
	if err != nil {
		return nil, domain_error.Raise(domain_error.CODE_INTERNAL_ERROR, "", err)
	}
	if project == nil {
		return nil, domain_error.Raise(domain_error.CODE_TASK_PROJECT_NOT_FOUND, "", nil)
	}

	now := time.Now()
	priority := input.Priority
	if priority == "" {
		priority = domain_task.PriorityMedium
	}

	task := &domain_task.Task{
		ID:         uuid.New(),
		ProjectID:  input.ProjectID,
		Title:      input.Title,
		Status:     domain_task.StatusTodo,
		Priority:   priority,
		AssigneeID: input.AssigneeID,
		DueDate:    input.DueDate,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := uc.tasks.Create(ctx, task); err != nil {
		return nil, domain_error.Raise(domain_error.CODE_INTERNAL_ERROR, "", err)
	}
	return task, nil
}

func (uc *useCase) List(ctx context.Context, projectID uuid.UUID, filter domain_task.ListFilter, p domain_task.Pagination) ([]*domain_task.Task, error) {
	tasks, err := uc.tasks.GetByProject(ctx, projectID, filter, p)
	if err != nil {
		return nil, domain_error.Raise(domain_error.CODE_INTERNAL_ERROR, "", err)
	}
	return tasks, nil
}

func (uc *useCase) Get(ctx context.Context, id uuid.UUID) (*domain_task.Task, error) {
	task, err := uc.tasks.GetByID(ctx, id)
	if err != nil {
		return nil, domain_error.Raise(domain_error.CODE_INTERNAL_ERROR, "", err)
	}
	if task == nil {
		return nil, domain_error.Raise(domain_error.CODE_TASK_NOT_FOUND, "", nil)
	}
	return task, nil
}

func (uc *useCase) Update(ctx context.Context, id uuid.UUID, callerID uuid.UUID, input domain_task.UpdateInput) (*domain_task.Task, error) {
	task, err := uc.tasks.GetByID(ctx, id)
	if err != nil {
		return nil, domain_error.Raise(domain_error.CODE_INTERNAL_ERROR, "", err)
	}
	if task == nil {
		return nil, domain_error.Raise(domain_error.CODE_TASK_NOT_FOUND, "", nil)
	}

	isAssignee := task.AssigneeID != nil && *task.AssigneeID == callerID
	project, err := uc.projects.GetByID(ctx, task.ProjectID)
	if err != nil {
		return nil, domain_error.Raise(domain_error.CODE_INTERNAL_ERROR, "", err)
	}
	isOwner := project != nil && project.OwnerID == callerID

	if !isOwner && !isAssignee {
		return nil, domain_error.Raise(domain_error.CODE_TASK_FORBIDDEN, "", nil)
	}

	updated, err := uc.tasks.Update(ctx, id, input)
	if err != nil {
		return nil, domain_error.Raise(domain_error.CODE_INTERNAL_ERROR, "", err)
	}
	return updated, nil
}

func (uc *useCase) Delete(ctx context.Context, id uuid.UUID, callerID uuid.UUID) error {
	task, err := uc.tasks.GetByID(ctx, id)
	if err != nil {
		return domain_error.Raise(domain_error.CODE_INTERNAL_ERROR, "", err)
	}
	if task == nil {
		return domain_error.Raise(domain_error.CODE_TASK_NOT_FOUND, "", nil)
	}

	isAssignee := task.AssigneeID != nil && *task.AssigneeID == callerID
	project, err := uc.projects.GetByID(ctx, task.ProjectID)
	if err != nil {
		return domain_error.Raise(domain_error.CODE_INTERNAL_ERROR, "", err)
	}
	isOwner := project != nil && project.OwnerID == callerID

	if !isOwner && !isAssignee {
		return domain_error.Raise(domain_error.CODE_TASK_FORBIDDEN, "", nil)
	}

	if err := uc.tasks.Delete(ctx, id); err != nil {
		return domain_error.Raise(domain_error.CODE_INTERNAL_ERROR, "", err)
	}
	return nil
}
