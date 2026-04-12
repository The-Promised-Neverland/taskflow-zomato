package tests

import (
	"net/http"
	"testing"

	"taskflow/delivery/http/routes"
	domain_project "taskflow/internal/domain/project"
	domain_task "taskflow/internal/domain/task"
	domain_user "taskflow/internal/domain/user"
	"taskflow/internal/usecase"
	authstub "taskflow/tests/auth"
	projectstub "taskflow/tests/projects"
	taskstub "taskflow/tests/tasks"

	"github.com/gin-gonic/gin"
)

type TestRouter struct {
	Router   http.Handler
	Auth     *authstub.Stub
	Projects *projectstub.Stub
	Tasks    *taskstub.Stub
}

func NewRouter() *TestRouter {
	auth := authstub.NewStub()
	projects := projectstub.NewStub()
	tasks := taskstub.NewStub()

	return &TestRouter{
		Router: routes.NewRouter(&usecase.UseCases{
			Auth:     auth,
			Projects: projects,
			Tasks:    tasks,
		}),
		Auth:     auth,
		Projects: projects,
		Tasks:    tasks,
	}
}

func InitTestMode(t *testing.T) {
	t.Helper()
	gin.SetMode(gin.TestMode)
}

var (
	_ domain_user.UseCase    = (*authstub.Stub)(nil)
	_ domain_project.UseCase = (*projectstub.Stub)(nil)
	_ domain_task.UseCase    = (*taskstub.Stub)(nil)
)
