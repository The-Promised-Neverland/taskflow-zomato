package task_controller

import (
	"net/http"
	"time"

	"taskflow/delivery/http/common"
	domain_error "taskflow/internal/domain/errors"
	domain_task "taskflow/internal/domain/task"
	"taskflow/utils/validator"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Controller struct {
	tasks domain_task.UseCase
}

func New(tasks domain_task.UseCase) *Controller {
	return &Controller{tasks: tasks}
}

// List returns tasks for a project.
// GET /v1/projects/{projectID}/tasks
func (c *Controller) List(ctx *gin.Context) {
	projectIDParam := ctx.Param("projectID")
	if projectIDParam == "" {
		projectIDParam = ctx.Param("id")
	}

	projectID, err := uuid.Parse(projectIDParam)
	if err != nil {
		common.SendAppError(ctx.Writer, domain_error.Raise(domain_error.CODE_INVALID_PAYLOAD, "invalid project id", err))
		return
	}

	filter := domain_task.ListFilter{}
	if s := ctx.Query("status"); s != "" {
		status := domain_task.Status(s)
		filter.Status = &status
	}
	if a := ctx.Query("assignee"); a != "" {
		aid, err := uuid.Parse(a)
		if err != nil {
			common.SendAppError(ctx.Writer, domain_error.Raise(domain_error.CODE_INVALID_PAYLOAD, "invalid assignee id", err))
			return
		}
		filter.AssigneeID = &aid
	}

	pagination := common.Pagination{Page: 1, PageSize: 20}
	if p, ok := ctx.Request.Context().Value(common.PaginationKey).(common.Pagination); ok {
		pagination = p
	}

	tasks, err := c.tasks.List(ctx.Request.Context(), projectID, filter, domain_task.Pagination{
		Page:     pagination.Page,
		PageSize: pagination.PageSize,
	})
	if err != nil {
		common.SendAppError(ctx.Writer, err)
		return
	}

	if tasks == nil {
		tasks = []*domain_task.Task{}
	}
	common.SendJSON(ctx.Writer, http.StatusOK, map[string]interface{}{"tasks": tasks})
}

// Create adds a task to the project.
// POST /v1/projects/{projectID}/tasks
func (c *Controller) Create(ctx *gin.Context) {
	callerID := ctx.Request.Context().Value(common.UserIDKey).(uuid.UUID)

	projectIDParam := ctx.Param("projectID")
	if projectIDParam == "" {
		projectIDParam = ctx.Param("id")
	}

	projectID, err := uuid.Parse(projectIDParam)
	if err != nil {
		common.SendAppError(ctx.Writer, domain_error.Raise(domain_error.CODE_INVALID_PAYLOAD, "invalid project id", err))
		return
	}

	var req createRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.SendAppError(ctx.Writer, domain_error.Raise(domain_error.CODE_INVALID_PAYLOAD, "", err))
		return
	}
	if err := validator.Validate(ctx.Request.Context(), req); err != nil {
		common.SendAppError(ctx.Writer, err)
		return
	}

	input := domain_task.CreateInput{
		ProjectID: projectID,
		Title:     req.Title,
		Priority:  req.Priority,
	}

	if req.AssigneeID != nil {
		aid, err := uuid.Parse(*req.AssigneeID)
		if err != nil {
			common.SendAppError(ctx.Writer, domain_error.Raise(domain_error.CODE_INVALID_PAYLOAD, "invalid assignee_id", err))
			return
		}
		input.AssigneeID = &aid
	}

	if req.DueDate != nil {
		t, err := time.Parse(time.DateOnly, *req.DueDate)
		if err != nil {
			common.SendAppError(ctx.Writer, domain_error.Raise(domain_error.CODE_INVALID_PAYLOAD, "invalid due_date, expected YYYY-MM-DD", err))
			return
		}
		input.DueDate = &t
	}

	task, err := c.tasks.Create(ctx.Request.Context(), callerID, input)
	if err != nil {
		common.SendAppError(ctx.Writer, err)
		return
	}

	common.SendJSON(ctx.Writer, http.StatusCreated, task)
}

// Update changes a task.
// PATCH /v1/tasks/{id}
func (c *Controller) Update(ctx *gin.Context) {
	callerID := ctx.Request.Context().Value(common.UserIDKey).(uuid.UUID)

	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		common.SendAppError(ctx.Writer, domain_error.Raise(domain_error.CODE_INVALID_PAYLOAD, "invalid task id", err))
		return
	}

	var req updateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.SendAppError(ctx.Writer, domain_error.Raise(domain_error.CODE_INVALID_PAYLOAD, "", err))
		return
	}

	input := domain_task.UpdateInput{
		Title:    req.Title,
		Status:   req.Status,
		Priority: req.Priority,
	}

	if req.AssigneeID != nil {
		aid, err := uuid.Parse(*req.AssigneeID)
		if err != nil {
			common.SendAppError(ctx.Writer, domain_error.Raise(domain_error.CODE_INVALID_PAYLOAD, "invalid assignee_id", err))
			return
		}
		input.AssigneeID = &aid
	}

	if req.DueDate != nil {
		t, err := time.Parse(time.DateOnly, *req.DueDate)
		if err != nil {
			common.SendAppError(ctx.Writer, domain_error.Raise(domain_error.CODE_INVALID_PAYLOAD, "invalid due_date, expected YYYY-MM-DD", err))
			return
		}
		input.DueDate = &t
	}

	task, err := c.tasks.Update(ctx.Request.Context(), id, callerID, input)
	if err != nil {
		common.SendAppError(ctx.Writer, err)
		return
	}

	common.SendJSON(ctx.Writer, http.StatusOK, task)
}

// Delete removes a task.
// DELETE /v1/tasks/{id}
func (c *Controller) Delete(ctx *gin.Context) {
	callerID := ctx.Request.Context().Value(common.UserIDKey).(uuid.UUID)

	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		common.SendAppError(ctx.Writer, domain_error.Raise(domain_error.CODE_INVALID_PAYLOAD, "invalid task id", err))
		return
	}

	if err := c.tasks.Delete(ctx.Request.Context(), id, callerID); err != nil {
		common.SendAppError(ctx.Writer, err)
		return
	}

	ctx.Status(http.StatusNoContent)
}
