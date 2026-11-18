package internal

import (
	"github.com/gin-gonic/gin"

	"ielts-practice/app/middlewares"
	"ielts-practice/app/modules/writing/dto"
	"ielts-practice/pkg/tokens"
)

type WritingRouter struct {
	presenter *WritingPresenter
	jwt       *tokens.JWTManager
}

func NewWritingRouter(presenter *WritingPresenter, jwt *tokens.JWTManager) *WritingRouter {
	return &WritingRouter{presenter: presenter, jwt: jwt}
}

func RegisterRoutes(router *gin.Engine, r *WritingRouter) {
	p := r.presenter
	// Public: figures are task content (no user data), and <img> tags can't refresh an expired session.
	router.GET("/api/v1/writing/figures/:file", p.ResolveFigure)

	w := router.Group("/api/v1/writing", middlewares.AuthMiddleware(r.jwt))
	w.GET("/status", p.GetStatus)
	w.GET("/tasks", middlewares.ValidateQueryParams[dto.TaskListQueryDto](), p.ListTasks)
	w.GET("/tasks/:taskId", middlewares.ValidateUUIDParams("taskId"), p.GetTask)
	// Starting may write a lesson (one AI call shared by everyone), so it is rate limited.
	w.POST("/tasks/:taskId/attempts", middlewares.MediumTermRateLimitMiddleware(), middlewares.ValidateUUIDParams("taskId"),
		middlewares.ValidateRequestBody[dto.StartAttemptRequestDto](), p.StartAttempt)
	w.GET("/attempts", p.ListAttempts)

	a := w.Group("/attempts/:attemptId", middlewares.ValidateUUIDParams("attemptId"))
	a.GET("", p.GetAttempt)
	// The first request for a lesson writes its vocabulary (one AI call shared by everyone), so it is rate limited.
	a.GET("/vocab", middlewares.MediumTermRateLimitMiddleware(), p.GetVocab)
	a.POST("/plan", middlewares.ValidateRequestBody[dto.PlanRequestDto](), p.AnswerPlan)
	// Daily limits (WRITING_DAILY_CHECKS / WRITING_DAILY_SUBMISSIONS) are enforced in the use cases.
	a.POST("/paragraphs/:key/check", middlewares.ShortTermRateLimitMiddleware(),
		middlewares.ValidateRequestBody[dto.CheckRequestDto](), p.CheckParagraph)
	a.POST("/submit", middlewares.ShortTermRateLimitMiddleware(), middlewares.ValidateRequestBody[dto.SubmitRequestDto](), p.SubmitAttempt)
}
