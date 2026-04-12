package task_controller

import domain_task "taskflow/internal/domain/task"

type createRequest struct {
	Title      string               `json:"title" validate:"required,min=1"`
	Priority   domain_task.Priority `json:"priority"`
	AssigneeID *string              `json:"assignee_id"`
	DueDate    *string              `json:"due_date"`
}

type updateRequest struct {
	Title      *string               `json:"title"`
	Status     *domain_task.Status   `json:"status"`
	Priority   *domain_task.Priority `json:"priority"`
	AssigneeID *string               `json:"assignee_id"`
	DueDate    *string               `json:"due_date"`
}
