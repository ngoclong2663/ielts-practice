package internal

import (
	"github.com/gin-gonic/gin"

	"ielts-practice/app/middlewares"
	"ielts-practice/app/modules/library/dto"
	"ielts-practice/pkg/tokens"
)

type LibraryRouter struct {
	presenter *LibraryPresenter
	jwt       *tokens.JWTManager
}

func NewLibraryRouter(presenter *LibraryPresenter, jwt *tokens.JWTManager) *LibraryRouter {
	return &LibraryRouter{presenter: presenter, jwt: jwt}
}

func RegisterRoutes(router *gin.Engine, r *LibraryRouter) {
	p := r.presenter
	g := router.Group("/api/v1/library", middlewares.AuthMiddleware(r.jwt))
	g.GET("/collections", p.ListCollections)
	g.GET("/collections/:slug/words", middlewares.ValidateQueryParams[dto.ListWordsQueryDto](), p.ListWords)
	g.POST("/add", middlewares.ShortTermRateLimitMiddleware(), middlewares.ValidateRequestBody[dto.AddWordsRequestDto](), p.AddWords)
}
