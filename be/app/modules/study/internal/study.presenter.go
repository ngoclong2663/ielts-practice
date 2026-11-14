package internal

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ielts-practice/app/middlewares"
	"ielts-practice/app/modules/study/dto"
	usecases "ielts-practice/app/modules/study/internal/use-cases"
	"ielts-practice/pkg"
)

type StudyPresenter struct {
	uc usecases.IStudyUseCases
}

func NewStudyPresenter(uc usecases.IStudyUseCases) *StudyPresenter {
	return &StudyPresenter{uc: uc}
}

func respond[T any](c *gin.Context, data T, err error) {
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, data)
}

func (p *StudyPresenter) GetQueue(c *gin.Context) {
	q, _ := pkg.GetQuery[dto.QueueQueryDto](c)
	data, err := p.uc.GetQueue(c.Request.Context(), middlewares.MustGetUserID(c), q)
	respond(c, data, err)
}

func (p *StudyPresenter) Answer(c *gin.Context) {
	req := pkg.MustGetBody[dto.AnswerRequestDto](c)
	data, err := p.uc.Answer(c.Request.Context(), middlewares.MustGetUserID(c), req)
	respond(c, data, err)
}

func (p *StudyPresenter) Undo(c *gin.Context) {
	req := pkg.MustGetBody[dto.UndoRequestDto](c)
	data, err := p.uc.Undo(c.Request.Context(), middlewares.MustGetUserID(c), req)
	respond(c, data, err)
}

func (p *StudyPresenter) GetCounts(c *gin.Context) {
	data, err := p.uc.GetCounts(c.Request.Context(), middlewares.MustGetUserID(c))
	respond(c, data, err)
}

func (p *StudyPresenter) GetStats(c *gin.Context) {
	data, err := p.uc.GetStats(c.Request.Context(), middlewares.MustGetUserID(c))
	respond(c, data, err)
}
