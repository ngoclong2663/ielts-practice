package internal

import (
	"github.com/gin-gonic/gin"

	"ielts-practice/app/middlewares"
	"ielts-practice/app/modules/vocab/dto"
	"ielts-practice/pkg/tokens"
)

type VocabRouter struct {
	presenter *VocabPresenter
	jwt       *tokens.JWTManager
}

func NewVocabRouter(presenter *VocabPresenter, jwt *tokens.JWTManager) *VocabRouter {
	return &VocabRouter{presenter: presenter, jwt: jwt}
}

func RegisterRoutes(router *gin.Engine, r *VocabRouter) {
	// Public: <img> tags can't refresh an expired session, and the endpoint only serves
	// photos we have already registered.
	router.GET("/api/v1/images/:provider/:id", r.presenter.ResolveImage)

	api := router.Group("/api/v1", middlewares.AuthMiddleware(r.jwt))
	vocab := api.Group("/vocab")
	{
		vocab.GET("/status", r.presenter.GetStatus)
		vocab.POST("/normalize", middlewares.ValidateRequestBody[dto.NormalizeRequestDto](), r.presenter.Normalize)
		vocab.GET("/jobs", r.presenter.ListJobs)
		vocab.POST("/jobs", middlewares.MediumTermRateLimitMiddleware(), middlewares.ValidateRequestBody[dto.CreateJobRequestDto](), r.presenter.CreateJob)
		vocab.GET("/jobs/:jobId", middlewares.ValidateUUIDParams("jobId"), r.presenter.GetJob)
		vocab.POST("/jobs/:jobId/commit", middlewares.ValidateUUIDParams("jobId"), middlewares.ValidateRequestBody[dto.CommitRequestDto](), r.presenter.CommitJob)
		items := vocab.Group("/jobs/:jobId/items/:itemId", middlewares.ValidateUUIDParams("jobId", "itemId"))
		items.PATCH("", middlewares.ValidateRequestBody[dto.UpdateItemRequestDto](), r.presenter.UpdateItem)
		items.POST("/retry", middlewares.ShortTermRateLimitMiddleware(), middlewares.ValidateRequestBody[dto.RetryItemRequestDto](), r.presenter.RetryItem)
		items.GET("/images", middlewares.ShortTermRateLimitMiddleware(), middlewares.ValidateQueryParams[dto.ImageSearchQueryDto](), r.presenter.SearchImages)
	}
	api.POST("/lexemes/:lexemeId/report", middlewares.ValidateUUIDParams("lexemeId"), middlewares.ValidateRequestBody[dto.ReportRequestDto](), r.presenter.ReportLexeme)
}
