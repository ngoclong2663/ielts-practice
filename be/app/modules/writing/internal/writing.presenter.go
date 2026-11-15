package internal

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"ielts-practice/app/middlewares"
	"ielts-practice/app/modules/writing/dto"
	usecases "ielts-practice/app/modules/writing/internal/use-cases"
	"ielts-practice/pkg"
)

type WritingPresenter struct {
	uc *usecases.WritingUseCases
}

func NewWritingPresenter(uc *usecases.WritingUseCases) *WritingPresenter {
	return &WritingPresenter{uc: uc}
}

func respond[T any](c *gin.Context, status int, data T, err error) {
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(status, data)
}

// ResolveFigure redirects to a presigned S3 link for a task figure.
func (p *WritingPresenter) ResolveFigure(c *gin.Context) {
	fig, err := p.uc.ResolveFigure(c.Request.Context(), c.Param("file"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.Header("Cache-Control", fmt.Sprintf("public, max-age=%d", int(usecases.FigureCacheMaxAge.Seconds())))
	if fig.RedirectURL == "" {
		c.Data(http.StatusOK, fig.ContentType, fig.Data)
		return
	}
	c.Redirect(http.StatusFound, fig.RedirectURL)
}

func (p *WritingPresenter) GetStatus(c *gin.Context) {
	data, err := p.uc.GetStatus(c.Request.Context(), middlewares.MustGetUserID(c))
	respond(c, http.StatusOK, data, err)
}

func (p *WritingPresenter) ListTasks(c *gin.Context) {
	q, _ := pkg.GetQuery[dto.TaskListQueryDto](c)
	data, err := p.uc.ListTasks(c.Request.Context(), middlewares.MustGetUserID(c), q)
	respond(c, http.StatusOK, data, err)
}

func (p *WritingPresenter) GetTask(c *gin.Context) {
	data, err := p.uc.GetTask(c.Request.Context(), middlewares.MustGetUserID(c), c.Param("taskId"))
	respond(c, http.StatusOK, data, err)
}

func (p *WritingPresenter) StartAttempt(c *gin.Context) {
	data, err := p.uc.StartAttempt(c.Request.Context(), middlewares.MustGetUserID(c), c.Param("taskId"),
		pkg.MustGetBody[dto.StartAttemptRequestDto](c))
	respond(c, http.StatusCreated, data, err)
}

func (p *WritingPresenter) ListAttempts(c *gin.Context) {
	data, err := p.uc.ListAttempts(c.Request.Context(), middlewares.MustGetUserID(c))
	respond(c, http.StatusOK, data, err)
}

func (p *WritingPresenter) GetAttempt(c *gin.Context) {
	data, err := p.uc.GetAttempt(c.Request.Context(), middlewares.MustGetUserID(c), c.Param("attemptId"))
	respond(c, http.StatusOK, data, err)
}

func (p *WritingPresenter) AnswerPlan(c *gin.Context) {
	data, err := p.uc.AnswerPlan(c.Request.Context(), middlewares.MustGetUserID(c), c.Param("attemptId"),
		pkg.MustGetBody[dto.PlanRequestDto](c))
	respond(c, http.StatusOK, data, err)
}

func (p *WritingPresenter) CheckParagraph(c *gin.Context) {
	data, err := p.uc.CheckParagraph(c.Request.Context(), middlewares.MustGetUserID(c), c.Param("attemptId"),
		c.Param("key"), pkg.MustGetBody[dto.CheckRequestDto](c))
	respond(c, http.StatusOK, data, err)
}

func (p *WritingPresenter) SubmitAttempt(c *gin.Context) {
	data, err := p.uc.SubmitAttempt(c.Request.Context(), middlewares.MustGetUserID(c), c.Param("attemptId"),
		pkg.MustGetBody[dto.SubmitRequestDto](c))
	respond(c, http.StatusOK, data, err)
}
