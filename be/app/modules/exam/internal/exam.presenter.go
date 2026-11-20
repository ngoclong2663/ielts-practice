package internal

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"ielts-practice/app/middlewares"
	"ielts-practice/app/modules/exam/dto"
	usecases "ielts-practice/app/modules/exam/internal/use-cases"
	"ielts-practice/pkg"
	"ielts-practice/pkg/exceptions"
)

type ExamPresenter struct {
	uc usecases.IExamUseCases
}

func NewExamPresenter(uc *usecases.ExamUseCases) *ExamPresenter {
	return &ExamPresenter{uc: uc}
}

func respond[T any](c *gin.Context, status int, data T, err error) {
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(status, data)
}

func (p *ExamPresenter) ListBooks(c *gin.Context) {
	q, _ := pkg.GetQuery[dto.SkillQueryDto](c)
	data, err := p.uc.ListBooks(c.Request.Context(), middlewares.MustGetUserID(c), q.Skill)
	respond(c, http.StatusOK, data, err)
}

func (p *ExamPresenter) GetTest(c *gin.Context) {
	q, _ := pkg.GetQuery[dto.SkillQueryDto](c)
	data, err := p.uc.GetTest(c.Request.Context(), c.Param("slug"), q.Skill)
	respond(c, http.StatusOK, data, err)
}

func (p *ExamPresenter) SubmitAttempt(c *gin.Context) {
	data, err := p.uc.SubmitAttempt(c.Request.Context(), middlewares.MustGetUserID(c), pkg.MustGetBody[dto.SubmitAttemptRequestDto](c))
	respond(c, http.StatusCreated, data, err)
}

func (p *ExamPresenter) GetAttempt(c *gin.Context) {
	data, err := p.uc.GetAttempt(c.Request.Context(), middlewares.MustGetUserID(c), c.Param("attemptId"))
	respond(c, http.StatusOK, data, err)
}

func (p *ExamPresenter) ListAttempts(c *gin.Context) {
	q, _ := pkg.GetQuery[dto.ListAttemptsQueryDto](c)
	data, err := p.uc.ListAttempts(c.Request.Context(), middlewares.MustGetUserID(c), q)
	respond(c, http.StatusOK, data, err)
}

func (p *ExamPresenter) GetStats(c *gin.Context) {
	q, _ := pkg.GetQuery[dto.SkillQueryDto](c)
	data, err := p.uc.GetStats(c.Request.Context(), middlewares.MustGetUserID(c), q.Skill)
	respond(c, http.StatusOK, data, err)
}

// ResolveMedia redirects to a presigned S3 link for an audio file or figure.
func (p *ExamPresenter) ResolveMedia(c *gin.Context) {
	url, err := p.uc.ResolveMedia(c.Request.Context(), c.Param("file"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.Header("Cache-Control", fmt.Sprintf("public, max-age=%d", int(usecases.MediaCacheMaxAge.Seconds())))
	c.Redirect(http.StatusFound, url)
}

// ResolveFigure serves a map or diagram compiled into the binary. The name is its content hash, so it never
// changes and can be cached for a year.
func (p *ExamPresenter) ResolveFigure(c *gin.Context) {
	b, ok := p.uc.ResolveFigure(c.Param("file"))
	if !ok {
		_ = c.Error(exceptions.NewException(http.StatusNotFound, exceptions.CodeExamMediaNotFound, "Figure not found", nil))
		return
	}
	c.Header("Cache-Control", "public, max-age=31536000, immutable")
	c.Header("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'")
	c.Data(http.StatusOK, "image/svg+xml", b)
}
