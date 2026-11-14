package internal

import (
	"github.com/gin-gonic/gin"

	"ielts-practice/app/middlewares"
	"ielts-practice/app/modules/study/dto"
	"ielts-practice/pkg/tokens"
)

type StudyRouter struct {
	presenter *StudyPresenter
	jwt       *tokens.JWTManager
}

func RegisterRoutes(router *gin.Engine, r *StudyRouter) {
	r.Register(router)
}

func NewStudyRouter(presenter *StudyPresenter, jwt *tokens.JWTManager) *StudyRouter {
	return &StudyRouter{presenter: presenter, jwt: jwt}
}

func (r *StudyRouter) Register(router *gin.Engine) {
	api := router.Group("/api/v1", middlewares.AuthMiddleware(r.jwt))
	study := api.Group("/study")
	{
		study.GET("/queue", middlewares.ValidateQueryParams[dto.QueueQueryDto](), r.presenter.GetQueue)
		study.POST("/answer", middlewares.ValidateRequestBody[dto.AnswerRequestDto](), r.presenter.Answer)
		study.POST("/undo", middlewares.ValidateRequestBody[dto.UndoRequestDto](), r.presenter.Undo)
		study.GET("/counts", r.presenter.GetCounts)
	}
	api.GET("/stats/overview", r.presenter.GetStats)
}
