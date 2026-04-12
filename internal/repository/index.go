package repository

import (
	domain_project "taskflow/internal/domain/project"
	domain_task "taskflow/internal/domain/task"
	domain_user "taskflow/internal/domain/user"
	project_repository "taskflow/internal/repository/project"
	task_repository "taskflow/internal/repository/task"
	user_repository "taskflow/internal/repository/user"
	postgres "taskflow/utils/database/postgres"
)

type Repository struct {
	Users    domain_user.Repository
	Projects domain_project.Repository
	Tasks    domain_task.Repository
}

func New(db *postgres.DBConnector) *Repository {
	return &Repository{
		Users:    user_repository.New(db),
		Projects: project_repository.New(db),
		Tasks:    task_repository.New(db),
	}
}
