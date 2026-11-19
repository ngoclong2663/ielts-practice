package internal

import (
	"github.com/gin-gonic/gin"

	"ielts-practice/app/middlewares"
	adminin "ielts-practice/app/modules/admin/ports/in"
	"ielts-practice/app/modules/library/dto"
	"ielts-practice/pkg/tokens"
)

type LibraryRouter struct {
	presenter *LibraryPresenter
	jwt       *tokens.JWTManager
	session   adminin.IAdminSession
}

func NewLibraryRouter(presenter *LibraryPresenter, jwt *tokens.JWTManager, session adminin.IAdminSession) *LibraryRouter {
	return &LibraryRouter{presenter: presenter, jwt: jwt, session: session}
}

func RegisterRoutes(router *gin.Engine, r *LibraryRouter) {
	p := r.presenter
	g := router.Group("/api/v1/library", middlewares.AuthMiddleware(r.jwt))
	g.GET("/collections", p.ListCollections)
	g.GET("/collections/:slug/words", middlewares.ValidateQueryParams[dto.ListWordsQueryDto](), p.ListWords)
	g.POST("/add", middlewares.ShortTermRateLimitMiddleware(), middlewares.ValidateRequestBody[dto.AddWordsRequestDto](), p.AddWords)

	// Library candidates (Epic 9): the owner approves AI-made cards learners keep into the library.
	admin := router.Group("/api/v1/admin/library", middlewares.AdminGuard(r.session)...)
	admin.GET("/candidates", p.ListCandidates)
	admin.POST("/candidates/approve", middlewares.ValidateRequestBody[dto.ApproveCandidatesRequestDto](), p.ApproveCandidates)
	admin.POST("/candidates/reject", middlewares.ValidateRequestBody[dto.RejectCandidatesRequestDto](), p.RejectCandidates)
}
