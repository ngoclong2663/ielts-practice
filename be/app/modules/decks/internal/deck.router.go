package internal

import (
	"github.com/gin-gonic/gin"

	"ielts-practice/app/middlewares"
	"ielts-practice/app/modules/decks/dto"
	"ielts-practice/pkg/tokens"
)

type DeckRouter struct {
	presenter *DeckPresenter
	jwt       *tokens.JWTManager
}

func RegisterRoutes(router *gin.Engine, r *DeckRouter) {
	r.Register(router)
}

func NewDeckRouter(presenter *DeckPresenter, jwt *tokens.JWTManager) *DeckRouter {
	return &DeckRouter{presenter: presenter, jwt: jwt}
}

func (r *DeckRouter) Register(router *gin.Engine) {
	api := router.Group("/api/v1", middlewares.AuthMiddleware(r.jwt))

	decks := api.Group("/decks")
	{
		decks.GET("", r.presenter.ListDecks)
		decks.POST("", middlewares.ValidateRequestBody[dto.CreateDeckRequestDto](), r.presenter.CreateDeck)
		decks.GET("/:deckId", middlewares.ValidateUUIDParams("deckId"), r.presenter.GetDeck)
		decks.PATCH("/:deckId", middlewares.ValidateUUIDParams("deckId"), middlewares.ValidateRequestBody[dto.UpdateDeckRequestDto](), r.presenter.UpdateDeck)
		decks.DELETE("/:deckId", middlewares.ValidateUUIDParams("deckId"), r.presenter.DeleteDeck)
		decks.GET("/:deckId/cards", middlewares.ValidateUUIDParams("deckId"), middlewares.ValidateQueryParams[dto.ListCardsQueryDto](), r.presenter.ListCards)
		decks.POST("/:deckId/cards", middlewares.ValidateUUIDParams("deckId"), middlewares.ValidateRequestBody[dto.CreateCardRequestDto](), r.presenter.CreateCard)
	}

	cards := api.Group("/cards")
	{
		cards.POST("/bulk", middlewares.ValidateRequestBody[dto.BulkCardsRequestDto](), r.presenter.BulkCards)
		cards.GET("/:cardId", middlewares.ValidateUUIDParams("cardId"), r.presenter.GetCard)
		cards.PATCH("/:cardId", middlewares.ValidateUUIDParams("cardId"), middlewares.ValidateRequestBody[dto.UpdateCardRequestDto](), r.presenter.UpdateCard)
	}
}
