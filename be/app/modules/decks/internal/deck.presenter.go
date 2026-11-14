package internal

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ielts-practice/app/middlewares"
	"ielts-practice/app/modules/decks/dto"
	usecases "ielts-practice/app/modules/decks/internal/use-cases"
	"ielts-practice/pkg"
)

type DeckPresenter struct {
	uc *usecases.DeckUseCases
}

func NewDeckPresenter(uc *usecases.DeckUseCases) *DeckPresenter {
	return &DeckPresenter{uc: uc}
}

func respond[T any](c *gin.Context, status int, data T, err error) {
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(status, data)
}

func (p *DeckPresenter) ListDecks(c *gin.Context) {
	data, err := p.uc.ListDecks(c.Request.Context(), middlewares.MustGetUserID(c))
	respond(c, http.StatusOK, data, err)
}

func (p *DeckPresenter) GetDeck(c *gin.Context) {
	data, err := p.uc.GetDeck(c.Request.Context(), middlewares.MustGetUserID(c), c.Param("deckId"))
	respond(c, http.StatusOK, data, err)
}

func (p *DeckPresenter) CreateDeck(c *gin.Context) {
	req := pkg.MustGetBody[dto.CreateDeckRequestDto](c)
	data, err := p.uc.CreateDeck(c.Request.Context(), middlewares.MustGetUserID(c), req)
	respond(c, http.StatusCreated, data, err)
}

func (p *DeckPresenter) UpdateDeck(c *gin.Context) {
	req := pkg.MustGetBody[dto.UpdateDeckRequestDto](c)
	data, err := p.uc.UpdateDeck(c.Request.Context(), middlewares.MustGetUserID(c), c.Param("deckId"), req)
	respond(c, http.StatusOK, data, err)
}

func (p *DeckPresenter) DeleteDeck(c *gin.Context) {
	if err := p.uc.ArchiveDeck(c.Request.Context(), middlewares.MustGetUserID(c), c.Param("deckId")); err != nil {
		_ = c.Error(err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (p *DeckPresenter) ListCards(c *gin.Context) {
	q, _ := pkg.GetQuery[dto.ListCardsQueryDto](c)
	data, err := p.uc.ListCards(c.Request.Context(), middlewares.MustGetUserID(c), c.Param("deckId"), q)
	respond(c, http.StatusOK, data, err)
}

func (p *DeckPresenter) CreateCard(c *gin.Context) {
	req := pkg.MustGetBody[dto.CreateCardRequestDto](c)
	data, err := p.uc.CreateCard(c.Request.Context(), middlewares.MustGetUserID(c), c.Param("deckId"), req)
	respond(c, http.StatusCreated, data, err)
}

func (p *DeckPresenter) GetCard(c *gin.Context) {
	data, err := p.uc.GetCard(c.Request.Context(), middlewares.MustGetUserID(c), c.Param("cardId"))
	respond(c, http.StatusOK, data, err)
}

func (p *DeckPresenter) UpdateCard(c *gin.Context) {
	req := pkg.MustGetBody[dto.UpdateCardRequestDto](c)
	data, err := p.uc.UpdateCard(c.Request.Context(), middlewares.MustGetUserID(c), c.Param("cardId"), req)
	respond(c, http.StatusOK, data, err)
}

func (p *DeckPresenter) BulkCards(c *gin.Context) {
	req := pkg.MustGetBody[dto.BulkCardsRequestDto](c)
	data, err := p.uc.BulkCards(c.Request.Context(), middlewares.MustGetUserID(c), req)
	respond(c, http.StatusOK, data, err)
}
