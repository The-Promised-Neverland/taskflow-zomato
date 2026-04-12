package routes

import (
	"net/http"
	"time"

	"taskflow/delivery/http/common"
	auth_controller "taskflow/delivery/http/controller/auth"
	project_controller "taskflow/delivery/http/controller/project"
	task_controller "taskflow/delivery/http/controller/task"
	"taskflow/delivery/http/middleware"
	"taskflow/internal/usecase"

	"github.com/gin-gonic/gin"
)

func NewRouter(uc *usecase.UseCases) http.Handler {
	r := gin.New()

	// Shared middleware
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())

	api := r.Group("/api")

	// Open endpoints
	api.GET("/health", func(c *gin.Context) {
		common.SendJSON(c.Writer, http.StatusOK, map[string]interface{}{
			"status":    "ok",
			"service":   "taskflow",
			"timestamp": time.Now().Unix(),
		})
	})

	authCtrl := auth_controller.New(uc.Auth)
	v1 := api.Group("/v1")

	// Authentication endpoints
	authGroup := v1.Group("/auth")
	authGroup.POST("/register", authCtrl.Register)
	authGroup.POST("/login", authCtrl.Login)
	authGroup.POST("/refresh", authCtrl.Refresh)

	// Session management endpoints
	authProtected := v1.Group("/auth")
	authProtected.Use(middleware.Authenticate(uc.Auth))
	authProtected.POST("/logout", authCtrl.Logout)
	authProtected.POST("/logout-all", authCtrl.LogoutAll)

	// Protected v1 API routes
	v1.Use(middleware.Authenticate(uc.Auth))
	v1.Use(middleware.Pagination())

	projectCtrl := project_controller.New(uc.Projects)
	v1.GET("/projects", projectCtrl.List)
	v1.POST("/projects", projectCtrl.Create)
	v1.GET("/projects/:id", projectCtrl.Get)
	v1.PATCH("/projects/:id", projectCtrl.Update)
	v1.DELETE("/projects/:id", projectCtrl.Delete)
	v1.GET("/projects/:id/stats", projectCtrl.Stats)

	taskCtrl := task_controller.New(uc.Tasks)
	v1.GET("/projects/:id/tasks", taskCtrl.List)
	v1.POST("/projects/:id/tasks", taskCtrl.Create)
	v1.PATCH("/tasks/:id", taskCtrl.Update)
	v1.DELETE("/tasks/:id", taskCtrl.Delete)

	return r
}
