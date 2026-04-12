package project_controller

import (
	"net/http"

	"taskflow/delivery/http/common"
	domain_error "taskflow/internal/domain/errors"
	domain_project "taskflow/internal/domain/project"
	"taskflow/utils/validator"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Controller struct {
	projects domain_project.UseCase
}

func New(projects domain_project.UseCase) *Controller {
	return &Controller{projects: projects}
}

// List returns the caller's accessible projects.
// GET /v1/projects
func (c *Controller) List(ctx *gin.Context) {
	callerID := ctx.Request.Context().Value(common.UserIDKey).(uuid.UUID)

	pagination := common.Pagination{Page: 1, PageSize: 20}
	if p, ok := ctx.Request.Context().Value(common.PaginationKey).(common.Pagination); ok {
		pagination = p
	}

	projects, err := c.projects.List(ctx.Request.Context(), callerID, domain_project.Pagination{
		Page:     pagination.Page,
		PageSize: pagination.PageSize,
	})
	if err != nil {
		common.SendAppError(ctx.Writer, err)
		return
	}

	if projects == nil {
		projects = []*domain_project.Project{}
	}
	common.SendJSON(ctx.Writer, http.StatusOK, map[string]interface{}{"projects": projects})
}

// Create adds a new project for the caller.
// POST /v1/projects
func (c *Controller) Create(ctx *gin.Context) {
	callerID := ctx.Request.Context().Value(common.UserIDKey).(uuid.UUID)

	var req createRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.SendAppError(ctx.Writer, domain_error.Raise(domain_error.CODE_INVALID_PAYLOAD, "", err))
		return
	}
	if err := validator.Validate(ctx.Request.Context(), req); err != nil {
		common.SendAppError(ctx.Writer, err)
		return
	}

	project, err := c.projects.Create(ctx.Request.Context(), domain_project.CreateInput{
		Name:        req.Name,
		Description: req.Description,
		OwnerID:     callerID,
	})
	if err != nil {
		common.SendAppError(ctx.Writer, err)
		return
	}

	common.SendJSON(ctx.Writer, http.StatusCreated, project)
}

// Get fetches a project by ID.
// GET /v1/projects/{id}
func (c *Controller) Get(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		common.SendAppError(ctx.Writer, domain_error.Raise(domain_error.CODE_INVALID_PAYLOAD, "invalid project id", err))
		return
	}

	project, err := c.projects.Get(ctx.Request.Context(), id)
	if err != nil {
		common.SendAppError(ctx.Writer, err)
		return
	}

	common.SendJSON(ctx.Writer, http.StatusOK, project)
}

// Stats summarizes task counts for a project.
// GET /v1/projects/{id}/stats
func (c *Controller) Stats(ctx *gin.Context) {
	callerID := ctx.Request.Context().Value(common.UserIDKey).(uuid.UUID)

	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		common.SendAppError(ctx.Writer, domain_error.Raise(domain_error.CODE_INVALID_PAYLOAD, "invalid project id", err))
		return
	}

	stats, err := c.projects.Stats(ctx.Request.Context(), id, callerID)
	if err != nil {
		common.SendAppError(ctx.Writer, err)
		return
	}

	common.SendJSON(ctx.Writer, http.StatusOK, stats)
}

// Update edits an existing project.
// PATCH /v1/projects/{id}
func (c *Controller) Update(ctx *gin.Context) {
	callerID := ctx.Request.Context().Value(common.UserIDKey).(uuid.UUID)

	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		common.SendAppError(ctx.Writer, domain_error.Raise(domain_error.CODE_INVALID_PAYLOAD, "invalid project id", err))
		return
	}

	var req updateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.SendAppError(ctx.Writer, domain_error.Raise(domain_error.CODE_INVALID_PAYLOAD, "", err))
		return
	}

	project, err := c.projects.Update(ctx.Request.Context(), id, callerID, domain_project.UpdateInput{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		common.SendAppError(ctx.Writer, err)
		return
	}

	common.SendJSON(ctx.Writer, http.StatusOK, project)
}

// Delete removes a project.
// DELETE /v1/projects/{id}
func (c *Controller) Delete(ctx *gin.Context) {
	callerID := ctx.Request.Context().Value(common.UserIDKey).(uuid.UUID)

	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		common.SendAppError(ctx.Writer, domain_error.Raise(domain_error.CODE_INVALID_PAYLOAD, "invalid project id", err))
		return
	}

	if err := c.projects.Delete(ctx.Request.Context(), id, callerID); err != nil {
		common.SendAppError(ctx.Writer, err)
		return
	}

	ctx.Status(http.StatusNoContent)
}
