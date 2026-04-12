package usecase

import (
	domain_project "taskflow/internal/domain/project"
	domain_task "taskflow/internal/domain/task"
	domain_user "taskflow/internal/domain/user"
	"taskflow/internal/repository"
	auth_usecase "taskflow/internal/usecase/auth"
	project_usecase "taskflow/internal/usecase/project"
	task_usecase "taskflow/internal/usecase/task"
	"taskflow/utils"
)

type UseCases struct {
	Auth     domain_user.UseCase
	Projects domain_project.UseCase
	Tasks    domain_task.UseCase
}

func Init(config *utils.Config, repos *repository.Repository) UseCases {
	return UseCases{
		Auth:     auth_usecase.New(config, repos.Users),
		Projects: project_usecase.New(repos.Projects),
		Tasks:    task_usecase.New(repos.Tasks, repos.Projects),
	}
}
