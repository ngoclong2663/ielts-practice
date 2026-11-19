package internal

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ielts-practice/app/middlewares"
	"ielts-practice/app/modules/library/dto"
	usecases "ielts-practice/app/modules/library/internal/use-cases"
	"ielts-practice/pkg"
)

type LibraryPresenter struct {
	uc *usecases.LibraryUseCases
}

func NewLibraryPresenter(uc *usecases.LibraryUseCases) *LibraryPresenter {
	return &LibraryPresenter{uc: uc}
}

func respond[T any](c *gin.Context, status int, data T, err error) {
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(status, data)
}

func (p *LibraryPresenter) ListCollections(c *gin.Context) {
	data, err := p.uc.ListCollections(c.Request.Context())
	respond(c, http.StatusOK, data, err)
}

func (p *LibraryPresenter) ListWords(c *gin.Context) {
	q, _ := pkg.GetQuery[dto.ListWordsQueryDto](c)
	data, err := p.uc.ListWords(c.Request.Context(), middlewares.MustGetUserID(c), c.Param("slug"), q)
	respond(c, http.StatusOK, data, err)
}

func (p *LibraryPresenter) AddWords(c *gin.Context) {
	data, err := p.uc.AddWords(c.Request.Context(), middlewares.MustGetUserID(c), pkg.MustGetBody[dto.AddWordsRequestDto](c))
	respond(c, http.StatusCreated, data, err)
}
