package internal

import (
	"github.com/gin-gonic/gin"

	"ielts-practice/app/middlewares"
	"ielts-practice/app/modules/guides/dto"
	"ielts-practice/pkg/tokens"
)

type GuidesRouter struct {
	presenter *GuidesPresenter
	jwt       *tokens.JWTManager
}

func NewGuidesRouter(presenter *GuidesPresenter, jwt *tokens.JWTManager) *GuidesRouter {
	return &GuidesRouter{presenter: presenter, jwt: jwt}
}

func RegisterRoutes(router *gin.Engine, r *GuidesRouter) {
	p := r.presenter
	// Public: the media are app content (no user data), and <video>/<img> tags can't refresh an expired session.
	router.GET("/api/v1/guides/media/:file", p.ResolveMedia)
	g := router.Group("/api/v1/guides", middlewares.AuthMiddleware(r.jwt))
	g.GET("", middlewares.ValidateQueryParams[dto.ListGuidesQueryDto](), p.ListGuides)
}
