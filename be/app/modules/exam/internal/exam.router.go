package internal

import (
	"github.com/gin-gonic/gin"

	"ielts-practice/app/middlewares"
	"ielts-practice/app/modules/exam/dto"
	"ielts-practice/pkg/tokens"
)

type ExamRouter struct {
	presenter *ExamPresenter
	jwt       *tokens.JWTManager
}

func NewExamRouter(presenter *ExamPresenter, jwt *tokens.JWTManager) *ExamRouter {
	return &ExamRouter{presenter: presenter, jwt: jwt}
}

func RegisterRoutes(router *gin.Engine, r *ExamRouter) {
	p := r.presenter
	// Public: figures and audio are test content (no user data) named by their content hash, and <img> and
	// <audio> elements cannot refresh an expired session in the middle of a test.
	router.GET("/api/v1/exam/figures/:file", p.ResolveFigure)
	router.GET("/api/v1/exam/media/:file", p.ResolveMedia)
	e := router.Group("/api/v1/exam", middlewares.AuthMiddleware(r.jwt))
	e.GET("/books", middlewares.ValidateQueryParams[dto.SkillQueryDto](), p.ListBooks)
	e.GET("/tests/:slug", middlewares.ValidateQueryParams[dto.SkillQueryDto](), p.GetTest)
	e.GET("/stats", middlewares.ValidateQueryParams[dto.SkillQueryDto](), p.GetStats)
	e.GET("/attempts", middlewares.ValidateQueryParams[dto.ListAttemptsQueryDto](), p.ListAttempts)
	e.POST("/attempts", middlewares.ShortTermRateLimitMiddleware(), middlewares.ValidateRequestBody[dto.SubmitAttemptRequestDto](), p.SubmitAttempt)
	e.GET("/attempts/:attemptId", middlewares.ValidateUUIDParams("attemptId"), p.GetAttempt)
}
