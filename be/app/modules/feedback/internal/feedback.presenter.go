package internal

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ielts-practice/app/middlewares"
	"ielts-practice/app/modules/feedback/dto"
	usecases "ielts-practice/app/modules/feedback/internal/use-cases"
	"ielts-practice/pkg"
)

type FeedbackPresenter struct {
	uc usecases.IFeedbackUseCases
}

func NewFeedbackPresenter(uc usecases.IFeedbackUseCases) *FeedbackPresenter {
	return &FeedbackPresenter{uc: uc}
}

func respond[T any](c *gin.Context, status int, data T, err error) {
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(status, data)
}

func (p *FeedbackPresenter) SendFeedback(c *gin.Context) {
	req := pkg.MustGetBody[dto.CreateFeedbackRequestDto](c)
	data, err := p.uc.SendFeedback(c.Request.Context(), middlewares.MustGetUserID(c), c.GetHeader("User-Agent"), req)
	respond(c, http.StatusCreated, data, err)
}

func (p *FeedbackPresenter) ListFeedback(c *gin.Context) {
	q, _ := pkg.GetQuery[dto.ListFeedbackQueryDto](c)
	data, err := p.uc.ListFeedback(c.Request.Context(), q)
	respond(c, http.StatusOK, data, err)
}

func (p *FeedbackPresenter) UpdateFeedback(c *gin.Context) {
	req := pkg.MustGetBody[dto.UpdateFeedbackRequestDto](c)
	data, err := p.uc.UpdateFeedback(c.Request.Context(), c.Param("id"), req)
	respond(c, http.StatusOK, data, err)
}
