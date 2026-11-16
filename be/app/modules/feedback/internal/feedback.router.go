package internal

import (
	"github.com/gin-gonic/gin"

	"ielts-practice/app/middlewares"
	adminin "ielts-practice/app/modules/admin/ports/in"
	"ielts-practice/app/modules/feedback/dto"
	"ielts-practice/pkg/tokens"
)

type FeedbackRouter struct {
	presenter *FeedbackPresenter
	jwt       *tokens.JWTManager
	session   adminin.IAdminSession
}

func RegisterRoutes(router *gin.Engine, r *FeedbackRouter) {
	r.Register(router)
}

func NewFeedbackRouter(presenter *FeedbackPresenter, jwt *tokens.JWTManager, session adminin.IAdminSession) *FeedbackRouter {
	return &FeedbackRouter{presenter: presenter, jwt: jwt, session: session}
}

func (r *FeedbackRouter) Register(router *gin.Engine) {
	learner := router.Group("/api/v1", middlewares.AuthMiddleware(r.jwt))
	learner.POST("/feedback", middlewares.ShortTermRateLimitMiddleware(),
		middlewares.ValidateRequestBody[dto.CreateFeedbackRequestDto](), r.presenter.SendFeedback)

	admin := router.Group("/api/v1/admin", middlewares.AdminGuard(r.session)...)
	admin.GET("/feedback", middlewares.ValidateQueryParams[dto.ListFeedbackQueryDto](), r.presenter.ListFeedback)
	admin.PATCH("/feedback/:id", middlewares.ValidateRequestBody[dto.UpdateFeedbackRequestDto](), r.presenter.UpdateFeedback)
}
